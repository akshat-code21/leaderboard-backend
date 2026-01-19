package controllers

import "matiks/leaderboard/internal/service"

type UpdateController struct {
	updateService *service.UpdateService
}

func NewUpdateController(updateService *service.UpdateService) *UpdateController {
	return &UpdateController{updateService: updateService}
}

func (c *UpdateController) SimulateUpdates(count int) error {
	return c.updateService.SimulateRandomUpdates(count)
}
