package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// Claims represents the JWT claims
type Claims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.StandardClaims
}

// AuthMiddleware validates JWT tokens and sets user info in the context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Check if the header has the Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		// Extract the token
		tokenString := parts[1]

		// Get the JWT secret from environment variables
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "ramaa212!" // Default fallback
		}

		// Parse and validate the token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		// Handle token validation errors
		if err != nil {
			log.Printf("Token validation error: %v", err)
			if ve, ok := err.(*jwt.ValidationError); ok {
				if ve.Errors&jwt.ValidationErrorExpired != 0 {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				}
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to parse token"})
			}
			c.Abort()
			return
		}

		// Check if the token is valid
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Check token expiration
		if claims.ExpiresAt < time.Now().Unix() {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
			c.Abort()
			return
		}

		// Set user information in the context
		c.Set("userId", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		// Continue to the next middleware/handler
		c.Next()
	}
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(userID, email, role string) (string, error) {
	// Get the JWT secret from environment variables
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "ramaa212!" // Default fallback
	}

	// Set token expiration time (24 hours)
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create the JWT claims
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "smart-parking-system",
		},
	}

	// Create the token with the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// CheckUserOwnership ensures the authenticated user is accessing their own data
func CheckUserOwnership(c *gin.Context) {
	// Get the target user ID from the URL parameter
	targetUserID := c.Param("id")

	// Get the authenticated user's ID from the context
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		c.Abort()
		return
	}

	// Check if the authenticated user is accessing their own data
	if userID != targetUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access to another user's data"})
		c.Abort()
		return
	}

	c.Next()
}
