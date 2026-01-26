package models

import (
	category_model "pos-go/models/category_model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Menu struct {
	ID          uuid.UUID               `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string                  `gorm:"not null" json:"name"`
	Description string                  `json:"description"`
	Price       float64                 `gorm:"not null" json:"price"`
	Image       string                  `json:"image"` // URL atau path ke gambar
	IsAvailable bool                    `json:"is_available"`
	CategoryID  uuid.UUID               `gorm:"type:uuid;not null" json:"category_id"`
	Category    category_model.Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	gorm.Model
}

// Hook untuk generate UUID sebelum insert
func (menu *Menu) BeforeCreate(tx *gorm.DB) (err error) {
	menu.ID = uuid.New()
	return
}
