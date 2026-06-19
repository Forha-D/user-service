package service

import (
	"context"
	"time"

	appErr "user-service/internal/errors"
	"user-service/internal/model"
	"user-service/internal/repository"
)

type UserService struct {
	repo    *repository.UserRepository
	timeout time.Duration
}

func NewUserService(r *repository.UserRepository, timeout time.Duration) *UserService {
	return &UserService{
		repo:    r,
		timeout: timeout,
	}
}

// Get user by email
func (s *UserService) GetProfile(ctx context.Context, email string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	return s.repo.FindByEmail(ctx, email)
}

// Create new user
func (s *UserService) CreateUser(ctx context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	existing, err := s.repo.FindByEmail(ctx, user.Email)
	if err != nil && err != appErr.ErrUserNotFound {
		return err
	}

	if existing != nil {
		return appErr.ErrUserAlreadyExists
	}

	return s.repo.CreateUser(ctx, user)
}

// Update user profile
func (s *UserService) UpdateProfile(ctx context.Context, email string, update map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	return s.repo.UpdateUser(ctx, email, update)
}
