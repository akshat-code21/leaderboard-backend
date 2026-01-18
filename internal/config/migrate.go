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

	log.Println("Creating index on username for search optimization...")
	err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_users_username
		ON users(username)
	`).Error
	if err != nil {
		log.Printf("Warning: Failed to create username index (may already exist): %v", err)
	}

	log.Println("Creating index on rating for sorting...")
	err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_users_rating
		ON users(rating DESC)
	`).Error
	if err != nil {
		log.Printf("Warning: Failed to create rating index (may already exist): %v", err)
	}

	log.Println("✅ Database migrations completed successfully")
	return nil
}
