package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/holycan/smart-parking-system/database"
	"github.com/holycan/smart-parking-system/middleware"
	"github.com/holycan/smart-parking-system/models"
)

// RegisterUser handles user models.registration
func RegisterUser(c *gin.Context) {
	var req models.UserRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user models.with this email already exists
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", req.Email).Scan(&count)
	if err != nil {
		log.Printf("Error checking for existing user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check for existing user"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Generate a new UUID for the user
	userID := uuid.New().String()

	// Insert the new user models.into the database
	_, err = database.DB.Exec(
		"INSERT INTO users (id, email, password, first_name, last_name, phone, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		userID, req.Email, string(hashedPassword), req.FirstName, req.LastName, req.Phone, time.Now(), time.Now(),
	)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(userID, req.Email, "user")
	if err != nil {
		log.Printf("Error generating token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	// Return the user models.and token
	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":        userID,
			"email":     req.Email,
			"firstName": req.FirstName,
			"lastName":  req.LastName,
			"phone":     req.Phone,
		},
		"token": token,
	})
}

// LoginUser handles user models.login
func LoginUser(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the user models.by email
	var user models.User
	err := database.DB.QueryRow(
		"SELECT id, email, password, first_name, last_name, phone, role FROM users WHERE email = $1",
		req.Email,
	).Scan(&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.Phone, &user.Role)

	if err != nil {
		log.Printf("Error finding user: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		log.Printf("Invalid password attempt for user models.%s: %v", user.Email, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	// Return the user models.and token
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"id":        user.ID,
			"email":     user.Email,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"phone":     user.Phone,
		},
		"token": token,
	})
}

// LogoutUser handles user models.logout
func LogoutUser(c *gin.Context) {
	// Since we're using JWT tokens and not storing them server-side,
	// we don't need to do anything special here.
	// The client will remove the token.
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// ValidateToken checks if a token is valid
func ValidateToken(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the JWT secret from environment variables
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "ramaa212!" // Default fallback
	}

	// Parse and validate the token
	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(req.Token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusOK, gin.H{"valid": false})
		return
	}

	// Check token expiration
	if claims.ExpiresAt < time.Now().Unix() {
		c.JSON(http.StatusOK, gin.H{"valid": false})
		return
	}

	// Token is valid
	c.JSON(http.StatusOK, gin.H{"valid": true})
}

// GetUserProfile retrieves the current user's profile
func GetUserProfile(c *gin.Context) {
	// Get user models.ID from the context (set by the auth middleware)
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Fetch user models.from database
	var user models.User
	err := database.DB.QueryRow(
		"SELECT id, email, first_name, last_name, phone, created_at, updated_at FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Phone, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		log.Printf("Error fetching user models.profile: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user models.profile"})
		return
	}

	// Return the user models.profile
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":        user.ID,
			"email":     user.Email,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"phone":     user.Phone,
			"createdAt": user.CreatedAt,
			"updatedAt": user.UpdatedAt,
		},
	})
}

// UpdateUserProfile updates the current user's profile
func UpdateUserProfile(c *gin.Context) {
	// Get user models.ID from the context (set by the auth middleware)
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.UserProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start a transaction
	tx, err := database.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}
	defer tx.Rollback()

	// Update user models.profile fields
	if req.FirstName != "" || req.LastName != "" || req.Phone != "" {
		// Prepare the update query
		query := "UPDATE users SET updated_at = $1"
		params := []interface{}{time.Now()}
		paramCount := 2

		// Add fields to update
		if req.FirstName != "" {
			query += fmt.Sprintf(", first_name = $%d", paramCount)
			params = append(params, req.FirstName)
			paramCount++
		}

		if req.LastName != "" {
			query += fmt.Sprintf(", last_name = $%d", paramCount)
			params = append(params, req.LastName)
			paramCount++
		}

		if req.Phone != "" {
			query += fmt.Sprintf(", phone = $%d", paramCount)
			params = append(params, req.Phone)
			paramCount++
		}

		// Add WHERE clause
		query += fmt.Sprintf(" WHERE id = $%d", paramCount)
		params = append(params, userID)

		// Execute the update
		_, err = tx.Exec(query, params...)
		if err != nil {
			log.Printf("Error updating user models.profile: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}
	}

	// Update password if provided
	if req.NewPassword != "" && req.CurrentPassword != "" {
		// Verify current password
		var passwordHash string
		err := tx.QueryRow("SELECT password FROM users WHERE id = $1", userID).Scan(&passwordHash)
		if err != nil {
			log.Printf("Error fetching user models.password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify current password"})
			return
		}

		// Compare the provided password with the stored hash
		err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.CurrentPassword))
		if err != nil {
			log.Printf("Invalid current password: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is incorrect"})
			return
		}

		// Hash the new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing new password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process new password"})
			return
		}

		// Update the password
		_, err = tx.Exec("UPDATE users SET password = $1, updated_at = $2 WHERE id = $3",
			string(hashedPassword), time.Now(), userID)
		if err != nil {
			log.Printf("Error updating password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
			return
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	// Fetch updated user models.from database
	var user models.User
	err = database.DB.QueryRow(
		"SELECT id, email, first_name, last_name, phone, created_at, updated_at FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Phone, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		log.Printf("Error fetching updated user models.profile: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Profile updated but failed to fetch updated profile"})
		return
	}

	// Return the updated user models.profile
	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user": gin.H{
			"id":        user.ID,
			"email":     user.Email,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"phone":     user.Phone,
			"createdAt": user.CreatedAt,
			"updatedAt": user.UpdatedAt,
		},
	})
}

