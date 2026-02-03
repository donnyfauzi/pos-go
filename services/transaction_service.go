package services

import (
	"errors"
	"pos-go/config"
	"pos-go/dto"
	menu_model "pos-go/models/menu_model"
	promo_model "pos-go/models/promo_model"
	transaction_model "pos-go/models/transaction_model"
	"strings"
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
	ErrInvalidStatus       = errors.New("Status pesanan tidak valid untuk role ini")
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
	var appliedPromoID *uuid.UUID
	var discount float64

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

	// Jika ada promo_code, validasi dan hitung diskon
	if req.PromoCode != "" {
		promoService := NewPromoService()
		promo, disc, err := promoService.ValidatePromo(req.PromoCode, subtotal)
		if err != nil {
			tx.Rollback()
			return nil, "", "", err
		}
		appliedPromoID = &promo.ID
		discount = disc
		subtotal = subtotal - discount
	}

	// Hitung PPN 10% otomatis dari subtotal setelah diskon
	tax := subtotal * 0.10
	totalAmount := subtotal + tax

	// Validasi order type & table number
	if req.OrderType == "dine_in" {
		if req.TableNumber == nil || *req.TableNumber <= 0 {
			tx.Rollback()
			return nil, "", "", errors.New("Nomor meja wajib untuk makan di tempat")
		}
	} else {
		// take_away: pastikan table_number kosong
		req.TableNumber = nil
	}

	// Create transaction
	var paymentStatus, orderStatus string
	var expiredAt *time.Time

	// LOGIC BARU (best practice kasir):
	// - Cash: pending dulu, dibayar di kasir -> kasir update jadi paid
	// - Non-cash: pending dulu, jadi paid via webhook Midtrans
	paymentStatus = "pending"
	orderStatus = "pending"
	if req.PaymentMethod != "cash" {
		// Set expired 24 jam dari sekarang untuk non-cash
		expirationTime := time.Now().Add(24 * time.Hour)
		expiredAt = &expirationTime
	}

	transaction := transaction_model.Transaction{
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		OrderType:     req.OrderType,
		TableNumber:   req.TableNumber,
		PromoCode:     req.PromoCode,
		Discount:      discount,
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

	// Untuk non-cash: dapatkan snap token SEBELUM commit. Jika gagal, rollback (data tidak masuk DB).
	var snapToken, snapURL string
	if req.PaymentMethod != "cash" {
		transaction.Items = items
		var errSnap error
		snapToken, snapURL, errSnap = s.GenerateSnapToken(transaction)
		if errSnap != nil {
			tx.Rollback()
			return nil, "", "", errSnap
		}
	}

	// Commit transaction (hanya setelah token berhasil untuk non-cash)
	if err := tx.Commit().Error; err != nil {
		return nil, "", "", ErrDatabaseError
	}

	// Jika ada promo yang terpakai: increment usage_count secara atomik, nonaktifkan jika habis
	if appliedPromoID != nil {
		// Increment atomik agar concurrent request tidak saling timpa
		config.DB.Model(&promo_model.Promo{}).Where("id = ?", *appliedPromoID).Update("usage_count", gorm.Expr("usage_count + ?", 1))
		var promo promo_model.Promo
		if err := config.DB.First(&promo, "id = ?", *appliedPromoID).Error; err == nil && promo.UsageLimit > 0 && promo.UsageCount >= promo.UsageLimit {
			config.DB.Model(&promo_model.Promo{}).Where("id = ?", *appliedPromoID).Update("is_active", false)
		}
	}

	// Load items for response
	transaction.Items = items

	return &transaction, snapToken, snapURL, nil
}

// GenerateSnapToken untuk request ke Midtrans. Mengembalikan error jika Midtrans gagal.
// Penting: sum(item_details) harus persis sama dengan GrossAmt (Midtrans validation).
func (s TransactionService) GenerateSnapToken(transaction transaction_model.Transaction) (string, string, error) {
	// Prepare items for Midtrans (harga per item * qty = subtotal per line)
	var itemDetails []midtrans.ItemDetails
	for _, item := range transaction.Items {
		itemDetails = append(itemDetails, midtrans.ItemDetails{
			ID:    item.MenuID.String(),
			Name:  item.MenuName,
			Price: int64(item.MenuPrice),
			Qty:   int32(item.Quantity),
		})
	}

	// Jika ada diskon: tambahkan line item negatif agar sum(item_details) = TotalAmount
	// sum(items) = subtotal sebelum diskon; TotalAmount = subtotal_setelah_diskon + tax
	if transaction.Discount > 0 {
		itemDetails = append(itemDetails, midtrans.ItemDetails{
			ID:    "DISC",
			Name:  "Diskon",
			Price: -int64(transaction.Discount),
			Qty:   1,
		})
	}

	// Hitung sum item_details sebelum PPN (agar bisa set PPN = sisa ke GrossAmt)
	var sumBeforeTax int64
	for _, it := range itemDetails {
		sumBeforeTax += it.Price * int64(it.Qty)
	}
	grossAmt := int64(transaction.TotalAmount)
	taxLineAmount := grossAmt - sumBeforeTax
	// PPN 10% sebagai line item (jumlah diset agar sum(item_details) = GrossAmt)
	if taxLineAmount != 0 {
		itemDetails = append(itemDetails, midtrans.ItemDetails{
			ID:    "TAX-PPN",
			Name:  "PPN 10%",
			Price: taxLineAmount,
			Qty:   1,
		})
	}

	// Prepare request
	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  transaction.ID.String(),
			GrossAmt: grossAmt,
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: transaction.CustomerName,
			Email: "",
			Phone: transaction.CustomerPhone,
		},
		Items: &itemDetails,
	}

	// Request snap token
	snapResp, err := config.MidtransClient.CreateTransaction(req)
	if err != nil {
		return "", "", err
	}
	if snapResp == nil || snapResp.Token == "" {
		return "", "", errors.New("gagal mendapatkan token pembayaran dari Midtrans")
	}

	return snapResp.Token, snapResp.RedirectURL, nil
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

