// Package service defines the business-logic (service) layer contracts and
// implementations for the smartcontact application.
//
// MIGRATION_NOTE: The Java source, UserService, was a Spring service-layer
// interface implemented by a @Service-annotated class. This file migrates the
// interface contract to a Go interface (UserService) and provides a concrete
// implementation (userService) constructed via NewUserService. Dependency
// injection is explicit through the constructor rather than Spring's
// annotation-driven container.
//
// MIGRATION_NOTE: Every I/O-bound method takes a context.Context as its first
// parameter, per Go convention, even though the Java interface had no such
// parameter. This enables cancellation and deadline propagation down to the
// repository/database layer.
//
// MIGRATION_NOTE: The Java methods returned values directly and, in one case,
// declared a checked UserNotFoundException. Idiomatic Go returns (T, error)
// and models the "not found" condition with the ErrUserNotFound sentinel (see
// internal/smartcontact/error). Callers use errors.Is to detect it.
package service

import (
	"context"
	"fmt"

	smartError "migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/repository"
)

// UserService defines the business-layer operations for managing users.
//
// MIGRATION_NOTE: This mirrors the Java UserService interface. Method names
// are adjusted to Go conventions (e.g. saveUser -> SaveUser) and each I/O
// method accepts a context.Context. Lookup methods return an error rather than
// throwing/returning null.
type UserService interface {
	// SaveUser persists a new user and returns the stored representation.
	SaveUser(ctx context.Context, user model.User) (model.User, error)

	// FetchUserList returns all users.
	FetchUserList(ctx context.Context) ([]model.User, error)

	// FetchUserByID looks up a user by its numeric id. It returns
	// ErrUserNotFound (wrapped) when no matching user exists.
	FetchUserByID(ctx context.Context, id int) (model.User, error)

	// DeleteUser removes the user with the given id.
	DeleteUser(ctx context.Context, id int) error

	// UpdateUser updates the user with the given id using the supplied data
	// and returns the updated representation.
	UpdateUser(ctx context.Context, id int, user model.User) (model.User, error)

	// GetUserByName looks up a user by name.
	//
	// MIGRATION_NOTE: The Java method was getUserNameByName and did NOT declare
	// UserNotFoundException. Per the migration debate notes, the service layer
	// owns the by-name no-match policy decision (HTTP 200 + null vs 404), which
	// is still pending capture from the running source app. This implementation
	// propagates whatever the repository returns (including a plain
	// sql.ErrNoRows-style not-found), and does NOT map it to ErrUserNotFound,
	// preserving the observable difference from the by-id path until the policy
	// is confirmed. REQUIRES MANUAL REVIEW.
	GetUserByName(ctx context.Context, name string) (model.User, error)
}

// userService is the concrete UserService implementation backed by a
// UserRepository.
//
// MIGRATION_NOTE: This replaces the Spring @Service implementation class. The
// repository dependency is injected via NewUserService.
type userService struct {
	repo repository.UserRepository
}

// NewUserService constructs a UserService backed by the given UserRepository.
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

// SaveUser validates and persists a new user, returning the stored user.
func (s *userService) SaveUser(ctx context.Context, user model.User) (model.User, error) {
	if err := user.Validate(); err != nil {
		return model.User{}, fmt.Errorf("save user: validation failed: %w", err)
	}

	saved, err := s.repo.Save(ctx, user)
	if err != nil {
		return model.User{}, fmt.Errorf("save user: %w", err)
	}
	return saved, nil
}

// FetchUserList returns all users.
func (s *userService) FetchUserList(ctx context.Context) ([]model.User, error) {
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch user list: %w", err)
	}
	return users, nil
}

// FetchUserByID looks up a user by id, returning a wrapped ErrUserNotFound
// when no matching user exists.
//
// MIGRATION_NOTE: This is the confirmed 404 path from the migration notes. The
// repository's not-found error is normalized to ErrUserNotFound so HTTP
// handlers can map it to 404 via errors.Is.
func (s *userService) FetchUserByID(ctx context.Context, id int) (model.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if smartError.IsUserNotFound(err) {
			return model.User{}, smartError.WrapUserNotFound(err)
		}
		return model.User{}, fmt.Errorf("fetch user by id %d: %w", id, err)
	}
	return user, nil
}

// DeleteUser removes the user with the given id.
func (s *userService) DeleteUser(ctx context.Context, id int) error {
	if err := s.repo.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("delete user %d: %w", id, err)
	}
	return nil
}

// UpdateUser updates the user identified by id with the supplied data and
// returns the updated user.
//
// MIGRATION_NOTE: The Java interface's updateUser returned void. Go convention
// favors returning the updated entity so callers need not re-fetch. The method
// first confirms the user exists (mapping absence to ErrUserNotFound), applies
// the id, validates, then persists.
func (s *userService) UpdateUser(ctx context.Context, id int, user model.User) (model.User, error) {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		if smartError.IsUserNotFound(err) {
			return model.User{}, smartError.WrapUserNotFound(err)
		}
		return model.User{}, fmt.Errorf("update user %d: %w", id, err)
	}

	user.ID = id
	if err := user.Validate(); err != nil {
		return model.User{}, fmt.Errorf("update user %d: validation failed: %w", id, err)
	}

	updated, err := s.repo.Save(ctx, user)
	if err != nil {
		return model.User{}, fmt.Errorf("update user %d: %w", id, err)
	}
	return updated, nil
}

// GetUserByName looks up a user by name.
//
// MIGRATION_NOTE: Per the debate notes, the by-name no-match policy (200+null
// vs 404) is unresolved. Unlike FetchUserByID, this does NOT normalize a
// not-found to ErrUserNotFound; it propagates the repository error verbatim so
// the eventual policy decision can be applied at a single, well-marked place.
// REQUIRES MANUAL REVIEW.
func (s *userService) GetUserByName(ctx context.Context, name string) (model.User, error) {
	user, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return model.User{}, fmt.Errorf("get user by name %q: %w", name, err)
	}
	return user, nil
}