// RefreshToken generates a new token for the user
func RefreshToken(c *gin.Context) {
	// Get user models.ID from the context (set by the auth middleware)
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Generate a new token
	token, err := middleware.GenerateToken(userID.(string), email.(string), role.(string))
	if err != nil {
		log.Printf("Error generating refresh token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
		"token":   token,
	})
}

// RequestPasswordReset initiates the password reset process
func RequestPasswordReset(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user models.exists
	var userID string
	err := database.DB.QueryRow("SELECT id FROM users WHERE email = $1", req.Email).Scan(&userID)
	if err != nil {
		// Don't reveal if email exists or not for security reasons
		c.JSON(http.StatusOK, gin.H{"message": "If your email is registered, you will receive password reset instructions"})
		return
	}

	// Generate a reset token
	resetToken := uuid.New().String()

	// Store the reset token in the database with expiration time (1 hour)
	expiryTime := time.Now().Add(1 * time.Hour)
	_, err = database.DB.Exec(
		"INSERT INTO password_reset_tokens (user_id, token, expires_at) VALUES ($1, $2, $3) ON CONFLICT (user_id) DO UPDATE SET token = $2, expires_at = $3",
		userID, resetToken, expiryTime,
	)

	if err != nil {
		log.Printf("Error storing reset token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password reset request"})
		return
	}

	// In a real application, send an email with the reset link
	resetLink := fmt.Sprintf("https://yourapp.com/reset-password?token=%s", resetToken)
	log.Printf("Password reset link for %s: %s", req.Email, resetLink)

	// For demo purposes, we'll just return success
	c.JSON(http.StatusOK, gin.H{
		"message": "If your email is registered, you will receive password reset instructions",
		// Only include this in development environment
		"dev_reset_link": resetLink,
	})
}

// ResetPassword completes the password reset process
func ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify the reset token
	var userID string
	err := database.DB.QueryRow(
		"SELECT user_id FROM password_reset_tokens WHERE token = $1 AND expires_at > $2",
		req.Token, time.Now(),
	).Scan(&userID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired reset token"})
		return
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Update the user's password
	_, err = database.DB.Exec(
		"UPDATE users SET password = $1, updated_at = $2 WHERE id = $3",
		string(hashedPassword), time.Now(), userID,
	)

	if err != nil {
		log.Printf("Error updating password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// Delete the used reset token
	_, err = database.DB.Exec("DELETE FROM password_reset_tokens WHERE token = $1", req.Token)
	if err != nil {
		log.Printf("Error deleting reset token: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password has been reset successfully"})
}

// VerifyEmail verifies a user's email address
func VerifyEmail(c *gin.Context) {
	token := c.Param("token")

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification token is required"})
		return
	}

	// Verify the email verification token
	var userID string
	err := database.DB.QueryRow(
		"SELECT user_id FROM email_verification_tokens WHERE token = $1 AND expires_at > $2",
		token, time.Now(),
	).Scan(&userID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired verification token"})
		return
	}

	// Update the user's email verification status
	_, err = database.DB.Exec(
		"UPDATE users SET email_verified = true, updated_at = $1 WHERE id = $2",
		time.Now(), userID,
	)

	if err != nil {
		log.Printf("Error updating email verification status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email"})
		return
	}

	// Delete the used verification token
	_, err = database.DB.Exec("DELETE FROM email_verification_tokens WHERE token = $1", token)
	if err != nil {
		log.Printf("Error deleting verification token: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email has been verified successfully"})
}

