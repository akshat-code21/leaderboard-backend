package service

import (
	"context"
	"log"
	"matiks/leaderboard/internal/models"
	"matiks/leaderboard/internal/repository"

	"github.com/redis/go-redis/v9"
)

type LeaderboardService struct {
	userRepo  *repository.UserRepository
	redisRepo *repository.RedisRepository
}

type leaderboardService interface {
	GetLeaderboard(page, limit int) (*models.LeaderboardResponse, error)
	calculateRanks(users []models.User) []models.LeaderboardEntry
}

// GetLeaderboard implements [leaderboardService].
func (s *LeaderboardService) GetLeaderboard(page, limit int) (*models.LeaderboardResponse, error) {
	if s.redisRepo == nil {
		return s.getLeaderboardFromDB(page, limit)
	}

	ctx := context.Background()

	offset := int64((page - 1) * limit)
	limit64 := int64(limit)

	totalRedis, err := s.redisRepo.GetTotalUsers(ctx)
	if err != nil {
		log.Printf("Redis check failed: %v, falling back to DB", err)
		return s.getLeaderboardFromDB(page, limit)
	}

	if totalRedis == 0 {
		log.Println("Redis is empty, falling back to DB")
		return s.getLeaderboardFromDB(page, limit)
	}

	redisEntries, err := s.redisRepo.GetLeaderboard(ctx, offset, limit64)
	if err != nil {
		log.Printf("Redis GetLeaderboard failed: %v, falling back to DB", err)
		return s.getLeaderboardFromDB(page, limit)
	}

	if len(redisEntries) == 0 {
		log.Println("Redis returned empty results, falling back to DB")
		return s.getLeaderboardFromDB(page, limit)
	}

	entries := s.convertRedisEntriesToLeaderboardEntries(redisEntries, offset)
	log.Printf("✅ Redis leaderboard hit - page %d, limit %d, total %d", page, limit, totalRedis)

	return &models.LeaderboardResponse{
		Entries: entries,
		Page:    page,
		Limit:   limit,
		Total:   int(totalRedis),
	}, nil
}

func (l *LeaderboardService) getLeaderboardFromDB(page, limit int) (*models.LeaderboardResponse, error) {

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

func NewLeaderboardService(userRepo *repository.UserRepository, redisRepo *repository.RedisRepository) leaderboardService {
	return &LeaderboardService{
		userRepo:  userRepo,
		redisRepo: redisRepo,
	}
}

func (s *LeaderboardService) convertRedisEntriesToLeaderboardEntries(redisEntries []redis.Z, offset int64) []models.LeaderboardEntry {
	entries := make([]models.LeaderboardEntry, 0, len(redisEntries))

	// ZRevRangeWithScores returns in descending order (highest score first)
	// Calculate ranks with tie-aware logic
	currentRank := int(offset) + 1

	for i, entry := range redisEntries {
		// If previous entry had different rating, update rank
		if i > 0 && redisEntries[i-1].Score != entry.Score {
			currentRank = int(offset) + i + 1
		}

		entries = append(entries, models.LeaderboardEntry{
			Rank:     currentRank,
			Username: entry.Member.(string),
			Rating:   int(entry.Score),
		})
	}

	return entries
}
