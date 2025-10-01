package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/pkg/chat"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		// *** DEV / TEST ***
		// saat buka HTML via file:// origin bisa "" atau "null"
		if origin == "" || origin == "null" {
			return true
		}
		switch origin {
		case "https://goagrolink.com",
			"http://localhost:3000",
			"http://localhost:5173",
			"http://localhost:8080",
			"http://localhost:5500": // kalau pakai http-server local
			return true
		default:
			log.Printf("WS blocked origin: %q", origin)
			return false
		}
	},
}

type ChatHandler struct {
	hub *chat.Hub
}

func NewChatHandler(hub *chat.Hub) *ChatHandler {
	return &ChatHandler{hub: hub}
}

func (h *ChatHandler) ServeWs(c *gin.Context) {
	// 1) Ambil user dari context (diset oleh AuthMiddleware)
	currentUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	user := currentUser.(*models.User)

	log.Printf("WS handshake OK user=%s origin=%s", user.ID, c.GetHeader("Origin"))

	// 2) Upgrade koneksi
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WS upgrade failed: %v", err)
		return
	}

	// 3) Register client & kirim self_id agar tester bisa menampilkan ID
	client := &chat.Client{ID: user.ID, Conn: conn, Hub: h.hub}
	h.hub.RegisterClient(client)

	_ = conn.WriteJSON(map[string]any{
		"type":    "self_id",
		"content": client.ID.String(),
	})

	// 4) Reader loop
	go h.clientReader(client)
}

func (h *ChatHandler) clientReader(client *chat.Client) {
	defer func() {
		client.Hub.UnregisterClient(client)
		client.Conn.Close()
	}()

	for {
		_, messageBytes, err := client.Conn.ReadMessage()
		if err != nil {
			log.Printf("WS read error (user=%s): %v", client.ID, err)
			break
		}

		var msg chat.Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("WS json unmarshal error: %v; raw=%s", err, string(messageBytes))
			continue
		}

		// server enforce sender
		msg.SenderID = client.ID

		client.Hub.RoutePrivateMessage(&msg)
	}
}
