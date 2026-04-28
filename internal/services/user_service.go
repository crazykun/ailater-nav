package services

import (
	"ai-later-nav/internal/database/repository"
	"ai-later-nav/internal/models"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUsernameExists  = errors.New("username already exists")
	ErrInvalidPassword = errors.New("invalid password")
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService() *UserService {
	return &UserService{
		userRepo: repository.NewUserRepository(),
	}
}

func (s *UserService) Register(username, password string) (*models.User, error) {
	if existing, _ := s.userRepo.GetByUsername(username); existing != nil {
		return nil, ErrUsernameExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:     username,
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

func (s *UserService) ChangePassword(userID int64, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrInvalidPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(userID, string(hash))
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

func (s *UserService) CountUsers() (int64, error) {
	return s.userRepo.CountUsers()
}

func (s *UserService) GetAllUsers() ([]*models.User, error) {
	return s.userRepo.GetAllUsers()
}

func (s *UserService) HasAnyUser() (bool, error) {
	count, err := s.userRepo.CountUsers()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *UserService) RegisterAdmin(username, password string) (*models.User, error) {
	hasUser, err := s.HasAnyUser()
	if err != nil {
		return nil, err
	}
	if hasUser {
		return nil, errors.New("admin already exists")
	}

	if existing, _ := s.userRepo.GetByUsername(username); existing != nil {
		return nil, ErrUsernameExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         "admin",
	}

	id, err := s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	user.ID = id
	return user, nil
}
