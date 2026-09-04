package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/repository"
)

type mockUserRepository struct {
	repository.UserRepository
	saveFunc      func(context.Context, *model.User) (*model.User, error)
	fetchAllFunc  func(context.Context) ([]*model.User, error)
	findByIDFunc  func(context.Context, int) (*model.User, error)
	deleteFunc    func(context.Context, int) error
	updateFunc    func(context.Context, int, *model.User) error
	findByNameFunc func(context.Context, string) (*model.User, error)
}

func (mur *mockUserRepository) Save(ctx context.Context, user *model.User) (*model.User, error) {
	if mur.saveFunc != nil {
		return mur.saveFunc(ctx, user)
	}
	return nil, nil
}

func (mur *mockUserRepository) FetchAll(ctx context.Context) ([]*model.User, error) {
	if mur.fetchAllFunc != nil {
		return mur.fetchAllFunc(ctx)
	}
	return nil, nil
}

func (mur *mockUserRepository) FindByID(ctx context.Context, id int) (*model.User, error) {
	if mur.findByIDFunc != nil {
		return mur.findByIDFunc(ctx, id)
	}
	return nil, nil
}

func (mur *mockUserRepository) Delete(ctx context.Context, id int) error {
	if mur.deleteFunc != nil {
		return mur.deleteFunc(ctx, id)
	}
	return nil
}

func (mur *mockUserRepository) Update(ctx context.Context, id int, user *model.User) error {
	if mur.updateFunc != nil {
		return mur.updateFunc(ctx, id, user)
	}
	return nil
}

func (mur *mockUserRepository) FindByName(ctx context.Context, name string) (*model.User, error) {
	if mur.findByNameFunc != nil {
		return mur.findByNameFunc(ctx, name)
	}
	return nil, nil
}

