package utils

import (
	"log"

	models "github.com/darmawguna/tirtaapp.git/model" // Adjust path if needed
	"gorm.io/gorm"
)

// ClearAllData deletes all records from the application's tables.
// USE WITH CAUTION!
func ClearAllData(db *gorm.DB) error {
	log.Println("WARNING: Attempting to delete all data from tables...")

	// Delete in an order that respects foreign key constraints (generally, delete dependent data first)
	modelsToDelete := []interface{}{
		&models.HemodialysisMonitoring{}, // Depends on HemodialysisSchedule
		&models.FluidBalanceLog{},      // Depends on User
		&models.Device{},               // Depends on User
		&models.ComplaintLog{},         // Depends on User
		&models.DrugSchedule{},         // Depends on User
		&models.ControlSchedule{},      // Depends on User
		&models.HemodialysisSchedule{}, // Depends on User
		&models.Quiz{},                 // Depends on User (CreatedBy)
		&models.Education{},            // Depends on User (CreatedBy)
		&models.User{},                 // Base table
	}

	// Iterate and delete data from each table
	// Using Unscoped() ensures soft deletes are also permanently removed if enabled
	for _, model := range modelsToDelete {
		log.Printf("Deleting data from table for model %T...", model)
		if err := db.Unscoped().Where("1 = 1").Delete(model).Error; err != nil {
			log.Printf("ERROR: Failed to delete data for model %T: %v", model, err)
			return err // Stop on first error
		}
	}

	log.Println("All specified table data deleted successfully.")
	return nil
}