package chat

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Hub mengelola semua client dan merutekan pesan pribadi.
type Hub struct {
	clients        map[uuid.UUID]*Client // Memetakan UserID ke Client
	privateMessage chan *Message         // Channel untuk pesan pribadi
	register       chan *Client
	unregister     chan *Client
}

func NewHub() *Hub {
	return &Hub{
		privateMessage: make(chan *Message),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		clients:        make(map[uuid.UUID]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.ID] = client
			log.Printf("💬 Client terhubung: %s", client.ID)

		case client := <-h.unregister:
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				log.Printf("🔌 Client terputus: %s", client.ID)
			}

		case message := <-h.privateMessage:
			// Cari koneksi milik penerima
			if recipient, ok := h.clients[message.RecipientID]; ok {
				log.Printf("📩 Mengirim pesan dari %s ke %s", message.SenderID, message.RecipientID)
				
				// Encode pesan ke JSON
				msgBytes, err := json.Marshal(message)
				if err != nil {
					log.Printf("Error marshalling message: %v", err)
					continue
				}

				// Kirim pesan hanya ke penerima
				if err := recipient.Conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
					log.Printf("Error mengirim pesan pribadi: %v", err)
				}
			} else {
				log.Printf("Penerima %s tidak ditemukan atau tidak online.", message.RecipientID)
				// Opsional: Kirim pesan error kembali ke pengirim
			}
		}
	}
}

// Metode publik untuk diakses dari handler
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

func (h *Hub) RoutePrivateMessage(message *Message) {
	h.privateMessage <- message
}