package routes

import (
	"pos-go/controllers"
	"pos-go/middleware"

	"github.com/gin-gonic/gin"
)

func MenuRoutes(r *gin.Engine) {
	menu := r.Group("/menu")
	{
		// Admin only - semua menu
		menu.GET("", middleware.AuthMiddleware(), middleware.RequireRole("admin"), controllers.GetAllMenus)
		
		// Public - hanya menu yang available
		menu.GET("/public", controllers.GetPublicMenus)
		
		// Admin only - create menu
		menu.POST("", middleware.AuthMiddleware(), middleware.RequireRole("admin"), controllers.CreateMenu)
		
		// Admin only - update menu
		menu.PUT("/:id", middleware.AuthMiddleware(), middleware.RequireRole("admin"), controllers.UpdateMenu)
		
		// Admin only - delete menu
		menu.DELETE("/:id", middleware.AuthMiddleware(), middleware.RequireRole("admin"), controllers.DeleteMenu)
	}
}

