package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type AdminRole string
type AdminStatus string

const (
	AdminRoleSuperAdmin AdminRole = "super_admin"
	AdminRoleAdmin      AdminRole = "admin"
	AdminRoleEditor     AdminRole = "editor"
)

const (
	AdminStatusActive  AdminStatus = "active"
	AdminStatusPassive AdminStatus = "passive"
)

type Admin struct {
	gorm.Model
	Email     string      `gorm:"type:varchar(255);unique;not null" json:"email"`
	Name      string      `gorm:"type:varchar(100);not null" json:"name"`
	Password  string      `gorm:"type:varchar(255);not null" json:"-"`
	Role      AdminRole   `gorm:"type:admin_role;not null" json:"role"`
	Status    AdminStatus `gorm:"type:admin_status;not null;default:'active'" json:"status"`
	LastLogin time.Time   `gorm:"type:timestamp with time zone" json:"last_login"`
}

// BeforeCreate ensures all timestamps are in UTC
func (a *Admin) BeforeCreate(tx *gorm.DB) error {
	a.CreatedAt = a.CreatedAt.UTC()
	a.UpdatedAt = a.UpdatedAt.UTC()
	if !a.LastLogin.IsZero() {
		a.LastLogin = a.LastLogin.UTC()
	}
	return nil
}

// BeforeUpdate ensures all timestamps are in UTC
func (a *Admin) BeforeUpdate(tx *gorm.DB) error {
	a.UpdatedAt = a.UpdatedAt.UTC()
	if !a.LastLogin.IsZero() {
		a.LastLogin = a.LastLogin.UTC()
	}
	return nil
}

type CreateAdminRequest struct {
	Email    string      `json:"email" binding:"required,email"`
	Name     string      `json:"name" binding:"required,min=2,max=100"`
	Password string      `json:"password" binding:"required,min=6"`
	Role     AdminRole   `json:"role" binding:"required"`
	Status   AdminStatus `json:"status" binding:"required"`
}

type UpdateAdminRequest struct {
	Email    string      `json:"email" binding:"omitempty,email"`
	Name     string      `json:"name" binding:"omitempty,min=2,max=100"`
	Password string      `json:"password" binding:"omitempty,min=6"`
	Role     AdminRole   `json:"role" binding:"omitempty"`
	Status   AdminStatus `json:"status" binding:"omitempty"`
}

type AdminLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// ValidateRole kontrol eder admin rolünün geçerli olup olmadığını
func (r AdminRole) ValidateRole() bool {
	switch r {
	case AdminRoleSuperAdmin, AdminRoleAdmin, AdminRoleEditor:
		return true
	}
	return false
}

// ValidateStatus kontrol eder admin statusünün geçerli olup olmadığını
func (s AdminStatus) ValidateStatus() bool {
	switch s {
	case AdminStatusActive, AdminStatusPassive:
		return true
	}
	return false
}

// CanDeleteAdmin kontrol eder admin'in silme yetkisi olup olmadığını
func (a *Admin) CanDeleteAdmin() bool {
	return a.Role == AdminRoleSuperAdmin
}

// IsActive kontrol eder admin'in aktif olup olmadığını
func (a *Admin) IsActive() bool {
	return a.Status == AdminStatusActive
}

// IsFirstSuperAdmin kontrol eder admin'in ilk oluşturulan super admin olup olmadığını
func (a *Admin) IsFirstSuperAdmin(db *gorm.DB) bool {
	var firstAdmin Admin
	if err := db.Where("role = ?", AdminRoleSuperAdmin).Order("created_at ASC").First(&firstAdmin).Error; err != nil {
		return false
	}
	return a.ID == firstAdmin.ID
}

// CanUpdateRole kontrol eder admin'in rol güncelleyip güncelleyemeyeceğini
func (a *Admin) CanUpdateRole() bool {
	return a.Role == AdminRoleSuperAdmin
}

// CanUpdateStatus kontrol eder admin'in status güncelleyip güncelleyemeyeceğini
func (a *Admin) CanUpdateStatus() bool {
	return a.Role == AdminRoleSuperAdmin
}

var (
	ErrInvalidRole     = errors.New("invalid admin role")
	ErrInvalidStatus   = errors.New("invalid admin status")
	ErrNotSuperAdmin   = errors.New("super admin permission required")
	ErrAdminNotActive  = errors.New("admin account is not active")
	ErrEmailExists     = errors.New("email already exists")
	ErrInvalidPassword = errors.New("invalid password")
	ErrAdminNotFound   = errors.New("admin not found")
)
