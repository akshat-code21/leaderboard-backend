package repository

import (
	"matiks/leaderboard/internal/models"

	"gorm.io/gorm"
)

// UserRepository handles all database operations for users
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetLeaderboard retrieves users ordered by rating DESC with pagination
// Returns users without rank calculation (rank is calculated in service layer)
func (r *UserRepository) GetLeaderboard(page, limit int) ([]models.User, error) {
	offset := (page - 1) * limit
	var users []models.User

	err := r.db.
		Order("rating DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	return users, err
}

// SearchUsers searches for users by username pattern (case-insensitive)
func (r *UserRepository) SearchUsers(query string, limit int) ([]models.User, error) {
	var users []models.User
	pattern := "%" + query + "%"

	err := r.db.
		Where("username ILIKE ?", pattern).
		Order("rating DESC").
		Limit(limit).
		Find(&users).Error

	return users, err
}

// GetUserByUsername retrieves a single user by username
func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetTotalUsers returns the total count of users in the database
func (r *UserRepository) GetTotalUsers() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Count(&count).Error
	return count, err
}

// CountUsersWithHigherRating counts users with rating greater than the given rating
// Used for rank calculation
func (r *UserRepository) CountUsersWithHigherRating(rating int) (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("rating > ?", rating).Count(&count).Error
	return count, err
}

// CountUsersWithRating counts users with a specific rating
// Used for tie-aware ranking
func (r *UserRepository) CountUsersWithRating(rating int) (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("rating = ?", rating).Count(&count).Error
	return count, err
}
