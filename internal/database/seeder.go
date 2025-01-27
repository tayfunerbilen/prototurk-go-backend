package database

import (
	"log"

	"prototurk/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedDefaultAdmin varsayılan super admin'i ekler
func SeedDefaultAdmin(db *gorm.DB) error {
	var count int64
	db.Model(&models.Admin{}).Count(&count)
	if count > 0 {
		return nil // Eğer admin varsa ekleme
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := models.Admin{
		Email:    "tayfunerbilen@gmail.com",
		Name:     "Tayfun Erbilen",
		Password: string(hashedPassword),
		Role:     models.AdminRoleSuperAdmin,
		Status:   models.AdminStatusActive,
		// LastLogin alanını boş bırak, ilk girişte güncellenecek
	}

	if err := db.Create(&admin).Error; err != nil {
		return err
	}

	log.Println("Default super admin created successfully!")
	return nil
}
