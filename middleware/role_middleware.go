package middleware

import (
	"pos-go/utils"

	"github.com/gin-gonic/gin"
)

// RequireRole untuk check apakah user punya role yang diizinkan
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil role dari context (dari AuthMiddleware)
		role, exists := c.Get("role")
		if !exists {
			utils.ErrorResponseUnauthorized(c, "Role tidak ditemukan")
			c.Abort()
			return
		}

		roleStr := role.(string)

		// Cek apakah role user ada di allowedRoles
		for _, allowedRole := range allowedRoles {
			if roleStr == allowedRole {
				c.Next()
				return
			}
		}

		// Role tidak diizinkan
		utils.ErrorResponseForbidden(c, "Akses ditolak anda tidak bisa mengakses endpoint ini")
		c.Abort()
	}
}
