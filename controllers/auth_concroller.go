package controllers

import (
	"net/http"
	"pos-go/dto"
	"pos-go/services"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var authService = services.NewAuthService()

func Register(c *gin.Context) {
	var input dto.RegisterDTO

	if err := c.ShouldBindJSON(&input); err != nil {
		errors := make(map[string]string)

		if ve, ok := err.(validator.ValidationErrors); ok {
			for _, fe := range ve {
				field := fe.Field()

				if fe.Tag() == "required" {
					errors[field] = "Nama wajib diisi"
				} else if fe.Tag() == "email" {
					errors[field] = "Format email tidak valid"
				} else if fe.Tag() == "min" {
					errors[field] = field + " minimal harus " + fe.Param() + " karakter"
				} else if fe.Tag() == "oneof" {
					errors[field] = field + " harus salah satu dari: " + fe.Param()
				} else {
					errors[field] = "Field tidak valid"
				}
			}

			c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
			return
		}

		// Kalau bukan error dari validator
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Jalankan logic register
	user, err := authService.Register(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Sukses register
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
