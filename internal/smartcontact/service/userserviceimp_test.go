package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Assuming UserDao is an interface that has methods for CRUD operations on users.
type UserDao interface {
	Save(user *User) (*User, error)
	FetchAll() ([]*User, error)
	FetchByID(id int) (*User, error)
	FetchByName(name string) (*User, error)
	Delete(id int) error
	Update(user *User) error
}

type User struct {
	ID   int
	Name string
}

type UserService struct {
	userDao UserDao
}

func NewUserService(userDao UserDao) *UserService {
	return &UserService{userDao: userDao}
}

func TestSaveUser(t *testing.T) {
	type args struct {
		user *User
	}
	tests := []struct {
		name         string
		args         args
		mockBehavior func(mockDao *mockUserDao)
		want         *User
		wantErr      bool
	}{
		{
			name: "save new user",
			args: args{user: &User{Name: "John Doe"}},
			mockBehavior: func(mockDao *mockUserDao) {
				mockDao.On("Save", &User{Name: "John Doe"}).Return(&User{ID: 1, Name: "John Doe"}, nil)
			},
			want:    &User{ID: 1, Name: "John Doe"},
			wantErr: false,
		},
		{
			name: "update existing user",
			args: args{user: &User{ID: 1, Name: "Jane Doe"}},
			mockBehavior: func(mockDao *mockUserDao) {
				mockDao.On("Save", &User{ID: 1, Name: "Jane Doe"}).Return(&User{ID: 1, Name: "Jane Doe"}, nil)
			},
			want:    &User{ID: 1, Name: "Jane Doe"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDao := &mockUserDao{}
			tt.mockBehavior(mockDao)
			svc := NewUserService(mockDao)
			got, err := svc.SaveUser(tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFetchUserList(t *testing.T) {
	tests := []struct {
		name         string
		mockBehavior func(mockDao *mockUserDao)
		want         []*User
		wantErr      bool
	}{
		{
			name: "fetch users when users exist",
			mockBehavior: func(mockDao *mockUserDao) {
				mockDao.On("FetchAll").Return([]*User{{ID: 1, Name: "John Doe"}, {ID: 2, Name: "Jane Doe"}}, nil)
			},
			want:    []*User{{ID: 1, Name: "John Doe"}, {ID: 2, Name: "Jane Doe"}},
			wantErr: false,
		},
		{
			name: "fetch users when no users exist",
			mockBehavior: func(mockDao *mockUserDao) {
				mockDao.On("FetchAll").Return([]*User{}, nil)
			},
			want:    []*User{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDao := &mockUserDao{}
			tt.mockBehavior(mockDao)
			svc := NewUserService(mockDao)
			got, err := svc.FetchUserList()
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchUserList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFetchUserById(t *testing.T) {
	tests := []struct {
		name         string
		id           int
		mockBehavior func(mockDao *mockUserDao)
		want         *User
		wantErr      bool
	}{
		{
			name: "fetch valid user",
			id:   1,
			mockBehavior: func(mockDao *mockUserDao) {
				mockDao.On("FetchByID", 1).Return(&User{ID: 1, Name: "John Doe"}, nil)
			},
			want:    &User{ID: 1, Name: "John Doe"},
			wantErr: false,
		},
		{
			name: "fetch invalid user",
			id:   999,
			mockBehavior: func(mockDao *mockUserDao) {
				mockDao.On("FetchByID", 999).Return(nil, ErrUserNotFound)
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDao := &mockUserDao{}
			tt.mockBehavior(mockDao)
			svc := NewUserService(mockDao)
			got, err := svc.FetchUserById(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchUserById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name         string
		id           int
		mockBehavior func(mockDao *mockUserDao)
		wantErr      bool
	}{
		{
			name: "delete valid user",
			id:   1,
			mockBehavior: func(mockDao *mockUserDao) {
				mockDao.On("Delete", 1).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "delete invalid user",
			id:   999,
			mockBehavior: func(mockDao *mockUserDao) {
				mockDao.On("Delete", 999).Return(ErrUserNotFound)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDao := &mockUserDao{}
			tt.mockBehavior(mockDao)
			svc := NewUserService(mockDao)
			err := svc.DeleteUser(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	type args struct {
		id   int
		user *User
	}
	tests := []struct {
		name         string
		args         args
		mockBehavior func(mockDao *mockUserDao)
		wantErr      bool
	}{
		{
			name: "update valid user",
			args: args{id: 1, user: &User{Name: "Jane Doe"}},
			mockBehavior: func(mockDao *mockUserDao) {
				mockDao.On("Update", &User{ID: 1, Name: "Jane Doe"}).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "update invalid user",
			args: args{id: 999, user: &User{Name: "Invalid User"}},
			mockBehavior: func(mockDao *mockUserDao) {
				mockDao.On("Update", &User{ID: 999, Name: "Invalid User"}).Return(ErrUserNotFound)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDao := &mockUserDao{}
			tt.mockBehavior(mockDao)
			svc := NewUserService(mockDao)
			tt.args.user.ID = tt.args.id
			err := svc.UpdateUser(tt.args.id, tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetUserNameByName(t *testing.T) {
	tests := []struct {
		name         string
		name         string
		mockBehavior func(mockDao *mockUserDao)
		want         *User
		wantErr      bool
	}{
		{
			name: "fetch valid user by name",
			name: "John Doe",
			mockBehavior: func(mockDao *mockUserDao) {
				mockDao.On("FetchByName", "John Doe").Return(&User{ID: 1, Name: "John Doe"}, nil)
			},
			want:    &User{ID: 1, Name: "John Doe"},
			wantErr: false,
		},
		{
			name: "fetch invalid user by name",
			name: "Nonexistent User",
			mockBehavior: func(mockDao *mockUserDao) {
				mockDao.On("FetchByName", "Nonexistent User").Return(nil, ErrUserNotFound)
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDao := &mockUserDao{}
			tt.mockBehavior(mockDao)
			svc := NewUserService(mockDao)
			got, err := svc.GetUserNameByName(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserNameByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

// Mock implementation of UserDao for testing purposes.
type mockUserDao struct {
	mock.Mock
}

func (m *mockUserDao) Save(user *User) (*User, error) {
	args := m.Called(user)
	return args.Get(0).(*User), args.Error(1)
}

func (m *mockUserDao) FetchAll() ([]*User, error) {
	args := m.Called()
	return args.Get(0).([]*User), args.Error(1)
}

func (m *mockUserDao) FetchByID(id int) (*User, error) {
	args := m.Called(id)
	return args.Get(0).(*User), args.Error(1)
}

func (m *mockUserDao) FetchByName(name string) (*User, error) {
	args := m.Called(name)
	return args.Get(0).(*User), args.Error(1)
}

func (m *mockUserDao) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockUserDao) Update(user *User) error {
	args := m.Called(user)
	return args.Error(0)
}

var ErrUserNotFound = &UserNotFoundException{"User are not available"}

type UserNotFoundException struct {
	msg string
}

func (e *UserNotFoundException) Error() string {
	return e.msg
}