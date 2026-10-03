// Package service contains the business-logic layer for User CRUD operations.
//
// This file holds both the UserService contract (the source's Spring
// UserService interface) and its concrete implementation (the counterpart of
// the source's UserServiceImp). In Go there is no separate interface/impl
// file split, so the two live together. The implementation is wired
// explicitly via NewUserService in cmd/server instead of @Service component
// scanning and @Autowired field injection.
package service

import (
	"context"
	"errors"
	"fmt"

	"migrated-app/internal/model"
)

// userNotAvailableMessage is the exact detail message the source passes to
// UserNotFoundException when a lookup by id misses. It is surfaced verbatim
// in the HTTP error body by the centralized error mapping in httpapi.
const userNotAvailableMessage = "User are not available"

// UserService is the service-layer contract for User CRUD operations and
// lookup by name. It replaces the source's Spring UserService interface.
//
// User ids are int32 to match model.User.ID and the store layer (Java int).
type UserService interface {
	// SaveUser persists u with merge semantics and returns the saved copy,
	// including the store-generated identifier.
	SaveUser(ctx context.Context, u *model.User) (*model.User, error)

	// FetchUserList returns every stored user. The result is never nil.
	FetchUserList(ctx context.Context) ([]model.User, error)

	// FetchUserByID returns the user with the given id, or a
	// *UserNotFoundError (detail "User are not available") when absent.
	FetchUserByID(ctx context.Context, id int32) (*model.User, error)

	// DeleteUser removes the user with the given id. A missing id is reported
	// by the repository (Spring's EmptyResultDataAccessException analogue),
	// not as a not-found error.
	DeleteUser(ctx context.Context, id int32) error

	// UpdateUser sets id on u and saves it, upserting the user at that id.
	// It returns u (the request object with the path id applied), not the
	// entity returned by the repository.
	UpdateUser(ctx context.Context, id int32, u *model.User) (*model.User, error)

	// GetUserNameByName returns the user whose name equals name. When no user
	// matches it returns (nil, nil), mirroring the source returning null.
	GetUserNameByName(ctx context.Context, name string) (*model.User, error)
}

// UserRepository is the persistence contract the service depends on. It is
// satisfied by the database/sql-backed repository in internal/store and by
// in-memory fakes in tests. It replaces the source's Spring Data JPA UserDao.
type UserRepository interface {
	// Save persists u with JPA merge semantics (insert when the id is new,
	// update otherwise) and returns the stored entity.
	Save(ctx context.Context, u *model.User) (*model.User, error)

	// FindAll returns all stored users.
	FindAll(ctx context.Context) ([]model.User, error)

	// FindByID returns the user with the given id and true, or (nil, false,
	// nil) when no such user exists.
	FindByID(ctx context.Context, id int32) (*model.User, bool, error)

	// DeleteByID deletes the user with the given id.
	DeleteByID(ctx context.Context, id int32) error

	// FindByName returns the unique user with the given name, or (nil, nil)
	// when none matches.
	FindByName(ctx context.Context, name string) (*model.User, error)
}

// userService is the default UserService implementation (UserServiceImp).
type userService struct {
	repo UserRepository
}

// NewUserService constructs a UserService backed by repo. It replaces Spring's
// @Service registration plus @Autowired injection of UserDao.
func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

// SaveUser persists u and returns the entity the repository returns.
func (s *userService) SaveUser(ctx context.Context, u *model.User) (*model.User, error) {
	if u == nil {
		// MIGRATION_NOTE: Spring Data's save(null) throws
		// IllegalArgumentException (surfacing as HTTP 500); we return an error.
		return nil, errors.New("service: save user: user must not be nil")
	}
	saved, err := s.repo.Save(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("service: save user: %w", err)
	}
	return saved, nil
}

// FetchUserList returns all users stored in the repository. The returned
// slice is never nil so it encodes as [] rather than null.
func (s *userService) FetchUserList(ctx context.Context) ([]model.User, error) {
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: fetch user list: %w", err)
	}
	if users == nil {
		users = []model.User{}
	}
	return users, nil
}

// FetchUserByID looks up a user by id, returning a *UserNotFoundError with
// detail "User are not available" when the user is absent.
func (s *userService) FetchUserByID(ctx context.Context, id int32) (*model.User, error) {
	u, ok, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service: fetch user %d: %w", id, err)
	}
	if !ok || u == nil {
		return nil, NewUserNotFoundError(userNotAvailableMessage)
	}
	return u, nil
}

// DeleteUser deletes the user with the given id via the repository.
func (s *userService) DeleteUser(ctx context.Context, id int32) error {
	if err := s.repo.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("service: delete user %d: %w", id, err)
	}
	return nil
}

// UpdateUser sets id on u and saves it, effectively upserting the user at
// that id.
//
// MIGRATION_NOTE: the source's updateUser is void and discards the saved
// entity. We preserve that quirk by returning the request object u with the
// path id applied, not the repository's returned entity. No validation is
// performed here, matching the source.
func (s *userService) UpdateUser(ctx context.Context, id int32, u *model.User) (*model.User, error) {
	if u == nil {
		// MIGRATION_NOTE: the source would NPE on user.setId(id) (HTTP 500).
		return nil, errors.New("service: update user: user must not be nil")
	}
	u.ID = id
	if _, err := s.repo.Save(ctx, u); err != nil {
		return nil, fmt.Errorf("service: update user %d: %w", id, err)
	}
	return u, nil
}

// GetUserNameByName returns the user whose name matches name, delegating to
// the repository's FindByName. A miss yields (nil, nil), like the source's
// null return; multiple matches surface the repository's error.
func (s *userService) GetUserNameByName(ctx context.Context, name string) (*model.User, error) {
	u, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("service: find user by name %q: %w", name, err)
	}
	return u, nil
}
