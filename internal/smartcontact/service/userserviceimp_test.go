package service

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/repository"
)

type mockUserRepository struct {
	repository.UserRepository
	saveFunc          func(context.Context, *model.User) (*model.User, error)
	fetchAllFunc      func(context.Context) ([]*model.User, error)
	findByIDFunc      func(context.Context, int) (*model.User, error)
	deleteFunc        func(context.Context, int) error
	updateFunc        func(context.Context, *model.User) error
	findByNameFunc    func(context.Context, string) (*model.User, error)
}

func (mur *mockUserRepository) Save(ctx context.Context, user *model.User) (*model.User, error) {
	return mur.saveFunc(ctx, user)
}

func (mur *mockUserRepository) FindAll(ctx context.Context) ([]*model.User, error) {
	return mur.fetchAllFunc(ctx)
}

func (mur *mockUserRepository) FindByID(ctx context.Context, id int) (*model.User, error) {
	return mur.findByIDFunc(ctx, id)
}

func (mur *mockUserRepository) Delete(ctx context.Context, id int) error {
	return mur.deleteFunc(ctx, id)
}

func (mur *mockUserRepository) Update(ctx context.Context, user *model.User) error {
	return mur.updateFunc(ctx, user)
}

func (mur *mockUserRepository) FindByName(ctx context.Context, name string) (*model.User, error) {
	return mur.findByNameFunc(ctx, name)
}

var saveUserTests = []struct {
	name         string
	input        *model.User
	mockBehavior func(mur *mockUserRepository)
	expectedUser *model.User
	expectedErr  error
}{
	{"valid user", &model.User{Name: "John Doe"}, func(mur *mockUserRepository) { mur.saveFunc = func(ctx context.Context, user *model.User) (*model.User, error) { return user, nil } }, &model.User{Name: "John Doe"}, nil},
	{"invalid user", &model.User{Name: ""}, func(mur *mockUserRepository) { mur.saveFunc = func(ctx context.Context, user *model.User) (*model.User, error) { return nil, errors.New("validation failed") } }, nil, errors.New("validation failed")},
}

func TestSaveUser(t *testing.T) {
	for _, tt := range saveUserTests {
		t.Run(tt.name, func(t *testing.T) {
			mur := &mockUserRepository{}
			tt.mockBehavior(mur)
			us := newUserServiceImpl(mur)
			user, err := us.SaveUser(context.Background(), tt.input)
			assert.Equal(t, tt.expectedUser, user)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

var fetchUserListTests = []struct {
	name         string
	mockBehavior func(mur *mockUserRepository)
	expectedList []*model.User
	expectedErr  error
}{
	{"users present", func(mur *mockUserRepository) { mur.fetchAllFunc = func(ctx context.Context) ([]*model.User, error) { return []*model.User{{Name: "Alice"}}, nil } }, []*model.User{{Name: "Alice"}}, nil},
	{"no users", func(mur *mockUserRepository) { mur.fetchAllFunc = func(ctx context.Context) ([]*model.User, error) { return []*model.User{}, nil } }, []*model.User{}, nil},
}

func TestFetchUserList(t *testing.T) {
	for _, tt := range fetchUserListTests {
		t.Run(tt.name, func(t *testing.T) {
			mur := &mockUserRepository{}
			tt.mockBehavior(mur)
			us := newUserServiceImpl(mur)
			list, err := us.FetchUserList(context.Background())
			assert.Equal(t, tt.expectedList, list)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

var fetchUserByIDTests = []struct {
	name         string
	id           int
	mockBehavior func(mur *mockUserRepository)
	expectedUser *model.User
	expectedErr  error
}{
	{"valid user", 1, func(mur *mockUserRepository) { mur.findByIDFunc = func(ctx context.Context, id int) (*model.User, error) { return &model.User{ID: id, Name: "Bob"}, nil } }, &model.User{ID: 1, Name: "Bob"}, nil},
	{"invalid user", 2, func(mur *mockUserRepository) { mur.findByIDFunc = func(ctx context.Context, id int) (*model.User, error) { return nil, sql.ErrNoRows } }, nil, error.NewUserNotFoundError("User not found", nil)},
}

func TestFetchUserByID(t *testing.T) {
	for _, tt := range fetchUserByIDTests {
		t.Run(tt.name, func(t *testing.T) {
			mur := &mockUserRepository{}
			tt.mockBehavior(mur)
			us := newUserServiceImpl(mur)
			user, err := us.FetchUserByID(context.Background(), tt.id)
			assert.Equal(t, tt.expectedUser, user)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

var deleteUserTests = []struct {
	name         string
	id           int
	mockBehavior func(mur *mockUserRepository)
	expectedErr  error
}{
	{"valid user", 3, func(mur *mockUserRepository) { mur.deleteFunc = func(ctx context.Context, id int) error { return nil } }, nil},
	{"invalid user", 4, func(mur *mockUserRepository) { mur.deleteFunc = func(ctx context.Context, id int) error { return sql.ErrNoRows } }, nil},
}

func TestDeleteUser(t *testing.T) {
	for _, tt := range deleteUserTests {
		t.Run(tt.name, func(t *testing.T) {
			mur := &mockUserRepository{}
			tt.mockBehavior(mur)
			us := newUserServiceImpl(mur)
			err := us.DeleteUser(context.Background(), tt.id)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

var updateUserTests = []struct {
	name         string
	id           int
	input        *model.User
	mockBehavior func(mur *mockUserRepository)
	expectedErr  error
}{
	{"valid user update", 5, &model.User{Name: "Charlie"}, func(mur *mockUserRepository) { mur.updateFunc = func(ctx context.Context, user *model.User) error { return nil } }, nil},
	{"invalid user update", 6, &model.User{Name: ""}, func(mur *mockUserRepository) { mur.updateFunc = func(ctx context.Context, user *model.User) error { return errors.New("validation failed") } }, errors.New("validation failed")},
}

func TestUpdateUser(t *testing.T) {
	for _, tt := range updateUserTests {
		t.Run(tt.name, func(t *testing.T) {
			mur := &mockUserRepository{}
			tt.mockBehavior(mur)
			us := newUserServiceImpl(mur)
			err := us.UpdateUser(context.Background(), tt.id, tt.input)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

var findByNameTests = []struct {
	name         string
	name         string
	mockBehavior func(mur *mockUserRepository)
	expectedUser *model.User
	expectedErr  error
}{
	{"valid user name", "Eve", func(mur *mockUserRepository) { mur.findByNameFunc = func(ctx context.Context, name string) (*model.User, error) { return &model.User{Name: name}, nil } }, &model.User{Name: "Eve"}, nil},
	{"invalid user name", "NotExisting", func(mur *mockUserRepository) { mur.findByNameFunc = func(ctx context.Context, name string) (*model.User, error) { return nil, sql.ErrNoRows } }, nil, nil},
}

func TestFindByName(t *testing.T) {
	for _, tt := range findByNameTests {
		t.Run(tt.name, func(t *testing.T) {
			mur := &mockUserRepository{}
			tt.mockBehavior(mur)
			us := newUserServiceImpl(mur)
			user, err := us.FindByName(context.Background(), tt.name)
			assert.Equal(t, tt.expectedUser, user)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}