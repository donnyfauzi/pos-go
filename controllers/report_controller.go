package controllers

import (
	"errors"
	"pos-go/services"
	"pos-go/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var reportService services.ReportService = services.NewReportService()

// GetReportByDate mengembalikan laporan harian (agregasi + list transaksi) untuk tanggal yang diberikan.
// Query: date (YYYY-MM-DD). Akses: admin (semua transaksi) & kasir (hanya transaksi yang ditutup oleh kasir itu).
func GetReportByDate(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		utils.ErrorResponseBadRequest(c, "Parameter date (YYYY-MM-DD) wajib diisi", nil)
		return
	}

	var cashierID *uuid.UUID
	roleVal, hasRole := c.Get("role")
	if hasRole {
		role, _ := roleVal.(string)
		if role == "kasir" || role == "Kasir" {
			uidVal, exists := c.Get("user_id")
			if exists && uidVal != nil {
				switch v := uidVal.(type) {
				case string:
					parsed, err := uuid.Parse(v)
					if err == nil {
						cashierID = &parsed
					}
				case uuid.UUID:
					cashierID = &v
				}
			}
		}
	}

	report, err := reportService.GetReportByDate(dateStr, cashierID)
	if err != nil {
		if errors.Is(err, services.ErrDatabaseError) {
			utils.ErrorResponseInternal(c, "Gagal mengambil laporan")
			return
		}
		utils.ErrorResponseBadRequest(c, "Tanggal tidak valid. Gunakan format YYYY-MM-DD", nil)
		return
	}

	utils.SuccessResponseOK(c, "Laporan harian berhasil diambil", report)
}

// GetReportCharts mengembalikan data grafik harian dan bulanan. Query: days (default 7), months (default 6). Akses: admin.
func GetReportCharts(c *gin.Context) {
	days := 7
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 90 {
			days = parsed
		}
	}
	months := 6
	if m := c.Query("months"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed > 0 && parsed <= 24 {
			months = parsed
		}
	}

	charts, err := reportService.GetReportCharts(days, months)
	if err != nil {
		if errors.Is(err, services.ErrDatabaseError) {
			utils.ErrorResponseInternal(c, "Gagal mengambil data grafik")
			return
		}
		utils.ErrorResponseInternal(c, "Gagal mengambil data grafik")
		return
	}

	utils.SuccessResponseOK(c, "Data grafik berhasil diambil", charts)
}

