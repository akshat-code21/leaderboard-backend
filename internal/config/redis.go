package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client

func ConnectRedis() (*redis.Client, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}
	// Get redis URL from environment
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		return nil, fmt.Errorf("REDIS_URL environment variable is not set")
	}
	opt, err := redis.ParseURL(redisUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	Redis = redis.NewClient(opt)
	return Redis, nil
}

func GetRedis() *redis.Client {
	return Redis
}
