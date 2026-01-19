package handlers

import (
	"net/http"
	"strconv"

	"matiks/leaderboard/internal/controllers"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	controller *controllers.UserController
}

func NewUserHandler(controller *controllers.UserController) *UserHandler {
	return &UserHandler{controller: controller}
}

// SearchUsers handles GET /api/v1/users/search
func (h *UserHandler) SearchUsers(c *gin.Context) {
	// 1. Extract query parameters
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	// 2. Call controller
	response, err := h.controller.SearchUsers(query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search users"})
		return
	}

	// 3. Return response
	c.JSON(http.StatusOK, response)
}

// GetUserRank handles GET /api/v1/users/:username/rank
func (h *UserHandler) GetUserRank(c *gin.Context) {
	// 1. Extract path parameter
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	// 2. Call controller
	response, err := h.controller.GetUserRank(username)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user rank"})
		return
	}

	// 3. Return response
	c.JSON(http.StatusOK, response)
}
