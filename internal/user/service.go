package user

import "context"

// Service is the application-layer contract for User operations
// (source: com.smartContact.service.UserService). HTTP handlers depend on this
// interface rather than on a concrete implementation, replacing Spring's
// interface-based @Autowired injection with explicit constructor injection.
//
// Every method takes a context.Context as its first parameter so request
// deadlines and cancellation reach the persistence layer.
//
// MIGRATION_NOTE: the source declared only one checked exception
// (UserNotFoundException on fetchUserById). In Go every method returns an
// error; not-found conditions are reported with errors that match
// ErrNotFound via errors.Is (or *NotFoundError via errors.As), which the HTTP
// layer maps to 404 with the verbatim NotFoundMessage. Transactional
// boundaries (@Transactional on the implementation, if any) are the
// implementation's responsibility, not part of this contract.
//
// MIGRATION_NOTE: the concrete implementation (source: UserServiceImp) is
// migrated in its own file in this package; this file intentionally declares
// only the contract to avoid duplicate declarations.
type Service interface {
	// SaveUser persists u and returns the saved instance, including any
	// generated fields such as the id (source: saveUser).
	SaveUser(ctx context.Context, u *User) (*User, error)

	// FetchUserList returns all users currently stored. The returned slice
	// is never nil, so it encodes as [] rather than null (source:
	// fetchUserList).
	FetchUserList(ctx context.Context) ([]User, error)

	// FetchUserByID returns the user with the given id. When no such user
	// exists the error matches ErrNotFound and its message is
	// NotFoundMessage (source: fetchUserById throwing UserNotFoundException).
	FetchUserByID(ctx context.Context, id int) (*User, error)

	// DeleteUser removes the user with the given id (source: deleteUser).
	// When no such user exists the error matches ErrEmptyResult, mirroring
	// Spring Data's EmptyResultDataAccessException from deleteById.
	DeleteUser(ctx context.Context, id int) error

	// UpdateUser stores the field values of u under the given id, overwriting
	// the existing row (source: updateUser, which set u.id = id and called
	// save, i.e. Hibernate merge semantics). Implementations set u.ID to id
	// before persisting, as the source mutated the passed entity.
	UpdateUser(ctx context.Context, id int, u *User) error

	// GetUserByName looks up the single user whose name equals name, typically
	// to resolve the authenticated principal (source: getUserNameByName). The
	// bool is false when no user matches (the source returned null). When
	// several users share the name the error matches ErrIncorrectResultSize.
	GetUserByName(ctx context.Context, name string) (*User, bool, error)
}
