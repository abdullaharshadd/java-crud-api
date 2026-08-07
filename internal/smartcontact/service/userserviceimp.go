// Package service defines the business-logic (service) layer contracts and
// implementations for the smartcontact application.
//
// MIGRATION_NOTE: The Java source, UserServiceImp, was the @Service-annotated
// implementation of the UserService interface. In Go the interface
// (UserService) and its concrete constructor (NewUserService) already live in
// userservice.go. To avoid redeclaring the concrete type in the same package,
// this file supplies the method set on that existing type. Spring's field
// injection (@Autowired UserDao) is replaced by explicit constructor injection
// of a repository.UserRepository (handled in userservice.go).
package service

import (
	"context"
	"errors"
	"fmt"

	smartcontacterror "migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
)

// SaveUser persists a new user and returns the saved entity, including any
// database-generated identifier.
//
// MIGRATION_NOTE: The Java saveUser returned the repository's save() result,
// which carries the DB-generated id. We preserve that behaviour here so the
// caller (HTTP layer) can choose to return either a text message or the saved
// User as JSON once the source-of-truth response shape is confirmed.

// FetchUserList returns all users.

// FetchUserByID returns the user with the given id.
//
// MIGRATION_NOTE: The Java method threw a checked UserNotFoundException when the
// Optional returned by findById was empty. Here we translate a missing row into
// the sentinel ErrUserNotFound (wrapped with context) so callers can detect it
// with errors.Is and map it to a 404. The repository is expected to signal
// absence via ErrUserNotFound (or sql.ErrNoRows wrapped as such).

// DeleteUser removes the user with the given id.

// UpdateUser sets the id on the supplied user and persists it, returning the
// saved entity.
//
// MIGRATION_NOTE: The Java updateUser returned void and mutated the passed User
// by calling setId(id) before save(). We keep that assignment but return the
// saved User (and an error) so callers have the persisted state without an
// out-of-band mutation contract.

// GetUserByName returns the user matching the given name.
//
// MIGRATION_NOTE: The Java getUserNameByName delegated to the Spring Data
// derived query findByName and returned the User directly (possibly null). Here
// we surface any lookup error explicitly; a missing user is reported via
// ErrUserNotFound so the caller can distinguish absence from other failures.
