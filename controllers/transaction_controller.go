package controllers

import (
	"errors"
	"pos-go/dto"
	"pos-go/services"
	"pos-go/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var transactionService services.TransactionService = services.NewTransactionService()

func CreateTransaction(c *gin.Context) {
	var req dto.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorsMap := make(map[string]string)

		if ve, ok := err.(validator.ValidationErrors); ok {
			for _, fe := range ve {
				field := fe.Field()
				switch fe.Tag() {
				case "required":
					errorsMap[field] = "Field wajib diisi"
				case "email":
					errorsMap[field] = "Email tidak valid"
				case "oneof":
					errorsMap[field] = "Pilihan tidak valid"
				case "min":
					errorsMap[field] = "Nilai minimal tidak terpenuhi"
				default:
					errorsMap[field] = "Field tidak valid"
				}
			}
			utils.ErrorResponseBadRequest(c, "Format data tidak valid", errorsMap)
			return
		}

		utils.ErrorResponseBadRequest(c, "Format data tidak valid", nil)
		return
	}

	transaction, snapToken, snapURL, err := transactionService.CreateTransaction(req)
	if err != nil {
		if errors.Is(err, services.ErrMenuNotFound) {
			utils.ErrorResponseNotFound(c, "Menu tidak ditemukan atau tidak tersedia")
			return
		}
		utils.ErrorResponseInternal(c, "Gagal membuat transaksi")
		return
	}

	// Response dengan snap token untuk non-cash
	response := dto.TransactionResponseWithSnap{
		ID:            transaction.ID,
		CustomerName:  transaction.CustomerName,
		CustomerPhone: transaction.CustomerPhone,
		CustomerEmail: transaction.CustomerEmail,
		TableNumber:   transaction.TableNumber,
		Subtotal:      transaction.Subtotal,
		Tax:           transaction.Tax,
		TotalAmount:   transaction.TotalAmount,
		PaymentMethod: transaction.PaymentMethod,
		PaymentStatus: transaction.PaymentStatus,
		OrderStatus:   transaction.OrderStatus,
		Notes:         transaction.Notes,
		SnapToken:     snapToken,
		SnapURL:       snapURL,
		CreatedAt:     transaction.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     transaction.UpdatedAt.Format(time.RFC3339),
	}

	// Add expired_at if exists
	if transaction.ExpiredAt != nil {
		expiredAtStr := transaction.ExpiredAt.Format(time.RFC3339)
		response.ExpiredAt = &expiredAtStr
	}

	// Convert items
	var items []dto.TransactionItemResponse
	for _, item := range transaction.Items {
		items = append(items, dto.TransactionItemResponse{
			ID:        item.ID,
			MenuID:    item.MenuID,
			MenuName:  item.MenuName,
			MenuPrice: item.MenuPrice,
			Quantity:  item.Quantity,
			Subtotal:  item.Subtotal,
		})
	}
	response.Items = items

	utils.SuccessResponseCreated(c, "Transaksi berhasil dibuat", response)
}

// HandleMidtransNotification untuk webhook dari Midtrans
func HandleMidtransNotification(c *gin.Context) {
	var notification dto.MidtransNotification
	if err := c.ShouldBindJSON(&notification); err != nil {
		utils.ErrorResponseBadRequest(c, "Invalid notification", nil)
		return
	}

	// Verify signature dari Midtrans (penting untuk security!)
	// TODO: Implementasi signature verification

	// Parse transaction ID
	transactionID, err := uuid.Parse(notification.OrderID)
	if err != nil {
		utils.ErrorResponseBadRequest(c, "Invalid transaction ID", nil)
		return
	}

	// Update status berdasarkan response Midtrans
	var paymentStatus, orderStatus string

	switch notification.TransactionStatus {
	case "capture", "settlement":
		paymentStatus = "paid"
		orderStatus = "completed"
	case "pending":
		paymentStatus = "pending"
		orderStatus = "pending"
	case "expire":
		// Transaction expired (lewat 24 jam)
		paymentStatus = "expired"
		orderStatus = "cancelled"
	case "deny", "cancel":
		// Transaction cancelled by user or system
		paymentStatus = "cancelled"
		orderStatus = "cancelled"
	default:
		paymentStatus = "pending"
		orderStatus = "pending"
	}

	_, err = transactionService.UpdateTransactionStatus(transactionID, paymentStatus, orderStatus)
	if err != nil {
		utils.ErrorResponseInternal(c, "Failed to update transaction")
		return
	}

	c.JSON(200, gin.H{"status": "success"})
}

func GetAllTransactions(c *gin.Context) {
	transactions, err := transactionService.GetAllTransactions()
	if err != nil {
		utils.ErrorResponseInternal(c, "Gagal mengambil data transaksi")
		return
	}

	utils.SuccessResponseOK(c, "Berhasil mengambil data transaksi", transactions)
}

func GetTransactionByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponseBadRequest(c, "ID transaksi tidak valid", nil)
		return
	}

	transaction, err := transactionService.GetTransactionByID(id)
	if err != nil {
		if errors.Is(err, services.ErrTransactionNotFound) {
			utils.ErrorResponseNotFound(c, "Transaksi tidak ditemukan")
			return
		}
		utils.ErrorResponseInternal(c, "Gagal mengambil data transaksi")
		return
	}

	utils.SuccessResponseOK(c, "Berhasil mengambil data transaksi", transaction)
}