// SetupMFA sets up multi-factor authentication for a user
func SetupMFA(c *gin.Context) {
	// Get user models.ID from the context (set by the auth middleware)
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		Method string `json:"method" binding:"required,oneof=totp sms email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate a secret key for TOTP or a verification code for SMS/email
	var secret string
	var qrCodeURL string

	switch req.Method {
	case "totp":
		// In a real implementation, use a proper TOTP library
		secret = fmt.Sprintf("TOTP_SECRET_%s", uuid.New().String())
		qrCodeURL = fmt.Sprintf("otpauth://totp/SmartParkingSystem:%s?secret=%s&issuer=SmartParkingSystem", userID, secret)
	case "sms", "email":
		// Generate a 6-digit verification code
		secret = fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}

	// Store the MFA method and secret in the database
	_, err := database.DB.Exec(
		"INSERT INTO user_mfa (user_id, method, secret, enabled, created_at) VALUES ($1, $2, $3, false, $4) ON CONFLICT (user_id) DO UPDATE SET method = $2, secret = $3, enabled = false, created_at = $4",
		userID, req.Method, secret, time.Now(),
	)

	if err != nil {
		log.Printf("Error setting up MFA: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set up MFA"})
		return
	}

	// In a real application, send the verification code via SMS or email if applicable
	if req.Method == "sms" || req.Method == "email" {
		log.Printf("MFA verification code for user models.%s: %s", userID, secret)
	}

	response := gin.H{
		"message": fmt.Sprintf("MFA setup initiated with method: %s", req.Method),
		"method":  req.Method,
	}

	if req.Method == "totp" {
		response["secret"] = secret
		response["qrCodeUrl"] = qrCodeURL
	}

	c.JSON(http.StatusOK, response)
}

// VerifyMFA verifies a multi-factor authentication code
func VerifyMFA(c *gin.Context) {
	var req struct {
		Code      string `json:"code" binding:"required"`
		SessionID string `json:"sessionId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If sessionId is provided, this is part of the login flow
	// Otherwise, it's enabling MFA for an authenticated user
	var userID string
	var method string
	var secret string
	var enabled bool

	if req.SessionID != "" {
		// Verify MFA during login
		err := database.DB.QueryRow(
			"SELECT u.id, m.method, m.secret, m.enabled FROM mfa_sessions s JOIN users u ON s.user_id = u.id JOIN user_mfa m ON u.id = m.user_id WHERE s.session_id = $1 AND s.expires_at > $2",
			req.SessionID, time.Now(),
		).Scan(&userID, &method, &secret, &enabled)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired MFA session"})
			return
		}

		if !enabled {
			c.JSON(http.StatusBadRequest, gin.H{"error": "MFA not enabled for this user"})
			return
		}
	} else {
		// Get user models.ID from the context (set by the auth middleware)
		userIDFromContext, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		userID = userIDFromContext.(string)

		// Get the user's MFA method and secret
		err := database.DB.QueryRow(
			"SELECT method, secret, enabled FROM user_mfa WHERE user_id = $1",
			userID,
		).Scan(&method, &secret, &enabled)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "MFA not set up for this user"})
			return
		}
	}

	// Verify the MFA code
	var codeValid bool

	switch method {
	case "totp":
		// In a real implementation, use a proper TOTP library to validate the code
		// For demo purposes, we'll just check if the code is "123456"
		codeValid = req.Code == "123456"
	case "sms", "email":
		// Check if the code matches the stored secret
		codeValid = req.Code == secret
	}

	if !codeValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification code"})
		return
	}

	// If this is part of enabling MFA, mark it as enabled
	if !enabled && req.SessionID == "" {
		_, err := database.DB.Exec(
			"UPDATE user_mfa SET enabled = true, updated_at = $1 WHERE user_id = $2",
			time.Now(), userID,
		)

		if err != nil {
			log.Printf("Error enabling MFA: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enable MFA"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "MFA has been enabled successfully"})
		return
	}

	// If this is part of the login flow, generate a token and return user models.info
	if req.SessionID != "" {
		// Get user models.information
		var user models.User
		err := database.DB.QueryRow(
			"SELECT id, email, first_name, last_name, phone, role FROM users WHERE id = $1",
			userID,
		).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Phone, &user.Role)

		if err != nil {
			log.Printf("Error fetching user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user models.information"})
			return
		}

		// Generate JWT token
		token, err := middleware.GenerateToken(user.ID, user.Email, user.Role)
		if err != nil {
			log.Printf("Error generating token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
			return
		}

		// Delete the MFA session
		_, err = database.DB.Exec("DELETE FROM mfa_sessions WHERE session_id = $1", req.SessionID)
		if err != nil {
			log.Printf("Error deleting MFA session: %v", err)
		}

		// Return the user models.and token
		c.JSON(http.StatusOK, gin.H{
			"message": "MFA verification successful",
			"user": gin.H{
				"id":        user.ID,
				"email":     user.Email,
				"firstName": user.FirstName,
				"lastName":  user.LastName,
				"phone":     user.Phone,
				"role":      user.Role,
			},
			"token": token,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Verification successful"})
}
