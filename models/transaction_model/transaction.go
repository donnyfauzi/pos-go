package transaction_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Transaction represents an order/transaction
type Transaction struct {
	ID            uuid.UUID         `gorm:"type:uuid;primaryKey" json:"id"`
	CustomerName  string            `gorm:"type:varchar(255);not null" json:"customer_name"`
	CustomerPhone string            `gorm:"type:varchar(20);not null" json:"customer_phone"`
	OrderType     string            `gorm:"type:varchar(20);not null;default:'take_away'" json:"order_type"` // dine_in | take_away
	TableNumber   *int              `gorm:"type:int" json:"table_number"`
	PromoCode     string            `gorm:"type:varchar(50)" json:"promo_code"`
	Discount      float64           `gorm:"type:decimal(15,2);default:0" json:"discount"`
	Subtotal      float64           `gorm:"type:decimal(15,2);not null;default:0" json:"subtotal"`             // Total sebelum pajak
	Tax           float64           `gorm:"type:decimal(15,2);not null;default:0" json:"tax"`                  // Pajak (PPN 10%)
	TotalAmount   float64           `gorm:"type:decimal(15,2);not null;default:0" json:"total_amount"`         // Total setelah pajak
	PaymentMethod string            `gorm:"type:varchar(50);not null" json:"payment_method"`                   // cash, credit_card, debit_card, e_wallet
	PaymentStatus string            `gorm:"type:varchar(50);not null;default:'pending'" json:"payment_status"` // pending, paid, cancelled, expired
	OrderStatus   string            `gorm:"type:varchar(50);not null;default:'pending'" json:"order_status"`   // pending, processing, completed, cancelled
	Notes         string            `gorm:"type:text" json:"notes"`
	ExpiredAt     *time.Time        `gorm:"type:timestamp" json:"expired_at"` // Waktu kadaluarsa untuk non-cash payment
	Items         []TransactionItem `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE" json:"items"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	DeletedAt     gorm.DeletedAt    `gorm:"index" json:"-"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// TransactionItem represents items in a transaction
type TransactionItem struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TransactionID uuid.UUID      `gorm:"type:uuid;not null" json:"transaction_id"`
	MenuID        uuid.UUID      `gorm:"type:uuid;not null" json:"menu_id"`
	MenuName      string         `gorm:"type:varchar(255);not null" json:"menu_name"`
	MenuPrice     float64        `gorm:"type:decimal(15,2);not null" json:"menu_price"`
	Quantity      int            `gorm:"type:int;not null" json:"quantity"`
	Subtotal      float64        `gorm:"type:decimal(15,2);not null" json:"subtotal"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (ti *TransactionItem) BeforeCreate(tx *gorm.DB) error {
	if ti.ID == uuid.Nil {
		ti.ID = uuid.New()
	}
	return nil
}
