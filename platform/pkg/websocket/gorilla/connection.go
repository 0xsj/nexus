package gorilla

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	gorillaWs "github.com/gorilla/websocket"

	ws "github.com/0xsj/nexus/platform/pkg/websocket"
)

// Connection wraps a Gorilla WebSocket connection with read/write pump goroutines.
type Connection struct {
	id   string
	conn *gorillaWs.Conn
	cfg  ws.Config

	send chan ws.Message
	done chan struct{}

	closeOnce sync.Once
}

// NewConnection wraps a Gorilla WebSocket conn with managed read/write pumps.
// The caller should call ReadPump() and WritePump() in separate goroutines.
func NewConnection(conn *gorillaWs.Conn, cfg ws.Config) *Connection {
	return &Connection{
		id:   uuid.NewString(),
		conn: conn,
		cfg:  cfg,
		send: make(chan ws.Message, cfg.SendBufferSize),
		done: make(chan struct{}),
	}
}

// ID returns the unique connection identifier.
func (c *Connection) ID() string {
	return c.id
}

// Send queues a message for delivery. Returns an error if the connection is closed.
func (c *Connection) Send(msg ws.Message) error {
	select {
	case c.send <- msg:
		return nil
	case <-c.done:
		return ws.ErrConnectionClosed("gorilla.Connection.Send", c.id)
	}
}

// Close gracefully shuts down the connection.
func (c *Connection) Close() error {
	c.closeOnce.Do(func() {
		close(c.done)
		c.conn.Close()
	})
	return nil
}

// Done returns a channel closed when the connection terminates.
func (c *Connection) Done() <-chan struct{} {
	return c.done
}

// WritePump pumps messages from the send channel to the WebSocket connection.
// Should be run in its own goroutine. Handles ping/pong keepalive.
func (c *Connection) WritePump() {
	ticker := time.NewTicker(c.cfg.PingInterval)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(c.cfg.WriteTimeout))
			if !ok {
				c.conn.WriteMessage(gorillaWs.CloseMessage, nil)
				return
			}

			data, err := json.Marshal(msg)
			if err != nil {
				continue
			}

			if err := c.conn.WriteMessage(gorillaWs.TextMessage, data); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(c.cfg.WriteTimeout))
			if err := c.conn.WriteMessage(gorillaWs.PingMessage, nil); err != nil {
				return
			}

		case <-c.done:
			return
		}
	}
}

// ReadPump pumps messages from the WebSocket connection.
// Should be run in its own goroutine. Calls onMessage for each received message.
// When the read loop ends, it closes the connection.
func (c *Connection) ReadPump(onMessage func(conn *Connection, msg ws.Message)) {
	defer c.Close()

	c.conn.SetReadLimit(int64(c.cfg.ReadBufferSize) * 1024)
	c.conn.SetReadDeadline(time.Now().Add(c.cfg.PongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(c.cfg.PongWait))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		var msg ws.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			errMsg := ws.NewErrorMessage("invalid message format")
			c.Send(errMsg)
			continue
		}

		if onMessage != nil {
			onMessage(c, msg)
		}
	}
}
