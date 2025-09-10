package models

import (
	// "gorm/io/gorm"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Payout merepresentasikan distribusi dana keluar ke setiap pekerja.
type Payout struct {
	ID            uuid.UUID `gorm:"type:char(36);primary_key"`
	TransactionID uuid.UUID `gorm:"type:char(36);not null"`
	WorkerID      uuid.UUID `gorm:"type:char(36);not null"`
	Amount        float64   `gorm:"type:decimal(12,2)"`
	Status        string    `gorm:"type:enum('pending_disbursement','completed','failed');default:'pending_disbursement'"`
	ReleasedAt    time.Time

	// Relasi yang benar HANYA ke Transaction dan Worker
	Transaction Transaction `gorm:"foreignKey:TransactionID"`
	Worker      Worker      `gorm:"foreignKey:WorkerID"`
}

func (p *Payout) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	p.ReleasedAt = time.Now()
	return
}