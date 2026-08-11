// Package service defines the business-logic contracts for the smartcontact
// application. It sits between the HTTP handlers and the repository layer.
//
// MIGRATION_NOTE: The Java source was a Spring service interface (UserService)
// with six methods, typically implemented by an @Service class and injected
// into controllers. In Go we keep the interface as the abstraction seam so
// handlers depend on behavior, not a concrete type. The interface lives in its
// own file; the concrete implementation (the Java @Service class) is migrated
// separately and constructed via NewUserService.
//
// Idiomatic Go changes from the source:
//   - Every method that performs I/O takes a context.Context as its first
//     parameter for cancellation/deadline propagation.
//   - Methods return (T, error) instead of throwing checked exceptions.
//     Java's `throws UserNotFoundException` becomes an error return that
//     callers inspect with errors.As against apperr.UserNotFound.
//   - Java's void methods (deleteUser, updateUser) return a bare error.
//   - Java's `int id` primary key is mapped to int64 to match the repository
//     layer and PostgreSQL identity columns.
package service

import (
	"context"

	"github.com/smartContact/internal/smartcontact/model"
)

// UserService describes the user-related business operations exposed to the
// HTTP layer. Implementations coordinate the repository and translate
// persistence errors into domain errors (e.g. apperr.UserNotFound).
type UserService interface {
	// SaveUser persists a new user and returns the stored representation
	// (including its database-assigned ID). Mirrors Java saveUser(User).
	SaveUser(ctx context.Context, user *model.User) (*model.User, error)

	// FetchUserList returns all users. Mirrors Java fetchUserList().
	FetchUserList(ctx context.Context) ([]*model.User, error)

	// FetchUserByID looks up a user by its primary key. It returns an error
	// satisfying errors.As(&apperr.UserNotFound) when no user exists with the
	// given id. Mirrors Java fetchUserById(int) throws UserNotFoundException.
	FetchUserByID(ctx context.Context, id int64) (*model.User, error)

	// DeleteUser removes the user with the given id. Mirrors Java
	// deleteUser(int) (void). Returns an error if the delete fails.
	DeleteUser(ctx context.Context, id int64) error

	// UpdateUser applies the fields of user to the record identified by id.
	// Mirrors Java updateUser(int, User) (void). Returns an error if the
	// target user does not exist or the update fails.
	UpdateUser(ctx context.Context, id int64, user *model.User) error

	// GetUserByName looks up a user by name. Mirrors Java
	// getUserNameByName(String).
	//
	// MIGRATION_NOTE: the source method was named getUserNameByName but
	// returns a full User; the Go name reflects its actual behavior.
	GetUserByName(ctx context.Context, name string) (*model.User, error)
}
