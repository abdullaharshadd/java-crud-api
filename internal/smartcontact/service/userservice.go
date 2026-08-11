// Package service provides the business-logic layer for the Smart Contact
// service. It is the Go equivalent of the source project's
// com.smartContact.service package.
//
// MIGRATION_NOTE: The Java source was a Spring service-layer interface
// (UserService) declaring CRUD operations over the User entity. A concrete
// @Service implementation (elsewhere in the source project) was injected into
// controllers and delegated to the Spring Data JPA UserDao repository.
//
// In idiomatic Go we keep the interface as a small abstraction, but we also
// provide a concrete implementation in this same file because the source
// interface has no additional behavior beyond delegating to the repository —
// there is no separate Java "UserServiceImp" file to migrate. Every method
// takes a context.Context (for cancellation/deadline propagation through the
// I/O in the repository layer) and returns an explicit error, replacing Java's
// checked-exception propagation (UserNotFoundException).
//
// Key translation decisions:
//
//   - saveUser(User)            -> SaveUser(ctx, User) (*model.User, error)
//   - fetchUserList()           -> FetchUserList(ctx) ([]model.User, error)
//   - fetchUserById(int) throws -> FetchUserByID(ctx, int) (*model.User, error)
//                                  (returns error.ErrUserNotFound when absent)
//   - deleteUser(int)           -> DeleteUser(ctx, int) error
//   - updateUser(int, User)     -> UpdateUser(ctx, int, User) error (void in
//                                  Java; informs the handler's empty-200-body
//                                  default)
//   - getUserNameByName(String) -> GetUserByName(ctx, string) (*model.User,
//                                  error)
package service

import (
	"context"
	"fmt"

	smartError "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/repository"
)

// UserService defines the service-layer contract for user management
// operations. It mirrors the source project's com.smartContact.service
// UserService interface. All I/O-bound methods accept a context.Context as
// their first parameter and return an explicit error.
type UserService interface {
	// SaveUser persists the supplied user and returns the stored record,
	// including any database-generated identifier.
	SaveUser(ctx context.Context, user model.User) (*model.User, error)

	// FetchUserList returns every user in the system.
	FetchUserList(ctx context.Context) ([]model.User, error)

	// FetchUserByID returns the user with the given id. It returns an error
	// wrapping error.ErrUserNotFound when no such user exists.
	FetchUserByID(ctx context.Context, id int) (*model.User, error)

	// DeleteUser removes the user with the given id.
	DeleteUser(ctx context.Context, id int) error

	// UpdateUser applies the supplied user data to the record identified by
	// id. It returns an error wrapping error.ErrUserNotFound when no such
	// user exists. The Java source returned void, which informs the HTTP
	// handler's empty-200-body default.
	UpdateUser(ctx context.Context, id int, user model.User) error

	// GetUserByName returns the user matching the given name. It returns an
	// error wrapping error.ErrUserNotFound when no such user exists.
	GetUserByName(ctx context.Context, name string) (*model.User, error)
}

// userService is the default UserService implementation. It delegates
// persistence to a repository.UserRepository, replacing the Spring @Service
// bean that was injected with a UserDao.
type userService struct {
	repo repository.UserRepository
}

// NewUserService constructs a UserService backed by the given
// repository.UserRepository. It returns an error if repo is nil so callers
// fail fast during wiring rather than panicking at request time.
func NewUserService(repo repository.UserRepository) (UserService, error) {
	if repo == nil {
		return nil, fmt.Errorf("service: user repository must not be nil")
	}
	return &userService{repo: repo}, nil
}

// SaveUser validates and persists the supplied user, returning the stored
// record.
func (s *userService) SaveUser(ctx context.Context, user model.User) (*model.User, error) {
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("service: save user: %w", err)
	}
	saved, err := s.repo.Save(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("service: save user: %w", err)
	}
	return saved, nil
}

// FetchUserList returns every user in the system.
func (s *userService) FetchUserList(ctx context.Context) ([]model.User, error) {
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: fetch user list: %w", err)
	}
	return users, nil
}

// FetchUserByID returns the user with the given id, or an error wrapping
// error.ErrUserNotFound when it does not exist.
func (s *userService) FetchUserByID(ctx context.Context, id int) (*model.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service: fetch user by id %d: %w", id, err)
	}
	return user, nil
}

// DeleteUser removes the user with the given id.
func (s *userService) DeleteUser(ctx context.Context, id int) error {
	if err := s.repo.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("service: delete user %d: %w", id, err)
	}
	return nil
}

// UpdateUser applies the supplied user data to the record identified by id.
// It returns an error wrapping error.ErrUserNotFound when the target user does
// not exist. The Java method returned void; here we surface persistence and
// not-found errors explicitly.
func (s *userService) UpdateUser(ctx context.Context, id int, user model.User) error {
	// Confirm the record exists first so callers receive a clear
	// not-found signal (Java relied on the repository throwing).
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

	// Preserve the target identifier regardless of the request body.
	user.ID = id
	if _, err := s.repo.Save(ctx, user); err != nil {
		return fmt.Errorf("service: update user %d: %w", id, err)
	}
	return nil
}

// GetUserByName returns the user matching the given name, or an error wrapping
// error.ErrUserNotFound when none exists.
func (s *userService) GetUserByName(ctx context.Context, name string) (*model.User, error) {
	user, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("service: get user by name %q: %w", name, err)
	}
	return user, nil
}
