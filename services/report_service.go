package services

import (
	"pos-go/config"
	"pos-go/dto"
	transaction_model "pos-go/models/transaction_model"
	user_model "pos-go/models/user_model"
	"time"

	"github.com/google/uuid"
)

type ReportService struct{}

func NewReportService() ReportService {
	return ReportService{}
}

// GetReportByDate mengembalikan laporan harian: agregasi + list transaksi (completed & paid) untuk tanggal tertentu.
// Jika cashierID != nil, hanya transaksi yang closed_by_user_id = cashierID (laporan per kasir).
// dateStr format: YYYY-MM-DD.
func (s ReportService) GetReportByDate(dateStr string, cashierID *uuid.UUID) (*dto.ReportResponse, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, err
	}
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.AddDate(0, 0, 1)

	var transactions []transaction_model.Transaction
	q := config.DB.Where(
		"created_at >= ? AND created_at < ? AND order_status = ? AND payment_status = ?",
		startOfDay, endOfDay, "completed", "paid",
	)
	if cashierID != nil {
		q = q.Where("closed_by_user_id = ?", *cashierID)
	}
	if err := q.Order("created_at ASC").Find(&transactions).Error; err != nil {
		return nil, ErrDatabaseError
	}

	// Agregasi
	var totalSales, totalCash, totalNonCash, totalDiscount, totalTax float64
	for _, t := range transactions {
		totalSales += t.TotalAmount
		totalDiscount += t.Discount
		totalTax += t.Tax
		if t.PaymentMethod == "cash" {
			totalCash += t.TotalAmount
		} else {
			totalNonCash += t.TotalAmount
		}
	}

	summary := dto.ReportSummary{
		Date:               dateStr,
		TotalTransactions:  len(transactions),
		TotalSales:         totalSales,
		TotalCash:          totalCash,
		TotalNonCash:       totalNonCash,
		TotalDiscount:      totalDiscount,
		TotalTax:           totalTax,
	}

	// Map closed_by_user_id -> name
	userIDs := make([]uuid.UUID, 0)
	for _, t := range transactions {
		if t.ClosedByUserID != nil {
			userIDs = append(userIDs, *t.ClosedByUserID)
		}
	}
	userNameMap := make(map[uuid.UUID]string)
	if len(userIDs) > 0 {
		var users []user_model.User
		if err := config.DB.Where("id IN ?", userIDs).Find(&users).Error; err == nil {
			for _, u := range users {
				userNameMap[u.ID] = u.Name
			}
		}
	}

	// Build list transaksi dengan nama kasir
	list := make([]dto.ReportTransactionItem, 0, len(transactions))
	for _, t := range transactions {
		closedByName := "-"
		if t.ClosedByUserID != nil {
			if name, ok := userNameMap[*t.ClosedByUserID]; ok {
				closedByName = name
			}
		}
		list = append(list, dto.ReportTransactionItem{
			ID:               t.ID,
			CustomerName:     t.CustomerName,
			OrderType:        t.OrderType,
			PaymentMethod:    t.PaymentMethod,
			TotalAmount:      t.TotalAmount,
			ClosedByUserID:   t.ClosedByUserID,
			ClosedByUserName: closedByName,
			CreatedAt:        t.CreatedAt.Format(time.RFC3339),
		})
	}

	return &dto.ReportResponse{
		Summary:      summary,
		Transactions: list,
	}, nil
}

// GetReportCharts mengembalikan data untuk grafik: harian (N hari terakhir) dan bulanan (N bulan terakhir).
func (s ReportService) GetReportCharts(days, months int) (*dto.ChartResponse, error) {
	if days <= 0 {
		days = 7
	}
	if months <= 0 {
		months = 6
	}

	now := time.Now().UTC()
	startDaily := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -days+1)
	endMonthly := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -months, 0)

	var txDaily []transaction_model.Transaction
	if err := config.DB.Where(
		"created_at >= ? AND order_status = ? AND payment_status = ?",
		startDaily, "completed", "paid",
	).Find(&txDaily).Error; err != nil {
		return nil, ErrDatabaseError
	}

	// Agregasi per hari
	dailyMap := make(map[string]struct {
		count int
		sales float64
	})
	for d := 0; d < days; d++ {
		day := startDaily.AddDate(0, 0, d)
		key := day.Format("2006-01-02")
		dailyMap[key] = struct {
			count int
			sales float64
		}{0, 0}
	}
	for _, t := range txDaily {
		key := t.CreatedAt.UTC().Format("2006-01-02")
		if v, ok := dailyMap[key]; ok {
			v.count++
			v.sales += t.TotalAmount
			dailyMap[key] = v
		} else {
			dailyMap[key] = struct {
				count int
				sales float64
			}{1, t.TotalAmount}
		}
	}
	dailyList := make([]dto.ChartDailyItem, 0, days)
	for d := 0; d < days; d++ {
		day := startDaily.AddDate(0, 0, d)
		key := day.Format("2006-01-02")
		v := dailyMap[key]
		dailyList = append(dailyList, dto.ChartDailyItem{
			Date:              key,
			TotalTransactions: v.count,
			TotalSales:        v.sales,
		})
	}

	// Transaksi untuk bulanan (last N months)
	var txMonthly []transaction_model.Transaction
	if err := config.DB.Where(
		"created_at >= ? AND order_status = ? AND payment_status = ?",
		endMonthly, "completed", "paid",
	).Find(&txMonthly).Error; err != nil {
		return nil, ErrDatabaseError
	}
	monthlyMap := make(map[string]struct {
		count int
		sales float64
	})
	for m := 0; m < months; m++ {
		month := endMonthly.AddDate(0, m, 0)
		key := month.Format("2006-01")
		monthlyMap[key] = struct {
			count int
			sales float64
		}{0, 0}
	}
	for _, t := range txMonthly {
		key := t.CreatedAt.UTC().Format("2006-01")
		if v, ok := monthlyMap[key]; ok {
			v.count++
			v.sales += t.TotalAmount
			monthlyMap[key] = v
		} else {
			monthlyMap[key] = struct {
				count int
				sales float64
			}{1, t.TotalAmount}
		}
	}
	monthlyList := make([]dto.ChartMonthlyItem, 0, months)
	for m := 0; m < months; m++ {
		month := endMonthly.AddDate(0, m, 0)
		key := month.Format("2006-01")
		v := monthlyMap[key]
		monthlyList = append(monthlyList, dto.ChartMonthlyItem{
			Month:             key,
			TotalTransactions: v.count,
			TotalSales:        v.sales,
		})
	}

	return &dto.ChartResponse{
		Daily:   dailyList,
		Monthly: monthlyList,
	}, nil
}

