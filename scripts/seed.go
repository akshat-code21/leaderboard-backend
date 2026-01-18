package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"matiks/leaderboard/internal/config"
	"matiks/leaderboard/internal/models"

	"gorm.io/gorm"
)

const (
	// Total number of users to seed
	totalUsers = 10000
	// Batch size for inserting users (for performance)
	batchSize = 500
)

func main() {
	log.Println("🌱 Starting database seeding...")

	// 1. Connect to database
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 2. Run migrations (ensure table exists)
	if err := config.AutoMigrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// 3. Check if users already exist
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		log.Printf("⚠️  Database already contains %d users", count)
		log.Println("Do you want to clear existing data and reseed? (y/n)")
		var response string
		fmt.Scanln(&response)
		if response == "y" || response == "Y" {
			log.Println("🗑️  Clearing existing users...")
			db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
			log.Println("✅ Existing users cleared")
		} else {
			log.Println("Seeding cancelled. Exiting...")
			return
		}
	}

	// 4. Seed users
	log.Printf("📊 Generating %d users...", totalUsers)
	startTime := time.Now()

	users := generateUsers(totalUsers)

	log.Printf("💾 Inserting users in batches of %d...", batchSize)
	if err := insertUsersInBatches(db, users, batchSize); err != nil {
		log.Fatal("Failed to insert users:", err)
	}

	elapsed := time.Since(startTime)
	log.Printf("✅ Successfully seeded %d users in %v", totalUsers, elapsed)

	// 5. Verify seeding
	var finalCount int64
	db.Model(&models.User{}).Count(&finalCount)
	log.Printf("📈 Total users in database: %d", finalCount)

	// 6. Show some statistics
	showStatistics(db)
}

// generateUsers creates a slice of users with random usernames and ratings
func generateUsers(count int) []models.User {
	rand.Seed(time.Now().UnixNano())
	users := make([]models.User, 0, count)

	// Create a map to track usernames and ensure uniqueness
	usernameMap := make(map[string]bool)

	for i := 0; i < count; i++ {
		// Generate unique username
		username := generateUniqueUsername(usernameMap, i)

		// Generate rating (100-5000)
		// We'll create some ties by grouping ratings
		rating := generateRatingWithTies(i, count)

		user := models.User{
			Username: username,
			Rating:   rating,
		}

		users = append(users, user)
	}

	return users
}

// generateUniqueUsername creates a unique username
func generateUniqueUsername(usernameMap map[string]bool, index int) string {
	var username string
	for {
		// Generate username patterns
		patterns := []string{
			fmt.Sprintf("user_%d", index+1),
			fmt.Sprintf("player_%d", index+1),
			fmt.Sprintf("gamer_%d", index+1),
			fmt.Sprintf("user%d", index+1),
			fmt.Sprintf("player%d", index+1),
		}

		// Randomly select a pattern
		pattern := patterns[rand.Intn(len(patterns))]

		// Add some variation with random suffix
		if rand.Float32() < 0.3 {
			username = fmt.Sprintf("%s_%d", pattern, rand.Intn(1000))
		} else {
			username = pattern
		}

		// Ensure uniqueness
		if !usernameMap[username] {
			usernameMap[username] = true
			break
		}
		// If collision, add random suffix
		username = fmt.Sprintf("%s_%d", username, rand.Intn(10000))
		if !usernameMap[username] {
			usernameMap[username] = true
			break
		}
	}

	return username
}

