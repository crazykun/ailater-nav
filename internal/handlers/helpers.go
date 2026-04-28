package handlers

import (
	"ai-later-nav/internal/middleware"
	"ai-later-nav/internal/models"
)

func generateToken(user *models.User) (string, error) {
	return middleware.GenerateToken(user)
}
