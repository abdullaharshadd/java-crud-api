// Package service defines the business-logic contracts and implementations for
// the smartcontact application. It sits between the HTTP handlers and the
// repository layer.
//
// MIGRATION_NOTE: The Java source was the Spring @Service implementation class
// UserServiceImp implementing the UserService interface. In Go the interface
// lives in userservice.go; this file supplies the concrete implementation.
// Spring's field injection (@Autowired UserDao) is replaced by explicit
// constructor injection via NewUserService. The repository dependency is
// declared as an interface so the service can be tested with a fake DAO.
//
// Idiomatic Go changes from the source:
//   - Every I/O method takes a context.Context as its first parameter for
//     cancellation/deadline propagation.
//   - fetchUserById's "not present" branch (Java Optional + thrown
//     UserNotFoundException) becomes an explicit (*User, error) return using
//     apperr.NewUserNotFound.
//   - updateUser still sets the id from the path variable on the incoming user
//     before delegating to Save, preserving the original business logic.
package service

import (
	"context"
	"fmt"

	apperr "migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
)

// userRepository is the subset of repository.UserDao behavior that the service
// depends on. Declaring it here (accepting an interface, returning a struct)
// keeps the service decoupled from the concrete data-access implementation and
// makes it trivial to substitute a fake in tests.
//
// The concrete *repository.UserDao satisfies this interface.
type userRepository interface {
	Save(ctx context.Context, user *model.User) (*model.User, error)
	FindAll(ctx context.Context) ([]model.User, error)
	FindByID(ctx context.Context, id int) (*model.User, error)
	DeleteByID(ctx context.Context, id int) error
	FindByName(ctx context.Context, name string) (*model.User, error)
}

// userService is the concrete implementation of the UserService interface.
// It delegates all persistence to the injected userRepository.
type userService struct {
	dao userRepository
}

// NewUserService constructs a UserService backed by the given repository.
// This replaces Spring's @Autowired field injection with explicit constructor
// injection. It returns the UserService interface so callers depend on the
// behavior contract rather than the concrete type.
func NewUserService(dao userRepository) UserService {
	return &userService{dao: dao}
}

// SaveUser persists a new user and returns the stored representation.
func (s *userService) SaveUser(ctx context.Context, user *model.User) (*model.User, error) {
	saved, err := s.dao.Save(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("save user: %w", err)
	}
	return saved, nil
}

// FetchUserList returns every user in the store.
func (s *userService) FetchUserList(ctx context.Context) ([]*model.User, error) {
	users, err := s.dao.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch user list: %w", err)
	}
	// Convert []model.User to []*model.User
	result := make([]*model.User, len(users))
	for i := range users {
		u := users[i]
		result[i] = &u
	}
	return result, nil
}

// FetchUserByID returns the user with the given id. If no such user exists it
// returns a *apperr.UserNotFound error, mirroring the Java source which threw
// UserNotFoundException("User are not available") when the Optional was empty.
func (s *userService) FetchUserByID(ctx context.Context, id int64) (*model.User, error) {
	user, err := s.dao.FindByID(ctx, int(id))
	if err != nil {
		// surface repository-level not-found as domain not-found
		return nil, fmt.Errorf("fetch user by id %d: %w", id, err)
	}
	if user == nil {
		return nil, apperr.NewUserNotFound("User are not available")
	}
	return user, nil
}

// DeleteUser removes the user with the given id.
func (s *userService) DeleteUser(ctx context.Context, id int64) error {
	if err := s.dao.DeleteByID(ctx, int(id)); err != nil {
		return fmt.Errorf("delete user %d: %w", id, err)
	}
	return nil
}

// UpdateUser sets the id from the path variable on the incoming user and
// persists it, preserving the Java source's behavior of forcing the entity id
// to match the URL id before delegating to save.
func (s *userService) UpdateUser(ctx context.Context, id int64, user *model.User) error {
	if user == nil {
		return fmt.Errorf("update user %d: user must not be nil", id)
	}
	user.ID = int(id)
	if _, err := s.dao.Save(ctx, user); err != nil {
		return fmt.Errorf("update user %d: %w", id, err)
	}
	return nil
}

// GetUserByName returns the user matching the given name.
func (s *userService) GetUserByName(ctx context.Context, name string) (*model.User, error) {
	user, err := s.dao.FindByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get user by name %q: %w", name, err)
	}
	return user, nil
}
