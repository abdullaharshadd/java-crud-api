package service

// Package service defines the service-layer contracts for the SmartContact
// application. This file corresponds to the original Spring Java interface
// com.smartContact.service.UserService.
//
// MIGRATION_NOTE: The Java source was a plain interface whose concrete
// implementation (typically a @Service-annotated UserServiceImp) lived in a
// separate file and was wired by Spring's DI container. In idiomatic Go we keep
// the interface here to enable dependency inversion, and the concrete
// implementation is provided via a constructor (NewUserService) that accepts
// its collaborators explicitly. Java's checked exception (UserNotFoundException)
// becomes an ordinary returned error (errors.ErrUserNotFound /
// *errors.UserNotFoundError), and every I/O-bound method takes a
// context.Context as its first parameter for cancellation/deadline propagation.

import (
	"context"
	"fmt"

	smarterr "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/repository"
)

// UserService defines the user-management operations exposed by the service
// layer: creating, listing, fetching, updating, and deleting users.
//
// It mirrors the surface of the original Java UserService interface. All
// methods that perform I/O accept a context.Context and return an error so
// callers can handle failures (including a not-found condition) explicitly.
type UserService interface {
	// SaveUser persists the given user and returns the stored representation
	// (including any server-assigned identifier).
	SaveUser(ctx context.Context, user model.User) (model.User, error)

	// FetchUserList returns all users.
	FetchUserList(ctx context.Context) ([]model.User, error)

	// FetchUserByID returns the user with the given id. It returns a
	// *smarterr.UserNotFoundError (wrapping smarterr.ErrUserNotFound) when no
	// user exists with that id.
	FetchUserByID(ctx context.Context, id int) (model.User, error)

	// DeleteUser removes the user with the given id.
	DeleteUser(ctx context.Context, id int) error

	// UpdateUser updates the user with the given id using the supplied values.
	// It returns a *smarterr.UserNotFoundError when no such user exists.
	UpdateUser(ctx context.Context, id int, user model.User) error

	// GetUserByName returns the user with the given name. It returns a
	// *smarterr.UserNotFoundError when no user exists with that name.
	GetUserByName(ctx context.Context, name string) (model.User, error)
}

// userService is the default UserService implementation. It delegates
// persistence to a repository.UserDao.
//
// MIGRATION_NOTE: The Java interface had no method body; the behavior lived in
// a separate @Service implementation. Because the interface half of this pair
// carries no logic of its own, this file also provides a concrete, DI-friendly
// implementation so the migrated package is directly usable. Adjust the
// delegation below if the original UserServiceImp contained additional
// business rules (validation, mapping, transactions) that require manual
// porting.
type userService struct {
	dao repository.UserDao
}

// NewUserService constructs a UserService backed by the given UserDao.
func NewUserService(dao repository.UserDao) UserService {
	return &userService{dao: dao}
}

// SaveUser persists the given user and returns the stored representation.
func (s *userService) SaveUser(ctx context.Context, user model.User) (model.User, error) {
	saved, err := s.dao.Save(ctx, user)
	if err != nil {
		return model.User{}, fmt.Errorf("save user: %w", err)
	}
	return saved, nil
}

// FetchUserList returns all users.
func (s *userService) FetchUserList(ctx context.Context) ([]model.User, error) {
	users, err := s.dao.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch user list: %w", err)
	}
	return users, nil
}

// FetchUserByID returns the user with the given id, or a
// *smarterr.UserNotFoundError when it does not exist.
func (s *userService) FetchUserByID(ctx context.Context, id int) (model.User, error) {
	user, err := s.dao.FindByID(ctx, id)
	if err != nil {
		return model.User{}, fmt.Errorf("fetch user by id %d: %w", id, err)
	}
	return user, nil
}

// DeleteUser removes the user with the given id.
func (s *userService) DeleteUser(ctx context.Context, id int) error {
	if err := s.dao.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("delete user %d: %w", id, err)
	}
	return nil
}

// UpdateUser updates the user identified by id with the supplied values.
//
// MIGRATION_NOTE: The original Java signature returned void and relied on JPA's
// managed-entity dirty checking to flush changes. Here we make the update
// explicit: verify the record exists (surfacing a UserNotFoundError otherwise),
// stamp the identifier onto the incoming value, and persist via Save. Confirm
// this matches the semantics of the original UserServiceImp.updateUser.
func (s *userService) UpdateUser(ctx context.Context, id int, user model.User) error {
	if _, err := s.dao.FindByID(ctx, id); err != nil {
		return fmt.Errorf("update user %d: %w", id, err)
	}

	user.ID = id
	if _, err := s.dao.Save(ctx, user); err != nil {
		return fmt.Errorf("update user %d: %w", id, err)
	}
	return nil
}

// GetUserByName returns the user with the given name, or a
// *smarterr.UserNotFoundError when no such user exists.
func (s *userService) GetUserByName(ctx context.Context, name string) (model.User, error) {
	user, err := s.dao.FindByName(ctx, name)
	if err != nil {
		return model.User{}, fmt.Errorf("get user by name %q: %w", name, err)
	}
	return user, nil
}

// Ensure the concrete type satisfies the interface at compile time.
var _ UserService = (*userService)(nil)

// Silence unused import if the error package is not referenced directly in a
// future refactor; it documents the not-found contract in the method godocs.
var _ = smarterr.ErrUserNotFound
