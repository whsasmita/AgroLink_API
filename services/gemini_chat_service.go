package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/whsasmita/AgroLink_API/config"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
)

var ErrDailyLimitExceeded = errors.New("daily limit exceeded")

const agrolinkScopeRefusal = "Wah, maaf banget ya. Untuk sekarang aku cuma bisa bantu hal-hal seputar penggunaan aplikasi AgroLink, seperti login, profil, worker, project, produk, delivery, pembayaran, notifikasi, dan premium chat. Kalau mau, aku bisa bantu arahkan ke fitur AgroLink yang kamu butuhkan."

type GeminiChatService interface {
	ChatPublic(ipAddress string, request dto.GeminiChatRequest) (*dto.GeminiChatResponse, error)
	ChatPrivate(user *models.User, request dto.GeminiChatRequest) (*dto.GeminiChatResponse, error)
	GetPremiumStatus(userID uuid.UUID) (*dto.GeminiPremiumStatusResponse, error)
	InitiatePremiumCheckout(user *models.User) (*dto.PaymentInitiationResponse, error)
	HandlePremiumWebhook(payload map[string]interface{}) error
	GenerateReply(message string, history []dto.GeminiChatMessage) (string, error)
}

type geminiChatService struct {
	repo   repositories.GeminiChatRepository
	apiKey string
	model  string
	client *http.Client
}

type geminiGenerateRequest struct {
	SystemInstruction geminiContent   `json:"systemInstruction"`
	Contents          []geminiContent `json:"contents"`
	GenerationConfig  geminiConfig    `json:"generationConfig"`
}

type geminiConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerateResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewGeminiChatService(repo repositories.GeminiChatRepository) GeminiChatService {
	model := strings.TrimSpace(os.Getenv("GEMINI_MODEL"))
	if model == "" {
		model = "gemini-3.5-flash"
	}

	return &geminiChatService{
		repo:   repo,
		apiKey: strings.TrimSpace(os.Getenv("GEMINI_API_KEY")),
		model:  model,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (s *geminiChatService) ChatPublic(ipAddress string, request dto.GeminiChatRequest) (*dto.GeminiChatResponse, error) {
	return s.handleChat(models.AIChatScopePublic, nil, ipAddress, request)
}

func (s *geminiChatService) ChatPrivate(user *models.User, request dto.GeminiChatRequest) (*dto.GeminiChatResponse, error) {
	if user == nil {
		return nil, errors.New("user is required")
	}
	return s.handleChat(models.AIChatScopePrivate, &user.ID, "", request)
}

func (s *geminiChatService) GetPremiumStatus(userID uuid.UUID) (*dto.GeminiPremiumStatusResponse, error) {
	now := time.Now().Local()
	privateLimit := getEnvInt("AI_PRIVATE_DAILY_LIMIT", 10)
	usedToday, err := s.repo.CountTurnsSince(models.AIChatScopePrivate, &userID, "", startOfDay(now))
	if err != nil {
		return nil, err
	}

	subscription, err := s.repo.FindSubscriptionByUserID(userID)
	if err != nil {
		return nil, err
	}

	status := "inactive"
	isPremium := false
	limit := privateLimit
	var amount float64
	var currency string
	var orderID string
	var snapToken string
	var redirectURL string
	var startsAt *time.Time
	var expiresAt *time.Time

	if subscription != nil {
		status = subscription.Status
		amount = subscription.Amount
		currency = subscription.Currency
		orderID = subscription.MidtransOrderID
		snapToken = subscription.SnapToken
		redirectURL = subscription.RedirectURL
		startsAt = subscription.StartsAt
		expiresAt = subscription.ExpiresAt
		if isSubscriptionActive(subscription, now) {
			isPremium = true
			limit = getEnvInt("AI_PREMIUM_DAILY_LIMIT", 200)
		}
	}

	remaining := limit - int(usedToday)
	if remaining < 0 {
		remaining = 0
	}

	return &dto.GeminiPremiumStatusResponse{
		IsPremium:      isPremium,
		Status:         status,
		PlanName:       "premium",
		Amount:         amount,
		Currency:       currency,
		OrderID:        orderID,
		SnapToken:      snapToken,
		RedirectURL:    redirectURL,
		StartsAt:       startsAt,
		ExpiresAt:      expiresAt,
		DailyLimit:     limit,
		RemainingToday: remaining,
	}, nil
}

func (s *geminiChatService) InitiatePremiumCheckout(user *models.User) (*dto.PaymentInitiationResponse, error) {
	if user == nil {
		return nil, errors.New("user is required")
	}
	if s.apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY is not configured")
	}

	now := time.Now().Local()
	existing, err := s.repo.FindSubscriptionByUserID(user.ID)
	if err != nil {
		return nil, err
	}
	if isSubscriptionActive(existing, now) {
		return nil, errors.New("premium subscription is already active")
	}

	amount := float64(getEnvInt("AI_PREMIUM_PRICE_IDR", 150000))
	orderID := fmt.Sprintf("ai-premium-%s", uuid.NewString())

	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: int64(amount),
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: user.Name,
			Email: user.Email,
		},
	}

	response, midtransErr := config.SnapClient.CreateTransaction(snapReq)
	if midtransErr != nil {
		return nil, fmt.Errorf("failed to create midtrans snap token: %s", midtransErr.Message)
	}

	subscription := &models.AIChatPremiumSubscription{
		UserID:          user.ID,
		Status:          models.AIChatSubscriptionStatusPending,
		PlanName:        "premium",
		Amount:          amount,
		Currency:        "IDR",
		MidtransOrderID: orderID,
		SnapToken:       response.Token,
		RedirectURL:     response.RedirectURL,
	}

	if err := s.repo.UpsertSubscription(subscription); err != nil {
		return nil, err
	}

	return &dto.PaymentInitiationResponse{
		SnapToken:   response.Token,
		OrderID:     orderID,
		Amount:      amount,
		RedirectURL: response.RedirectURL,
	}, nil
}

