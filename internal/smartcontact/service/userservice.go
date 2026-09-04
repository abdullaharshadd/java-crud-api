package service

import (
	"context"
	"errors"
	"migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/repository"
)

// UserService defines the methods for interacting with user-related operations.
type UserService interface {
	SaveUser(ctx context.Context, user *model.User) (*model.User, error)
	FetchUserList(ctx context.Context) ([]*model.User, error)
	FetchUserByID(ctx context.Context, id int) (*model.User, error)
	DeleteUser(ctx context.Context, id int) error
	UpdateUser(ctx context.Context, id int, user *model.User) error
	FindByName(ctx context.Context, name string) (*model.User, error)
}

// newUserService creates a new UserService instance.
func newUserService(ur repository.UserRepository) UserService {
	return &userService{
		userRepository: ur,
	}
}

// userService is the implementation of UserService.
type userService struct {
	userRepository repository.UserRepository
}

// SaveUser saves a user to the database.
func (us *userService) SaveUser(ctx context.Context, user *model.User) (*model.User, error) {
	if err := user.Validate(); err != nil {
		return nil, err
	}
	return us.userRepository.Save(ctx, user)
}

// FetchUserList retrieves a list of users from the database.
func (us *userService) FetchUserList(ctx context.Context) ([]*model.User, error) {
	return us.userRepository.FetchAll(ctx)
}

// FetchUserByID retrieves a user by their ID.
func (us *userService) FetchUserByID(ctx context.Context, id int) (*model.User, error) {
	user, err := us.userRepository.FindByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, error.NewUserNotFoundError("User not found", nil)
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// DeleteUser deletes a user from the database by their ID.
func (us *userService) DeleteUser(ctx context.Context, id int) error {
	return us.userRepository.Delete(ctx, id)
}

// UpdateUser updates a user in the database by their ID.
func (us *userService) UpdateUser(ctx context.Context, id int, user *model.User) error {
	if err := user.Validate(); err != nil {
		return err
	}
	return us.userRepository.Update(ctx, id, user)
}

// FindByName retrieves a user by their name.
func (us *userService) FindByName(ctx context.Context, name string) (*model.User, error) {
	return us.userRepository.FindByName(ctx, name)
}