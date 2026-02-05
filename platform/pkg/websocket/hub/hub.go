package hub

import (
	"context"
	"sync"

	"github.com/0xsj/nexus/platform/pkg/observability/log"
	ws "github.com/0xsj/nexus/platform/pkg/websocket"
)

// Hub manages WebSocket connections and rooms using a channel-based event loop.
type Hub struct {
	logger log.Logger

	// Channels for the main event loop.
	register   chan ws.Connection
	unregister chan ws.Connection
	broadcast  chan ws.Message
	roomCast   chan roomMessage
	joinRoom   chan roomAction
	leaveRoom  chan roomAction

	// State protected by the event loop (no direct access outside loop).
	connections map[string]ws.Connection
	rooms       map[string]map[string]ws.Connection

	done chan struct{}
	once sync.Once
}

type roomMessage struct {
	room string
	msg  ws.Message
}

type roomAction struct {
	room string
	conn ws.Connection
}

// New creates a new Hub.
func New(logger log.Logger) *Hub {
	if logger == nil {
		logger = log.New()
	}

	return &Hub{
		logger:      logger.With(log.Component("websocket.hub")),
		register:    make(chan ws.Connection, 64),
		unregister:  make(chan ws.Connection, 64),
		broadcast:   make(chan ws.Message, 256),
		roomCast:    make(chan roomMessage, 256),
		joinRoom:    make(chan roomAction, 64),
		leaveRoom:   make(chan roomAction, 64),
		connections: make(map[string]ws.Connection),
		rooms:       make(map[string]map[string]ws.Connection),
		done:        make(chan struct{}),
	}
}

// Start begins the hub's event loop. Blocks until ctx is cancelled or Close is called.
func (h *Hub) Start(ctx context.Context) error {
	h.logger.Info("websocket hub started")

	for {
		select {
		case conn := <-h.register:
			h.connections[conn.ID()] = conn
			h.logger.Debug("connection registered", log.String("conn_id", conn.ID()))

		case conn := <-h.unregister:
			if _, ok := h.connections[conn.ID()]; ok {
				delete(h.connections, conn.ID())
				h.removeFromAllRooms(conn)
				conn.Close()
				h.logger.Debug("connection unregistered", log.String("conn_id", conn.ID()))
			}

		case msg := <-h.broadcast:
			for id, conn := range h.connections {
				if err := conn.Send(msg); err != nil {
					h.logger.Debug("broadcast send failed, removing",
						log.String("conn_id", id),
						log.Err(err),
					)
					delete(h.connections, id)
					h.removeFromAllRooms(conn)
					conn.Close()
				}
			}

		case rm := <-h.roomCast:
			if members, ok := h.rooms[rm.room]; ok {
				for id, conn := range members {
					if err := conn.Send(rm.msg); err != nil {
						h.logger.Debug("room broadcast send failed, removing",
							log.String("conn_id", id),
							log.String("room", rm.room),
							log.Err(err),
						)
						delete(members, id)
						delete(h.connections, id)
						conn.Close()
					}
				}
			}

		case action := <-h.joinRoom:
			members, ok := h.rooms[action.room]
			if !ok {
				members = make(map[string]ws.Connection)
				h.rooms[action.room] = members
			}
			members[action.conn.ID()] = action.conn
			h.logger.Debug("joined room",
				log.String("conn_id", action.conn.ID()),
				log.String("room", action.room),
			)

		case action := <-h.leaveRoom:
			if members, ok := h.rooms[action.room]; ok {
				delete(members, action.conn.ID())
				if len(members) == 0 {
					delete(h.rooms, action.room)
				}
			}

		case <-ctx.Done():
			h.shutdown()
			return ctx.Err()

		case <-h.done:
			h.shutdown()
			return nil
		}
	}
}

// Close signals the hub to stop.
func (h *Hub) Close() error {
	h.once.Do(func() {
		close(h.done)
	})
	return nil
}

// Register adds a connection to the hub.
func (h *Hub) Register(conn ws.Connection) {
	h.register <- conn
}

// Unregister removes a connection from the hub.
func (h *Hub) Unregister(conn ws.Connection) {
	h.unregister <- conn
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(msg ws.Message) {
	h.broadcast <- msg
}

// BroadcastToRoom sends a message to all connections in a room.
func (h *Hub) BroadcastToRoom(room string, msg ws.Message) {
	h.roomCast <- roomMessage{room: room, msg: msg}
}

// JoinRoom adds a connection to a named room.
func (h *Hub) JoinRoom(room string, conn ws.Connection) {
	h.joinRoom <- roomAction{room: room, conn: conn}
}

// LeaveRoom removes a connection from a named room.
func (h *Hub) LeaveRoom(room string, conn ws.Connection) {
	h.leaveRoom <- roomAction{room: room, conn: conn}
}

// shutdown closes all active connections.
func (h *Hub) shutdown() {
	for id, conn := range h.connections {
		conn.Close()
		delete(h.connections, id)
	}
	h.rooms = make(map[string]map[string]ws.Connection)
	h.logger.Info("websocket hub stopped")
}

// removeFromAllRooms removes a connection from every room it belongs to.
func (h *Hub) removeFromAllRooms(conn ws.Connection) {
	for room, members := range h.rooms {
		delete(members, conn.ID())
		if len(members) == 0 {
			delete(h.rooms, room)
		}
	}
}
