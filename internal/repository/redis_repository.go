package repository

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{client: client}
}

func (r *RedisRepository) AddToLeaderboard(ctx context.Context, username string, rating int) error {
	return r.client.ZAdd(ctx, "leaderboard:ratings", redis.Z{
		Score:  float64(rating),
		Member: username,
	}).Err()
}

func (r *RedisRepository) GetLeaderboard(ctx context.Context, offset, limit int64) ([]redis.Z, error) {
	// ZRevRangeWithScores returns in DESCENDING order (highest score first)
	// This is what we want for leaderboard (rank 1 = highest rating)
	return r.client.ZRevRangeWithScores(ctx, "leaderboard:ratings", offset, offset+limit-1).Result()
}

func (r *RedisRepository) GetUserRank(ctx context.Context, username string) (int64, error) {
	return r.client.ZRevRank(ctx, "leaderboard:ratings", username).Result()
}

// CountUsersWithHigherRating counts users with rating greater than the given rating
// Used for tie-aware rank calculation
func (r *RedisRepository) CountUsersWithHigherRating(ctx context.Context, rating int) (int64, error) {
	// Use ZCOUNT to count members with score > rating
	// In Redis sorted sets, score is the rating
	// Format: "(rating" means > rating (exclusive), "+inf" means positive infinity
	ratingStr := strconv.FormatFloat(float64(rating), 'f', -1, 64)
	return r.client.ZCount(ctx, "leaderboard:ratings",
		"("+ratingStr, "+inf").Result()
}

func (r *RedisRepository) GetTotalUsers(ctx context.Context) (int64, error) {
	return r.client.ZCard(ctx, "leaderboard:ratings").Result()
}
