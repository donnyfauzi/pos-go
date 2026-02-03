package dto

import "github.com/google/uuid"

// CreateTransactionRequest represents the request to create a new transaction
type CreateTransactionRequest struct {
	CustomerName  string                         `json:"customer_name" binding:"required"`
	CustomerPhone string                         `json:"customer_phone" binding:"required"`
	OrderType     string                         `json:"order_type" binding:"required,oneof=dine_in take_away"`
	TableNumber   *int                           `json:"table_number" binding:"omitempty"`
	PaymentMethod string                         `json:"payment_method" binding:"required,oneof=cash credit_card debit_card e_wallet"`
	Notes         string                         `json:"notes" binding:"omitempty"`
	PromoCode     string                         `json:"promo_code" binding:"omitempty"`
	Items         []CreateTransactionItemRequest `json:"items" binding:"required,min=1,dive"`
}

// CreateTransactionItemRequest represents an item in the transaction
type CreateTransactionItemRequest struct {
	MenuID   uuid.UUID `json:"menu_id" binding:"required"`
	Quantity int       `json:"quantity" binding:"required,min=1"`
}

// TransactionResponse represents the response for a transaction
type TransactionResponse struct {
	ID            uuid.UUID                 `json:"id"`
	CustomerName  string                    `json:"customer_name"`
	CustomerPhone string                    `json:"customer_phone"`
	OrderType     string                    `json:"order_type"`
	TableNumber   *int                      `json:"table_number"`
	PromoCode     string                    `json:"promo_code"`
	Discount      float64                   `json:"discount"`
	Subtotal      float64                   `json:"subtotal"`      // Total sebelum pajak
	Tax           float64                   `json:"tax"`           // Pajak PPN 10%
	TotalAmount   float64                   `json:"total_amount"`  // Total setelah pajak
	PaymentMethod string                    `json:"payment_method"`
	PaymentStatus string                    `json:"payment_status"`
	OrderStatus   string                    `json:"order_status"`
	Notes         string                    `json:"notes"`
	ExpiredAt     *string                   `json:"expired_at,omitempty"` // Waktu kadaluarsa
	Items         []TransactionItemResponse `json:"items"`
	CreatedAt     string                    `json:"created_at"`
	UpdatedAt     string                    `json:"updated_at"`
}

// TransactionItemResponse represents an item in the transaction response
type TransactionItemResponse struct {
	ID        uuid.UUID `json:"id"`
	MenuID    uuid.UUID `json:"menu_id"`
	MenuName  string    `json:"menu_name"`
	MenuPrice float64   `json:"menu_price"`
	Quantity  int       `json:"quantity"`
	Subtotal  float64   `json:"subtotal"`
}

// GetTransactionsResponse represents the response for getting all transactions
type GetTransactionsResponse struct {
	Data    []TransactionResponse `json:"data"`
	Message string                `json:"message"`
}

// GetTransactionResponse represents the response for getting a single transaction
type GetTransactionResponse struct {
	Data    TransactionResponse `json:"data"`
	Message string              `json:"message"`
}

// CreateTransactionResponse represents the response after creating a transaction
type CreateTransactionResponse struct {
	Data    TransactionResponse `json:"data"`
	Message string              `json:"message"`
}

// TransactionResponseWithSnap untuk response yang include snap token (untuk non-cash payment)
type TransactionResponseWithSnap struct {
	ID            uuid.UUID                 `json:"id"`
	CustomerName  string                    `json:"customer_name"`
	CustomerPhone string                    `json:"customer_phone"`
	OrderType     string                    `json:"order_type"`
	TableNumber   *int                      `json:"table_number"`
	PromoCode     string                    `json:"promo_code"`
	Discount      float64                   `json:"discount"`
	Subtotal      float64                   `json:"subtotal"`             // Total sebelum pajak
	Tax           float64                   `json:"tax"`                  // Pajak PPN 10%
	TotalAmount   float64                   `json:"total_amount"`         // Total setelah pajak
	PaymentMethod string                    `json:"payment_method"`
	PaymentStatus string                    `json:"payment_status"`
	OrderStatus   string                    `json:"order_status"`
	Notes         string                    `json:"notes"`
	ExpiredAt     *string                   `json:"expired_at,omitempty"` // Waktu kadaluarsa
	Items         []TransactionItemResponse `json:"items"`
	SnapToken     string                    `json:"snap_token,omitempty"` // Untuk non-cash
	SnapURL       string                    `json:"snap_url,omitempty"`   // Untuk non-cash
	CreatedAt     string                    `json:"created_at"`
	UpdatedAt     string                    `json:"updated_at"`
}

// MidtransNotification untuk terima notifikasi dari Midtrans webhook
type MidtransNotification struct {
	TransactionID     string `json:"transaction_id"`
	OrderID           string `json:"order_id"`
	GrossAmount       string `json:"gross_amount"`
	PaymentType       string `json:"payment_type"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	SignatureKey      string `json:"signature_key"`
}

// UpdateOrderStatusRequest untuk update status pesanan (kasir / koki)
type UpdateOrderStatusRequest struct {
	OrderStatus string `json:"order_status" binding:"required,oneof=pending cooking ready completed cancelled"`
}

// ReceiptItemResponse item untuk struk
type ReceiptItemResponse struct {
	MenuName  string  `json:"menu_name"`
	Quantity  int     `json:"quantity"`
	MenuPrice float64 `json:"menu_price"`
	Subtotal  float64 `json:"subtotal"`
}

// ReceiptResponse data struk untuk print (GET /transaction/:id/receipt)
type ReceiptResponse struct {
	ID                 uuid.UUID              `json:"id"`
	CreatedAt          string                 `json:"created_at"`
	CustomerName       string                 `json:"customer_name"`
	CustomerPhone      string                 `json:"customer_phone"`
	OrderType          string                 `json:"order_type"`
	TableNumber        *int                   `json:"table_number,omitempty"`
	Items              []ReceiptItemResponse `json:"items"`
	Subtotal           float64                `json:"subtotal"`
	Discount           float64                `json:"discount"`
	Tax                float64                `json:"tax"`
	TotalAmount        float64                `json:"total_amount"`
	PaymentMethod      string                 `json:"payment_method"`
	PaymentStatus      string                 `json:"payment_status"`
	ClosedByUserName   string                 `json:"closed_by_user_name"`
}
