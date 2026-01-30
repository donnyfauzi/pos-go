package main

import (
	"log"
	"pos-go/config"
	database "pos-go/database/migrations"
	"pos-go/routes"
	"pos-go/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Koneksi ke database
	config.ConnectDatabase()

	// Inisialisasi Midtrans
	config.InitMidtrans()

	// Migrasi seed database untuk admin awal
	database.SeedAdmin()

	// Set Gin mode (hilangkan debug mode warning) - HARUS SEBELUM gin.Default()
	gin.SetMode(gin.ReleaseMode)

	// Inisialisasi Gin
	r := gin.Default()

	// Static file handler untuk serve uploaded images
	r.Static("/uploads", "./uploads")

	// CORS Middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,      //untuk cookie!
		MaxAge:           12 * 3600, // 12 jam
	}))

	// Set trusted proxies (hilangkan proxy warning)
	r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	// Setup routes
	routes.AuthRoutes(r)
	routes.CategoryRoutes(r)
	routes.MenuRoutes(r)
	routes.TransactionRoutes(r)
	routes.PromoRoutes(r)

	r.GET("/ping", func(c *gin.Context) {
		utils.SuccessResponseOK(c, "API sukses berjalan", nil)
	})

	log.Println("Server berjalan di http://localhost:8080")
	r.Run(":8080")
}
