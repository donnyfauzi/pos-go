package services

import (
	"errors"
	"pos-go/config"
	"pos-go/dto"
	settlement_model "pos-go/models/settlement_model"
	transaction_model "pos-go/models/transaction_model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrSettlementAlreadyExists = errors.New("Settlement untuk tanggal ini sudah ada")

type SettlementService struct{}

func NewSettlementService() SettlementService {
	return SettlementService{}
}

// expectedCashFromTransactions menghitung total tunai yang seharusnya (transaksi cash, paid, completed, di tanggal tersebut)
func expectedCashFromTransactions(startOfDay, endOfDay time.Time) (float64, error) {
	var total float64
	err := config.DB.Model(&transaction_model.Transaction{}).
		Select("COALESCE(SUM(total_amount), 0)").
		Where("created_at >= ? AND created_at < ? AND order_status = ? AND payment_status = ? AND payment_method = ?",
			startOfDay, endOfDay, "completed", "paid", "cash").
		Scan(&total).Error
	if err != nil {
		return 0, ErrDatabaseError
	}
	return total, nil
}

// CreateSettlement menyimpan settlement (tutup kasir). Satu settlement per (date, user_id).
func (s SettlementService) CreateSettlement(userID uuid.UUID, dateStr string, actualCash float64) (*dto.SettlementResponse, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, err
	}
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.AddDate(0, 0, 1)

	expectedCash, err := expectedCashFromTransactions(startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}

	discrepancy := actualCash - expectedCash

	// Cek sudah ada settlement untuk (date, user_id)?
	var existing settlement_model.Settlement
	if err := config.DB.Where("date = ? AND user_id = ?", startOfDay, userID).First(&existing).Error; err == nil {
		return nil, ErrSettlementAlreadyExists
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrDatabaseError
	}

	settlement := settlement_model.Settlement{
		Date:         startOfDay,
		UserID:       userID,
		ExpectedCash: expectedCash,
		ActualCash:   actualCash,
		Discrepancy:  discrepancy,
	}
	if err := config.DB.Create(&settlement).Error; err != nil {
		return nil, ErrDatabaseError
	}

	return toSettlementResponse(&settlement), nil
}

// GetSettlementByDateAndUser mengembalikan settlement untuk tanggal dan user (kasir) tertentu. Nil jika belum ada.
func (s SettlementService) GetSettlementByDateAndUser(dateStr string, userID uuid.UUID) (*dto.SettlementResponse, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, err
	}
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	var settlement settlement_model.Settlement
	if err := config.DB.Where("date = ? AND user_id = ?", startOfDay, userID).First(&settlement).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // belum ada settlement
		}
		return nil, ErrDatabaseError
	}
	return toSettlementResponse(&settlement), nil
}

// GetSettlementWithExpected mengembalikan expected_cash untuk tanggal + settlement jika sudah ada (untuk GET /settlement?date=)
func (s SettlementService) GetSettlementWithExpected(dateStr string, userID uuid.UUID) (*dto.GetSettlementResponse, error) {
	expectedCash, err := s.GetExpectedCashForDate(dateStr)
	if err != nil {
		return nil, err
	}
	settlement, err := s.GetSettlementByDateAndUser(dateStr, userID)
	if err != nil {
		return nil, err
	}
	return &dto.GetSettlementResponse{
		ExpectedCash: expectedCash,
		Settlement:   settlement,
	}, nil
}

// GetExpectedCashForDate mengembalikan expected cash untuk tanggal (tanpa simpan settlement). Untuk tampilan form.
func (s SettlementService) GetExpectedCashForDate(dateStr string) (float64, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return 0, err
	}
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.AddDate(0, 0, 1)
	return expectedCashFromTransactions(startOfDay, endOfDay)
}

func toSettlementResponse(s *settlement_model.Settlement) *dto.SettlementResponse {
	dateStr := s.Date.Format("2006-01-02")
	return &dto.SettlementResponse{
		ID:           s.ID.String(),
		Date:         dateStr,
		UserID:       s.UserID.String(),
		ExpectedCash: s.ExpectedCash,
		ActualCash:   s.ActualCash,
		Discrepancy:  s.Discrepancy,
		CreatedAt:    s.CreatedAt.Format(time.RFC3339),
	}
}
