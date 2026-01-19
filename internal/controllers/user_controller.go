package controllers

import (
	"errors"
	"matiks/leaderboard/internal/models"
	"matiks/leaderboard/internal/service"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController(userService *service.UserService) *UserController {
	return &UserController{userService: userService}
}

func (c *UserController) SearchUsers(query string, page, limit int) (*models.UserSearchResponse, error) {
	if query == "" {
		return nil, errors.New("query is required")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		return nil, errors.New("limit must be between 1 and 100")
	}
	response, err := c.userService.SearchUsers(query, page, limit)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *UserController) GetUserRank(username string) (*models.UserRankResponse, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}
	response, err := c.userService.GetUserRank(username)
	if err != nil {
		return nil, err
	}
	return response, nil
}
