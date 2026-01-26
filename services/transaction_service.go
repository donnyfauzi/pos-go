package services

import (
	"errors"
	"pos-go/config"
	"pos-go/dto"
	menu_model "pos-go/models/menu_model"
	transaction_model "pos-go/models/transaction_model"
	"time"

	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"gorm.io/gorm"
)

// Sentinel errors
var (
	ErrTransactionNotFound = errors.New("Transaksi tidak ditemukan")
	ErrDatabaseError       = errors.New("Database error")
)

type TransactionService struct {
}

func NewTransactionService() TransactionService {
	return TransactionService{}
}

// CreateTransaction creates a new transaction with items
func (s TransactionService) CreateTransaction(req dto.CreateTransactionRequest) (*transaction_model.Transaction, string, string, error) {
	// Start transaction
	tx := config.DB.Begin()
	if tx.Error != nil {
		return nil, "", "", ErrDatabaseError
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Calculate subtotal and prepare items
	var subtotal float64
	var items []transaction_model.TransactionItem

	for _, itemReq := range req.Items {
		var menu menu_model.Menu
		if err := tx.First(&menu, "id = ? AND is_available = ?", itemReq.MenuID, true).Error; err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return nil, "", "", ErrMenuNotFound
			}
			return nil, "", "", ErrDatabaseError
		}

		itemSubtotal := menu.Price * float64(itemReq.Quantity)
		subtotal += itemSubtotal

		item := transaction_model.TransactionItem{
			MenuID:    menu.ID,
			MenuName:  menu.Name,
			MenuPrice: menu.Price,
			Quantity:  itemReq.Quantity,
			Subtotal:  itemSubtotal,
		}
		items = append(items, item)
	}

	// Hitung PPN 10% otomatis
	tax := subtotal * 0.10
	totalAmount := subtotal + tax

	// Create transaction
	var paymentStatus, orderStatus string
	var expiredAt *time.Time

	// LOGIC: Cash = langsung paid, Non-cash = pending
	if req.PaymentMethod == "cash" {
		paymentStatus = "paid"
		orderStatus = "completed"
		// Cash tidak perlu expired time
	} else {
		paymentStatus = "pending"
		orderStatus = "pending"
		// Set expired 24 jam dari sekarang untuk non-cash
		expirationTime := time.Now().Add(24 * time.Hour)
		expiredAt = &expirationTime
	}

	transaction := transaction_model.Transaction{
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		CustomerEmail: req.CustomerEmail,
		TableNumber:   req.TableNumber,
		Subtotal:      subtotal,
		Tax:           tax,
		TotalAmount:   totalAmount,
		PaymentMethod: req.PaymentMethod,
		PaymentStatus: paymentStatus,
		OrderStatus:   orderStatus,
		ExpiredAt:     expiredAt,
		Notes:         req.Notes,
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return nil, "", "", ErrDatabaseError
	}

	// Create transaction items
	for i := range items {
		items[i].TransactionID = transaction.ID
		if err := tx.Create(&items[i]).Error; err != nil {
			tx.Rollback()
			return nil, "", "", ErrDatabaseError
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, "", "", ErrDatabaseError
	}

	// Load items for response
	transaction.Items = items

	// Generate Snap Token untuk non-cash
	var snapToken, snapURL string
	if req.PaymentMethod != "cash" {
		snapToken, snapURL = s.GenerateSnapToken(transaction)
	}

	return &transaction, snapToken, snapURL, nil
}

// GenerateSnapToken untuk request ke Midtrans
func (s TransactionService) GenerateSnapToken(transaction transaction_model.Transaction) (string, string) {
	// Prepare items for Midtrans
	var itemDetails []midtrans.ItemDetails
	for _, item := range transaction.Items {
		itemDetails = append(itemDetails, midtrans.ItemDetails{
			ID:    item.MenuID.String(),
			Name:  item.MenuName,
			Price: int64(item.MenuPrice),
			Qty:   int32(item.Quantity),
		})
	}

	// Tambahkan PPN 10% sebagai line item terpisah
	// Ini penting agar gross_amount = sum of item_details (Midtrans validation)
	if transaction.Tax > 0 {
		itemDetails = append(itemDetails, midtrans.ItemDetails{
			ID:    "TAX-PPN",
			Name:  "PPN 10%",
			Price: int64(transaction.Tax),
			Qty:   1,
		})
	}

	// Prepare request
	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  transaction.ID.String(),
			GrossAmt: int64(transaction.TotalAmount),
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: transaction.CustomerName,
			Email: transaction.CustomerEmail,
			Phone: transaction.CustomerPhone,
		},
		Items: &itemDetails,
	}

	// Request snap token
	snapResp, err := config.MidtransClient.CreateTransaction(req)
	if err != nil {
		return "", ""
	}

	return snapResp.Token, snapResp.RedirectURL
}

// UpdateTransactionStatus updates payment and order status
func (s TransactionService) UpdateTransactionStatus(id uuid.UUID, paymentStatus, orderStatus string) (*transaction_model.Transaction, error) {
	var transaction transaction_model.Transaction
	if err := config.DB.First(&transaction, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTransactionNotFound
		}
		return nil, ErrDatabaseError
	}

	updates := map[string]interface{}{}
	if paymentStatus != "" {
		updates["payment_status"] = paymentStatus
	}
	if orderStatus != "" {
		updates["order_status"] = orderStatus
	}

	if err := config.DB.Model(&transaction).Updates(updates).Error; err != nil {
		return nil, ErrDatabaseError
	}

	// Reload with items
	if err := config.DB.Preload("Items").First(&transaction, "id = ?", id).Error; err != nil {
		return nil, ErrDatabaseError
	}

	return &transaction, nil
}

// GetAllTransactions retrieves all transactions with items
func (s TransactionService) GetAllTransactions() ([]transaction_model.Transaction, error) {
	var transactions []transaction_model.Transaction
	if err := config.DB.Preload("Items").Order("created_at DESC").Find(&transactions).Error; err != nil {
		return nil, ErrDatabaseError
	}
	return transactions, nil
}

// GetTransactionByID retrieves a transaction by ID with items
func (s TransactionService) GetTransactionByID(id uuid.UUID) (*transaction_model.Transaction, error) {
	var transaction transaction_model.Transaction
	if err := config.DB.Preload("Items").First(&transaction, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTransactionNotFound
		}
		return nil, ErrDatabaseError
	}
	return &transaction, nil
}
