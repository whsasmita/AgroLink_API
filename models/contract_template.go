package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ContractTemplate struct {
	ID        uuid.UUID `gorm:"type:char(36);primary_key"`
	Name      string    `gorm:"type:varchar(100);not null"`
	Content   string    `gorm:"type:text;not null"`
	IsDefault bool      `gorm:"default:false"`
}

func (ct *ContractTemplate) BeforeCreate(tx *gorm.DB) (err error) {
	if ct.ID == uuid.Nil {
		ct.ID = uuid.New()
	}
	return
}