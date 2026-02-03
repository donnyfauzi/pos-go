package dto

// CreateSettlementRequest body untuk POST settlement (tutup kasir)
type CreateSettlementRequest struct {
	Date       string  `json:"date" binding:"required"`        // YYYY-MM-DD
	ActualCash float64 `json:"actual_cash" binding:"required"`  // uang tunai yang diserahkan kasir
}

// SettlementResponse response settlement (untuk GET / POST)
type SettlementResponse struct {
	ID           string  `json:"id"`
	Date         string  `json:"date"`           // YYYY-MM-DD
	UserID       string  `json:"user_id"`
	ExpectedCash float64 `json:"expected_cash"`
	ActualCash   float64 `json:"actual_cash"`
	Discrepancy  float64 `json:"discrepancy"`
	CreatedAt    string  `json:"created_at"`
}

// GetSettlementResponse response GET /settlement?date= (expected_cash + settlement jika sudah ada)
type GetSettlementResponse struct {
	ExpectedCash float64             `json:"expected_cash"` // uang tunai yang seharusnya (dari transaksi)
	Settlement   *SettlementResponse `json:"settlement"`   // null jika belum tutup kasir
}

// SettlementStatusItem satu baris status settlement per kasir (untuk admin GET /settlement/status-by-date)
type SettlementStatusItem struct {
	UserID       string   `json:"user_id"`
	UserName     string   `json:"user_name"`
	ExpectedCash float64  `json:"expected_cash"`
	Settlement   *SettlementResponse `json:"settlement,omitempty"` // null jika belum tutup kasir
}

// GetSettlementStatusByDateResponse response GET /settlement/status-by-date (admin)
type GetSettlementStatusByDateResponse struct {
	Date    string                `json:"date"`
	Items   []SettlementStatusItem `json:"items"`
}
