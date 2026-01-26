package services

import (
	"errors"
	"pos-go/config"
	"pos-go/dto"
	menu_model "pos-go/models/menu_model"
	transaction_model "pos-go/models/transaction_model"

	"github.com/google/uuid"
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
func (s TransactionService) CreateTransaction(req dto.CreateTransactionRequest) (*transaction_model.Transaction, error) {
	// Start transaction
	tx := config.DB.Begin()
	if tx.Error != nil {
		return nil, ErrDatabaseError
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Calculate total amount and prepare items
	var totalAmount float64
	var items []transaction_model.TransactionItem

	for _, itemReq := range req.Items {
		// Get menu details
		var menu menu_model.Menu
		if err := tx.First(&menu, "id = ? AND is_available = ?", itemReq.MenuID, true).Error; err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return nil, ErrMenuNotFound
			}
			return nil, ErrDatabaseError
		}

		// Calculate subtotal
		subtotal := menu.Price * float64(itemReq.Quantity)
		totalAmount += subtotal

		// Create transaction item
		item := transaction_model.TransactionItem{
			MenuID:    menu.ID,
			MenuName:  menu.Name,
			MenuPrice: menu.Price,
			Quantity:  itemReq.Quantity,
			Subtotal:  subtotal,
		}
		items = append(items, item)
	}

	// Create transaction
	// Untuk resto cepat saji: langsung set completed & paid
	transaction := transaction_model.Transaction{
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		CustomerEmail: req.CustomerEmail,
		TableNumber:   req.TableNumber,
		TotalAmount:   totalAmount,
		PaymentMethod: req.PaymentMethod,
		PaymentStatus: "paid",      // Langsung paid untuk fast food
		OrderStatus:   "completed", // Langsung completed untuk fast food
		Notes:         req.Notes,
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return nil, ErrDatabaseError
	}

	// Create transaction items
	for i := range items {
		items[i].TransactionID = transaction.ID
		if err := tx.Create(&items[i]).Error; err != nil {
			tx.Rollback()
			return nil, ErrDatabaseError
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, ErrDatabaseError
	}

	// Load items for response
	transaction.Items = items

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

