package config

import (
	"log"

	models "github.com/darmawguna/tirtaapp.git/model"
	"gorm.io/gorm"
)

// RunMigration menjalankan auto-migrasi GORM.
func RunMigration(db *gorm.DB) {
	err := db.AutoMigrate(&models.User{}, &models.Quiz{}, &models.Education{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migration successful.")
}