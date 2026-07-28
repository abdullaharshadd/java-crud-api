// Package service defines the service-layer contracts and business logic for
// the smartContact application. It sits between the HTTP handler layer and the
// data-access (repository) layer, decoupling controllers from concrete
// business-logic implementations.
//
// MIGRATION_NOTE: The Java source was a Spring service interface (UserService)
// consumed via dependency injection. In Go the idiomatic equivalent is a plain
// interface declared in the package that consumes it; the concrete
// implementation is provided by a constructor and injected explicitly. This
// file declares only the interface contract — the concrete implementation
// (formerly UserServiceImp) is migrated separately and must satisfy this
// interface.
//
// MIGRATION_NOTE: Every method that performs I/O (all of them, since they hit
// the repository/database) takes a context.Context as its first parameter for
// cancellation and deadline propagation — something the Java signatures had no
// equivalent for.
//
// MIGRATION_NOTE: Per the agent debate, all user-returning methods now return
// the shared wire shape model.UserResponse rather than the internal write shape
// model.User. The single write-shape input (saveUser/updateUser) still accepts
// model.User.
//
// MIGRATION_NOTE: The Java `throws UserNotFoundException` checked exception is
// replaced by Go's standard (T, error) return convention. Callers inspect the
// error with errors.Is(err, smartcontacterror.ErrUserNotFound) or errors.As.
// Methods that returned void now return error so failures are observable.
package service

import (
	"context"

	"migrated-app/internal/smartcontact/model"
)

// UserService defines the business-logic contract for User CRUD operations.
// Implementations coordinate the repository layer and translate persistence
// results into client-facing responses.
//
// MIGRATION_NOTE: Java's `int id` primary key maps to Go's int here to preserve
// the source contract. If the underlying identity column can exceed int32 range
// on 32-bit platforms this should be revisited (int64 recommended).
type UserService interface {
	// SaveUser persists a new user and returns the created user as a response.
	// It returns a non-nil error if persistence fails.
	SaveUser(ctx context.Context, user model.User) (model.UserResponse, error)

	// FetchUserList returns all users as response shapes. It returns a non-nil
	// error if the underlying query fails. An empty result is not an error.
	FetchUserList(ctx context.Context) ([]model.UserResponse, error)

	// FetchUserByID returns the user with the given id. If no such user exists
	// it returns an error satisfying errors.Is(err, smartcontacterror.ErrUserNotFound).
	FetchUserByID(ctx context.Context, id int) (model.UserResponse, error)

	// DeleteUser removes the user with the given id. It returns a non-nil error
	// if the deletion fails.
	//
	// MIGRATION_NOTE: The Java method returned void; it now returns error so
	// callers can react to failures rather than silently ignoring them.
	DeleteUser(ctx context.Context, id int) error

	// UpdateUser updates the user identified by id with the supplied fields.
	// It returns an error satisfying errors.Is(err, smartcontacterror.ErrUserNotFound)
	// if the user does not exist, or another error if the update fails.
	//
	// MIGRATION_NOTE: The Java method returned void; it now returns error.
	UpdateUser(ctx context.Context, id int, user model.User) error

	// GetUserByName returns the user matching the given name.
	//
	// MIGRATION_NOTE: renamed from the Java getUserNameByName, which returned a
	// User (not a name) despite its name — the Go name reflects the actual
	// behaviour. Returns an error satisfying errors.Is(err,
	// smartcontacterror.ErrUserNotFound) when no user matches.
	GetUserByName(ctx context.Context, name string) (model.UserResponse, error)
}
