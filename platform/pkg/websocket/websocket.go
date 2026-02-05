package websocket

import (
	"context"
	"io"
	"net/http"
)

// Connection represents a single WebSocket connection.
type Connection interface {
	// ID returns the unique identifier for this connection.
	ID() string

	// Send queues a message for delivery to the client.
	Send(msg Message) error

	// Close gracefully closes the connection.
	Close() error

	// Done returns a channel that is closed when the connection is terminated.
	Done() <-chan struct{}
}

// Upgrader upgrades an HTTP request to a WebSocket connection.
type Upgrader interface {
	// Upgrade upgrades the HTTP connection to a WebSocket connection.
	Upgrade(w http.ResponseWriter, r *http.Request) (Connection, error)
}

// Hub manages WebSocket connections, rooms, and message broadcasting.
type Hub interface {
	// Register adds a connection to the hub.
	Register(conn Connection)

	// Unregister removes a connection from the hub.
	Unregister(conn Connection)

	// Broadcast sends a message to all connected clients.
	Broadcast(msg Message)

	// BroadcastToRoom sends a message to all connections in a room.
	BroadcastToRoom(room string, msg Message)

	// JoinRoom adds a connection to a named room.
	JoinRoom(room string, conn Connection)

	// LeaveRoom removes a connection from a named room.
	LeaveRoom(room string, conn Connection)

	// Start begins the hub's event loop.
	Start(ctx context.Context) error

	io.Closer
}
