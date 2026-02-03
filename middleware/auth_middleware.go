package middleware

import (
	"pos-go/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// getTokenFromRequest mengambil token dari Cookie "token" (untuk frontend) atau header Authorization: Bearer <token> (untuk Postman/API client).
func getTokenFromRequest(c *gin.Context) string {
	// 1. Cek cookie (frontend kirim otomatis dengan withCredentials)
	if token, err := c.Cookie("token"); err == nil && token != "" {
		return token
	}
	// 2. Cek header Authorization: Bearer <token> (Postman / API client)
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimSpace(auth[7:])
	}
	return ""
}

// AuthMiddleware untuk memvalidasi JWT token dari cookie atau Authorization header
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := getTokenFromRequest(c)
		if token == "" {
			utils.ErrorResponseUnauthorized(c, "Token tidak ditemukan. Kirim via cookie 'token' atau header Authorization: Bearer <token>")
			c.Abort()
			return
		}

		// Validasi token
		claims, err := utils.ValidateToken(token)
		if err != nil {
			utils.ErrorResponseUnauthorized(c, "Token tidak valid")
			c.Abort()
			return
		}

		// Simpan claims ke context (bisa dipakai di handler)
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}
