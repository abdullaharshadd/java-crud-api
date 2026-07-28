// Package service defines the business-logic contracts for the smartcontact
// application. Services sit between the HTTP handlers and the repository layer,
// orchestrating validation and persistence for domain models.
//
// MIGRATION_NOTE: The Java source was a Spring service-layer interface
// (UserService) whose implementation (@Service UserServiceImpl) was expected
// elsewhere. In Go the idiomatic split is a small interface declared here for
// dependency injection into handlers, plus a concrete constructor-backed
// implementation that wraps the repository. Both live in the same package.
//
// Notable signature changes from the Java contract:
//   - Every I/O-bound method now takes a context.Context as its first argument
//     for cancellation/deadline propagation.
//   - Every method returns an explicit error rather than relying on unchecked
//     runtime exceptions or the single checked UserNotFoundException.
//   - getUserNameByName (a lookup by name) becomes GetUserByName and returns
//     (User, bool, error): the bool threads the found/not-found result up to
//     the handler, replacing Java's habit of returning null.
//   - fetchUserById maps to GetUserByID and surfaces a not-found condition as
//     an apperror.ErrUserNotFound the handler can inspect via errors.Is.
package service

import (
	"context"
	"errors"
	"fmt"

	apperror "github.com/smartcontact/internal/smartcontact/error"
	"github.com/smartcontact/internal/smartcontact/model"
	"github.com/smartcontact/internal/smartcontact/repository"
)

// UserService is the service-layer contract for user management operations.
// It is intended to be injected into HTTP handlers, allowing them to depend on
// this abstraction rather than a concrete implementation (which eases testing
// with mocks/fakes).
type UserService interface {
	// SaveUser validates and persists a new user, returning the stored user
	// (including any database-generated fields such as the ID).
	SaveUser(ctx context.Context, user model.User) (model.User, error)

	// FetchUserList returns all users. An empty slice (not an error) is
	// returned when there are none.
	FetchUserList(ctx context.Context) ([]model.User, error)

	// GetUserByID returns the user with the given id. It returns an error that
	// satisfies apperror.IsUserNotFound when no such user exists.
	GetUserByID(ctx context.Context, id int) (model.User, error)

	// DeleteUser removes the user with the given id. It returns an error that
	// satisfies apperror.IsUserNotFound when no such user exists.
	DeleteUser(ctx context.Context, id int) error

	// UpdateUser applies the fields of user to the existing record identified
	// by id. It returns the updated user, or an error satisfying
	// apperror.IsUserNotFound when no such user exists.
	UpdateUser(ctx context.Context, id int, user model.User) (model.User, error)

	// GetUserByName looks up a user by name. The boolean reports whether a
	// matching user was found; error is non-nil only for genuine failures
	// (a missing user is a normal (zero, false, nil) result).
	GetUserByName(ctx context.Context, name string) (model.User, bool, error)
}

// userService is the default UserService implementation. It delegates
// persistence to a repository and performs domain validation before writes.
//
// MIGRATION_NOTE: The repository is held behind the repository.Querier
// interface (rather than the concrete *repository.UserRepository) so the
// service can be unit-tested with a fake repository.
type userService struct {
	repo repository.Querier
}

// NewUserService constructs a UserService backed by the given repository.
func NewUserService(repo repository.Querier) UserService {
	return &userService{repo: repo}
}

// SaveUser validates the incoming user and persists it.
func (s *userService) SaveUser(ctx context.Context, user model.User) (model.User, error) {
	if err := user.Validate(); err != nil {
		return model.User{}, fmt.Errorf("save user: %w", err)
	}

	saved, err := s.repo.Save(ctx, user)
	if err != nil {
		return model.User{}, fmt.Errorf("save user: %w", err)
	}
	return saved, nil
}

// FetchUserList returns every user in the store.
func (s *userService) FetchUserList(ctx context.Context) ([]model.User, error) {
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch user list: %w", err)
	}
	return users, nil
}

// GetUserByID returns the user with the given id.
func (s *userService) GetUserByID(ctx context.Context, id int) (model.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if apperror.IsUserNotFound(err) {
			return model.User{}, err
		}
		return model.User{}, fmt.Errorf("get user by id %d: %w", id, err)
	}
	return user, nil
}

// DeleteUser removes the user with the given id, first confirming it exists so
// the caller receives a not-found error rather than a silent no-op.
func (s *userService) DeleteUser(ctx context.Context, id int) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		if apperror.IsUserNotFound(err) {
			return err
		}
		return fmt.Errorf("delete user %d: %w", id, err)
	}

	if err := s.repo.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("delete user %d: %w", id, err)
	}
	return nil
}

// UpdateUser applies user's fields to the record identified by id.
//
// MIGRATION_NOTE: The Java interface declared updateUser as void. Because the
// underlying repository exposes an upsert-style Save, this loads the existing
// record to confirm it exists (surfacing ErrUserNotFound), applies the mutable
// fields, re-validates, and persists — returning the updated user so handlers
// can echo it back without a second read.
func (s *userService) UpdateUser(ctx context.Context, id int, user model.User) (model.User, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if apperror.IsUserNotFound(err) {
			return model.User{}, err
		}
		return model.User{}, fmt.Errorf("update user %d: %w", id, err)
	}

	// Preserve the identity of the existing record; overlay the mutable fields.
	user.ID = existing.ID

	if err := user.Validate(); err != nil {
		return model.User{}, fmt.Errorf("update user %d: %w", id, err)
	}

	updated, err := s.repo.Save(ctx, user)
	if err != nil {
		return model.User{}, fmt.Errorf("update user %d: %w", id, err)
	}
	return updated, nil
}

// GetUserByName looks up a user by name, reporting found/not-found via the
// boolean rather than an error.
func (s *userService) GetUserByName(ctx context.Context, name string) (model.User, bool, error) {
	user, err := s.repo.FindByName(ctx, name)
	if err != nil {
		if apperror.IsUserNotFound(err) || errors.Is(err, apperror.ErrUserNotFound) {
			return model.User{}, false, nil
		}
		return model.User{}, false, fmt.Errorf("get user by name %q: %w", name, err)
	}
	return user, true, nil
}
