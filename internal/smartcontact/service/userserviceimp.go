package service

// MIGRATION_NOTE: The Java source UserServiceImp.java was the Spring
// @Service implementation of the UserService interface, using @Autowired
// field injection for its UserDao dependency. The idiomatic Go translation is:
//
//   - A concrete userService struct that holds the UserRepository dependency,
//     injected explicitly via the NewUserService constructor (defined in
//     userservice.go). This replaces Spring's field injection / component
//     scanning.
//   - Every I/O operation takes a context.Context as its first parameter for
//     cancellation and deadline propagation.
//   - Methods return explicit (T, error) pairs. The Java
//     UserNotFoundException checked-exception path (thrown by fetchUserById
//     when Optional.isPresent() is false) becomes the migrated
//     error.ErrUserNotFound sentinel returned from GetByID.
//
// The exported UserService interface and the NewUserService constructor live
// in userservice.go; this file provides the concrete implementation.
//
// MIGRATION_NOTE (Change 12 verification gate): The Java deleteUser delegated
// directly to userDao.deleteById(id) with NO existence guard. Spring Data JPA
// throws EmptyResultDataAccessException when the id does not exist, which the
// original code did not catch or map to a 404. That behaviour is preserved
// here: Delete forwards straight to the repository's DeleteByID, and the
// repository's ErrEmptyResultDelete propagates up unmapped (the handler layer
// surfaces it as a 500), matching the Java contract.

import (
	"context"
	"errors"
	"fmt"

	smartcontacterror "internal/smartcontact/error"
	"internal/smartcontact/model"
	"internal/smartcontact/repository"
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
func (s *userService) Save(ctx context.Context, user *model.User) (*model.User, error) {
	if user == nil {
		return nil, fmt.Errorf("save user: %w", errors.New("user must not be nil"))
	}
	saved, err := s.repo.Save(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("save user: %w", err)
	}
	return saved, nil
}

// List returns all persisted users.
//
// MIGRATION_NOTE: Mirrors UserServiceImp.fetchUserList, which delegated
// directly to UserDao.findAll.
func (s *userService) List(ctx context.Context) ([]*model.User, error) {
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

// GetByID returns the user with the given id.
//
// MIGRATION_NOTE: Mirrors UserServiceImp.fetchUserById. The Java code wrapped
// the repository lookup in an Optional and threw UserNotFoundException when the
// value was absent. Here the not-found case is expressed with the migrated
// error.ErrUserNotFound sentinel (matchable via errors.Is). A repository
// sql.ErrNoRows-equivalent is normalised to the not-found sentinel to preserve
// the original "User are not available" semantics.
func (s *userService) GetByID(ctx context.Context, id int) (*model.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, smartcontacterror.ErrUserNotFound) {
			return nil, smartcontacterror.NewUserNotFoundWithCause("User are not available", err)
		}
		return nil, fmt.Errorf("get user by id %d: %w", id, err)
	}
	if user == nil {
		return nil, smartcontacterror.NewUserNotFound("User are not available")
	}
	return user, nil
}

// Delete removes the user with the given id.
//
// MIGRATION_NOTE (Change 12): The Java deleteUser delegated directly to
// UserDao.deleteById(id) with no existence guard. That behaviour is preserved:
// any ErrEmptyResultDelete from the repository propagates unmapped so callers
// observe the same failure the Java code did (surfaced as a 500 upstream).
func (s *userService) Delete(ctx context.Context, id int) error {
	if err := s.repo.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("delete user %d: %w", id, err)
	}
	return nil
}

// Update sets the id on the supplied user and persists it, returning the
// stored entity.
//
// MIGRATION_NOTE: Mirrors UserServiceImp.updateUser, which set the id on the
// incoming entity and re-saved it (JPA merge semantics). The Java @NotNull on
// the primitive int parameter was a no-op at runtime; in Go the int is
// non-nullable by construction. The original method returned void; here we
// return the saved user for a more useful Go API.
func (s *userService) Update(ctx context.Context, id int, user *model.User) (*model.User, error) {
	if user == nil {
		return nil, fmt.Errorf("update user %d: %w", id, errors.New("user must not be nil"))
	}
	user.ID = id
	saved, err := s.repo.Save(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("update user %d: %w", id, err)
	}
	return saved, nil
}

// GetByName returns the user matching the given name.
//
// MIGRATION_NOTE: Mirrors UserServiceImp.getUserNameByName, which delegated to
// the Spring Data derived query UserDao.findByName. The Java version could
// return null; here a missing user is reported explicitly via error.
func (s *userService) GetByName(ctx context.Context, name string) (*model.User, error) {
	user, err := s.repo.FindByName(ctx, name)
	if err != nil {
		if errors.Is(err, smartcontacterror.ErrUserNotFound) {
			return nil, smartcontacterror.NewUserNotFoundWithCause("User are not available", err)
		}
		return nil, fmt.Errorf("get user by name %q: %w", name, err)
	}
	if user == nil {
		return nil, smartcontacterror.NewUserNotFound("User are not available")
	}
	return user, nil
}
