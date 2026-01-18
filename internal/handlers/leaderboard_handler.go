package handlers

import (
	"net/http"
	"strconv"

	"matiks/leaderboard/internal/controllers"

	"github.com/gin-gonic/gin"
)

type LeaderboardHandler struct {
	controller *controllers.LeaderboardController
}

func NewLeaderboardHandler(controller *controllers.LeaderboardController) *LeaderboardHandler {
	return &LeaderboardHandler{controller: controller}
}

// GetLeaderboard handles GET /api/v1/leaderboard
func (h *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	// 1. Extract query parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "50")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page parameter"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	// 2. Call controller
	response, err := h.controller.GetLeaderboard(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get leaderboard"})
		return
	}

	// 3. Return response
	c.JSON(http.StatusOK, response)
}
