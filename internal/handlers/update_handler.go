package handlers

import (
	"net/http"
	"strconv"

	"matiks/leaderboard/internal/controllers"

	"github.com/gin-gonic/gin"
)

type UpdateHandler struct {
	controller *controllers.UpdateController
}

func NewUpdateHandler(controller *controllers.UpdateController) *UpdateHandler {
	return &UpdateHandler{controller: controller}
}

// SimulateUpdates handles POST /api/v1/admin/simulate-updates
func (h *UpdateHandler) SimulateUpdates(c *gin.Context) {
	countStr := c.DefaultQuery("count", "10")
	count, err := strconv.Atoi(countStr)
	if err != nil || count < 1 || count > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "count must be between 1 and 100",
		})
		return
	}

	// Queue updates (non-blocking)
	if err := h.controller.SimulateUpdates(count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Updates queued successfully",
		"count":   count,
	})
}
