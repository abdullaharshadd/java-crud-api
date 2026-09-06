package service

import (
	"context"
	"testing"
	"migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByName(ctx context.Context, name string) (*model.User, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// TestUserServiceImp tests the UserServiceImp implementation.
func TestUserServiceImp(t *testing.T) {
	userService := service.NewUserService(&MockUserRepository{})
	user := &model.User{
		Name: "hemraj",
		Email: "hemrajmalhi1234@gmail.com",
		About: "Sr",
		Password: "root",
		Role: "java developer",
		ID: 3,
	}
	userService.(*service.UserServiceImp).userRepository = &MockUserRepository{mock.Mock{} }
	userService.(*service.UserServiceImp).userRepository.On("FindByName", mock.Anything, "hemraj").Return(user, nil)

	ctx := context.Background()
	name := "hemraj"
	foundUser, err := userService.FindByName(ctx, name)
	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, name, foundUser.Name)
}
