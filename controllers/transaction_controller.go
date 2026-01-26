package controllers

import (
	"errors"
	"pos-go/dto"
	"pos-go/services"
	"pos-go/utils"

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

	transaction, err := transactionService.CreateTransaction(req)
	if err != nil {
		if errors.Is(err, services.ErrMenuNotFound) {
			utils.ErrorResponseNotFound(c, "Menu tidak ditemukan atau tidak tersedia")
			return
		}
		utils.ErrorResponseInternal(c, "Gagal membuat transaksi")
		return
	}

	utils.SuccessResponseCreated(c, "Transaksi berhasil dibuat", transaction)
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

