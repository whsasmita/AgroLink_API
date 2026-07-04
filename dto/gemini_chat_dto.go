package dto

import "time"

// GeminiChatMessage represents one message in a chat history.
type GeminiChatMessage struct {
	Role    string `json:"role" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// GeminiChatRequest is the payload for the Gemini chat endpoint.
type GeminiChatRequest struct {
	Message string              `json:"message" binding:"required"`
	History []GeminiChatMessage `json:"history,omitempty"`
}

// GeminiChatResponse is returned by the chat endpoint.
type GeminiChatResponse struct {
	Reply            string     `json:"reply"`
	Scope            string     `json:"scope"`
	DailyUsed        int64      `json:"daily_used"`
	DailyLimit       int        `json:"daily_limit"`
	Remaining        int        `json:"remaining"`
	IsPremium        bool       `json:"is_premium"`
	PremiumExpiresAt *time.Time `json:"premium_expires_at,omitempty"`
	TurnID           string     `json:"turn_id,omitempty"`
	Model            string     `json:"model,omitempty"`
}

// GeminiPremiumStatusResponse tells the client the current premium state.
type GeminiPremiumStatusResponse struct {
	IsPremium      bool       `json:"is_premium"`
	Status         string     `json:"status"`
	PlanName       string     `json:"plan_name"`
	Amount         float64    `json:"amount"`
	Currency       string     `json:"currency"`
	OrderID        string     `json:"order_id,omitempty"`
	SnapToken      string     `json:"snap_token,omitempty"`
	RedirectURL    string     `json:"redirect_url,omitempty"`
	StartsAt       *time.Time `json:"starts_at,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	DailyLimit     int        `json:"daily_limit"`
	RemainingToday int        `json:"remaining_today"`
}
