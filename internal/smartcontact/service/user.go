// Package service implements the user-related business operations for the
// SmartContact application.
package service

import (
	"context"
	"fmt"

	"migrated-app/internal/model"
)

// UserRepository is the subset of persistence operations UserService depends on.
type UserRepository interface {
	Save(ctx context.Context, u *model.User) (*model.User, error)
	FindAll(ctx context.Context) ([]model.User, error)
	DeleteByID(ctx context.Context, id int) error
}

// UserService provides user-related business operations.
type UserService struct {
	repo UserRepository
}

// NewUserService constructs a UserService backed by the given repository.
func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

// SaveUser persists a new user entity and returns the saved user.
func (s *UserService) SaveUser(ctx context.Context, user *model.User) (*model.User, error) {
	saved, err := s.repo.Save(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("save user: %w", err)
	}
	return saved, nil
}
