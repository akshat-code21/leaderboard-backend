package service

import (
	"errors"
	"matiks/leaderboard/internal/models"
	"matiks/leaderboard/internal/repository"
)

type UserService struct {
	UserRepository repository.UserRepository
}

type userService interface {
	SearchUsers(query string, limit int) (*models.UserSearchResponse, error)
	GetUserRank(username string) (*models.UserRankResponse, error)
}

func NewUserService(userRepository *repository.UserRepository) userService {
	return &UserService{UserRepository: *userRepository}
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
	entries := make([]models.LeaderboardEntry, 0)
	for _, user := range users {
		// Calculate rank for each user
		rank, err := s.calculateUserRank(user.Rating)
		if err != nil {
			return nil, err
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
