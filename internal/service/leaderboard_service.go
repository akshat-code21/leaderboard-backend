package service

import (
	"matiks/leaderboard/internal/models"
	"matiks/leaderboard/internal/repository"
)

type LeaderboardService struct {
	userRepo *repository.UserRepository
}

type leaderboardService interface {
	GetLeaderboard(page, limit int) (*models.LeaderboardResponse, error)
	calculateRanks(users []models.User) []models.LeaderboardEntry
}

// GetLeaderboard implements [leaderboardService].
func (l *LeaderboardService) GetLeaderboard(page int, limit int) (*models.LeaderboardResponse, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 100 {
		limit = 50
	}

	users, err := l.userRepo.GetLeaderboard(page, limit)

	if err != nil {
		return nil, err
	}

	entries := l.calculateRanks(users)

	total, err := l.userRepo.GetTotalUsers()
	if err != nil {
		return nil, err
	}

	return &models.LeaderboardResponse{
		Entries: entries,
		Page:    page,
		Limit:   limit,
		Total:   int(total),
	}, nil
}

// calculateRanks implements [leaderboardService].
func (l *LeaderboardService) calculateRanks(users []models.User) []models.LeaderboardEntry {
	entries := make([]models.LeaderboardEntry, 0)
	currentRank := 1

	for i, user := range users {
		if i > 0 && users[i-1].Rating != user.Rating {
			currentRank = i + 1
		}
		entries = append(entries, models.LeaderboardEntry{
			Rank:     currentRank,
			Username: user.Username,
			Rating:   user.Rating,
		})
	}
	return entries
}

func NewLeaderboardService(userRepo *repository.UserRepository) leaderboardService {
	return &LeaderboardService{userRepo: userRepo}
}
