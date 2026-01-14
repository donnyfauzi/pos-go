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
		auth.GET("/me", middleware.AuthMiddleware(), controllers.GetCurrentUser)
		auth.POST("/register", middleware.AuthMiddleware(), middleware.RequireRole("admin"), controllers.Register)
		auth.PUT("/change-password", middleware.AuthMiddleware(), controllers.ChangePassword)
	}

	user := r.Group("/user")
	{
		user.GET("", middleware.AuthMiddleware(), middleware.RequireRole("admin"), controllers.GetAllUsers)
		user.DELETE("/:id", middleware.AuthMiddleware(), middleware.RequireRole("admin"), controllers.DeleteUser)
	}
}


