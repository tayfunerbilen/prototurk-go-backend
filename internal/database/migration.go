package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunMigrations tüm migration dosyalarını sırayla çalıştırır
func RunMigrations(config *Config) error {
	// Database bağlantısını bir kere oluştur
	db, err := NewConnection(config)
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}

	// Migration klasörünü oku
	files, err := os.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("error reading migrations directory: %v", err)
	}

	// SQL dosyalarını filtrele ve sırala
	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}
	sort.Strings(sqlFiles)

	// Her migration dosyasını çalıştır
	for _, file := range sqlFiles {
		// Migration daha önce çalıştırılmış mı kontrol et
		var count int64
		if err := db.Table("migrations").Where("name = ?", file).Count(&count).Error; err == nil && count > 0 {
			log.Printf("Skipping migration %s: already executed", file)
			continue
		}

		log.Printf("Running migration: %s", file)

		// Migration dosyasını oku
		content, err := os.ReadFile(filepath.Join("migrations", file))
		if err != nil {
			return fmt.Errorf("error reading migration file %s: %v", file, err)
		}

		// Migration'ı transaction içinde çalıştır
		tx := db.Begin()

		// Migration'ı çalıştır
		if err := tx.Exec(string(content)).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("error executing migration %s: %v", file, err)
		}

		// Migration kaydını ekle
		if err := tx.Table("migrations").Create(map[string]interface{}{
			"name": file,
		}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("error recording migration %s: %v", file, err)
		}

		tx.Commit()
		log.Printf("Migration completed: %s", file)
	}

	return nil
}
