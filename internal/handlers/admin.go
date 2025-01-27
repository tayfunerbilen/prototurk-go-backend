package handlers

import (
	"net/http"
	"strconv"
	"time"

	"prototurk/internal/models"
	"prototurk/pkg/response"
	"prototurk/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminHandler struct {
	db *gorm.DB
}

func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

// Create yeni bir admin oluşturur (Sadece super admin yapabilir)
func (h *AdminHandler) Create(c *gin.Context) {
	admin := c.MustGet("admin").(models.Admin)
	if admin.Role != models.AdminRoleSuperAdmin {
		c.JSON(http.StatusForbidden, response.Error("FORBIDDEN", "Super admin permission required", nil))
		return
	}

	var req models.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid request", err.Error()))
		return
	}

	// Role ve status validasyonu
	if !req.Role.ValidateRole() {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid admin role", nil))
		return
	}

	// İlk super admin'in rolünü kontrol et
	var firstSuperAdmin models.Admin
	if err := h.db.Where("role = ?", models.AdminRoleSuperAdmin).Order("created_at ASC").First(&firstSuperAdmin).Error; err == nil {
		// Eğer ilk super admin varsa ve yeni admin super admin olarak oluşturulmaya çalışılıyorsa
		if req.Role == models.AdminRoleSuperAdmin {
			c.JSON(http.StatusForbidden, response.Error("FORBIDDEN", "Cannot create another super admin while first super admin exists", nil))
			return
		}
	}

	if !req.Status.ValidateStatus() {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid admin status", nil))
		return
	}

	// Email kontrolü - sadece aktif (silinmemiş) admin'lerde kontrol et
	var existingAdmin models.Admin
	if err := h.db.Unscoped().Where("email = ? AND deleted_at IS NULL", req.Email).First(&existingAdmin).Error; err == nil {
		c.JSON(http.StatusConflict, response.Error("EMAIL_EXISTS", "Email already exists", nil))
		return
	}

	// Parolayı hashle
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("SERVER_ERROR", "Error processing request", nil))
		return
	}

	newAdmin := models.Admin{
		Email:    req.Email,
		Name:     req.Name,
		Password: string(hashedPassword),
		Role:     req.Role,
		Status:   req.Status,
	}

	if err := h.db.Create(&newAdmin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("SERVER_ERROR", "Error creating admin", nil))
		return
	}

	c.JSON(http.StatusCreated, response.Success(newAdmin))
}

// List tüm adminleri listeler
func (h *AdminHandler) List(c *gin.Context) {
	var admins []models.Admin
	if err := h.db.Find(&admins).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("SERVER_ERROR", "Error listing admins", nil))
		return
	}

	c.JSON(http.StatusOK, response.Success(admins))
}

// Get tek bir admin'i getirir
func (h *AdminHandler) Get(c *gin.Context) {
	id := c.Param("id")

	var admin models.Admin
	if err := h.db.First(&admin, id).Error; err != nil {
		c.JSON(http.StatusNotFound, response.Error("NOT_FOUND", "Admin not found", nil))
		return
	}

	c.JSON(http.StatusOK, response.Success(admin))
}

