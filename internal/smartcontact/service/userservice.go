package service

import (
	"context"
	"fmt"

	smartError "migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/repository"
)

type UserService interface {
	SaveUser(ctx context.Context, user model.User) (*model.User, error)
	FetchUserList(ctx context.Context) ([]model.User, error)
	FetchUserByID(ctx context.Context, id int) (*model.User, error)
	DeleteUser(ctx context.Context, id int) error
	UpdateUser(ctx context.Context, id int, user model.User) error
	GetUserByName(ctx context.Context, name string) (*model.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) (UserService, error) {
	if repo == nil {
		return nil, fmt.Errorf("service: user repository must not be nil")
	}
	return &userService{repo: repo}, nil
}

func (s *userService) SaveUser(ctx context.Context, user model.User) (*model.User, error) {
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("service: save user: %w", err)
	}
	saved, err := s.repo.Save(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("service: save user: %w", err)
	}
	return saved, nil
}

func (s *userService) FetchUserList(ctx context.Context) ([]model.User, error) {
	ptrs, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: fetch user list: %w", err)
	}
	users := make([]model.User, 0, len(ptrs))
	for _, p := range ptrs {
		if p != nil {
			users = append(users, *p)
		}
	}
	return users, nil
}

func (s *userService) FetchUserByID(ctx context.Context, id int) (*model.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service: fetch user by id %d: %w", id, err)
	}
	return user, nil
}

func (s *userService) DeleteUser(ctx context.Context, id int) error {
	if err := s.repo.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("service: delete user %d: %w", id, err)
	}
	return nil
}

func (s *userService) UpdateUser(ctx context.Context, id int, user model.User) error {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("service: update user %d: %w", id, err)
	}
	if existing == nil {
		return fmt.Errorf("service: update user %d: %w",
			id, smartError.NewUserNotFoundErrorf("user %d not found", id))
	}

	if err := user.Validate(); err != nil {
		return fmt.Errorf("service: update user %d: %w", id, err)
	}

	user.ID = id
	if _, err := s.repo.Save(ctx, &user); err != nil {
		return fmt.Errorf("service: update user %d: %w", id, err)
	}
	return nil
}

func (s *userService) GetUserByName(ctx context.Context, name string) (*model.User, error) {
	user, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("service: get user by name %q: %w", name, err)
	}
	return user, nil
}