func (s *geminiChatService) HandlePremiumWebhook(payload map[string]interface{}) error {
	orderID := getString(payload, "order_id")
	if !strings.HasPrefix(orderID, "ai-premium-") {
		return errors.New("not a premium order")
	}

	subscription, err := s.repo.FindSubscriptionByOrderID(orderID)
	if err != nil {
		return err
	}
	if subscription == nil {
		return errors.New("premium subscription not found")
	}

	status := strings.ToLower(getString(payload, "transaction_status"))
	now := time.Now().Local()
	durationDays := getEnvInt("AI_PREMIUM_DURATION_DAYS", 30)

	switch status {
	case "settlement", "capture":
		base := now
		if subscription.ExpiresAt != nil && subscription.ExpiresAt.After(now) {
			base = *subscription.ExpiresAt
		}
		expiresAt := base.AddDate(0, 0, durationDays)
		startsAt := now
		if subscription.StartsAt != nil && subscription.Status == models.AIChatSubscriptionStatusActive {
			startsAt = *subscription.StartsAt
		}
		subscription.Status = models.AIChatSubscriptionStatusActive
		subscription.StartsAt = &startsAt
		subscription.ExpiresAt = &expiresAt
		subscription.PaidAt = &now
	case "expire":
		subscription.Status = models.AIChatSubscriptionStatusExpired
	case "deny", "cancel", "failure":
		subscription.Status = models.AIChatSubscriptionStatusCancelled
	default:
		return nil
	}

	return s.repo.UpdateSubscription(subscription)
}

