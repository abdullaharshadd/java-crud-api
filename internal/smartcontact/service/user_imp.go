package service

import (
	"context"
	"fmt"

	"migrated-app/internal/apperr"
	"migrated-app/internal/model"
)

// SaveUser persists the given user via the repository and returns the saved user.
//
// Equivalent to UserServiceImp.saveUser.
func (s *UserService) SaveUser(ctx context.Context, user *model.User) (*model.User, error) {
	saved, err := s.repo.Save(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("save user: %w", err)
	}
	return saved, nil
}

// FetchUserList retrieves all users from the repository.
//
// Equivalent to UserServiceImp.fetchUserList.
func (s *UserService) FetchUserList(ctx context.Context) ([]*model.User, error) {
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch user list: %w", err)
	}
	return users, nil
}

// FetchUserByID retrieves a single user by its id, returning
// apperr.ErrUserNotFound if no such user exists.
//
// Equivalent to UserServiceImp.fetchUserById, which threw UserNotFoundException.
func (s *UserService) FetchUserByID(ctx context.Context, id int) (*model.User, error) {
	user, found, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetch user by id: %w", err)
	}
	if !found {
		return nil, apperr.ErrUserNotFound
	}
	return user, nil
}

// DeleteUser deletes the user with the given id via the repository.
//
// Equivalent to UserServiceImp.deleteUser.
func (s *UserService) DeleteUser(ctx context.Context, id int) error {
	if err := s.repo.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

// UpdateUser sets the id on the provided user and persists it via the repository.
//
// Equivalent to UserServiceImp.updateUser.
func (s *UserService) UpdateUser(ctx context.Context, id int, user *model.User) error {
	user.SetID(id)
	if _, err := s.repo.Save(ctx, user); err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

// GetUserNameByName retrieves a user by its name via the repository.
//
// Equivalent to UserServiceImp.getUserNameByName.
func (s *UserService) GetUserNameByName(ctx context.Context, name string) (*model.User, error) {
	user, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get user by name: %w", err)
	}
	return user, nil
}