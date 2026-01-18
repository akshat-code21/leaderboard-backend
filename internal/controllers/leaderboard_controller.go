package controllers

import (
	"matiks/leaderboard/internal/models"
	"matiks/leaderboard/internal/service"
)

type LeaderboardController struct {
	leaderboardService *service.LeaderboardService
}

func NewLeaderboardController(leaderboardService *service.LeaderboardService) *LeaderboardController {
	return &LeaderboardController{leaderboardService: leaderboardService}
}

func (c *LeaderboardController) GetLeaderboard(page, limit int) (*models.LeaderboardResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	response, err := c.leaderboardService.GetLeaderboard(page, limit)
	if err != nil {
		return nil, err
	}
	return response, nil
}