func (s *geminiChatService) handleChat(scope string, userID *uuid.UUID, ipAddress string, request dto.GeminiChatRequest) (*dto.GeminiChatResponse, error) {
	if s.apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY is not configured")
	}

	message := strings.TrimSpace(request.Message)
	if message == "" {
		return nil, errors.New("message cannot be empty")
	}

	now := time.Now().Local()
	dayStart := startOfDay(now)
	historyLimit := getEnvInt("AI_CHAT_HISTORY_LIMIT", 10)
	dailyLimit := getEnvInt("AI_PRIVATE_DAILY_LIMIT", 10)
	isPremium := false
	var premiumExpiresAt *time.Time

	if scope == models.AIChatScopePublic {
		dailyLimit = getEnvInt("AI_PUBLIC_DAILY_LIMIT", 3)
	} else {
		subscription, err := s.repo.FindSubscriptionByUserID(*userID)
		if err != nil {
			return nil, err
		}
		if isSubscriptionActive(subscription, now) {
			isPremium = true
			premiumExpiresAt = subscription.ExpiresAt
			dailyLimit = getEnvInt("AI_PREMIUM_DAILY_LIMIT", 200)
		}
	}

	usedToday, err := s.repo.CountTurnsSince(scope, userID, ipAddress, dayStart)
	if err != nil {
		return nil, err
	}
	if int(usedToday) >= dailyLimit {
		return nil, ErrDailyLimitExceeded
	}

	if !isAgroLinkScopeQuery(message) {
		return s.storeRefusalTurn(scope, userID, ipAddress, message, isPremium, premiumExpiresAt, dailyLimit, usedToday)
	}

	turns, err := s.repo.FindRecentTurns(scope, userID, ipAddress, historyLimit)
	if err != nil {
		return nil, err
	}

	history := turnsToGeminiHistory(turns)
	if len(history) == 0 && len(request.History) > 0 {
		history = append(history, request.History...)
	}

	reply, err := s.GenerateReply(message, history)
	if err != nil {
		return nil, err
	}

	turn := &models.AIChatTurn{
		Scope:       scope,
		UserMessage: message,
		AIReply:     reply,
		Model:       s.model,
	}
	if userID != nil {
		turn.UserID = userID
	}
	if trimmedIP := strings.TrimSpace(ipAddress); trimmedIP != "" {
		turn.IPAddress = &trimmedIP
	}

	if err := s.repo.CreateTurn(turn); err != nil {
		return nil, err
	}

	remaining := dailyLimit - int(usedToday) - 1
	if remaining < 0 {
		remaining = 0
	}

	return &dto.GeminiChatResponse{
		Reply:            reply,
		Scope:            scope,
		DailyUsed:        usedToday + 1,
		DailyLimit:       dailyLimit,
		Remaining:        remaining,
		IsPremium:        isPremium,
		PremiumExpiresAt: premiumExpiresAt,
		TurnID:           turn.ID.String(),
		Model:            s.model,
	}, nil
}

func (s *geminiChatService) storeRefusalTurn(scope string, userID *uuid.UUID, ipAddress, message string, isPremium bool, premiumExpiresAt *time.Time, dailyLimit int, usedToday int64) (*dto.GeminiChatResponse, error) {
	turn := &models.AIChatTurn{
		Scope:       scope,
		UserMessage: message,
		AIReply:     agrolinkScopeRefusal,
		Model:       "policy-refusal",
	}
	if userID != nil {
		turn.UserID = userID
	}
	if trimmedIP := strings.TrimSpace(ipAddress); trimmedIP != "" {
		turn.IPAddress = &trimmedIP
	}

	if err := s.repo.CreateTurn(turn); err != nil {
		return nil, err
	}

	remaining := dailyLimit - int(usedToday) - 1
	if remaining < 0 {
		remaining = 0
	}

	return &dto.GeminiChatResponse{
		Reply:            agrolinkScopeRefusal,
		Scope:            scope,
		DailyUsed:        usedToday + 1,
		DailyLimit:       dailyLimit,
		Remaining:        remaining,
		IsPremium:        isPremium,
		PremiumExpiresAt: premiumExpiresAt,
		TurnID:           turn.ID.String(),
		Model:            "policy-refusal",
	}, nil
}

