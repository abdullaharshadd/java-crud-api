```go
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperr "github.com/smartContact/internal/smartcontact/error/apperror"
	"github.com/smartContact/internal/smartcontact/handler"
	"github.com/smartContact/internal/smartcontact/model"
)

// ---------------------------------------------------------------------------
// Mock service
// ---------------------------------------------------------------------------

type mockUserService struct {
	saveUserFn          func(ctx context.Context, u *model.User) error
	getAllUsersFn        func(ctx context.Context) ([]*model.User, error)
	fetchUserByIdFn     func(ctx context.Context, id int) (*model.User, error)
	deleteUserFn        func(ctx context.Context, id int) error
	updateUserFn        func(ctx context.Context, id int, u *model.User) error
	getUserNameByNameFn func(ctx context.Context, name string) (*model.User, error)
}

func (m *mockUserService) SaveUser(ctx context.Context, u *model.User) error {
	if m.saveUserFn != nil {
		return m.saveUserFn(ctx, u)
	}
	return nil
}

func (m *mockUserService) GetAllUsers(ctx context.Context) ([]*model.User, error) {
	if m.getAllUsersFn != nil {
		return m.getAllUsersFn(ctx)
	}
	return nil, nil
}

func (m *mockUserService) FetchUserById(ctx context.Context, id int) (*model.User, error) {
	if m.fetchUserByIdFn != nil {
		return m.fetchUserByIdFn(ctx, id)
	}
	return nil, nil
}

func (m *mockUserService) DeleteUser(ctx context.Context, id int) error {
	if m.deleteUserFn != nil {
		return m.deleteUserFn(ctx, id)
	}
	return nil
}

func (m *mockUserService) UpdateUser(ctx context.Context, id int, u *model.User) error {
	if m.updateUserFn != nil {
		return m.updateUserFn(ctx, id, u)
	}
	return nil
}

func (m *mockUserService) GetUserNameByName(ctx context.Context, name string) (*model.User, error) {
	if m.getUserNameByNameFn != nil {
		return m.getUserNameByNameFn(ctx, name)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// Helper — build a mux with routes registered
// ---------------------------------------------------------------------------

func newMux(svc *mockUserService) *http.ServeMux {
	mux := http.NewServeMux()
	h := handler.NewUserHandler(svc)
	h.RegisterRoutes(mux)
	return mux
}

// ---------------------------------------------------------------------------
// SaveUser tests
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	type tc struct {
		name           string
		body           string
		setupSvc       func(*mockUserService)
		wantStatus     int
		wantBodyContains string
	}

	tests := []tc{
		{
			name: "valid payload — 200 and success message",
			body: `{"id":1,"name":"Alice","email":"alice@example.com"}`,
			setupSvc: func(m *mockUserService) {
				m.saveUserFn = func(_ context.Context, u *model.User) error {
					return nil
				}
			},
			wantStatus:       http.StatusOK,
			wantBodyContains: "User data saved successfully!",
		},
		{
			name:             "invalid JSON body — 400",
			body:             `{not-json`,
			setupSvc:         func(m *mockUserService) {},
			wantStatus:       http.StatusBadRequest,
			wantBodyContains: "",
		},
		{
			name: "validation failure — 400",
			// model.User.Validate() should return error for empty/missing required fields
			body: `{}`,
			setupSvc: func(m *mockUserService) {
				// service should NOT be called
				m.saveUserFn = func(_ context.Context, u *model.User) error {
					t.Error("saveUser should not be called when validation fails")
					return nil
				}
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error — propagated as error response",
			body: `{"id":2,"name":"Bob","email":"bob@example.com"}`,
			setupSvc: func(m *mockUserService) {
				m.saveUserFn = func(_ context.Context, u *model.User) error {
					return errors.New("db failure")
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}

			mux := newMux(svc)
			req := httptest.NewRequest(http.MethodPost, "/save_user_data", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code, "unexpected status code")
			if tt.wantBodyContains != "" {
				assert.Contains(t, rec.Body.String(), tt.wantBodyContains)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserList tests
// ---------------------------------------------------------------------------

func TestFetchUserList(t *testing.T) {
	type tc struct {
		name       string
		setupSvc   func(*mockUserService)
		wantStatus int
		wantUsers  []*model.User
	}

	tests := []tc{
		{
			name: "returns list of users — 200",
			setupSvc: func(m *mockUserService) {
				m.getAllUsersFn = func(_ context.Context) ([]*model.User, error) {
					return []*model.User{
						{ID: 1, Name: "Alice", Email: "alice@example.com"},
						{ID: 2, Name: "Bob", Email: "bob@example.com"},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantUsers: []*model.User{
				{ID: 1, Name: "Alice", Email: "alice@example.com"},
				{ID: 2, Name: "Bob", Email: "bob@example.com"},
			},
		},
		{
			name: "no users — returns empty list 200",
			setupSvc: func(m *mockUserService) {
				m.getAllUsersFn = func(_ context.Context) ([]*model.User, error) {
					return []*model.User{}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantUsers:  []*model.User{},
		},
		{
			name: "service error — propagated",
			setupSvc: func(m *mockUserService) {
				m.getAllUsersFn = func(_ context.Context) ([]*model.User, error) {
					return nil, errors.New("db error")
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}

			mux := newMux(svc)
			req := httptest.NewRequest(http.MethodGet, "/get_user_data", nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantUsers != nil {
				assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
				var got []*model.User
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
				assert.Equal(t, tt.wantUsers, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserById tests
// ---------------------------------------------------------------------------

func TestFetchUserById(t *testing.T) {
	type tc struct {
		name       string
		pathID     string
		setupSvc   func(*mockUserService)
		wantStatus int
		wantUser   *model.User
	}

	tests := []tc{
		{
			name:   "existing user — 200",
			pathID: "1",
			setupSvc: func(m *mockUserService) {
				m.fetchUserByIdFn = func(_ context.Context, id int) (*model.User, error) {
					assert.Equal(t, 1, id)
					return &model.User{ID: 1, Name: "Alice", Email: "alice@example.com"}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantUser:   &model.User{ID: 1, Name: "Alice", Email: "alice@example.com"},
		},
		{
			name:   "user not found — propagates UserNotFoundException",
			pathID: "999",
			setupSvc: func(m *mockUserService) {
				m.fetchUserByIdFn = func(_ context.Context, id int) (*model.User, error) {
					return nil, apperr.NewNotFoundError("user not found")
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "non-integer id — 400",
			pathID:     "abc",
			setupSvc:   func(m *mockUserService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "service error — propagated",
			pathID: "2",
			setupSvc: func(m *mockUserService) {
				m.fetchUserByIdFn = func(_ context.Context, id int) (*model.User, error) {
					return nil, errors.New("db error")
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}

			mux := newMux(svc)
			req := httptest.NewRequest(http.MethodGet, "/get_user_data/"+tt.pathID, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantUser != nil {
				var got model.User
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
				assert.Equal(t, *tt.wantUser, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteUser tests
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	type tc struct {
		name             string
		pathID           string
		setupSvc         func(*mockUserService)
		wantStatus       int
		wantBodyContains string
	}

	tests := []tc{
		{
			name:   "successful deletion — 200 and success message",
			pathID: "1",
			setupSvc: func(m *mockUserService) {
				m.deleteUserFn = func(_ context.Context, id int) error {
					assert.Equal(t, 1, id)
					return nil
				}
			},
			wantStatus:       http.StatusOK,
			wantBodyContains: "user data deleted Successfully",
		},
		{
			name:       "non-integer id — 400",
			pathID:     "xyz",
			setupSvc:   func(m *mockUserService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "service error propagated",
			pathID: "2",
			setupSvc: func(m *mockUserService) {
				m.deleteUserFn = func(_ context.Context, id int) error {
					return errors.New("delete failed")
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:   "user not found — service returns not-found error",
			pathID: "42",
			setupSvc: func(m *mockUserService) {
				m.deleteUserFn = func(_ context.Context, id int) error {
					return apperr.NewNotFoundError("user not found")
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}

			mux := newMux(svc)
			req := httptest.NewRequest(http.MethodDelete, "/delete_user_data/"+tt.pathID, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantBodyContains != "" {
				assert.Contains(t, rec.Body.String(), tt.wantBodyContains)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// UpdateUser tests
// ---------------------------------------------------------------------------

func TestUpdateUser(t *testing.T) {
	type tc struct {
		name       string
		pathID     string
		body       string
		setupSvc   func(*mockUserService)
		wantStatus int
		wantUser   *model.User
	}

	tests := []tc{
		{
			name:   "successful update — returns request payload 200",
			pathID: "1",
			body:   `{"id":1,"name":"Alice Updated","email":"alice@example.com"}`,
			setupSvc: func(m *mockUserService) {
				m.updateUserFn = func(_ context.Context, id int, u *model.User) error {
					assert.Equal(t, 1, id)
					assert.Equal(t, "Alice Updated", u.Name)
					return nil
				}
			},
			wantStatus: http.StatusOK,
			wantUser:   &model.User{ID: 1, Name: "Alice Updated", Email: "alice@example.com"},
		},
		{
			name:       "non-integer id — 400",
			pathID:     "notanid",
			body:       `{"id":1,"name":"Alice"}`,
			setupSvc:   func(m *mockUserService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid JSON body — 400",
			pathID:     "1",
			body:       `{badjson`,
			setupSvc:   func(m *mockUserService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "service error — propagated",
			pathID: "1",
			body:   `{"id":1,"name":"Alice","email":"alice@example.com"}`,
			setupSvc: func(m *mockUserService) {
				m.updateUserFn = func(_ context.Context, id int, u *model.User) error {
					return errors.New("update failed")
				}