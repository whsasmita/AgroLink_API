package repositories

import (
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type InvoiceRepository interface {
	Create(invoice *models.Invoice) error
	FindByID(id string) (*models.Invoice, error)
	FindByProjectID(projectID string) (*models.Invoice, error)
	UpdateStatus(id string, status string) error
}

type invoiceRepository struct{ db *gorm.DB }

func NewInvoiceRepository(db *gorm.DB) InvoiceRepository {
	return &invoiceRepository{db: db}
}

func (r *invoiceRepository) Create(invoice *models.Invoice) error {
	return r.db.Create(invoice).Error
}

func (r *invoiceRepository) FindByID(id string) (*models.Invoice, error) {
	var invoice models.Invoice
	err := r.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

func (r *invoiceRepository) FindByProjectID(projectID string) (*models.Invoice, error) {
	var invoice models.Invoice
	err := r.db.Where("project_id = ?", projectID).First(&invoice).Error
	return &invoice, err
}

func (r *invoiceRepository) UpdateStatus(id string, status string) error {
	return r.db.Model(&models.Invoice{}).Where("id = ?", id).Update("status", status).Error
}