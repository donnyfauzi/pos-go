package routes

import (
	"pos-go/controllers"
	"pos-go/middleware"

	"github.com/gin-gonic/gin"
)

func PromoRoutes(r *gin.Engine) {
	promo := r.Group("/promo")
	{
		// Public - get active promos dan validate promo code
		promo.GET("/active", controllers.GetActivePromos)
		promo.POST("/validate", controllers.ValidatePromo)

		// Admin only - CRUD promo
		promo.POST("", middleware.AuthMiddleware(), middleware.RequireRole("admin"), controllers.CreatePromo)
		promo.GET("", middleware.AuthMiddleware(), middleware.RequireRole("admin"), controllers.GetAllPromos)
		promo.GET("/:id", middleware.AuthMiddleware(), middleware.RequireRole("admin"), controllers.GetPromoByID)
		promo.PUT("/:id", middleware.AuthMiddleware(), middleware.RequireRole("admin"), controllers.UpdatePromo)
		promo.DELETE("/:id", middleware.AuthMiddleware(), middleware.RequireRole("admin"), controllers.DeletePromo)
	}
}