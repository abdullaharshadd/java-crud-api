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

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartContact/internal/smartcontact/error/apperror"
	"github.com/smartContact/internal/smartcontact/handler"
	"github.com/smartContact/internal/smartcontact/model"
)

// ---------------------------------------------------------------------------
// Mock UserService
// ---------------------------------------------------------------------------

type mockUserService struct {
	saveUserFn        func(ctx context.Context, user *model.User) error
	fetchUserListFn   func(ctx context.Context) ([]*model.User, error)
	fetchUserByIDFn   func(ctx context.Context, id int) (*model.User, error)
	deleteUserFn      func(ctx context.Context, id int) error
	updateUserFn      func(ctx context.Context, id int, user *model.User) error
	getUserByNameFn   func(ctx context.Context, name string) (*model.User, error)
}

func (m *mockUserService) SaveUser(ctx context.Context, user *model.User) error {
	if m.saveUserFn != nil {
		return m.saveUserFn(ctx, user)
	}
	return nil
}

func (m *mockUserService) FetchUserList(ctx context.Context) ([]*model.User, error) {
	if m.fetchUserListFn != nil {
		return m.fetchUserListFn(ctx)
	}
	return nil, nil
}

func (m *mockUserService) FetchUserByID(ctx context.Context, id int) (*model.User, error) {
	if m.fetchUserByIDFn != nil {
		return m.fetchUserByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockUserService) DeleteUser(ctx context.Context, id int) error {
	if m.deleteUserFn != nil {
		return m.deleteUserFn(ctx, id)
	}
	return nil
}

func (m *mockUserService) UpdateUser(ctx context.Context, id int, user *model.User) error {
	if m.updateUserFn != nil {
		return m.updateUserFn(ctx, id, user)
	}
	return nil
}

func (m *mockUserService) GetUserByName(ctx context.Context, name string) (*model.User, error) {
	if m.getUserByNameFn != nil {
		return m.getUserByNameFn(ctx, name)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newRouter wires the controller onto a chi router and returns it, ready for
// httptest recording.
func newRouter(svc *mockUserService) chi.Router {
	ctrl := handler.NewUserController(svc, validator.New())
	r := chi.NewRouter()
	ctrl.RegisterRoutes(r)
	return r
}

// mustMarshal marshals v to JSON or fails the test immediately.
func mustMarshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

// decodeJSON decodes the response body into dst.
func decodeJSON(t *testing.T, body *bytes.Buffer, dst interface{}) {
	t.Helper()
	require.NoError(t, json.NewDecoder(body).Decode(dst))
}

// ---------------------------------------------------------------------------
// apperror stub — the package is imported so WriteError must be resolvable.
// If apperror.WriteError inspects the error for a BadRequest() bool method or
// a UserNotFound marker, we provide matching sentinel errors here.
// ---------------------------------------------------------------------------

// notFoundError is a sentinel that apperror maps to HTTP 404.
type notFoundError struct{ msg string }

func (e *notFoundError) Error() string    { return e.msg }
func (e *notFoundError) NotFound() bool   { return true }

// serviceError is a generic 500-level error from the service.
type serviceError struct{ msg string }

func (e *serviceError) Error() string { return e.msg }

// ---------------------------------------------------------------------------
// SaveUser tests  POST /save_user_data
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	type tc struct {
		name           string
		body           interface{}
		rawBody        string // used when body is invalid JSON
		useRawBody     bool
		setupMock      func(svc *mockUserService)
		wantStatus     int
		wantBodyExact  string
		wantBodySubstr string
		assertMock     func(t *testing.T, svc *mockUserService)
	}

	var savedUser *model.User

	tests := []tc{
		{
			name: "valid user body returns 200 with success message",
			body: model.User{
				// Populate required fields so that validator.Struct passes.
				// Adjust field names / tags to match your actual model.User.
				Name:  "Alice",
				Email: "alice@example.com",
			},
			setupMock: func(svc *mockUserService) {
				svc.saveUserFn = func(ctx context.Context, user *model.User) error {
					savedUser = user
					return nil
				}
			},
			wantStatus:    http.StatusOK,
			wantBodyExact: "User data saved successfully!",
			assertMock: func(t *testing.T, svc *mockUserService) {
				require.NotNil(t, savedUser, "UserService.SaveUser must have been called")
			},
		},
		{
			name:       "invalid JSON body returns 400",
			useRawBody: true,
			rawBody:    `{not valid json`,
			setupMock:  func(svc *mockUserService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error propagates",
			body: model.User{
				Name:  "Bob",
				Email: "bob@example.com",
			},
			setupMock: func(svc *mockUserService) {
				svc.saveUserFn = func(ctx context.Context, user *model.User) error {
					return &serviceError{"db error"}
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{}
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}
			router := newRouter(svc)

			var bodyReader *strings.Reader
			if tt.useRawBody {
				bodyReader = strings.NewReader(tt.rawBody)
			} else {
				bodyReader = strings.NewReader(string(mustMarshal(t, tt.body)))
			}

			req := httptest.NewRequest(http.MethodPost, "/save_user_data", bodyReader)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.wantBodyExact != "" {
				assert.Equal(t, tt.wantBodyExact, strings.TrimSpace(rr.Body.String()))
			}
			if tt.wantBodySubstr != "" {
				assert.Contains(t, rr.Body.String(), tt.wantBodySubstr)
			}
			if tt.assertMock != nil {
				tt.assertMock(t, svc)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserList tests  GET /get_user_data
// ---------------------------------------------------------------------------

func TestFetchUserList(t *testing.T) {
	type tc struct {
		name       string
		setupMock  func(svc *mockUserService)
		wantStatus int
		checkBody  func(t *testing.T, body *bytes.Buffer)
	}

	tests := []tc{
		{
			name: "returns list of users when users exist",
			setupMock: func(svc *mockUserService) {
				svc.fetchUserListFn = func(ctx context.Context) ([]*model.User, error) {
					return []*model.User{
						{Name: "Alice", Email: "alice@example.com"},
						{Name: "Bob", Email: "bob@example.com"},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body *bytes.Buffer) {
				var users []*model.User
				decodeJSON(t, body, &users)
				assert.Len(t, users, 2)
				assert.Equal(t, "Alice", users[0].Name)
				assert.Equal(t, "Bob", users[1].Name)
			},
		},
		{
			name: "returns empty list when no users exist",
			setupMock: func(svc *mockUserService) {
				svc.fetchUserListFn = func(ctx context.Context) ([]*model.User, error) {
					return []*model.User{}, nil
				}
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body *bytes.Buffer) {
				var users []*model.User
				decodeJSON(t, body, &users)
				assert.Empty(t, users)
			},
		},
		{
			name: "service error propagates",
			setupMock: func(svc *mockUserService) {
				svc.fetchUserListFn = func(ctx context.Context) ([]*model.User, error) {
					return nil, &serviceError{"db error"}
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{}
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}
			router := newRouter(svc)

			req := httptest.NewRequest(http.MethodGet, "/get_user_data", nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.checkBody != nil {
				tt.checkBody(t, rr.Body)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserByID tests  GET /get_user_data/{id}
// ---------------------------------------------------------------------------

func TestFetchUserByID(t *testing.T) {
	type tc struct {
		name       string
		pathID     string
		setupMock  func(svc *mockUserService)
		wantStatus int
		checkBody  func(t *testing.T, body *bytes.Buffer)
	}

	tests := []tc{
		{
			name:   "valid id with existing user returns 200",
			pathID: "42",
			setupMock: func(svc *mockUserService) {
				svc.fetchUserByIDFn = func(ctx context.Context, id int) (*model.User, error) {
					assert.Equal(t, 42, id)
					return &model.User{Name: "Alice", Email: "alice@example.com"}, nil
				}
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body *bytes.Buffer) {
				var u model.User
				decodeJSON(t, body, &u)
				assert.Equal(t, "Alice", u.Name)
			},
		},
		{
			name:   "non-numeric id returns 400",
			pathID: "abc",
			setupMock: func(svc *mockUserService) {
				// Should never be called.
				svc.fetchUserByIDFn = func(ctx context.Context, id int) (*model.User, error) {
					t.Fatal("FetchUserByID should not have been called")
					return nil, nil
				}
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "id with no matching user returns 404",
			pathID: "99",
			setupMock: func(svc *mockUserService) {
				svc.fetchUserByIDFn = func(ctx context.Context, id int) (*model.User, error) {
					return nil, &notFoundError{"user not found"}
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{}
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}
			router := newRouter(svc)

			req := httptest.NewRequest(http.MethodGet, "/get_user_data/"+tt.pathID, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.checkBody != nil {
				tt.checkBody(t, rr.Body)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteUser tests  DELETE /delete_user_data/{id}
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	type tc struct {
		name           string
		pathID         string
		setupMock      func(svc *mockUserService)
		wantStatus     int
		wantBodyExact  string
		assertMock     func(t *testing.T, svc *mockUserService)
	}

	var deletedID int

	tests := []tc{
		{
			name:   "valid id deletes user and returns 200 with success message",
			pathID: "7",
			setupMock: func(svc *mockUserService) {
				svc.deleteUserFn = func(ctx context.Context, id int) error {
					deletedID = id
					return nil
				}
			},
			wantStatus:    http.StatusOK,
			wantBodyExact: "user data deleted Successfully",
			assertMock: func(t *testing.T, svc *mockUserService) {
				assert.Equal(t, 7, deletedID)
			},
		},
		{
			name:   "non-numeric id returns 400",
			pathID: "xyz",
			setupMock: func(svc *mockUserService) {
				svc.deleteUserFn = func(ctx context.Context, id int) error {
					t.Fatal("DeleteUser should not have been called")
					return nil
				}
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "service error for nonexistent id propagates",
			pathID: "99",
			setupMock: func(svc *mockUserService) {
				svc.deleteUserFn = func(ctx context.Context, id int) error {
					return &notFoundError{"user not found"}
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{}
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}
			router := newRouter(svc)

			req := httptest.NewRequest(http.MethodDelete, "/delete_user_data/"+tt.pathID,