// generateRatingWithTies generates ratings with intentional ties for testing
func generateRatingWithTies(index, total int) int {
	// Create rating distribution with intentional ties
	// Higher ratings are less common (like a real leaderboard)

	// 20% chance to have a tie (same rating as nearby users)
	if rand.Float32() < 0.2 && index > 0 {
		// Return a rating that might create a tie
		// Use common rating values
		tieRatings := []int{1000, 1500, 2000, 2500, 3000, 3500, 4000, 4500}
		return tieRatings[rand.Intn(len(tieRatings))]
	}

	// Generate rating with distribution:
	// - 10% chance: 4500-5000 (top tier)
	// - 20% chance: 3500-4499 (high tier)
	// - 30% chance: 2500-3499 (mid-high tier)
	// - 30% chance: 1500-2499 (mid tier)
	// - 10% chance: 100-1499 (low tier)

	randVal := rand.Float32()

	switch {
	case randVal < 0.1:
		// Top tier: 4500-5000
		return 4500 + rand.Intn(501)
	case randVal < 0.3:
		// High tier: 3500-4499
		return 3500 + rand.Intn(1000)
	case randVal < 0.6:
		// Mid-high tier: 2500-3499
		return 2500 + rand.Intn(1000)
	case randVal < 0.9:
		// Mid tier: 1500-2499
		return 1500 + rand.Intn(1000)
	default:
		// Low tier: 100-1499
		return 100 + rand.Intn(1400)
	}
}

// insertUsersInBatches inserts users in batches for better performance
func insertUsersInBatches(db *gorm.DB, users []models.User, batchSize int) error {
	totalBatches := (len(users) + batchSize - 1) / batchSize

	for i := 0; i < len(users); i += batchSize {
		end := i + batchSize
		if end > len(users) {
			end = len(users)
		}

		batch := users[i:end]
		batchNum := (i / batchSize) + 1

		if err := db.Create(&batch).Error; err != nil {
			return fmt.Errorf("failed to insert batch %d/%d: %w", batchNum, totalBatches, err)
		}

		if batchNum%10 == 0 || batchNum == totalBatches {
			log.Printf("  ✓ Inserted batch %d/%d (%d users)", batchNum, totalBatches, end)
		}
	}

	return nil
}

// showStatistics displays some statistics about the seeded data
func showStatistics(db *gorm.DB) {
	log.Println("\n📊 Database Statistics:")

	// Total users
	var totalCount int64
	db.Model(&models.User{}).Count(&totalCount)
	log.Printf("  Total users: %d", totalCount)

	// Rating statistics
	var minRating, maxRating, avgRating int
	db.Model(&models.User{}).Select("MIN(rating)").Scan(&minRating)
	db.Model(&models.User{}).Select("MAX(rating)").Scan(&maxRating)
	db.Model(&models.User{}).Select("AVG(rating)").Scan(&avgRating)
	log.Printf("  Rating range: %d - %d (avg: %d)", minRating, maxRating, avgRating)

	// Count users by rating tier
	var topTier, highTier, midTier, lowTier int64
	db.Model(&models.User{}).Where("rating >= ?", 4500).Count(&topTier)
	db.Model(&models.User{}).Where("rating >= ? AND rating < ?", 3500, 4500).Count(&highTier)
	db.Model(&models.User{}).Where("rating >= ? AND rating < ?", 2500, 3500).Count(&midTier)
	db.Model(&models.User{}).Where("rating < ?", 2500).Count(&lowTier)

	log.Println("\n  Rating Distribution:")
	log.Printf("    Top tier (4500-5000): %d users", topTier)
	log.Printf("    High tier (3500-4499): %d users", highTier)
	log.Printf("    Mid tier (2500-3499): %d users", midTier)
	log.Printf("    Low tier (100-2499): %d users", lowTier)

	// Find some example ties
	log.Println("\n  Example Ties (users with same rating):")
	var tieExamples []models.User
	db.Where("rating IN (SELECT rating FROM users GROUP BY rating HAVING COUNT(*) > 1)").
		Order("rating DESC").
		Limit(10).
		Find(&tieExamples)

	if len(tieExamples) > 0 {
		ratingGroups := make(map[int][]string)
		for _, user := range tieExamples {
			ratingGroups[user.Rating] = append(ratingGroups[user.Rating], user.Username)
		}

		count := 0
		for rating, usernames := range ratingGroups {
			if count >= 5 {
				break
			}
			if len(usernames) > 1 {
				log.Printf("    Rating %d: %d users (e.g., %s, %s)", rating, len(usernames), usernames[0], usernames[1])
				count++
			}
		}
	}

	log.Println("\n✅ Seeding completed successfully!")
	log.Println("🚀 You can now start the server with: go run cmd/server/main.go")
}
