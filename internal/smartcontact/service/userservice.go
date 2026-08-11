// Package service defines the business-logic layer contracts and
// implementations migrated from com.smartContact.service.
//
// MIGRATION_NOTE: The Java UserService was a Spring @Service interface
// with no implementation body (the concrete UserServiceImpl was a
// separate class). This file migrates the interface contract to an
// idiomatic Go interface plus a concrete database/sql-backed
// implementation, since the migration notes call for it to be injected
// into NewUserHandler via a constructor. The method set is preserved:
// SaveUser, GetAllUsers, FetchUserById, DeleteUser, UpdateUser and
// GetUserNameByName. Each I/O method takes a context.Context as its
// first parameter per Go conventions.
package service

import (
	"context"
	"fmt"

	apperr "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/repository"
)

// UserService defines the user-related business operations exposed to the
// HTTP/handler layer. It mirrors the Spring UserService interface.
type UserService interface {
	// SaveUser persists a new user and returns the stored representation
	// (including any generated identifier).
	SaveUser(ctx context.Context, user *model.User) (*model.User, error)

	// GetAllUsers returns every user.
	//
	// MIGRATION_NOTE: mirrors the Java fetchUserList().
	GetAllUsers(ctx context.Context) ([]*model.User, error)

	// FetchUserById returns the user with the given id, or a
	// *apperr.UserNotFoundError if no such user exists.
	FetchUserById(ctx context.Context, id int) (*model.User, error)

	// DeleteUser removes the user with the given id.
	DeleteUser(ctx context.Context, id int) error

	// UpdateUser applies the fields of user to the record identified by id.
	UpdateUser(ctx context.Context, id int, user *model.User) error

	// GetUserNameByName looks up a user by name.
	//
	// MIGRATION_NOTE: mirrors the Java getUserNameByName(String name).
	GetUserNameByName(ctx context.Context, name string) (*model.User, error)
}

// userService is the default UserService implementation backed by a
// repository.UserRepository.
type userService struct {
	repo repository.UserRepository
}

// NewUserService constructs a UserService backed by the given repository.
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

// SaveUser validates and persists a new user.
func (s *userService) SaveUser(ctx context.Context, user *model.User) (*model.User, error) {
	if user == nil {
		return nil, fmt.Errorf("save user: user must not be nil")
	}
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("save user: %w", err)
	}
	saved, err := s.repo.Save(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("save user: %w", err)
	}
	return saved, nil
}

// GetAllUsers returns all persisted users.
func (s *userService) GetAllUsers(ctx context.Context) ([]*model.User, error) {
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all users: %w", err)
	}
	return users, nil
}

// FetchUserById returns the user with the given id. A missing row is
// translated into a *apperr.UserNotFoundError, mirroring the Java method's
// declared throws UserNotFoundException.
func (s *userService) FetchUserById(ctx context.Context, id int) (*model.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil, apperr.NewUserNotFoundErrorf("user with id %d not found", id)
		}
		return nil, fmt.Errorf("fetch user by id: %w", err)
	}
	return user, nil
}

// DeleteUser removes the user with the given id.
func (s *userService) DeleteUser(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if isNotFound(err) {
			return apperr.NewUserNotFoundErrorf("user with id %d not found", id)
		}
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

// UpdateUser applies user's fields to the record identified by id.
//
// MIGRATION_NOTE: The Java updateUser mutated a managed JPA entity and
// relied on Hibernate dirty-checking to flush the changes. Go has no
// such session, so the update is expressed by loading the existing row,
// applying non-nil fields and re-saving. If your repository exposes a
// dedicated Update method, prefer wiring it here.
func (s *userService) UpdateUser(ctx context.Context, id int, user *model.User) error {
	if user == nil {
		return fmt.Errorf("update user: user must not be nil")
	}
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return apperr.NewUserNotFoundErrorf("user with id %d not found", id)
		}
		return fmt.Errorf("update user: %w", err)
	}

	if user.Name != nil {
		existing.Name = user.Name
	}
	if user.Email != nil {
		existing.Email = user.Email
	}
	if user.Password != nil {
		existing.Password = user.Password
	}
	if user.Role != nil {
		existing.Role = user.Role
	}
	if user.About != nil {
		existing.About = user.About
	}

	if err := existing.Validate(); err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if _, err := s.repo.Save(ctx, existing); err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

// GetUserNameByName looks up a user by name.
func (s *userService) GetUserNameByName(ctx context.Context, name string) (*model.User, error) {
	user, err := s.repo.FindByName(ctx, name)
	if err != nil {
		if isNotFound(err) {
			return nil, apperr.NewUserNotFoundErrorf("user with name %q not found", name)
		}
		return nil, fmt.Errorf("get user by name: %w", err)
	}
	return user, nil
}

// isNotFound reports whether err represents a missing-row condition from
// the repository layer.
func isNotFound(err error) bool {
	return errorsIs(err, repository.ErrUserNotFound)
}
