package service

import (
	"context"
	"errors"
	"migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/repository"
)

// userServiceImpl is the implementation of UserService.
// It interacts with the user repository to perform CRUD operations.
type userServiceImpl struct {
	userRepository repository.UserRepository
}

// newUserServiceImpl creates a new userServiceImpl instance.
func newUserServiceImpl(ur repository.UserRepository) UserService {
	return &userServiceImpl{
		userRepository: ur,
	}
}

// SaveUser saves a user to the database.
func (us *userServiceImpl) SaveUser(ctx context.Context, user *model.User) (*model.User, error) {
	if err := user.Validate(); err != nil {
		return nil, err
	}
	return us.userRepository.Save(ctx, user)
}

// FetchUserList fetches all users from the database.
func (us *userServiceImpl) FetchUserList(ctx context.Context) ([]*model.User, error) {
	return us.userRepository.FindAll(ctx)
}

// FetchUserByID fetches a user by their ID.
func (us *userServiceImpl) FetchUserByID(ctx context.Context, id int) (*model.User, error) {
	user, err := us.userRepository.FindByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, error.NewUserNotFoundError("User not found", nil)
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// DeleteUser deletes a user by their ID.
func (us *userServiceImpl) DeleteUser(ctx context.Context, id int) error {
	return us.userRepository.Delete(ctx, id)
}

// UpdateUser updates a user's information in the database.
func (us *userServiceImpl) UpdateUser(ctx context.Context, id int, user *model.User) error {
	user.ID = id
	if err := user.Validate(); err != nil {
		return err
	}
	return us.userRepository.Update(ctx, user)
}

// FindByName finds a user by their name.
func (us *userServiceImpl) FindByName(ctx context.Context, name string) (*model.User, error) {
	return us.userRepository.FindByName(ctx, name)
}
