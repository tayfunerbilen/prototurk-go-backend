package models

import (
	"time"

	"gorm.io/gorm"
)

type UserStatus string

const (
	UserStatusActive  UserStatus = "active"
	UserStatusPassive UserStatus = "passive"
	UserStatusBanned  UserStatus = "banned"
)

type User struct {
	gorm.Model
	Username      string     `gorm:"type:varchar(32);unique;not null" json:"username"`
	Email         string     `gorm:"type:varchar(255);unique;not null" json:"email"`
	Password      string     `gorm:"type:varchar(255);not null" json:"-"`
	Status        UserStatus `gorm:"type:user_status;default:'active'" json:"status"`
	LastLoginDate time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"last_login_date"`
}

// BeforeCreate ensures all timestamps are in UTC
func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.CreatedAt = u.CreatedAt.UTC()
	u.UpdatedAt = u.UpdatedAt.UTC()
	u.LastLoginDate = u.LastLoginDate.UTC()
	return nil
}

// BeforeUpdate ensures all timestamps are in UTC
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = u.UpdatedAt.UTC()
	if !u.LastLoginDate.IsZero() {
		u.LastLoginDate = u.LastLoginDate.UTC()
	}
	return nil
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

// UpdateProfileRequest represents the request body for profile updates
type UpdateProfileRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=32"`
	Email    string `json:"email" binding:"omitempty,email"`
}

// UpdatePasswordRequest represents the request body for password updates
type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=6"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// Validate checks if at least one field is provided for update
func (r *UpdateProfileRequest) Validate() bool {
	return r.Username != "" || r.Email != ""
}
