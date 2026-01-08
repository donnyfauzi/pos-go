package middleware

import (
	"pos-go/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware untuk memvalidasi JWT token dari cookie
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil token dari cookie
		token, err := c.Cookie("token")
		if err != nil {
			utils.ErrorResponseUnauthorized(c, "Token tidak ditemukan")
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

