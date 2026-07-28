// Package service provides the business-logic layer for the SmartContact
// application. It replaces the Java Spring service layer under
// com.smartContact.service.
//
// MIGRATION_NOTE: The Java source was split into an interface (UserService) and
// an implementation (a @Service-annotated class, migrated elsewhere). Idiomatic
// Go does not need this split: dependencies are injected explicitly via a
// constructor and consumers (e.g. HTTP handlers) may define their own narrow
// interface at the point of use if they need to mock the service. Therefore
// this file collapses interface + implementation into a single concrete
// UserService type backed by the repository.UserDao. There is no Spring
// component scanning or @Autowired wiring; the caller must call NewUserService.
//
// The original Java methods map as follows:
//
//	saveUser(User)              -> SaveUser(ctx, req)
//	fetchUserList()            -> FetchUserList(ctx)
//	fetchUserById(int)         -> FetchUserByID(ctx, id)
//	deleteUser(int)            -> DeleteUser(ctx, id)
//	updateUser(int, User)      -> UpdateUser(ctx, id, req)
//	getUserNameByName(String)  -> GetUserByName(ctx, name)
//
// The Java checked exception UserNotFoundException becomes the sentinel error
// error.ErrUserNotFound, propagated (and wrapped) through the (T, error) return
// convention rather than thrown.
package service

import (
	"context"
	"errors"
	"fmt"

	smartcontacterror "internal/smartcontact/error"
	"internal/smartcontact/model"
	"internal/smartcontact/repository"
)

// UserService implements the user-related business operations of the
// SmartContact application. It decouples HTTP handlers from the underlying
// data-access layer.
type UserService struct {
	dao repository.UserDao
}

// NewUserService constructs a UserService backed by the given UserDao.
// It replaces Spring's dependency injection of the concrete service.
func NewUserService(dao repository.UserDao) *UserService {
	return &UserService{dao: dao}
}

// SaveUser validates the incoming request, converts it to a model.User and
// persists it, returning the stored representation. It replaces the Java
// saveUser(User) method.
//
// MIGRATION_NOTE: The Java method accepted and returned a JPA entity directly.
// Here we accept a CreateUserRequest (the API-facing DTO) and return a
// UserResponse, keeping the persistence entity internal to the service and
// repository layers.
func (s *UserService) SaveUser(ctx context.Context, req model.CreateUserRequest) (model.UserResponse, error) {
	if err := req.Validate(); err != nil {
		return model.UserResponse{}, fmt.Errorf("save user: %w", err)
	}

	user := req.ToUser()
	saved, err := s.dao.Save(ctx, user)
	if err != nil {
		return model.UserResponse{}, fmt.Errorf("save user: %w", err)
	}

	return saved.ToResponse(), nil
}

// FetchUserList returns all users. It replaces the Java fetchUserList() method.
func (s *UserService) FetchUserList(ctx context.Context) ([]model.UserResponse, error) {
	users, err := s.dao.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch user list: %w", err)
	}

	responses := make([]model.UserResponse, 0, len(users))
	for _, u := range users {
		responses = append(responses, u.ToResponse())
	}
	return responses, nil
}

// FetchUserByID returns the user with the given id. If no such user exists it
// returns error.ErrUserNotFound (wrapped), replacing the Java
// fetchUserById(int) method that threw UserNotFoundException.
func (s *UserService) FetchUserByID(ctx context.Context, id int) (model.UserResponse, error) {
	user, err := s.dao.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, smartcontacterror.ErrUserNotFound) {
			return model.UserResponse{}, fmt.Errorf("fetch user by id %d: %w", id, smartcontacterror.ErrUserNotFound)
		}
		return model.UserResponse{}, fmt.Errorf("fetch user by id %d: %w", id, err)
	}
	return user.ToResponse(), nil
}

// DeleteUser removes the user with the given id. It replaces the Java
// deleteUser(int) method.
//
// MIGRATION_NOTE: The Java method returned void and silently ignored missing
// rows. Here we surface a wrapped ErrUserNotFound if the repository reports the
// row did not exist, so callers can respond appropriately.
func (s *UserService) DeleteUser(ctx context.Context, id int) error {
	if err := s.dao.DeleteByID(ctx, id); err != nil {
		if errors.Is(err, smartcontacterror.ErrUserNotFound) {
			return fmt.Errorf("delete user %d: %w", id, smartcontacterror.ErrUserNotFound)
		}
		return fmt.Errorf("delete user %d: %w", id, err)
	}
	return nil
}

// UpdateUser applies the given request to the user identified by id and returns
// the updated representation. It replaces the Java updateUser(int, User) method.
//
// MIGRATION_NOTE: The Java method returned void. In idiomatic Go we return the
// updated UserResponse (and an error) so callers do not need a second lookup.
func (s *UserService) UpdateUser(ctx context.Context, id int, req model.CreateUserRequest) (model.UserResponse, error) {
	if err := req.Validate(); err != nil {
		return model.UserResponse{}, fmt.Errorf("update user %d: %w", id, err)
	}

	user := req.ToUser()
	user.ID = id

	updated, err := s.dao.Update(ctx, user)
	if err != nil {
		if errors.Is(err, smartcontacterror.ErrUserNotFound) {
			return model.UserResponse{}, fmt.Errorf("update user %d: %w", id, smartcontacterror.ErrUserNotFound)
		}
		return model.UserResponse{}, fmt.Errorf("update user %d: %w", id, err)
	}
	return updated.ToResponse(), nil
}

// GetUserByName returns the user with the given name. If no such user exists it
// returns error.ErrUserNotFound (wrapped). It replaces the Java
// getUserNameByName(String) method.
//
// MIGRATION_NOTE: The Java derived query findByName/getUserNameByName returned
// null when no match was found, leaving null-handling to the caller. Here we
// convert a not-found result into an explicit, wrapped ErrUserNotFound.
func (s *UserService) GetUserByName(ctx context.Context, name string) (model.UserResponse, error) {
	user, err := s.dao.FindByName(ctx, name)
	if err != nil {
		if errors.Is(err, smartcontacterror.ErrUserNotFound) {
			return model.UserResponse{}, fmt.Errorf("get user by name %q: %w", name, smartcontacterror.ErrUserNotFound)
		}
		return model.UserResponse{}, fmt.Errorf("get user by name %q: %w", name, err)
	}
	return user.ToResponse(), nil
}
