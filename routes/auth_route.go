package routes

import (
	"pos-go/controllers"
	"pos-go/middleware"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(r *gin.Engine) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", controllers.Login)
		auth.POST("/register", middleware.AuthMiddleware(), middleware.RequireRole("admin"), controllers.Register)
		auth.PUT("/change-password", middleware.AuthMiddleware(), controllers.ChangePassword)
	}
}
