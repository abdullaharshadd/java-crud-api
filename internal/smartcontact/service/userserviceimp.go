package service

// MIGRATION_NOTE: The Java source was UserServiceImp, the Spring @Service
// implementation of the UserService interface. The interface itself was
// migrated separately (see userservice.go). In Go we do not redeclare the
// interface; this file provides the concrete implementation type and its
// constructor. Spring field injection (@Autowired UserDao) becomes explicit
// constructor injection of the repository dependency, expressed as the
// repository.UserRepository interface so the service is decoupled from the
// concrete UserDao.
//
// MIGRATION_NOTE: Every method that performs I/O takes a context.Context as
// its first parameter (Spring had no equivalent). Methods return (T, error)
// instead of throwing checked exceptions. The Optional<User> lookup in
// fetchUserById becomes an errors.Is check against ErrUserNotFound surfaced by
// the repository, preserving the original "user not available" semantics.
//
// MIGRATION_NOTE: The Java updateUser mutated the incoming User's id and
// re-persisted via save(); the create path (saveUser) and update path both
// funnel through the repository's Merge. Here updateUser sets the id on the
// model and delegates to Merge, matching the original upsert-via-save
// behaviour.

import (
	"context"
	"errors"
	"fmt"

	smartcontacterror "migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/repository"
)

// UserServiceImp is the concrete implementation of the UserService interface.
// It contains no business state of its own beyond the repository dependency it
// delegates to, mirroring the original stateless Spring @Service bean.
type UserServiceImp struct {
	userDao repository.UserRepository
}

// NewUserServiceImp constructs a UserServiceImp backed by the supplied
// repository. This replaces Spring's @Autowired field injection with explicit
// constructor injection. The returned value satisfies the UserService
// interface declared in userservice.go.
func NewUserServiceImp(userDao repository.UserRepository) *UserServiceImp {
	return &UserServiceImp{userDao: userDao}
}

// Compile-time assertion that *UserServiceImp satisfies UserService.
var _ UserService = (*UserServiceImp)(nil)

// SaveUser persists a new user, delegating to the repository. It corresponds
// to the Java saveUser method and returns the stored user (with any
// database-generated id populated) or an error.
func (s *UserServiceImp) SaveUser(ctx context.Context, user model.User) (model.UserResponse, error) {
	saved, err := s.userDao.Merge(ctx, user)
	if err != nil {
		return model.UserResponse{}, fmt.Errorf("save user: %w", err)
	}
	return saved, nil
}

// FetchUserList returns all users, delegating to the repository. It
// corresponds to the Java fetchUserList method.
func (s *UserServiceImp) FetchUserList(ctx context.Context) ([]model.UserResponse, error) {
	users, err := s.userDao.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch user list: %w", err)
	}
	return users, nil
}

// FetchUserByID returns the user with the given id. It corresponds to the Java
// fetchUserById method: when no user exists it returns ErrUserNotFound
// (equivalent to the original UserNotFoundException with the message
// "user are not available"). Callers can test for this with
// errors.Is(err, smartcontacterror.ErrUserNotFound).
func (s *UserServiceImp) FetchUserByID(ctx context.Context, id int) (model.UserResponse, error) {
	user, err := s.userDao.FindByID(ctx, id)
	if err != nil {
		if errorsIsNotFound(err) {
			return model.UserResponse{}, smartcontacterror.NewUserNotFoundError("user are not available")
		}
		return model.UserResponse{}, fmt.Errorf("fetch user by id %d: %w", id, err)
	}
	return user, nil
}

// DeleteUser removes the user with the given id, delegating to the repository.
// It corresponds to the Java deleteUser method.
func (s *UserServiceImp) DeleteUser(ctx context.Context, id int) error {
	if err := s.userDao.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("delete user %d: %w", id, err)
	}
	return nil
}

// UpdateUser assigns the supplied id to the user and persists it, mirroring the
// Java updateUser method which set the id then re-saved. It delegates to the
// repository's Merge (Spring's save() was an upsert).
func (s *UserServiceImp) UpdateUser(ctx context.Context, id int, user model.User) error {
	user.ID = id
	if _, err := s.userDao.Merge(ctx, user); err != nil {
		return fmt.Errorf("update user %d: %w", id, err)
	}
	return nil
}

// GetUserByName returns the user matching the given name, delegating to the
// repository's derived finder. It corresponds to the Java getUserNameByName
// method (which invoked the derived query findByName).
func (s *UserServiceImp) GetUserByName(ctx context.Context, name string) (model.UserResponse, error) {
	user, err := s.userDao.FindByName(ctx, name)
	if err != nil {
		return model.UserResponse{}, fmt.Errorf("get user by name %q: %w", name, err)
	}
	return user, nil
}

// errorsIsNotFound reports whether err represents a "user not found" condition
// surfaced by the repository. It is a thin helper kept local so callers of the
// service continue to receive the domain ErrUserNotFound sentinel.
func errorsIsNotFound(err error) bool {
	return err != nil && errors.Is(err, smartcontacterror.ErrUserNotFound)
}