func TestSaveUser(t *testing.T) {
	type args struct {
		ctx  context.Context
		user *model.User
	}
	tests := []struct {
		name     string
		args     args
		mockFunc func(mur *mockUserRepository)
		want     *model.User
		wantErr  bool
	}{
		{
			name: "valid user",
			args: args{
				ctx: context.Background(),
				user: &model.User{
					Name:  "John Doe",
					Email: "john.doe@example.com",
				},
			},
			mockFunc: func(mur *mockUserRepository) {
				mur.saveFunc = func(ctx context.Context, user *model.User) (*model.User, error) {
					user.ID = 1 // Simulate auto-generated ID
					return user, nil
				}
			},
			want: &model.User{
				ID:    1,
				Name:  "John Doe",
				Email: "john.doe@example.com",
			},
			wantErr: false,
		},
		{
			name: "invalid user",
			args: args{
				ctx: context.Background(),
				user: &model.User{
					Name:  "",
					Email: "john.doe@example.com",
				},
			},
			mockFunc: func(mur *mockUserRepository) {},
			want:     nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mur := &mockUserRepository{}
			if tt.mockFunc != nil {
				tt.mockFunc(mur)
			}
			us := newUserService(mur)
			got, err := us.SaveUser(tt.args.ctx, tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserService.SaveUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFetchUserList(t *testing.T) {
	tests := []struct {
		name     string
		mockFunc func(mur *mockUserRepository)
		want     []*model.User
		wantErr  bool
	}{
		{
			name: "empty list",
			mockFunc: func(mur *mockUserRepository) {
				mur.fetchAllFunc = func(ctx context.Context) ([]*model.User, error) {
					return []*model.User{}, nil
				}
			},
			want:    []*model.User{},
			wantErr: false,
		},
		{
			name: "multiple users",
			mockFunc: func(mur *mockUserRepository) {
				mur.fetchAllFunc = func(ctx context.Context) ([]*model.User, error) {
					return []*model.User{
						{ID: 1, Name: "John Doe", Email: "john.doe@example.com"},
						{ID: 2, Name: "Jane Doe", Email: "jane.doe@example.com"},
					}, nil
				}
			},
			want: []*model.User{
				{ID: 1, Name: "John Doe", Email: "john.doe@example.com"},
				{ID: 2, Name: "Jane Doe", Email: "jane.doe@example.com"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mur := &mockUserRepository{}
			if tt.mockFunc != nil {
				tt.mockFunc(mur)
			}
			us := newUserService(mur)
			got, err := us.FetchUserList(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("UserService.FetchUserList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFetchUserByID(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name     string
		args     args
		mockFunc func(mur *mockUserRepository)
		want     *model.User
		wantErr  bool
	}{
		{
			name: "valid user ID",
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			mockFunc: func(mur *mockUserRepository) {
				mur.findByIDFunc = func(ctx context.Context, id int) (*model.User, error) {
					return &model.User{ID: 1, Name: "John Doe", Email: "john.doe@example.com"}, nil
				}
			},
			want: &model.User{ID: 1, Name: "John Doe", Email: "john.doe@example.com"},
			wantErr: false,
		},
		{
			name: "non-existent user ID",
			args: args{
				ctx: context.Background(),
				id:  3,
			},
			mockFunc: func(mur *mockUserRepository) {
				mur.findByIDFunc = func(ctx context.Context, id int) (*model.User, error) {
					return nil, sql.ErrNoRows
				}
			},
			want: nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mur := &mockUserRepository{}
			if tt.mockFunc != nil {
				tt.mockFunc(mur)
			}
			us := newUserService(mur)
			got, err := us.FetchUserByID(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserService.FetchUserByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDeleteUser(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name     string
		args     args
		mockFunc func(mur *mockUserRepository)
		wantErr  bool
	}{
		{
			name: "valid user ID",
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			mockFunc: func(mur *mockUserRepository) {
				mur.deleteFunc = func(ctx context.Context, id int) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "non-existent user ID",
			args: args{
				ctx: context.Background(),
				id:  3,
			},
			mockFunc: func(mur *mockUserRepository) {
				mur.deleteFunc = func(ctx context.Context, id int) error {
					return error.NewUserNotFoundError("User not found", nil)
				}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mur := &mockUserRepository{}
			if tt.mockFunc != nil {
				tt.mockFunc(mur)
			}
			us := newUserService(mur)
			err := us.DeleteUser(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserService.DeleteUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	type args struct {
		ctx  context.Context
		id   int
		user *model.User
	}
	tests := []struct {
		name     string
		args     args
		mockFunc func(mur *mockUserRepository)
		wantErr  bool
	}{
		{
			name: "valid user ID and updated user",
			args: args{
				ctx: context.Background(),
				id:  1,
				user: &model.User{
					Name:  "John Doe",
					Email: "john.doe@newemail.com",
				},
			},
			mockFunc: func(mur *mockUserRepository) {
				mur.updateFunc = func(ctx context.Context, id int, user *model.User) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "non-existent user ID",
			args: args{
				ctx: context.Background(),
				id:  3,
				user: &model.User{
					Name:  "John Doe",
					Email: "john.doe@newemail.com",
				},
			},
			mockFunc: func(mur *mockUserRepository) {
				mur.updateFunc = func(ctx context.Context, id int, user *model.User) error {
					return error.NewUserNotFoundError("User not found", nil)
				}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mur := &mockUserRepository{}
			if tt.mockFunc != nil {
				tt.mockFunc(mur)
			}
			us := newUserService(mur)
			err := us.UpdateUser(tt.args.ctx, tt.args.id, tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserService.UpdateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFindByName(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name     string
		args     args
		mockFunc func(mur *mockUserRepository)
		want     *model.User
		wantErr  bool
	}{
		{
			name: "valid user name",
			args: args{
				ctx:  context.Background(),
				name: "John Doe",
			},
			mockFunc: func(mur *mockUserRepository) {
				mur.findByNameFunc = func(ctx context.Context, name string) (*model.User, error) {
					return &model.User{ID: 1, Name: "John Doe", Email: "john.doe@example.com"}, nil
				}
			},
			want: &model.User{ID: 1, Name: "John Doe", Email: "john.doe@example.com"},
			wantErr: false,
		},
		{
			name: "non-existent user name",
			args: args{
				ctx:  context.Background(),
				name: "Non Existent",
			},
			mockFunc: func(mur *mockUserRepository) {
				mur.findByNameFunc = func(ctx context.Context, name string) (*model.User, error) {
					return nil, nil
				}
			},
			want: nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mur := &mockUserRepository{}
			if tt.mockFunc != nil {
				tt.mockFunc(mur)
			}
			us := newUserService(mur)
			got, err := us.FindByName(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserService.FindByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}