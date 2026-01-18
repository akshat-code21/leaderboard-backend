package repository

import (
	"context"
	"matiks/leaderboard/internal/models"

	"github.com/redis/go-redis/v9"
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

func (r *UserRepository) SyncUserToRedis(ctx context.Context, redisRepo *RedisRepository, user models.User) error {
	return redisRepo.AddToLeaderboard(ctx, user.Username, user.Rating)
}

func (r *UserRepository) SyncAllUserToRedis(ctx context.Context, redisRepo *RedisRepository) error {
	// Clear existing Redis data first
	if err := redisRepo.client.Del(ctx, "leaderboard:ratings").Err(); err != nil {
		return err
	}

	var users []models.User
	if err := r.db.Order("rating DESC").Find(&users).Error; err != nil {
		return err
	}

	if len(users) == 0 {
		return nil
	}

	// Batch add to Redis (in chunks for large datasets)
	batchSize := 1000
	for i := 0; i < len(users); i += batchSize {
		end := i + batchSize
		if end > len(users) {
			end = len(users)
		}

		zMembers := make([]redis.Z, 0, end-i)
		for _, user := range users[i:end] {
			zMembers = append(zMembers, redis.Z{
				Score:  float64(user.Rating),
				Member: user.Username,
			})
		}

		if err := redisRepo.client.ZAdd(ctx, "leaderboard:ratings", zMembers...).Err(); err != nil {
			return err
		}
	}

	return nil
}
