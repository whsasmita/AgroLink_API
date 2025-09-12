package repositories

import (
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type ContractTemplateRepository interface {
	GetDefault() (*models.ContractTemplate, error)
}

type contractTemplateRepository struct{ db *gorm.DB }

func NewContractTemplateRepository(db *gorm.DB) ContractTemplateRepository {
	return &contractTemplateRepository{db: db}
}

func (r *contractTemplateRepository) GetDefault() (*models.ContractTemplate, error) {
	var template models.ContractTemplate
	err := r.db.Where("is_default = ?", true).First(&template).Error
	return &template, err
}