package routes

import (
	"pos-go/controllers"
	"pos-go/middleware"

	"github.com/gin-gonic/gin"
)

func ReportRoutes(r *gin.Engine) {
	report := r.Group("/report")
	report.Use(middleware.AuthMiddleware())
	{
		// GET /report?date=2026-01-30 — laporan harian (summary + list transaksi). Admin & Kasir.
		report.GET("", middleware.RequireRole("admin", "kasir"), controllers.GetReportByDate)
		// GET /report/charts?days=7&months=6 — data grafik harian & bulanan. Admin.
		report.GET("/charts", middleware.RequireRole("admin"), controllers.GetReportCharts)
	}
}
