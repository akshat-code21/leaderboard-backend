package service

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"matiks/leaderboard/internal/repository"
	"sync"
	"time"
)

type UpdateService struct {
	userRepo   *repository.UserRepository
	redisRepo  *repository.RedisRepository
	updateChan chan UpdateRequest
	workers    int
	wg         sync.WaitGroup
}

type UpdateRequest struct {
	Username  string
	NewRating int
}

func NewUpdateService(userRepo *repository.UserRepository, redisRepo *repository.RedisRepository) *UpdateService {
	service := &UpdateService{
		userRepo:   userRepo,
		redisRepo:  redisRepo,
		updateChan: make(chan UpdateRequest, 100), // Buffer for 100 updates
		workers:    5,                             // Number of concurrent workers
	}

	// Start worker goroutines
	for i := 0; i < service.workers; i++ {
		service.wg.Add(1)
		go service.worker(i)
	}

	return service
}

// worker processes updates from the channel
func (s *UpdateService) worker(id int) {
	defer s.wg.Done()
	ctx := context.Background()

	for update := range s.updateChan {
		log.Printf("Worker %d: Updating %s to rating %d", id, update.Username, update.NewRating)

		// Update database
		if err := s.userRepo.UpdateUserRating(ctx, update.Username, update.NewRating); err != nil {
			log.Printf("Worker %d: Failed to update DB for %s: %v", id, update.Username, err)
			continue
		}

		// Update Redis
		if s.redisRepo != nil {
			if err := s.redisRepo.AddToLeaderboard(ctx, update.Username, update.NewRating); err != nil {
				log.Printf("Worker %d: Failed to update Redis for %s: %v", id, update.Username, err)
			}
		}

		log.Printf("Worker %d: Successfully updated %s", id, update.Username)
	}
}

// QueueUpdate adds an update to the processing queue (non-blocking)
func (s *UpdateService) QueueUpdate(username string, newRating int) error {
	select {
	case s.updateChan <- UpdateRequest{Username: username, NewRating: newRating}:
		return nil
	default:
		return fmt.Errorf("update queue is full")
	}
}

// SimulateRandomUpdates updates random users with new ratings
func (s *UpdateService) SimulateRandomUpdates(count int) error {
	ctx := context.Background()

	// Get random users from database
	users, err := s.userRepo.GetRandomUsers(ctx, count)
	if err != nil {
		return err
	}

	rand.Seed(time.Now().UnixNano())

	for _, user := range users {
		// Generate random rating between 100 and 5000
		newRating := rand.Intn(4900) + 100

		// Queue update (non-blocking)
		if err := s.QueueUpdate(user.Username, newRating); err != nil {
			log.Printf("Failed to queue update for %s: %v", user.Username, err)
		}
	}

	log.Printf("Queued %d random updates", len(users))
	return nil
}

// Shutdown gracefully stops all workers
func (s *UpdateService) Shutdown() {
	close(s.updateChan)
	s.wg.Wait()
	log.Println("Update service shutdown complete")
}
