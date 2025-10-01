package chat

import "github.com/google/uuid"

// Message adalah struktur untuk pesan yang dikirim antara klien.
type Message struct {
	RecipientID uuid.UUID `json:"recipient_id"`
	Content     string    `json:"content"`
	SenderID    uuid.UUID `json:"sender_id"` // Diisi oleh server
}