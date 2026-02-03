package routes

import (
	"pos-go/controllers"
	"pos-go/middleware"

	"github.com/gin-gonic/gin"
)

func SettlementRoutes(r *gin.Engine) {
	settlement := r.Group("/settlement")
	settlement.Use(middleware.AuthMiddleware())
	{
		// GET /settlement/status-by-date?date= — admin only. Status settlement per kasir.
		settlement.GET("/status-by-date", middleware.RequireRole("admin"), controllers.GetSettlementStatusByDate)
		// GET /settlement?date=YYYY-MM-DD — expected_cash + settlement (jika sudah ada). Kasir & Admin.
		settlement.GET("", middleware.RequireRole("admin", "kasir"), controllers.GetSettlement)
		// POST /settlement — simpan settlement (tutup kasir). Kasir & Admin.
		settlement.POST("", middleware.RequireRole("admin", "kasir"), controllers.CreateSettlement)
	}
}
