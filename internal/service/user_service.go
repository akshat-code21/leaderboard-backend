package service

import (
	"context"
	"errors"
	"matiks/leaderboard/internal/models"
	"matiks/leaderboard/internal/repository"
)

type UserService struct {
	UserRepository repository.UserRepository
	redisRepo      repository.RedisRepository
}

type userService interface {
	SearchUsers(query string, limit int) (*models.UserSearchResponse, error)
	GetUserRank(username string) (*models.UserRankResponse, error)
}

func NewUserService(userRepository *repository.UserRepository, redisRepo *repository.RedisRepository) userService {
	return &UserService{UserRepository: *userRepository, redisRepo: *redisRepo}

}

func (s *UserService) SearchUsers(query string, limit int) (*models.UserSearchResponse, error) {
	if query == "" {
		return nil, errors.New("query is required")
	}
	if limit < 1 || limit > 100 {
		return nil, errors.New("limit must be between 1 and 100")
	}

	users, err := s.UserRepository.SearchUsers(query, limit)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	entries := make([]models.LeaderboardEntry, 0)

	for _, user := range users {
		// Calculate rank using Redis (much faster than DB COUNT queries)
		rank, err := s.calculateUserRankFromRedis(ctx, user.Rating)
		if err != nil {
			// Fallback to DB if Redis fails
			rank, err = s.calculateUserRank(user.Rating)
			if err != nil {
				return nil, err
			}
		}

		entries = append(entries, models.LeaderboardEntry{
			Rank:     rank,
			Username: user.Username,
			Rating:   user.Rating,
		})
	}

	return &models.UserSearchResponse{Users: entries, Count: len(entries)}, nil
}

func (s *UserService) GetUserRank(username string) (*models.UserRankResponse, error) {
	ctx := context.Background()

	user, err := s.UserRepository.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	// Try Redis first for rank calculation
	// Count users with higher rating (tie-aware ranking)
	rank, err := s.calculateUserRankFromRedis(ctx, user.Rating)
	if err == nil {
		return &models.UserRankResponse{
			Username: user.Username,
			Rating:   user.Rating,
			Rank:     rank,
		}, nil
	}

	return s.getUserRankFromDB(username)
}

func (s *UserService) calculateUserRankFromRedis(ctx context.Context, rating int) (int, error) {
	// Count users with rating > user's rating
	// This gives us tie-aware ranking
	count, err := s.redisRepo.CountUsersWithHigherRating(ctx, rating)
	if err != nil {
		return 0, err
	}
	return int(count) + 1, nil
}

func (s *UserService) getUserRankFromDB(username string) (*models.UserRankResponse, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}
	user, err := s.UserRepository.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	rank, err := s.calculateUserRank(user.Rating)
	if err != nil {
		return nil, err
	}
	return &models.UserRankResponse{Username: user.Username, Rating: user.Rating, Rank: rank}, nil
}

func (s *UserService) calculateUserRank(rating int) (int, error) {
	count, err := s.UserRepository.CountUsersWithHigherRating(rating)
	if err != nil {
		return 0, err
	}
	return int(count) + 1, nil
}
