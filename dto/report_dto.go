package dto

import "github.com/google/uuid"

// ReportSummary agregasi laporan harian (transaksi completed & paid)
type ReportSummary struct {
	Date            string  `json:"date"`             // YYYY-MM-DD
	TotalTransactions int   `json:"total_transactions"`
	TotalSales       float64 `json:"total_sales"`       // total_amount sum
	TotalCash        float64 `json:"total_cash"`        // sum where payment_method = cash
	TotalNonCash     float64 `json:"total_non_cash"`    // sum where payment_method != cash
	TotalDiscount    float64 `json:"total_discount"`
	TotalTax         float64 `json:"total_tax"`
}

// ReportTransactionItem satu transaksi dalam list laporan (dengan nama kasir)
type ReportTransactionItem struct {
	ID               uuid.UUID `json:"id"`
	CustomerName     string    `json:"customer_name"`
	OrderType        string    `json:"order_type"`
	PaymentMethod    string    `json:"payment_method"`
	TotalAmount      float64   `json:"total_amount"`
	ClosedByUserID   *uuid.UUID `json:"closed_by_user_id,omitempty"`
	ClosedByUserName string    `json:"closed_by_user_name"` // dari join users, atau "-" jika null
	CreatedAt        string    `json:"created_at"`
}

// ReportResponse response GET laporan harian (summary + list transaksi)
type ReportResponse struct {
	Summary      ReportSummary           `json:"summary"`
	Transactions []ReportTransactionItem `json:"transactions"`
}

// ChartDailyItem satu titik untuk grafik harian
type ChartDailyItem struct {
	Date               string  `json:"date"` // YYYY-MM-DD
	TotalTransactions  int     `json:"total_transactions"`
	TotalSales         float64 `json:"total_sales"`
}

// ChartMonthlyItem satu titik untuk grafik bulanan
type ChartMonthlyItem struct {
	Month              string  `json:"month"` // YYYY-MM
	TotalTransactions  int     `json:"total_transactions"`
	TotalSales         float64 `json:"total_sales"`
}

// ChartResponse response GET /report/charts (untuk dashboard admin)
type ChartResponse struct {
	Daily   []ChartDailyItem   `json:"daily"`
	Monthly []ChartMonthlyItem `json:"monthly"`
}
