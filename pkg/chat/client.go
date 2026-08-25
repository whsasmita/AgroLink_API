package chat

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Client merepresentasikan satu pengguna yang terhubung ke hub melalui WebSocket. Setiap client memiliki ID unik, koneksi WebSocket, dan referensi ke hub tempat mereka terhubung.
type Client struct {
	ID   uuid.UUID
	Conn *websocket.Conn
	Hub  *Hub
}
