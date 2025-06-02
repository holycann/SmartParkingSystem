package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// AuthWebSocketMiddleware authenticates WebSocket connections using JWT tokens
func AuthWebSocketMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the query parameter
		token := c.Query("token")
		if token == "" {
			// Try to get token from Authorization header
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if token == "" {
			log.Println("WebSocket connection attempted without authentication token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required for WebSocket connection"})
			return
		}

		// Parse and validate the token
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			log.Println("JWT_SECRET environment variable not set")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Server configuration error"})
			return
		}

		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !parsedToken.Valid {
			log.Printf("Invalid WebSocket authentication token: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token"})
			return
		}

		// Extract claims
		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			log.Println("Failed to extract claims from WebSocket authentication token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		// Set user ID in context
		userID, ok := claims["userId"].(string)
		if !ok {
			log.Println("User ID not found in WebSocket authentication token claims")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		c.Set("userId", userID)
		c.Next()
	}
}
