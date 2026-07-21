// Package service contains the SmartContact service's business-logic layer.
// It is the Go equivalent of the Java package com.smartContact.service.
package service

import (
	"context"
	"errors"
	"fmt"

	smartError "internal/smartcontact/error"
	"internal/smartcontact/model"
	"internal/smartcontact/repository"
)

// userServiceImp is the concrete implementation of UserService.
//
// MIGRATION_NOTE: The source UserServiceImp.java was a Spring @Service that
// used field injection (@Autowired) to obtain its UserDao dependency. In
// idiomatic Go the dependency is injected explicitly through the constructor
// (NewUserService) and stored as an unexported field. The struct is
// unexported because callers depend on the UserService interface, not the
// concrete type.
type userServiceImp struct {
	userDao repository.UserDao
}

// NewUserService constructs a UserService backed by the given UserDao.
//
// This replaces Spring's dependency-injection container: the repository
// dependency is passed in explicitly rather than resolved from a context.
func NewUserService(userDao repository.UserDao) UserService {
	return &userServiceImp{userDao: userDao}
}

// SaveUser persists the given user and returns the stored representation.
func (s *userServiceImp) SaveUser(ctx context.Context, user model.User) (model.User, error) {
	saved, err := s.userDao.Save(ctx, user)
	if err != nil {
		return model.User{}, fmt.Errorf("save user: %w", err)
	}
	return saved, nil
}

// FetchUserList returns all persisted users.
func (s *userServiceImp) FetchUserList(ctx context.Context) ([]model.User, error) {
	users, err := s.userDao.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch user list: %w", err)
	}
	return users, nil
}

// FetchUserByID returns the user with the given id.
//
// MIGRATION_NOTE: The Java version wrapped the lookup in an Optional and threw
// a checked UserNotFoundException when the value was absent. In Go the absence
// is surfaced as an error value: a smartError.UserNotFoundError is returned so
// callers can detect it with errors.Is/As. The original message
// ("User are not available") is preserved.
func (s *userServiceImp) FetchUserByID(ctx context.Context, id int) (model.User, error) {
	user, found, err := s.userDao.FindByID(ctx, id)
	if err != nil {
		return model.User{}, fmt.Errorf("fetch user by id %d: %w", id, err)
	}
	if !found {
		return model.User{}, smartError.NewUserNotFoundErrorMessage("User are not available")
	}
	return user, nil
}

// DeleteUser removes the user with the given id.
func (s *userServiceImp) DeleteUser(ctx context.Context, id int) error {
	if err := s.userDao.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("delete user %d: %w", id, err)
	}
	return nil
}

// UpdateUser sets the id on the given user and persists it.
//
// MIGRATION_NOTE: The Java version mutated the passed User via setId and
// returned void. Here the user is copied, its ID is set, and the persisted
// value is returned to keep the function free of hidden mutation. The
// @NotNull annotation on the primitive int parameter carried no runtime
// meaning in Java and has no Go analogue.
func (s *userServiceImp) UpdateUser(ctx context.Context, id int, user model.User) (model.User, error) {
	user.ID = id
	saved, err := s.userDao.Save(ctx, user)
	if err != nil {
		return model.User{}, fmt.Errorf("update user %d: %w", id, err)
	}
	return saved, nil
}

// GetUserByName returns the user with the given name.
//
// MIGRATION_NOTE: The Java findByName could return null with no existence
// check. To make the absent case explicit for callers, this returns a
// UserNotFoundError when no matching user is found, mirroring FetchUserByID.
func (s *userServiceImp) GetUserByName(ctx context.Context, name string) (model.User, error) {
	user, found, err := s.userDao.FindByName(ctx, name)
	if err != nil {
		if errors.Is(err, smartError.ErrUserNotFound) {
			return model.User{}, err
		}
		return model.User{}, fmt.Errorf("get user by name %q: %w", name, err)
	}
	if !found {
		return model.User{}, smartError.NewUserNotFoundErrorMessage("User are not available")
	}
	return user, nil
}