// Update admin bilgilerini günceller
func (h *AdminHandler) Update(c *gin.Context) {
	currentAdmin := c.MustGet("admin").(models.Admin)
	id := c.Param("id")

	var admin models.Admin
	if err := h.db.First(&admin, id).Error; err != nil {
		c.JSON(http.StatusNotFound, response.Error("NOT_FOUND", "Admin not found", nil))
		return
	}

	// İlk super admin'i güncellemeye çalışıyorsa ve kendisi değilse engelle
	if admin.IsFirstSuperAdmin(h.db) && currentAdmin.ID != admin.ID {
		c.JSON(http.StatusForbidden, response.Error("FORBIDDEN", "Cannot update first super admin", nil))
		return
	}

	// Sadece super admin başka bir admin'i güncelleyebilir
	if currentAdmin.ID != admin.ID && currentAdmin.Role != models.AdminRoleSuperAdmin {
		c.JSON(http.StatusForbidden, response.Error("FORBIDDEN", "Super admin permission required", nil))
		return
	}

	var req models.UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid request", err.Error()))
		return
	}

	updates := make(map[string]interface{})

	if req.Email != "" && req.Email != admin.Email {
		var existingAdmin models.Admin
		if err := h.db.Unscoped().Where("email = ? AND deleted_at IS NULL AND id != ?", req.Email, id).First(&existingAdmin).Error; err == nil {
			c.JSON(http.StatusConflict, response.Error("EMAIL_EXISTS", "Email already exists", nil))
			return
		}
		updates["email"] = req.Email
	}

	if req.Name != "" {
		updates["name"] = req.Name
	}

	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.Error("SERVER_ERROR", "Error processing request", nil))
			return
		}
		updates["password"] = string(hashedPassword)
	}

	// Role ve status güncellemelerini kontrol et
	if req.Role != "" {
		// İlk super admin rolünü değiştirmeye çalışıyorsa engelle
		if admin.IsFirstSuperAdmin(h.db) {
			c.JSON(http.StatusForbidden, response.Error("FORBIDDEN", "First super admin role cannot be changed", nil))
			return
		}

		// Kendisi super admin değilse veya kendi rolünü değiştirmeye çalışıyorsa engelle
		if !currentAdmin.CanUpdateRole() || (currentAdmin.ID == admin.ID && currentAdmin.Role != models.AdminRoleSuperAdmin) {
			c.JSON(http.StatusForbidden, response.Error("FORBIDDEN", "Cannot update role", nil))
			return
		}
		if !req.Role.ValidateRole() {
			c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid admin role", nil))
			return
		}
		updates["role"] = req.Role
	}

	if req.Status != "" {
		// İlk super admin statusünü değiştirmeye çalışıyorsa engelle
		if admin.IsFirstSuperAdmin(h.db) {
			c.JSON(http.StatusForbidden, response.Error("FORBIDDEN", "First super admin status cannot be changed", nil))
			return
		}

		// Kendisi super admin değilse veya kendi statusünü değiştirmeye çalışıyorsa engelle
		if !currentAdmin.CanUpdateStatus() || (currentAdmin.ID == admin.ID && currentAdmin.Role != models.AdminRoleSuperAdmin) {
			c.JSON(http.StatusForbidden, response.Error("FORBIDDEN", "Cannot update status", nil))
			return
		}
		if !req.Status.ValidateStatus() {
			c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid admin status", nil))
			return
		}
		updates["status"] = req.Status
	}

	if err := h.db.Model(&admin).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("SERVER_ERROR", "Error updating admin", nil))
		return
	}

	c.JSON(http.StatusOK, response.Success(admin))
}

// Delete bir admin'i siler (Sadece super admin yapabilir)
func (h *AdminHandler) Delete(c *gin.Context) {
	admin := c.MustGet("admin").(models.Admin)
	if !admin.CanDeleteAdmin() {
		c.JSON(http.StatusForbidden, response.Error("FORBIDDEN", "Super admin permission required", nil))
		return
	}

	id := c.Param("id")
	idUint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid admin ID", nil))
		return
	}

	// Admin'in kendisini silmesini engelle
	if admin.ID == uint(idUint) {
		c.JSON(http.StatusForbidden, response.Error("FORBIDDEN", "You cannot delete yourself", nil))
		return
	}

	var targetAdmin models.Admin
	if err := h.db.First(&targetAdmin, idUint).Error; err != nil {
		c.JSON(http.StatusNotFound, response.Error("NOT_FOUND", "Admin not found", nil))
		return
	}

	// İlk super admin'i silmeye çalışıyorsa engelle
	if targetAdmin.IsFirstSuperAdmin(h.db) {
		c.JSON(http.StatusForbidden, response.Error("FORBIDDEN", "Cannot delete first super admin", nil))
		return
	}

	if err := h.db.Delete(&targetAdmin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("SERVER_ERROR", "Error deleting admin", nil))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "Admin deleted successfully"}))
}

// Login admin girişi yapar
func (h *AdminHandler) Login(c *gin.Context) {
	var req models.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid request", err.Error()))
		return
	}

	var admin models.Admin
	if err := h.db.Where("email = ?", req.Email).First(&admin).Error; err != nil {
		c.JSON(http.StatusUnauthorized, response.Error("INVALID_CREDENTIALS", "Invalid email or password", nil))
		return
	}

	if !admin.IsActive() {
		c.JSON(http.StatusForbidden, response.Error("ACCOUNT_INACTIVE", "Admin account is not active", nil))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, response.Error("INVALID_CREDENTIALS", "Invalid email or password", nil))
		return
	}

	// Son giriş tarihini güncelle
	h.db.Model(&admin).Update("last_login", utils.Now())

	// JWT token oluştur
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin_id": admin.ID,
		"role":     admin.Role,
		"exp":      time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
	})

	tokenString, err := token.SignedString([]byte(c.MustGet("jwt_secret").(string)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("SERVER_ERROR", "Error generating token", nil))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"token": tokenString,
		"admin": admin,
	}))
}

// Me giriş yapmış admin bilgilerini getirir
func (h *AdminHandler) Me(c *gin.Context) {
	admin := c.MustGet("admin").(models.Admin)
	c.JSON(http.StatusOK, response.Success(admin))
}
