package user

import (
	"context"
	"errors"
	"fmt"
)

// Service is the application-layer contract for User operations
// (source: com.smartContact.service.UserService). HTTP handlers depend on this
// interface rather than on a concrete implementation, which replaces Spring's
// interface-based @Autowired injection with explicit constructor injection.
//
// Every method takes a context.Context as its first parameter so request
// deadlines and cancellation reach the persistence layer.
//
// MIGRATION_NOTE: the source declared one checked exception,
// UserNotFoundException on fetchUserById. In Go every method returns an
// error instead. Not-found conditions are reported with errors that match
// ErrNotFound via errors.Is, or *NotFoundError via errors.As. The HTTP layer
// maps these to 404 with the verbatim NotFoundMessage.
type Service interface {
	// SaveUser persists u (insert when new, merge when its ID exists) and
	// returns the entity as stored by the repository.
	SaveUser(ctx context.Context, u *User) (*User, error)

	// FetchUserList returns every stored user.
	FetchUserList(ctx context.Context) ([]User, error)

	// FetchUserByID returns the user with the given id. When no such user
	// exists it returns an error matching ErrNotFound.
	FetchUserByID(ctx context.Context, id int) (*User, error)

	// DeleteUser removes the user with the given id.
	DeleteUser(ctx context.Context, id int) error

	// UpdateUser sets id on u and saves it. This replaces the stored record
	// with that id, or creates one when none exists (JPA merge semantics).
	UpdateUser(ctx context.Context, id int, u *User) error

	// GetUserByName returns the full User whose name exactly matches name.
	GetUserByName(ctx context.Context, name string) (*User, error)
}

// service is the default Service implementation
// (source: com.smartContact.service.UserServiceImp). It is a thin
// pass-through over a Repository and has no validation of its own.
//
// MIGRATION_NOTE: the source had no @Transactional annotation, so each
// repository call runs in its own transaction. That behaviour is preserved:
// the service never opens a transaction spanning multiple repository calls.
type service struct {
	repo Repository
}

// Compile-time check that service satisfies Service.
var _ Service = (*service)(nil)

// NewService returns a Service backed by repo. It replaces the source's
// field injection (@Autowired private UserDao userDao) with explicit
// constructor injection.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// SaveUser persists u and returns the saved entity exactly as returned by the
// repository (source: saveUser).
func (s *service) SaveUser(ctx context.Context, u *User) (*User, error) {
	saved, err := s.repo.Save(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("save user: %w", err)
	}
	return saved, nil
}

// FetchUserList returns all users stored in the repository
// (source: fetchUserList).
func (s *service) FetchUserList(ctx context.Context) ([]User, error) {
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch user list: %w", err)
	}
	return users, nil
}

// FetchUserByID retrieves a user by id (source: fetchUserById). An absent user
// yields a *NotFoundError carrying the verbatim message "User are not
// available" (NotFoundMessage), which matches ErrNotFound via errors.Is.
//
// MIGRATION_NOTE: Optional<User> plus isPresent() becomes the repository
// reporting ErrNotFound. A (nil, nil) result is also defensively treated as
// absent.
func (s *service) FetchUserByID(ctx context.Context, id int) (*User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, NewNotFoundErrorMsg(NotFoundMessage)
		}
		return nil, fmt.Errorf("fetch user %d: %w", id, err)
	}
	if u == nil {
		return nil, NewNotFoundErrorMsg(NotFoundMessage)
	}
	return u, nil
}

// DeleteUser deletes the user with the given id through the repository's
// DeleteByID (source: deleteUser).
//
// MIGRATION_NOTE: Spring Data 2.7's deleteById throws
// EmptyResultDataAccessException for a missing id. The repository's error
// (e.g. ErrEmptyResult) is propagated unchanged, wrapped with context, so the
// HTTP layer can map it the same way the source did.
func (s *service) DeleteUser(ctx context.Context, id int) error {
	if err := s.repo.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("delete user %d: %w", id, err)
	}
	return nil
}

// UpdateUser sets id on u and saves it, replacing the stored record with that
// id or creating one (source: updateUser).
//
// MIGRATION_NOTE: the source mutated the caller's User (user.setId(id)). This
// version mutates u in place the same way. The saved entity is discarded, as
// in the source, which returned void.
func (s *service) UpdateUser(ctx context.Context, id int, u *User) error {
	if u == nil {
		return errors.New("update user: nil user")
	}
	u.ID = id
	if _, err := s.repo.Save(ctx, u); err != nil {
		return fmt.Errorf("update user %d: %w", id, err)
	}
	return nil
}

// GetUserByName looks up a single user by exact name and returns the full
// User as returned by the repository (source: getUserNameByName).
//
// MIGRATION_NOTE: renamed from getUserNameByName to GetUserByName because it
// returns the whole User, not just the name. In the source, findByName
// returned null when no user matched. Here the repository's not-found or
// result-size error is propagated, wrapped with context. Callers that relied
// on a null return must check errors.Is(err, ErrNotFound) instead.
func (s *service) GetUserByName(ctx context.Context, name string) (*User, error) {
	u, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get user by name %q: %w", name, err)
	}
	return u, nil
}