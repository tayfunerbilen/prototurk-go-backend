package middleware

import (
	"net/http"
	"os"
	"strings"

	"prototurk/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add JWT secret to context
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			c.JSON(http.StatusInternalServerError, response.Error("SERVER_ERROR", "JWT secret not configured", nil))
			c.Abort()
			return
		}
		c.Set("jwt_secret", jwtSecret)

		// Skip auth for login and register endpoints
		if c.Request.URL.Path == "/api/auth/login" || c.Request.URL.Path == "/api/auth/register" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, response.Error("UNAUTHORIZED", "No authorization header", nil))
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, response.Error("UNAUTHORIZED", "Invalid token", err.Error()))
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("user_id", uint(claims["user_id"].(float64)))
			c.Set("username", claims["username"].(string))
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, response.Error("UNAUTHORIZED", "Invalid token claims", nil))
			c.Abort()
			return
		}
	}
}
