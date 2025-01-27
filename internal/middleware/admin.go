package middleware

import (
	"net/http"
	"os"
	"strings"

	"prototurk/internal/models"
	"prototurk/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func AdminJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Public routes için middleware'i atla
		if c.Request.URL.Path == "/api/admin/login" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, response.Error("UNAUTHORIZED", "Authorization header required", nil))
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, response.Error("UNAUTHORIZED", "Invalid token", nil))
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, response.Error("UNAUTHORIZED", "Invalid token claims", nil))
			c.Abort()
			return
		}

		// Admin ID'yi context'e ekle
		adminID, ok := claims["admin_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, response.Error("UNAUTHORIZED", "Invalid admin token", nil))
			c.Abort()
			return
		}

		// Admin'i veritabanından kontrol et
		db := c.MustGet("db").(*gorm.DB)
		var admin models.Admin
		if err := db.First(&admin, uint(adminID)).Error; err != nil {
			c.JSON(http.StatusUnauthorized, response.Error("UNAUTHORIZED", "Admin not found", nil))
			c.Abort()
			return
		}

		// Admin aktif mi kontrol et
		if !admin.IsActive() {
			c.JSON(http.StatusForbidden, response.Error("FORBIDDEN", "Admin account is not active", nil))
			c.Abort()
			return
		}

		// Admin bilgilerini context'e ekle
		c.Set("admin_id", uint(adminID))
		c.Set("admin_role", admin.Role)
		c.Set("admin", admin)

		c.Next()
	}
}
