package user

import (
	"context"
	"errors"
	"fmt"
)

// ServiceImp is the default implementation of the User application service
// (source: com.smartContact.service.UserServiceImp). It is a thin pass-through
// over a Repository and adds no validation of its own, matching the source.
//
// MIGRATION_NOTE: the Service interface is declared once, in
// internal/user/service.go. This file does not redeclare it, because a second
// declaration in package user would fail to compile. That interface must
// declare GetUserByName as (*User, bool, error) to match the method below.
//
// MIGRATION_NOTE: the source had no @Transactional annotation, so each
// repository call runs in its own transaction. The service keeps that
// behaviour and never opens a transaction spanning several repository calls.
type ServiceImp struct {
	repo Repository
}

// NewService returns a ServiceImp backed by repo. It replaces the source's
// field injection (@Autowired private UserDao userDao) with explicit
// constructor injection. Wire it in cmd/server/main.go, for example
// user.NewService(user.NewRepository(db)), after the schema is ensured.
func NewService(repo Repository) *ServiceImp {
	return &ServiceImp{repo: repo}
}

// SaveUser persists u and returns the saved entity exactly as the repository
// returns it (source: saveUser). It inserts when u is new and merges when u's
// ID already exists.
func (s *ServiceImp) SaveUser(ctx context.Context, u *User) (*User, error) {
	saved, err := s.repo.Save(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("save user: %w", err)
	}
	return saved, nil
}

// FetchUserList returns every user stored in the repository
// (source: fetchUserList).
func (s *ServiceImp) FetchUserList(ctx context.Context) ([]User, error) {
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch user list: %w", err)
	}
	return users, nil
}

// FetchUserByID returns the user with the given id (source: fetchUserById).
// When no such user exists, it returns a *NotFoundError carrying the verbatim
// message "User are not available" (NotFoundMessage). That error matches
// ErrNotFound via errors.Is, and the HTTP layer maps it to 404.
//
// MIGRATION_NOTE: Optional<User> plus isPresent() becomes the repository
// reporting ErrNotFound. A (nil, nil) result is also treated as absent, as a
// defensive check.
func (s *ServiceImp) FetchUserByID(ctx context.Context, id int) (*User, error) {
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

// DeleteUser deletes the user with the given id (source: deleteUser).
//
// MIGRATION_NOTE: Spring Data's deleteById throws
// EmptyResultDataAccessException when the id is missing. Here the repository's
// error is wrapped with context and propagated unchanged, so the HTTP layer
// can map it the same way the source did.
func (s *ServiceImp) DeleteUser(ctx context.Context, id int) error {
	if err := s.repo.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("delete user %d: %w", id, err)
	}
	return nil
}

// UpdateUser sets id on u and saves it (source: updateUser). This replaces the
// stored record with that id, or creates one when none exists (JPA merge
// semantics).
//
// MIGRATION_NOTE: like the source (user.setId(id)), this mutates the caller's
// User in place. The saved entity is discarded because the source returned
// void. A nil u returns an error; in the source it would have caused a
// NullPointerException.
func (s *ServiceImp) UpdateUser(ctx context.Context, id int, u *User) error {
	if u == nil {
		return errors.New("update user: nil user")
	}
	u.ID = id
	if _, err := s.repo.Save(ctx, u); err != nil {
		return fmt.Errorf("update user %d: %w", id, err)
	}
	return nil
}

// GetUserByName returns the user whose name exactly matches name
// (source: getUserNameByName).
//
// It returns three values:
//   - (user, true, nil) when a matching user exists.
//   - (nil, false, nil) when no user has that name. This is not an error,
//     which mirrors the source's null return.
//   - (nil, false, err) when the repository query fails.
//
// MIGRATION_NOTE: renamed from getUserNameByName to GetUserByName because it
// returns the whole User, not just the name. The signature changed from
// returning a nullable User to (*User, bool, error), matching
// Repository.FindByName. Java null maps to (nil, false, nil). Callers must
// check the bool rather than compare against nil alone.
func (s *ServiceImp) GetUserByName(ctx context.Context, name string) (*User, bool, error) {
	u, found, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, false, fmt.Errorf("get user by name %q: %w", name, err)
	}
	if !found || u == nil {
		return nil, false, nil
	}
	return u, true, nil
}
