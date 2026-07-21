// Package service contains the SmartContact service's business-logic layer.
// It is the Go equivalent of the Java package com.smartContact.service.
//
// MIGRATION_NOTE: The source UserService.java was a Spring service-layer
// interface that decoupled controllers from a concrete @Service
// implementation. Spring wired the implementation via dependency injection.
//
// In idiomatic Go the interface is defined here (in the consuming package's
// domain) and a concrete implementation is provided by NewUserService, which
// takes the repository.UserDao dependency explicitly rather than through an
// injection container. This preserves the interface-based DI contract while
// following Go's constructor-based wiring.
//
// Additional idiomatic adjustments:
//   - context.Context is threaded through every method because each call
//     ultimately performs I/O against the data store.
//   - Methods return (T, error) instead of throwing checked exceptions.
//   - The Java int id becomes an int; callers convert as needed.
package service

import (
	"context"
	"errors"
	"fmt"

	smartcontacterror "internal/smartcontact/error"
	"internal/smartcontact/model"
	"internal/smartcontact/repository"
)

// UserService defines the service-layer contract for user-related business
// operations. It decouples the transport/handler layer from the concrete
// implementation, mirroring the Spring @Service abstraction.
type UserService interface {
	// SaveUser persists the given user and returns the stored representation.
	SaveUser(ctx context.Context, user *model.User) (*model.User, error)

	// FetchUserList returns all users.
	FetchUserList(ctx context.Context) ([]*model.User, error)

	// FetchUserByID returns the user with the given id. It returns a
	// *smartcontacterror.UserNotFoundError (wrapping ErrUserNotFound) when no
	// such user exists.
	FetchUserByID(ctx context.Context, id int) (*model.User, error)

	// DeleteUser removes the user with the given id.
	DeleteUser(ctx context.Context, id int) error

	// UpdateUser updates the user with the given id using the supplied data.
	UpdateUser(ctx context.Context, id int, user *model.User) error

	// GetUserByName returns the user matching the given name.
	//
	// MIGRATION_NOTE: the Java method was named getUserNameByName but returned
	// a full User, so this is renamed to GetUserByName for clarity.
	GetUserByName(ctx context.Context, name string) (*model.User, error)
}

// userService is the default UserService implementation backed by a
// repository.UserDao.
type userService struct {
	dao repository.UserDao
}

// NewUserService constructs a UserService backed by the given UserDao.
// It replaces Spring's dependency injection with explicit wiring.
func NewUserService(dao repository.UserDao) UserService {
	return &userService{dao: dao}
}

// SaveUser validates and persists the given user.
func (s *userService) SaveUser(ctx context.Context, user *model.User) (*model.User, error) {
	if user == nil {
		return nil, errors.New("service: user must not be nil")
	}
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("service: invalid user: %w", err)
	}

	saved, err := s.dao.Save(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("service: save user: %w", err)
	}
	return saved, nil
}

// FetchUserList returns all users.
func (s *userService) FetchUserList(ctx context.Context) ([]*model.User, error) {
	users, err := s.dao.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: fetch user list: %w", err)
	}
	return users, nil
}

// FetchUserByID returns the user with the given id, or a UserNotFoundError
// when the user does not exist.
func (s *userService) FetchUserByID(ctx context.Context, id int) (*model.User, error) {
	user, err := s.dao.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, smartcontacterror.ErrUserNotFound) {
			return nil, smartcontacterror.NewUserNotFoundErrorMessageCause(
				fmt.Sprintf("user not found for id %d", id), err)
		}
		return nil, fmt.Errorf("service: fetch user by id %d: %w", id, err)
	}
	if user == nil {
		return nil, smartcontacterror.NewUserNotFoundErrorMessage(
			fmt.Sprintf("user not found for id %d", id))
	}
	return user, nil
}

// DeleteUser removes the user with the given id.
func (s *userService) DeleteUser(ctx context.Context, id int) error {
	if err := s.dao.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("service: delete user %d: %w", id, err)
	}
	return nil
}

// UpdateUser updates the user with the given id using the supplied data.
//
// MIGRATION_NOTE: the Java interface declared updateUser but provided no
// body. The concrete Spring implementation (not part of this file) held the
// logic. Here we reconstruct the idiomatic behaviour: verify the target
// exists, apply the id, validate, and persist via Save (Spring Data's save
// performed an upsert). Review against the original implementation.
func (s *userService) UpdateUser(ctx context.Context, id int, user *model.User) error {
	if user == nil {
		return errors.New("service: user must not be nil")
	}

	existing, err := s.dao.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, smartcontacterror.ErrUserNotFound) {
			return smartcontacterror.NewUserNotFoundErrorMessageCause(
				fmt.Sprintf("user not found for id %d", id), err)
		}
		return fmt.Errorf("service: update user %d: lookup: %w", id, err)
	}
	if existing == nil {
		return smartcontacterror.NewUserNotFoundErrorMessage(
			fmt.Sprintf("user not found for id %d", id))
	}

	user.ID = id
	if err := user.Validate(); err != nil {
		return fmt.Errorf("service: invalid user: %w", err)
	}

	if _, err := s.dao.Save(ctx, user); err != nil {
		return fmt.Errorf("service: update user %d: save: %w", id, err)
	}
	return nil
}

// GetUserByName returns the user matching the given name.
func (s *userService) GetUserByName(ctx context.Context, name string) (*model.User, error) {
	user, err := s.dao.FindByName(ctx, name)
	if err != nil {
		if errors.Is(err, smartcontacterror.ErrUserNotFound) {
			return nil, smartcontacterror.NewUserNotFoundErrorMessageCause(
				fmt.Sprintf("user not found for name %q", name), err)
		}
		return nil, fmt.Errorf("service: get user by name %q: %w", name, err)
	}
	if user == nil {
		return nil, smartcontacterror.NewUserNotFoundErrorMessage(
			fmt.Sprintf("user not found for name %q", name))
	}
	return user, nil
}
