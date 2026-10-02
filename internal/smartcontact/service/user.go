// Package service implements the user-related business operations for the
// SmartContact application.
//
// MIGRATION_NOTE: The source UserService was a Spring interface whose six
// methods (saveUser, fetchUserList, fetchUserById, deleteUser, updateUser,
// getUserNameByName) were implemented by an auto-wired concrete class. Here the
// interface becomes a concrete struct backed by a narrow UserRepository
// interface declaring exactly the persistence operations these methods need.
package service

import (
	"context"
	"errors"
	"fmt"

	"migrated-app/internal/apperr"
	"migrated-app/internal/model"
)

// ErrUserNotFound is returned when a requested user does not exist.
//
// MIGRATION_NOTE: Replaces the checked UserNotFoundException thrown by
// fetchUserById in the source. It aliases the shared apperr sentinel so callers
// (e.g. the REST error handler) can match it with errors.Is regardless of which
// layer surfaced it.
var ErrUserNotFound = apperr.ErrUserNotFound

// UserRepository is the subset of persistence operations UserService depends on.
//
// MIGRATION_NOTE: Mirrors the Spring Data UserDao methods actually used by the
// service: Save, FindAll, FindByID, Update, DeleteByID and FindByName.
type UserRepository interface {
	Save(ctx context.Context, u *model.User) (*model.User, error)
	FindAll(ctx context.Context) ([]model.User, error)
	FindByID(ctx context.Context, id int) (*model.User, error)
	Update(ctx context.Context, id int, u *model.User) (*model.User, error)
	DeleteByID(ctx context.Context, id int) error
	FindByName(ctx context.Context, name string) (*model.User, error)
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
//
// Mirrors the source saveUser(User).
func (s *UserService) SaveUser(ctx context.Context, user *model.User) (*model.User, error) {
	saved, err := s.repo.Save(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("save user: %w", err)
	}
	return saved, nil
}

// FetchUserList returns all users.
//
// Mirrors the source fetchUserList().
func (s *UserService) FetchUserList(ctx context.Context) ([]model.User, error) {
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch user list: %w", err)
	}
	return users, nil
}

// FetchUserByID returns the user with the given id, or ErrUserNotFound if no
// such user exists.
//
// Mirrors the source fetchUserById(int) which threw UserNotFoundException.
func (s *UserService) FetchUserByID(ctx context.Context, id int) (*model.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("fetch user by id %d: %w", id, err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// DeleteUser removes the user with the given id.
//
// Mirrors the source deleteUser(int).
func (s *UserService) DeleteUser(ctx context.Context, id int) error {
	if err := s.repo.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("delete user %d: %w", id, err)
	}
	return nil
}

// UpdateUser updates the user identified by id with the provided data and
// returns the updated user.
//
// MIGRATION_NOTE: The source updateUser returned void; the Go variant returns
// the updated *model.User so callers can respond without a second read, while
// still faithfully mapping (int id, User user) -> update semantics.
func (s *UserService) UpdateUser(ctx context.Context, id int, user *model.User) (*model.User, error) {
	updated, err := s.repo.Update(ctx, id, user)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("update user %d: %w", id, err)
	}
	return updated, nil
}

// GetUserNameByName returns the user matching the given name.
//
// Mirrors the source getUserNameByName(String).
func (s *UserService) GetUserNameByName(ctx context.Context, name string) (*model.User, error) {
	user, err := s.repo.FindByName(ctx, name)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by name %q: %w", name, err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}
