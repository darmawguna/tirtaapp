package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil user role yang sudah di-set oleh AuthMiddleware
		userRole, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User role not found in token"})
			return
		}

		// Periksa apakah rolenya adalah "admin"
		if userRole.(string) != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "This action requires admin privileges"})
			return
		}

		// Jika admin, lanjutkan ke handler berikutnya
		c.Next()
	}
}