package routes

import (
	"pos-go/controllers"
	"pos-go/middleware"

	"github.com/gin-gonic/gin"
)

func TransactionRoutes(r *gin.Engine) {
	transaction := r.Group("/transaction")
	{
		// Public - create transaction (checkout untuk customer)
		transaction.POST("", controllers.CreateTransaction)

		// Admin & Kasir - lihat semua transaksi
		transaction.GET("", middleware.AuthMiddleware(), controllers.GetAllTransactions)

		// Admin & Kasir - lihat detail transaksi
		transaction.GET("/:id", middleware.AuthMiddleware(), controllers.GetTransactionByID)
	}
}
