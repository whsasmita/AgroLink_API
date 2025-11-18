package models

import (
	// "gorm/io/gorm"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Payout merepresentasikan distribusi dana keluar ke setiap pekerja.
// type Payout struct {
// 	ID            uuid.UUID `gorm:"type:char(36);primary_key"`
// 	TransactionID uuid.UUID `gorm:"type:char(36);not null"`
// 	PayeeID   uuid.UUID `gorm:"column:payee_id;type:char(36);not null"`
// 	PayeeType string    `gorm:"type:enum('worker','driver');not null"`
// 	Amount        float64   `gorm:"type:decimal(12,2)"`
// 	Status        string    `gorm:"type:enum('pending_disbursement','completed','failed');default:'pending_disbursement'"`
// 	ReleasedAt    time.Time
// 	TransferProofURL *string `gorm:"type:text"`

// 	// Relasi yang benar HANYA ke Transaction dan Worker
// 	Transaction Transaction `gorm:"foreignKey:TransactionID"`
// 	Worker      *Worker     `gorm:"foreignKey:PayeeID;constraint:false"`
//     Driver      *Driver     `gorm:"foreignKey:PayeeID;constraint:false"`
// }

type Payout struct {
    ID               uuid.UUID `gorm:"type:char(36);primary_key"`
    TransactionID    uuid.UUID `gorm:"type:char(36);not null"`
    PayeeID          uuid.UUID `gorm:"column:payee_id;type:char(36);not null"`
    PayeeType        string    `gorm:"type:enum('worker','driver');not null"`
    Amount           float64   `gorm:"type:decimal(12,2)"`
    Status           string    `gorm:"type:enum('pending_disbursement','completed','failed');default:'pending_disbursement'"`
    ReleasedAt       time.Time
    TransferProofURL *string   `gorm:"type:text"`

    // Relasi
    Transaction Transaction `gorm:"foreignKey:TransactionID"`
	Worker *Worker `gorm:"-"`
	Driver *Driver `gorm:"-"`

    // JANGAN gunakan pointer relasi polimorfik dengan FK
    // Gunakan asosiasi manual atau skip constraint sepenuhnya
}

func (p *Payout) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	p.ReleasedAt = time.Now()
	return
}

func (p *Payout) LoadPayee(db *gorm.DB) error {
	switch p.PayeeType {
	case "worker":
		var worker Worker
		// Muat data Worker dan relasi User-nya
		if err := db.Preload("User").Where("user_id = ?", p.PayeeID).First(&worker).Error; err != nil {
			return err
		}
		p.Worker = &worker
	case "driver":
		var driver Driver
		// Muat data Driver dan relasi User-nya
		if err := db.Preload("User").Where("user_id = ?", p.PayeeID).First(&driver).Error; err != nil {
			return err
		}
		p.Driver = &driver
	default:
		return fmt.Errorf("unknown payee type: %s", p.PayeeType)
	}
	return nil
}