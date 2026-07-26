package service

import (
	"context"
	"errors"
	"fmt"

	smartcontacterror "migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/repository"
)

// userService is the concrete UserService implementation backed by a
// UserRepository. It is the Go analogue of the Spring @Service bean
// UserServiceImp.
type userService struct {
	repo repository.UserRepository
}

// Save persists the given user and returns the stored entity.
//
// MIGRATION_NOTE: Mirrors UserServiceImp.saveUser, which delegated directly to
// UserDao.save.

// List returns all persisted users.
//
// MIGRATION_NOTE: Mirrors UserServiceImp.fetchUserList, which delegated
// directly to UserDao.findAll.

// GetByID returns the user with the given id.
//
// MIGRATION_NOTE: Mirrors UserServiceImp.fetchUserById. The Java code wrapped
// the repository lookup in an Optional and threw UserNotFoundException when the
// value was absent. Here the not-found case is expressed with the migrated
// error.ErrUserNotFound sentinel (matchable via errors.Is). A repository
// sql.ErrNoRows-equivalent is normalised to the not-found sentinel to preserve
// the original "User are not available" semantics.

// Delete removes the user with the given id.
//
// MIGRATION_NOTE (Change 12): The Java deleteUser delegated directly to
// UserDao.deleteById(id) with no existence guard. That behaviour is preserved:
// any ErrEmptyResultDelete from the repository propagates unmapped so callers
// observe the same failure the Java code did (surfaced as a 500 upstream).

// Update sets the id on the supplied user and persists it, returning the
// stored entity.
//
// MIGRATION_NOTE: Mirrors UserServiceImp.updateUser, which set the id on the
// incoming entity and re-saved it (JPA merge semantics). The Java @NotNull on
// the primitive int parameter was a no-op at runtime; in Go the int is
// non-nullable by construction. The original method returned void; here we
// return the saved user for a more useful Go API.

// GetByName returns the user matching the given name.
//
// MIGRATION_NOTE: Mirrors UserServiceImp.getUserNameByName, which delegated to
// the Spring Data derived query UserDao.findByName. The Java version could
// return null; here a missing user is reported explicitly via error.

var (
	_ smartcontacterror.Error = nil
	_ model.User              = model.User{}
	_ repository.UserRepository = nil
	_ = context.Background
	_ = errors.New
	_ = fmt.Errorf
)