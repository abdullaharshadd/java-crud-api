package service

import (
	"context"

	"migrated-app/internal/model"
)

// UserService is the service-layer contract for User CRUD operations and
// lookup by name. It replaces the source's Spring UserService interface; the
// concrete implementation (counterpart of UserServiceImp) is wired explicitly
// by constructor in cmd/server rather than by @Autowired injection.
//
// User ids are int32 to match model.User.ID and the store layer (Java int).
type UserService interface {
	// SaveUser persists u with merge semantics and returns the saved copy,
	// including the store-generated identifier.
	SaveUser(ctx context.Context, u *model.User) (*model.User, error)

	// FetchUserList returns every stored user. The result is never nil.
	FetchUserList(ctx context.Context) ([]model.User, error)

	// FetchUserByID returns the user with the given id. When no such user
	// exists the returned error matches ErrUserNotFound under errors.Is
	// (a *UserNotFoundError carrying UserNotFoundMessage), which the HTTP
	// layer maps to 404.
	FetchUserByID(ctx context.Context, id int32) (model.User, error)

	// DeleteUser removes the user with the given id. Deleting a missing id
	// yields a non-not-found error (Spring Data 2.x deleteById behaviour,
	// surfaced as 500).
	DeleteUser(ctx context.Context, id int32) error

	// UpdateUser overwrites the stored user identified by id with the values
	// in u (u's own id is replaced by id; nil fields are written as NULL).
	// If no user with that id exists, a new one is created, as JPA save does.
	UpdateUser(ctx context.Context, id int32, u *model.User) error

	// GetUserNameByName looks up the single user whose name equals name.
	// ok is false when no user matches (Java returned null); more than one
	// match is reported as an error.
	GetUserNameByName(ctx context.Context, name string) (u model.User, ok bool, err error)
}
