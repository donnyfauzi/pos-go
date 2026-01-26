package routes

import (
	"pos-go/controllers"
	"pos-go/middleware"

	"github.com/gin-gonic/gin"
)

func CategoryRoutes(r *gin.Engine) {
	category := r.Group("/category")
	{
		// Public endpoint - untuk customer lihat categories
		category.GET("/public", controllers.GetAllCategories)

		// Admin only endpoints
		category.POST("", middleware.AuthMiddleware(), middleware.RequireRole("admin"), controllers.CreateCategory)
		category.GET("", middleware.AuthMiddleware(), middleware.RequireRole("admin"), controllers.GetAllCategories)
	}
}
