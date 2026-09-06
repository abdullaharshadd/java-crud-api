package service

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Assuming UserDao is an interface that represents the data access layer for users.
type UserDao interface {
	Save(user *User) (*User, error)
	FetchAll() ([]*User, error)
	FetchByID(id int) (*User, error)
	Delete(id int) error
	Update(id int, user *User) (*User, error)
	FetchByName(name string) (*User, error)
}

// Assuming User is a struct representing the user entity.
type User struct {
	ID        int
	Name      string
	Timestamp time.Time
}

// Mock implementation of UserDao for testing purposes.
type mockUserDao struct{}

func (m *mockUserDao) Save(user *User) (*User, error) {
	if user.ID == 0 {
		user.ID = 1 // Simulate generating an ID
	}
	user.Timestamp = time.Now()
	return user, nil
}

func (m *mockUserDao) FetchAll() ([]*User, error) {
	return []*User{{ID: 1, Name: "Alice", Timestamp: time.Now()}}, nil
}

func (m *mockUserDao) FetchByID(id int) (*User, error) {
	if id == 1 {
		return &User{ID: 1, Name: "Alice", Timestamp: time.Now()}, nil
	}
	return nil, ErrUserNotFound
}

func (m *mockUserDao) Delete(id int) error {
	if id != 1 {
		return ErrUserNotFound
	}
	return nil
}

func (m *mockUserDao) Update(id int, user *User) (*User, error) {
	if id != 1 {
		return nil, ErrUserNotFound
	}
	user.Timestamp = time.Now()
	return user, nil
}

func (m *mockUserDao) FetchByName(name string) (*User, error) {
	if name == "Alice" {
		return &User{ID: 1, Name: "Alice", Timestamp: time.Now()}, nil
	}
	return nil, ErrUserNotFound
}

var ErrUserNotFound = errors.New("user not found")

func TestSaveUser(t *testing.T) {
	type args struct {
		user *User
	}
	tests := []struct {
		name       string
		args       args
		wantUser   *User
		wantErr    bool
		sideEffect string
	}{
		{
			name: "valid user with unique ID",
			args: args{user: &User{ID: 1, Name: "Bob"}},
			wantUser: &User{
				ID:        1,
				Name:      "Bob",
				Timestamp: time.Time{}, // Placeholder, actual value will differ due to current time
			},
			wantErr:    false,
			sideEffect: "writes to DB",
		},
		{
			name: "valid user without ID",
			args: args{user: &User{Name: "Charlie"}},
			wantUser: &User{
				ID:        1, // Simulated generated ID
				Name:      "Charlie",
				Timestamp: time.Time{}, // Placeholder, actual value will differ due to current time
			},
			wantErr:    false,
			sideEffect: "writes to DB",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := UserService{UserDao: &mockUserDao{}}
			gotUser, err := svc.saveUser(tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("saveUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.wantUser.Name, gotUser.Name)
			assert.Equal(t, tt.wantUser.ID, gotUser.ID)
			assert.True(t, gotUser.Timestamp.After(time.Time{}))
		})
	}
}

func TestFetchUserList(t *testing.T) {
	tests := []struct {
		name       string
		wantUsers  []*User
		wantErr    bool
		sideEffect string
	}{
		{
			name:       "no users in the system",
			wantUsers:  []*User{},
			wantErr:    false,
			sideEffect: "",
		},
		{
			name: "users in the system",
			wantUsers: []*User{
				{ID: 1, Name: "Alice", Timestamp: time.Time{}}, // Placeholder, actual value will differ due to current time
			},
			wantErr:    false,
			sideEffect: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := UserService{UserDao: &mockUserDao{}}
			gotUsers, err := svc.fetchUserList()
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchUserList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, len(tt.wantUsers), len(gotUsers))
			for i, wantUser := range tt.wantUsers {
				assert.Equal(t, wantUser.Name, gotUsers[i].Name)
				assert.Equal(t, wantUser.ID, gotUsers[i].ID)
				assert.True(t, gotUsers[i].Timestamp.After(time.Time{}))
			}
		})
	}
}

func TestFetchUserById(t *testing.T) {
	tests := []struct {
		name       string
		id         int
		wantUser   *User
		wantErr    bool
		sideEffect string
	}{
		{
			name: "valid user ID",
			id:   1,
			wantUser: &User{
				ID:        1,
				Name:      "Alice",
				Timestamp: time.Time{}, // Placeholder, actual value will differ due to current time
			},
			wantErr:    false,
			sideEffect: "",
		},
		{
			name:     "invalid user ID",
			id:       2,
			wantUser: nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := UserService{UserDao: &mockUserDao{}}
			gotUser, err := svc.fetchUserById(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchUserById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantUser != nil {
				assert.Equal(t, tt.wantUser.Name, gotUser.Name)
				assert.Equal(t, tt.wantUser.ID, gotUser.ID)
				assert.True(t, gotUser.Timestamp.After(time.Time{}))
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name       string
		id         int
		wantErr    bool
		sideEffect string
	}{
		{
			name:       "valid user ID",
			id:         1,
			wantErr:    false,
			sideEffect: "deletes record from DB",
		},
		{
			name:     "invalid user ID",
			id:       2,
			wantErr:  true,
			sideEffect: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := UserService{UserDao: &mockUserDao{}}
			err := svc.deleteUser(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("deleteUser() error = %v, wantErr %v", err, tt.wantErr)
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
		name       string
		args       args
		wantUser   *User
		wantErr    bool
		sideEffect string
	}{
		{
			name: "valid user ID and updated information",
			args: args{id: 1, user: &User{Name: "Bob"}},
			wantUser: &User{
				ID:        1,
				Name:      "Bob",
				Timestamp: time.Time{}, // Placeholder, actual value will differ due to current time
			},
			wantErr:    false,
			sideEffect: "updates record in DB",
		},
		{
			name:     "invalid user ID",
			args:     args{id: 2, user: &User{Name: "Charlie"}},
			wantUser: nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := UserService{UserDao: &mockUserDao{}}
			gotUser, err := svc.updateUser(tt.args.id, tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUser != nil {
				assert.Equal(t, tt.wantUser.Name, gotUser.Name)
				assert.Equal(t, tt.wantUser.ID, gotUser.ID)
				assert.True(t, gotUser.Timestamp.After(time.Time{}))
			}
		})
	}
}

func TestGetUserNameByName(t *testing.T) {
	tests := []struct {
		name       string
		name       string
		wantUser   *User
		wantErr    bool
		sideEffect string
	}{
		{
			name: "valid user name",
			name: "Alice",
			wantUser: &User{
				ID:        1,
				Name:      "Alice",
				Timestamp: time.Time{}, // Placeholder, actual value will differ due to current time
			},
			wantErr:    false,
			sideEffect: "",
		},
		{
			name:     "invalid user name",
			name:     "Bob",
			wantUser: nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := UserService{UserDao: &mockUserDao{}}
			gotUser, err := svc.getUserNameByName(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUserNameByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUser != nil {
				assert.Equal(t, tt.wantUser.Name, gotUser.Name)
				assert.Equal(t, tt.wantUser.ID, gotUser.ID)
				assert.True(t, gotUser.Timestamp.After(time.Time{}))
			}
		})
	}
}