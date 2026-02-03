package settlement_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Settlement rekonsiliasi uang tunai saat kasir tutup kasir (per tanggal, per user)
type Settlement struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Date          time.Time `gorm:"type:date;not null;uniqueIndex:idx_settlement_date_user" json:"date"` // tanggal tutup kasir
	UserID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_settlement_date_user" json:"user_id"`
	ExpectedCash  float64   `gorm:"type:decimal(15,2);not null;default:0" json:"expected_cash"`  // dari sum transaksi tunai paid hari itu
	ActualCash    float64   `gorm:"type:decimal(15,2);not null;default:0" json:"actual_cash"`    // input kasir
	Discrepancy   float64   `gorm:"type:decimal(15,2);default:0" json:"discrepancy"`             // actual - expected
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate set UUID
func (s *Settlement) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
