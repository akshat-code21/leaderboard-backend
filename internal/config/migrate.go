package config

import (
	"log"

	"matiks/leaderboard/internal/models"

	"gorm.io/gorm"
)

// AutoMigrate runs database migrations
func AutoMigrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(&models.User{})
	if err != nil {
		return err
	}

	log.Println("✅ Database migrations completed successfully")
	return nil
}
