package main

import (
	"log"
	"net/http"

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

	// 3. Initialize layers (bottom to top)
	// Repository layer
	userRepo := repository.NewUserRepository(db)

	// Service layer
	leaderboardServiceInterface := service.NewLeaderboardService(userRepo)
	userServiceInterface := service.NewUserService(userRepo)

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

	// Handler layer
	leaderboardHandler := handlers.NewLeaderboardHandler(leaderboardController)
	userHandler := handlers.NewUserHandler(userController)

	// 4. Setup Gin router
	router := gin.Default()

	// 5. Add CORS middleware
	router.Use(cors.Default())

	// 6. Setup routes
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "database": "connected"})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Leaderboard routes
		api.GET("/leaderboard", leaderboardHandler.GetLeaderboard)

		// User routes
		api.GET("/users/search", userHandler.SearchUsers)
		api.GET("/users/:username/rank", userHandler.GetUserRank)
	}

	// 7. Start server
	port := ":8080"
	log.Printf("🚀 Server starting on port %s", port)
	if err := router.Run(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
