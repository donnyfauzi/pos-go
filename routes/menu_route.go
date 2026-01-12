package routes

import (
	"pos-go/controllers"
	"pos-go/middleware"

	"github.com/gin-gonic/gin"
)

func MenuRoutes(r *gin.Engine) {
	menu := r.Group("/menu")
	{
		menu.POST("", middleware.AuthMiddleware(), middleware.RequireRole("admin"), controllers.CreateMenu)
	}
}

