package repositories

import (
	"context"

	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type WebhookLogRepository interface {
	Create(ctx context.Context, log *models.WebhookLog) error
	MarkProcessed(ctx context.Context, id string) error
	AttachError(ctx context.Context, id string, errMsg string) error
	UpdateSignatureValid(ctx context.Context, id string, valid bool) error
}

type webhookLogRepository struct {
	db *gorm.DB
}

func NewWebhookLogRepository(db *gorm.DB) WebhookLogRepository {
	return &webhookLogRepository{db: db}
}

func (r *webhookLogRepository) Create(ctx context.Context, logEntry *models.WebhookLog) error {
	return r.db.WithContext(ctx).Create(logEntry).Error
}

func (r *webhookLogRepository) MarkProcessed(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&models.WebhookLog{}).
		Where("id = ?", id).
		Update("processed", true).Error
}

func (r *webhookLogRepository) AttachError(ctx context.Context, id string, errMsg string) error {
	return r.db.WithContext(ctx).
		Model(&models.WebhookLog{}).
		Where("id = ?", id).
		Update("error_message", errMsg).Error
}

func (r *webhookLogRepository) UpdateSignatureValid(ctx context.Context, id string, valid bool) error {
	return r.db.WithContext(ctx).
		Model(&models.WebhookLog{}).
		Where("id = ?", id).
		Update("signature_valid", valid).Error
}
