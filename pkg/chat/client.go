package chat

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Client merepresentasikan satu pengguna yang terhubung melalui WebSocket.
type Client struct {
	ID   uuid.UUID
	Conn *websocket.Conn
	Hub  *Hub
}