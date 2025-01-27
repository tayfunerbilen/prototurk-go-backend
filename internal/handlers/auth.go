package handlers

import (
	"net/http"
	"time"

	"prototurk/internal/models"
	"prototurk/pkg/response"
	"prototurk/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid request body", err.Error()))
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("SERVER_ERROR", "Error processing request", nil))
		return
	}

	// Create user
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Status:   models.UserStatusActive,
	}

	result := h.db.Create(&user)
	if result.Error != nil {
		if h.db.Where("username = ?", req.Username).First(&models.User{}).Error == nil {
			c.JSON(http.StatusConflict, response.Error("USERNAME_EXISTS", "Username already exists", nil))
			return
		}
		if h.db.Where("email = ?", req.Email).First(&models.User{}).Error == nil {
			c.JSON(http.StatusConflict, response.Error("EMAIL_EXISTS", "Email already exists", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("SERVER_ERROR", "Error creating user", nil))
		return
	}

	c.JSON(http.StatusCreated, response.Success(user))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid request body", err.Error()))
		return
	}

	var user models.User
	// Try to find user by username or email
	if err := h.db.Where("username = ? OR email = ?", req.Identifier, req.Identifier).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, response.Error("INVALID_CREDENTIALS", "Invalid username/email or password", nil))
		return
	}

	if user.Status == models.UserStatusBanned {
		c.JSON(http.StatusForbidden, response.Error("USER_BANNED", "User is banned", nil))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, response.Error("INVALID_CREDENTIALS", "Invalid username/email or password", nil))
		return
	}

	// Update last login date
	h.db.Model(&user).Update("last_login_date", utils.Now())

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      utils.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
	})

	tokenString, err := token.SignedString([]byte(c.MustGet("jwt_secret").(string)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("SERVER_ERROR", "Error generating token", nil))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"token": tokenString,
		"user":  user,
	}))
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.Error("UNAUTHORIZED", "User not authenticated", nil))
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, response.Error("USER_NOT_FOUND", "User not found", nil))
		return
	}

	c.JSON(http.StatusOK, response.Success(user))
}

// UpdateProfile handles the profile update request
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.Error("UNAUTHORIZED", "User not authenticated", nil))
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid request body", err.Error()))
		return
	}

	if !req.Validate() {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "At least one field (username or email) must be provided", nil))
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, response.Error("USER_NOT_FOUND", "User not found", nil))
		return
	}

	// Check username uniqueness if provided
	if req.Username != "" && req.Username != user.Username {
		var existingUser models.User
		if err := h.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, response.Error("USERNAME_EXISTS", "Username already exists", nil))
			return
		}
	}

	// Check email uniqueness if provided
	if req.Email != "" && req.Email != user.Email {
		var existingUser models.User
		if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, response.Error("EMAIL_EXISTS", "Email already exists", nil))
			return
		}
	}

	// Update only provided fields
	updates := make(map[string]interface{})
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}

	if err := h.db.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("SERVER_ERROR", "Error updating profile", nil))
		return
	}

	c.JSON(http.StatusOK, response.Success(user))
}

// UpdatePassword handles the password update request
func (h *AuthHandler) UpdatePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.Error("UNAUTHORIZED", "User not authenticated", nil))
		return
	}

	var req models.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid request body", err.Error()))
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, response.Error("USER_NOT_FOUND", "User not found", nil))
		return
	}

	// Mevcut parolayı kontrol et
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, response.Error("INVALID_PASSWORD", "Current password is incorrect", nil))
		return
	}

	// Yeni parolayı hashle
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("SERVER_ERROR", "Error processing request", nil))
		return
	}

	// Parolayı güncelle
	if err := h.db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("SERVER_ERROR", "Error updating password", nil))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"message": "Password updated successfully",
	}))
}
