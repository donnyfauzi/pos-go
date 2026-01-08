package main

import (
	"log"
	"pos-go/config"
	"pos-go/routes"
	"pos-go/utils"
	"pos-go/database/migrations"

	"github.com/gin-gonic/gin"
)

func main() {
	// Koneksi ke database
	config.ConnectDatabase()

	// Migrasi seed database untuk admin awal 
	database.SeedAdmin()

	// Set Gin mode (hilangkan debug mode warning) - HARUS SEBELUM gin.Default()
	gin.SetMode(gin.ReleaseMode)

	// Inisialisasi Gin
	r := gin.Default()

	// Set trusted proxies (hilangkan proxy warning)
	r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	// Setup routes
	routes.AuthRoutes(r)

	r.GET("/ping", func(c *gin.Context) {
		utils.SuccessResponseOK(c, "API is running", nil)
	})

	log.Println("Server berjalan di http://localhost:8080")
	r.Run(":8080")
}
