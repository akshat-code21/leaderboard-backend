package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"matiks/leaderboard/internal/config"
	"matiks/leaderboard/internal/controllers"
	"matiks/leaderboard/internal/handlers"
	"matiks/leaderboard/internal/repository"
	"matiks/leaderboard/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Connect to database
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 2. Run migrations
	if err := config.AutoMigrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Connect to Redis
	redisClient, err := config.ConnectRedis()
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v (continuing without Redis)", err)
		redisClient = nil
	} else {
		if err := redisClient.Ping(context.Background()).Err(); err != nil {
			log.Printf("Warning: Redis ping failed: %v (continuing without Redis)", err)
			redisClient = nil
		} else {
			log.Println("Successfully connected to Redis")
		}
	}

	// 3. Initialize layers (bottom to top)
	// Repository layer
	userRepo := repository.NewUserRepository(db)

	// Initialize Redis repository (can be nil if Redis unavailable)
	var redisRepo *repository.RedisRepository
	if redisClient != nil {
		redisRepo = repository.NewRedisRepository(redisClient)

		// Sync data to Redis on startup (run in background)
		go func() {
			ctx := context.Background()
			log.Println("Syncing database to Redis...")
			if err := userRepo.SyncAllUserToRedis(ctx, redisRepo); err != nil {
				log.Printf("Failed to sync to Redis: %v", err)
			} else {
				count, err := redisRepo.GetTotalUsers(ctx)
				if err != nil {
					log.Printf("Failed to get Redis count: %v", err)
				} else {
					log.Printf("Successfully synced %d users to Redis", count)
				}
			}
		}()
	} else {
		log.Println("Running without Redis - using database only")
	}

	// Service layer
	leaderboardServiceInterface := service.NewLeaderboardService(userRepo, redisRepo)
	userServiceInterface := service.NewUserService(userRepo, redisRepo)
	updateService := service.NewUpdateService(userRepo, redisRepo)

	// Type assertions to get concrete types for controllers
	leaderboardService, ok := leaderboardServiceInterface.(*service.LeaderboardService)
	if !ok {
		log.Fatal("Failed to assert LeaderboardService type")
	}

	userService, ok := userServiceInterface.(*service.UserService)
	if !ok {
		log.Fatal("Failed to assert UserService type")
	}

	// Controller layer
	leaderboardController := controllers.NewLeaderboardController(leaderboardService)
	userController := controllers.NewUserController(userService)
	updateController := controllers.NewUpdateController(updateService)
	// Handler layer
	leaderboardHandler := handlers.NewLeaderboardHandler(leaderboardController)
	userHandler := handlers.NewUserHandler(userController)
	updateHandler := handlers.NewUpdateHandler(updateController)
	// 4. Setup Gin router
	router := gin.Default()

	// 5. Add CORS middleware
	router.Use(cors.Default())

	// 6. Setup routes
	// Health check
	router.GET("/health", func(c *gin.Context) {
		redisStatus := "disconnected"
		if redisClient != nil {
			if err := redisClient.Ping(context.Background()).Err(); err == nil {
				redisStatus = "connected"
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"status":   "healthy",
			"database": "connected",
			"redis":    redisStatus,
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Leaderboard routes
		api.GET("/leaderboard", leaderboardHandler.GetLeaderboard)

		// User routes
		api.GET("/users/search", userHandler.SearchUsers)
		api.GET("/users/:username/rank", userHandler.GetUserRank)

		// Admin routes (for syncing Redis)
		if redisRepo != nil {
			api.POST("/admin/sync-redis", func(c *gin.Context) {
				ctx := context.Background()
				log.Println("Manual Redis sync triggered...")
				if err := userRepo.SyncAllUserToRedis(ctx, redisRepo); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				count, _ := redisRepo.GetTotalUsers(ctx)
				c.JSON(http.StatusOK, gin.H{
					"message": "Redis synced successfully",
					"count":   count,
				})
			})
			api.POST("/admin/simulate-updates", updateHandler.SimulateUpdates)
		}
	}

	// 7. Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default for local development
	}
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}

	go func() {
		ticker := time.NewTicker(5 * time.Minute) // Update every 5 minutes
		defer ticker.Stop()

		for range ticker.C {
			log.Println("Running scheduled random updates...")
			if err := updateService.SimulateRandomUpdates(10); err != nil {
				log.Printf("Scheduled update failed: %v", err)
			}
		}
	}()
}
