package main

import (
	"log"
	"pos-go/config"
	"pos-go/routes"

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
		c.JSON(200, gin.H{
			"message": "API is running",
		})
	})

	log.Println("Server berjalan di http://localhost:8080")
	r.Run(":8080")
}
