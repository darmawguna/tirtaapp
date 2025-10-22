package config

import (
	"log"

	"gorm.io/gorm"
)

// RunMigration menjalankan auto-migrasi GORM.
func RunMigration(db *gorm.DB, models ...interface{}) {
	err := db.AutoMigrate(models...)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migration successful.")
}