// UpdateOrderStatusForRole mengubah order_status dengan aturan per role (kasir / koki). closedByUserID diisi saat kasir menandai completed (untuk laporan).
func (s TransactionService) UpdateOrderStatusForRole(id uuid.UUID, role string, newStatus string, closedByUserID *uuid.UUID) (*transaction_model.Transaction, error) {
	var tx transaction_model.Transaction
	if err := config.DB.First(&tx, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTransactionNotFound
		}
		return nil, ErrDatabaseError
	}

	current := tx.OrderStatus

	switch role {
	case "kasir":
		switch newStatus {
		case "cooking":
			if current != "pending" {
				return nil, ErrInvalidStatus
			}
		case "completed":
			if current != "ready" {
				return nil, ErrInvalidStatus
			}
			// Best practice: hanya boleh selesai jika sudah dibayar
			if tx.PaymentStatus != "paid" {
				return nil, errors.New("Transaksi belum berstatus paid, tidak dapat diselesaikan")
			}
		case "cancelled":
			// Kasir boleh batalkan pesanan yang belum selesai: pending, cooking, atau ready
			if current != "pending" && current != "cooking" && current != "ready" {
				return nil, ErrInvalidStatus
			}
		default:
			return nil, ErrInvalidStatus
		}
	case "koki":
		switch newStatus {
		case "ready":
			if current != "cooking" {
				return nil, ErrInvalidStatus
			}
		default:
			return nil, ErrInvalidStatus
		}
	default:
		return nil, ErrInvalidStatus
	}

	updates := map[string]interface{}{"order_status": newStatus}
	// Saat kasir menandai selesai (completed), simpan kasir yang memproses untuk laporan per kasir
	if newStatus == "completed" && closedByUserID != nil {
		updates["closed_by_user_id"] = closedByUserID
	}
	if err := config.DB.Model(&tx).Updates(updates).Error; err != nil {
		return nil, ErrDatabaseError
	}

	if err := config.DB.Preload("Items").First(&tx, "id = ?", id).Error; err != nil {
		return nil, ErrDatabaseError
	}

	return &tx, nil
}

// CancelOrder membatalkan pesanan. Kasir atau Admin; status saat ini harus pending, cooking, atau ready.
func (s TransactionService) CancelOrder(id uuid.UUID, role string) (*transaction_model.Transaction, error) {
	role = strings.ToLower(strings.TrimSpace(role))
	if role != "kasir" && role != "admin" {
		return nil, ErrInvalidStatus
	}
	// Admin diperlakukan seperti kasir untuk aksi batalkan
	effectiveRole := role
	if role == "admin" {
		effectiveRole = "kasir"
	}
	return s.UpdateOrderStatusForRole(id, effectiveRole, "cancelled", nil)
}

// ConfirmCashPaid - khusus kasir untuk konfirmasi pembayaran tunai. closedByUserID = user_id kasir yang login (untuk laporan per kasir).
func (s TransactionService) ConfirmCashPaid(id uuid.UUID, closedByUserID *uuid.UUID) (*transaction_model.Transaction, error) {
	var transaction transaction_model.Transaction
	if err := config.DB.First(&transaction, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTransactionNotFound
		}
		return nil, ErrDatabaseError
	}

	// Validasi: hanya untuk cash & masih pending
	if transaction.PaymentMethod != "cash" {
		return nil, errors.New("Hanya transaksi tunai yang bisa dikonfirmasi oleh kasir")
	}
	if transaction.PaymentStatus != "pending" {
		return nil, errors.New("Status pembayaran sudah bukan pending")
	}
	if transaction.OrderStatus == "cancelled" {
		return nil, errors.New("Transaksi sudah dibatalkan")
	}

	updates := map[string]interface{}{"payment_status": "paid"}
	if closedByUserID != nil {
		updates["closed_by_user_id"] = closedByUserID
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
