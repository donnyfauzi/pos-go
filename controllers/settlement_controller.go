package controllers

import (
	"errors"
	"pos-go/services"
	"pos-go/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var settlementService services.SettlementService = services.NewSettlementService()

// GetSettlement GET /settlement?date=YYYY-MM-DD — expected_cash + settlement (jika sudah ada) untuk user yang login.
func GetSettlement(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		utils.ErrorResponseBadRequest(c, "Parameter date (YYYY-MM-DD) wajib diisi", nil)
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists || userIDVal == nil {
		utils.ErrorResponseUnauthorized(c, "User tidak ditemukan")
		return
	}
	userID, err := uuid.Parse(userIDVal.(string))
	if err != nil {
		utils.ErrorResponseUnauthorized(c, "User ID tidak valid")
		return
	}

	resp, err := settlementService.GetSettlementWithExpected(dateStr, userID)
	if err != nil {
		if errors.Is(err, services.ErrDatabaseError) {
			utils.ErrorResponseInternal(c, "Gagal mengambil data settlement")
			return
		}
		utils.ErrorResponseBadRequest(c, "Tanggal tidak valid. Gunakan format YYYY-MM-DD", nil)
		return
	}

	utils.SuccessResponseOK(c, "Data settlement berhasil diambil", resp)
}

// CreateSettlement POST /settlement — simpan settlement (tutup kasir). Body: { date, actual_cash }.
func CreateSettlement(c *gin.Context) {
	var req struct {
		Date       string  `json:"date" binding:"required"`
		ActualCash float64 `json:"actual_cash" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponseBadRequest(c, "Format data tidak valid. Perlu date (YYYY-MM-DD) dan actual_cash", nil)
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists || userIDVal == nil {
		utils.ErrorResponseUnauthorized(c, "User tidak ditemukan")
		return
	}
	userID, err := uuid.Parse(userIDVal.(string))
	if err != nil {
		utils.ErrorResponseUnauthorized(c, "User ID tidak valid")
		return
	}

	settlement, err := settlementService.CreateSettlement(userID, req.Date, req.ActualCash)
	if err != nil {
		if errors.Is(err, services.ErrSettlementAlreadyExists) {
			utils.ErrorResponseBadRequest(c, "Settlement untuk tanggal ini sudah ada", nil)
			return
		}
		if errors.Is(err, services.ErrDatabaseError) {
			utils.ErrorResponseInternal(c, "Gagal menyimpan settlement")
			return
		}
		utils.ErrorResponseBadRequest(c, "Tanggal tidak valid. Gunakan format YYYY-MM-DD", nil)
		return
	}

	utils.SuccessResponseCreated(c, "Settlement berhasil disimpan", settlement)
}

// GetSettlementStatusByDate GET /settlement/status-by-date?date=YYYY-MM-DD — admin only. Daftar kasir + expected cash + status settlement.
func GetSettlementStatusByDate(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		utils.ErrorResponseBadRequest(c, "Parameter date (YYYY-MM-DD) wajib diisi", nil)
		return
	}

	resp, err := settlementService.GetSettlementStatusByDate(dateStr)
	if err != nil {
		if errors.Is(err, services.ErrDatabaseError) {
			utils.ErrorResponseInternal(c, "Gagal mengambil status settlement")
			return
		}
		utils.ErrorResponseBadRequest(c, "Tanggal tidak valid. Gunakan format YYYY-MM-DD", nil)
		return
	}

	utils.SuccessResponseOK(c, "Status settlement per kasir berhasil diambil", resp)
}
