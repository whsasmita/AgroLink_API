package repositories

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GeminiChatRepository interface {
	CreateTurn(turn *models.AIChatTurn) error
	CountTurnsSince(scope string, userID *uuid.UUID, ipAddress string, since time.Time) (int64, error)
	FindRecentTurns(scope string, userID *uuid.UUID, ipAddress string, limit int) ([]models.AIChatTurn, error)
	FindSubscriptionByUserID(userID uuid.UUID) (*models.AIChatPremiumSubscription, error)
	FindSubscriptionByOrderID(orderID string) (*models.AIChatPremiumSubscription, error)
	UpsertSubscription(subscription *models.AIChatPremiumSubscription) error
	UpdateSubscription(subscription *models.AIChatPremiumSubscription) error
}

type geminiChatRepository struct {
	db *gorm.DB
}

func NewGeminiChatRepository(db *gorm.DB) GeminiChatRepository {
	return &geminiChatRepository{db: db}
}

func (r *geminiChatRepository) CreateTurn(turn *models.AIChatTurn) error {
	return r.db.Create(turn).Error
}

func (r *geminiChatRepository) CountTurnsSince(scope string, userID *uuid.UUID, ipAddress string, since time.Time) (int64, error) {
	query := r.db.Model(&models.AIChatTurn{}).
		Where("scope = ? AND created_at >= ?", scope, since)

	if userID != nil {
		query = query.Where("user_id = ?", userID.String())
	} else {
		query = query.Where("ip_address = ?", ipAddress)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *geminiChatRepository) FindRecentTurns(scope string, userID *uuid.UUID, ipAddress string, limit int) ([]models.AIChatTurn, error) {
	query := r.db.Model(&models.AIChatTurn{}).
		Where("scope = ?", scope)

	if userID != nil {
		query = query.Where("user_id = ?", userID.String())
	} else {
		query = query.Where("ip_address = ?", ipAddress)
	}

	var turns []models.AIChatTurn
	if err := query.Order("created_at DESC").Limit(limit).Find(&turns).Error; err != nil {
		return nil, err
	}

	return turns, nil
}

func (r *geminiChatRepository) FindSubscriptionByUserID(userID uuid.UUID) (*models.AIChatPremiumSubscription, error) {
	var subscription models.AIChatPremiumSubscription
	err := r.db.Where("user_id = ?", userID.String()).First(&subscription).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &subscription, nil
}

func (r *geminiChatRepository) FindSubscriptionByOrderID(orderID string) (*models.AIChatPremiumSubscription, error) {
	var subscription models.AIChatPremiumSubscription
	err := r.db.Where("midtrans_order_id = ?", orderID).First(&subscription).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &subscription, nil
}

func (r *geminiChatRepository) UpsertSubscription(subscription *models.AIChatPremiumSubscription) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		UpdateAll: true,
	}).Create(subscription).Error
}

func (r *geminiChatRepository) UpdateSubscription(subscription *models.AIChatPremiumSubscription) error {
	return r.db.Save(subscription).Error
}
