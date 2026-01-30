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

		// Public - webhook dari Midtrans (PENTING!)
		transaction.POST("/notification", controllers.HandleMidtransNotification)

		// Admin & Kasir - lihat semua transaksi
		transaction.GET("", middleware.AuthMiddleware(), controllers.GetAllTransactions)

		// Admin & Kasir - lihat detail transaksi
		transaction.GET("/:id", middleware.AuthMiddleware(), controllers.GetTransactionByID)

		// Kasir only - konfirmasi pembayaran tunai
		transaction.PATCH("/:id/cash-paid", middleware.AuthMiddleware(), middleware.RequireRole("kasir"), controllers.ConfirmCashPaid)

		// Kasir & Koki - update order_status dengan aturan per role
		transaction.PATCH("/:id/order-status", middleware.AuthMiddleware(), middleware.RequireRole("kasir", "koki"), controllers.UpdateOrderStatus)
	}
}
