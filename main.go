package main

import (
	"log"
	"pos-go/config"
	"pos-go/routes"
	"pos-go/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// Koneksi ke database
	config.ConnectDatabase()

	// Inisialisasi Gin
	r := gin.Default()

	// Setup routes
	routes.AuthRoutes(r)

	r.GET("/ping", func(c *gin.Context) {
		utils.SuccessResponseOK(c, "API is running", nil)
	})

	log.Println("Server berjalan di http://localhost:8080")
	r.Run(":8080")
}