func (s *geminiChatService) GenerateReply(message string, history []dto.GeminiChatMessage) (string, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return "", errors.New("message cannot be empty")
	}

	payload := geminiGenerateRequest{
		SystemInstruction: geminiContent{
			Parts: []geminiPart{{Text: "You are AgroLink Assistant, a helpful AI chatbot for the AgroLink agricultural platform. Answer in Indonesian by default unless the user asks otherwise. Focus on agriculture, workers, farmers, logistics, orders, delivery, payments, and app usage. If the question is outside AgroLink scope, answer briefly and redirect back to the platform context."}},
		},
		GenerationConfig: geminiConfig{
			Temperature:     0.7,
			MaxOutputTokens: 1024,
		},
	}

	for _, item := range history {
		role := normalizeGeminiRole(item.Role)
		content := strings.TrimSpace(item.Content)
		if role == "" || content == "" {
			continue
		}
		payload.Contents = append(payload.Contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: content}},
		})
	}

	payload.Contents = append(payload.Contents, geminiContent{
		Role:  "user",
		Parts: []geminiPart{{Text: message}},
	})

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", s.model, s.apiKey)
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var geminiResp geminiGenerateResponse
	if err := json.Unmarshal(responseBody, &geminiResp); err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if geminiResp.Error != nil && geminiResp.Error.Message != "" {
			return "", fmt.Errorf("gemini api error: %s", geminiResp.Error.Message)
		}
		return "", fmt.Errorf("gemini api returned status %d: %s", resp.StatusCode, string(responseBody))
	}

	if len(geminiResp.Candidates) == 0 {
		return "", errors.New("no response from Gemini")
	}

	var reply strings.Builder
	for _, part := range geminiResp.Candidates[0].Content.Parts {
		if strings.TrimSpace(part.Text) != "" {
			reply.WriteString(part.Text)
		}
	}

	result := strings.TrimSpace(reply.String())
	if result == "" {
		return "", errors.New("empty response from Gemini")
	}

	return result, nil
}

func turnsToGeminiHistory(turns []models.AIChatTurn) []dto.GeminiChatMessage {
	if len(turns) == 0 {
		return nil
	}

	sortedTurns := make([]models.AIChatTurn, len(turns))
	copy(sortedTurns, turns)
	sort.Slice(sortedTurns, func(i, j int) bool {
		return sortedTurns[i].CreatedAt.Before(sortedTurns[j].CreatedAt)
	})

	history := make([]dto.GeminiChatMessage, 0, len(sortedTurns)*2)
	for _, turn := range sortedTurns {
		history = append(history, dto.GeminiChatMessage{Role: "user", Content: turn.UserMessage})
		history = append(history, dto.GeminiChatMessage{Role: "assistant", Content: turn.AIReply})
	}

	return history
}

func isAgroLinkScopeQuery(message string) bool {
	text := strings.ToLower(strings.TrimSpace(message))
	if text == "" {
		return false
	}

	allowedKeywords := []string{
		"agrolink", "login", "daftar", "register", "signup", "auth", "profil", "profile",
		"worker", "petani", "farmer", "driver", "ekspedisi", "delivery", "project", "produk", "product",
		"chat", "midtrans", "payment", "pembayaran", "invoice", "payout", "contract", "kontrak",
		"checkout", "premium", "admin", "review", "notification", "notifikasi", "farm", "lahan",
		"panen", "hasil panen", "logistik", "order", "cart", "checkout", "fitur", "bantuan", "cara pakai",
	}
	for _, keyword := range allowedKeywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}

	return false
}

func normalizeGeminiRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "user", "model":
		return strings.ToLower(strings.TrimSpace(role))
	case "assistant":
		return "model"
	default:
		return ""
	}
}

func isSubscriptionActive(subscription *models.AIChatPremiumSubscription, now time.Time) bool {
	if subscription == nil {
		return false
	}
	if subscription.Status != models.AIChatSubscriptionStatusActive {
		return false
	}
	if subscription.ExpiresAt == nil {
		return false
	}
	return subscription.ExpiresAt.After(now)
}

func startOfDay(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

func getEnvInt(key string, defaultValue int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
