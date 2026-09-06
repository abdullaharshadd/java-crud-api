package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Assuming UserDao is an interface that UserServiceImp interacts with.
type UserDao interface {
	GetUserByName(name string) (*User, error)
}

type UserServiceImp struct {
	UserDao UserDao
}

type User struct {
	ID       int
	Name     string
	Email    string
	About    string
	Password string
	Role     string
}

func (s *UserServiceImp) getUserNameByName(name string) (*User, error) {
	return s.UserDao.GetUserByName(name)
}

// Mock implementation of UserDao for testing purposes.
type mockUserDao struct{}

func (m *mockUserDao) GetUserByName(name string) (*User, error) {
	if name == "hemraj" {
		return &User{
			ID:       3,
			Name:     "hemraj",
			Email:    "hemrajmalhi1234@gmail.com",
			About:    "Sr",
			Password: "root",
			Role:     "java developer",
		}, nil
	}
	return nil, ErrUserNotFound
}

var ErrUserNotFound = errors.New("user not found")

func TestGetUserNameByName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name     string
		args     args
		wantUser *User
		wantErr  bool
	}{
		{
			name: "Valid user name",
			args: args{name: "hemraj"},
			wantUser: &User{
				ID:       3,
				Name:     "hemraj",
				Email:    "hemrajmalhi1234@gmail.com",
				About:    "Sr",
				Password: "root",
				Role:     "java developer",
			},
			wantErr: false,
		},
		{
			name:     "Invalid user name",
			args:     args{name: "nonexistentuser"},
			wantUser: nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &UserServiceImp{
				UserDao: &mockUserDao{},
			}
			gotUser, err := s.getUserNameByName(tt.args.name)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUser, gotUser)
			}
		})
	}
}