// Package service defines the service-layer contracts for the SmartContact
// application.
//
// MIGRATION_NOTE: The original Java type was `com.smartContact.service.UserService`,
// a Spring `@Service`-style interface with no implementation. Spring wired a
// concrete implementation into consumers via its DI container based on this
// abstraction.
//
// Go has no annotation-driven DI container. The idiomatic replacement is:
//   - A plain Go interface (UserService) that declares the service contract.
//   - A concrete implementation (constructed via a NewXxx function) that is
//     wired explicitly at composition time (e.g. in main or a wire provider).
//
// Additional idiomatic changes applied:
//   - Every I/O-bound method takes a context.Context as its first parameter for
//     cancellation and deadline propagation.
//   - Every method returns an explicit error rather than throwing. The Java
//     `throws UserNotFoundException` becomes a returned error that callers can
//     inspect with errors.Is(err, smartcontacterror.ErrUserNotFound).
//   - The Java primitive `int` id becomes a Go `int`; `String name` becomes
//     `string`.
//
// Business-logic asymmetry preserved from the source:
//   - FetchUserByID converts a nil/absent user into a UserNotFoundError.
//   - GetUserNameByName passes a nil/absent result through WITHOUT converting it
//     into a not-found error (it returns (nil, nil) in that case). This mirrors
//     the original Java, where fetchUserById threw UserNotFoundException but
//     getUserNameByName did not. Implementations MUST preserve this asymmetry.
package service

import (
	"context"

	"github.com/smartContact/internal/smartcontact/model"
)

// UserService defines the service-layer operations for managing users.
//
// It is the Go equivalent of the Spring `UserService` interface. A concrete
// implementation is expected to be constructed (via a NewXxx constructor) and
// wired into HTTP handlers or other consumers at application startup.
type UserService interface {
	// SaveUser persists the given user and returns the stored representation
	// (which may include a server-assigned ID). It returns a non-nil error if
	// the user cannot be saved.
	SaveUser(ctx context.Context, user *model.User) (*model.User, error)

	// FetchUserList returns all users. An empty slice (not nil) with a nil error
	// indicates that there are simply no users. A non-nil error indicates a
	// retrieval failure.
	FetchUserList(ctx context.Context) ([]*model.User, error)

	// FetchUserByID looks up a single user by its ID.
	//
	// MIGRATION_NOTE: The Java method declared `throws UserNotFoundException`.
	// Implementations MUST convert an absent/nil user into an error that
	// satisfies errors.Is(err, smartcontacterror.ErrUserNotFound) rather than
	// returning (nil, nil).
	FetchUserByID(ctx context.Context, id int) (*model.User, error)

	// DeleteUser removes the user with the given ID. It returns a non-nil error
	// if deletion fails.
	DeleteUser(ctx context.Context, id int) error

	// UpdateUser updates the user identified by id with the provided user data.
	// It returns a non-nil error if the update fails.
	UpdateUser(ctx context.Context, id int, user *model.User) error

	// GetUserNameByName looks up a user by name.
	//
	// MIGRATION_NOTE: Unlike FetchUserByID, the original Java method did NOT
	// throw UserNotFoundException. Implementations MUST preserve this asymmetry:
	// when no matching user exists, return (nil, nil) rather than a not-found
	// error. A non-nil error is reserved for genuine retrieval failures.
	GetUserNameByName(ctx context.Context, name string) (*model.User, error)
}
