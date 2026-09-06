package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func TestUserServiceImp_FindByName(t *testing.T) {
	testCases := []struct {
		name              string
		inputName         string
		expectedUser      *model.User
		expectedError     bool
		expectedSideEffect func(*MockUserRepository)
	}{
		{
			name: "valid user name",
			inputName: "hemraj",
			expectedUser: &model.User{
				Name: "hemraj",
				Email: "hemrajmalhi1234@gmail.com",
				About: "Sr",
				Password: "root",
				Role: "java developer",
				ID: 3,
			},
			expectedError: false,
			expectedSideEffect: func(repo *MockUserRepository) {
				repo.On("FindByName", mock.Anything, "hemraj").Return(&model.User{
					Name: "hemraj",
					Email: "hemrajmalhi1234@gmail.com",
					About: "Sr",
					Password: "root",
					Role: "java developer",
					ID: 3,
				}, nil)
			},
		},
		{
			name: "invalid user name",
			inputName: "nonexistent",
			expectedUser: nil,
			expectedError: true,
			expectedSideEffect: func(repo *MockUserRepository) {
				repo.On("FindByName", mock.Anything, "nonexistent").Return(nil, error.NewNotFoundError("User not found"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &MockUserRepository{}
			if tc.expectedSideEffect != nil {
				tc.expectedSideEffect(repo)
			}
			userService := NewUserService(repo)
			ctx := context.Background()

			foundUser, err := userService.FindByName(ctx, tc.inputName)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.expectedUser != nil {
				assert.NotNil(t, foundUser)
				assert.Equal(t, tc.expectedUser.Name, foundUser.Name)
				assert.Equal(t, tc.expectedUser.Email, foundUser.Email)
				assert.Equal(t, tc.expectedUser.About, foundUser.About)
				assert.Equal(t, tc.expectedUser.Role, foundUser.Role)
				assert.Equal(t, tc.expectedUser.ID, foundUser.ID)
			} else {
				assert.Nil(t, foundUser)
			}

			repo.AssertExpectations(t)
		})
	}
}