package services

import (
	"ai-later-nav/internal/database/repository"
	"ai-later-nav/internal/models"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUsernameExists    = errors.New("username already exists")
	ErrEmailExists       = errors.New("email already exists")
	ErrInvalidPassword   = errors.New("invalid password")
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService() *UserService {
	return &UserService{
		userRepo: repository.NewUserRepository(),
	}
}

func (s *UserService) Register(username, email, password string) (*models.User, error) {
	if existing, _ := s.userRepo.GetByUsername(username); existing != nil {
		return nil, ErrUsernameExists
	}

	if email != "" {
		if existing, _ := s.userRepo.GetByEmail(email); existing != nil {
			return nil, ErrEmailExists
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         "user",
	}

	id, err := s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	user.ID = id
	return user, nil
}

func (s *UserService) Login(username, password string) (*models.User, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	return user, nil
}

func (s *UserService) GetByID(id int64) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *UserService) AddFavorite(userID, siteID int64) error {
	return s.userRepo.AddFavorite(userID, siteID)
}

func (s *UserService) RemoveFavorite(userID, siteID int64) error {
	return s.userRepo.RemoveFavorite(userID, siteID)
}

func (s *UserService) IsFavorite(userID, siteID int64) (bool, error) {
	return s.userRepo.IsFavorite(userID, siteID)
}

func (s *UserService) GetFavoriteIDs(userID int64) ([]int64, error) {
	return s.userRepo.GetFavorites(userID)
}
