package service

import (
	appErr "user-service/internal/errors"
	"user-service/internal/model"
	"user-service/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(r *repository.UserRepository) *UserService {
	return &UserService{
		repo: r,
	}
}

// Get user by email
func (s *UserService) GetProfile(email string) (*model.User, error) {
	return s.repo.FindByEmail(email)
}

// Create new user
func (s *UserService) CreateUser(user *model.User) error {

	existing, _ := s.repo.FindByEmail(user.Email)

	if existing != nil {
		return appErr.ErrUserAlreadyExists
	}

	return s.repo.CreateUser(user)

}

// Update user profile
func (s *UserService) UpdateProfile(email string, update map[string]interface{}) error {
	return s.repo.UpdateUser(email, update)
}
