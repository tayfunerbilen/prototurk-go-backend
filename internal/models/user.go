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
	LastLoginDate time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"last_login_date"`
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
