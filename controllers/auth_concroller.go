package controllers

import (
	"net/http"
	user_model "pos-go/models/user_model"
	"pos-go/services"

	"github.com/gin-gonic/gin"
)

var authService = services.NewAuthService()

func Register(c *gin.Context) {
	var input user_model.User

	// Ambil data dari body
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Panggil service
	user, err := authService.Register(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Success
	c.JSON(http.StatusCreated, gin.H{
		"message": "User berhasil dibuat",
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}
