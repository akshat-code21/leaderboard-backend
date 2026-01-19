package models

import "time"

// User model
type User struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null"`
	Rating    int       `json:"rating" gorm:"not null;check:rating >= 100 AND rating <= 5000"`
	Rank      int       `json:"rank,omitempty" gorm:"-"` // Calculated field, not stored in DB
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// LeaderboardEntry represents a single entry in the leaderboard
type LeaderboardEntry struct {
	Rank     int    `json:"rank"`
	Username string `json:"username"`
	Rating   int    `json:"rating"`
}

// LeaderboardResponse represents the paginated leaderboard response
type LeaderboardResponse struct {
	Entries []LeaderboardEntry `json:"entries"`
	Page    int                `json:"page"`
	Limit   int                `json:"limit"`
	Total   int                `json:"total"`
}

// UserSearchResponse represents the user search response
type UserSearchResponse struct {
	Users []LeaderboardEntry `json:"users"`
	Page  int                `json:"page"`
	Limit int                `json:"limit"`
	Total int                `json:"total"`
}

// UserRankResponse represents a single user's rank information
type UserRankResponse struct {
	Username string `json:"username"`
	Rating   int    `json:"rating"`
	Rank     int    `json:"rank"`
}

// Generic response wrapper (optional, for error handling)
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
