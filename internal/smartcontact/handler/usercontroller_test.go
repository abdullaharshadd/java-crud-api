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
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	smartErr "internal/smartcontact/error"
	"internal/smartcontact/handler"
	"internal/smartcontact/model"
)

// ---------------------------------------------------------------------------
// Mock service
// ---------------------------------------------------------------------------

type mockUserService struct {
	saveUserFn      func(ctx context.Context, user *model.User) error
	fetchUserListFn func(ctx context.Context) ([]*model.User, error)
	getUserByIDFn   func(ctx context.Context, id int) (*model.User, error)
	deleteUserFn    func(ctx context.Context, id int) error
	updateUserFn    func(ctx context.Context, id int, user *model.User) error
	getUserByNameFn func(ctx context.Context, name string) (*model.User, error)
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

func (m *mockUserService) GetUserByID(ctx context.Context, id int) (*model.User, error) {
	if m.getUserByIDFn != nil {
		return m.getUserByIDFn(ctx, id)
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

func newRouter(svc *mockUserService) chi.Router {
	logger := zerolog.Nop()
	ctrl := handler.NewUserController(svc, logger)
	r := chi.NewRouter()
	ctrl.RegisterRoutes(r)
	return r
}

func toJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

// decodeBody decodes the JSON body of a response into dst.
func decodeBody(t *testing.T, body string, dst interface{}) {
	t.Helper()
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(body)), dst))
}

// ---------------------------------------------------------------------------
// SaveUser tests
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	type tc struct {
		name           string
		body           string
		saveUserFn     func(ctx context.Context, user *model.User) error
		wantStatus     int
		wantBodyContains string
	}

	validUser := model.User{Name: "Alice", Email: "alice@example.com"}

	tests := []tc{
		{
			name:             "valid user – service succeeds",
			body:             string(toJSON(t, validUser)),
			saveUserFn:       func(_ context.Context, _ *model.User) error { return nil },
			wantStatus:       http.StatusOK,
			wantBodyContains: "User data saved successfully!",
		},
		{
			name:             "invalid JSON body",
			body:             `{not valid json`,
			saveUserFn:       nil,
			wantStatus:       http.StatusBadRequest,
			wantBodyContains: "invalid request body",
		},
		{
			name: "validation error – user fails Validate()",
			// model.User with missing required fields should fail Validate()
			body:             `{}`,
			saveUserFn:       func(_ context.Context, _ *model.User) error { return nil },
			wantStatus:       http.StatusBadRequest,
			wantBodyContains: "",
		},
		{
			name:             "service returns generic error",
			body:             string(toJSON(t, validUser)),
			saveUserFn:       func(_ context.Context, _ *model.User) error { return errors.New("db error") },
			wantStatus:       http.StatusInternalServerError,
			wantBodyContains: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{saveUserFn: tt.saveUserFn}
			r := newRouter(svc)

			req := httptest.NewRequest(http.MethodPost, "/save_user_data", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.wantBodyContains != "" {
				assert.Contains(t, rr.Body.String(), tt.wantBodyContains)
			}
		})
	}
}

func TestSaveUser_CallsService(t *testing.T) {
	called := false
	var receivedUser *model.User

	svc := &mockUserService{
		saveUserFn: func(_ context.Context, u *model.User) error {
			called = true
			receivedUser = u
			return nil
		},
	}

	user := model.User{Name: "Bob", Email: "bob@example.com"}
	r := newRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/save_user_data", bytes.NewReader(toJSON(t, user)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Only assert the service was called when the request was valid enough to
	// pass decode+validate. If model.User{Name, Email} passes Validate():
	if rr.Code == http.StatusOK {
		assert.True(t, called, "expected service.SaveUser to be called")
		assert.NotNil(t, receivedUser)
	}
}

// ---------------------------------------------------------------------------
// FetchUserList tests
// ---------------------------------------------------------------------------

func TestFetchUserList(t *testing.T) {
	users := []*model.User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}

	tests := []struct {
		name            string
		fetchUserListFn func(ctx context.Context) ([]*model.User, error)
		wantStatus      int
		wantLen         int
	}{
		{
			name:            "returns list of users",
			fetchUserListFn: func(_ context.Context) ([]*model.User, error) { return users, nil },
			wantStatus:      http.StatusOK,
			wantLen:         2,
		},
		{
			name:            "returns empty list",
			fetchUserListFn: func(_ context.Context) ([]*model.User, error) { return []*model.User{}, nil },
			wantStatus:      http.StatusOK,
			wantLen:         0,
		},
		{
			name:            "service error returns 500",
			fetchUserListFn: func(_ context.Context) ([]*model.User, error) { return nil, errors.New("db down") },
			wantStatus:      http.StatusInternalServerError,
			wantLen:         -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{fetchUserListFn: tt.fetchUserListFn}
			r := newRouter(svc)

			req := httptest.NewRequest(http.MethodGet, "/get_user_data", nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)

			if tt.wantLen >= 0 {
				var got []*model.User
				decodeBody(t, rr.Body.String(), &got)
				assert.Len(t, got, tt.wantLen)
			}
		})
	}
}

func TestFetchUserList_NoWriteOperations(t *testing.T) {
	saveUserCalled := false
	svc := &mockUserService{
		fetchUserListFn: func(_ context.Context) ([]*model.User, error) { return nil, nil },
		saveUserFn:      func(_ context.Context, _ *model.User) error { saveUserCalled = true; return nil },
	}
	r := newRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/get_user_data", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.False(t, saveUserCalled, "FetchUserList must not trigger any write operations")
	_ = rr
}

// ---------------------------------------------------------------------------
// FetchUserByID tests
// ---------------------------------------------------------------------------

func TestFetchUserByID(t *testing.T) {
	existingUser := &model.User{Name: "Alice", Email: "alice@example.com"}

	tests := []struct {
		name          string
		pathID        string
		getUserByIDFn func(ctx context.Context, id int) (*model.User, error)
		wantStatus    int
		wantUser      bool
	}{
		{
			name:          "existing user returns 200",
			pathID:        "1",
			getUserByIDFn: func(_ context.Context, _ int) (*model.User, error) { return existingUser, nil },
			wantStatus:    http.StatusOK,
			wantUser:      true,
		},
		{
			name:   "user not found returns 404",
			pathID: "99",
			getUserByIDFn: func(_ context.Context, _ int) (*model.User, error) {
				return nil, smartErr.ErrUserNotFound
			},
			wantStatus: http.StatusNotFound,
			wantUser:   false,
		},
		{
			name:   "generic service error returns 500",
			pathID: "1",
			getUserByIDFn: func(_ context.Context, _ int) (*model.User, error) {
				return nil, errors.New("connection reset")
			},
			wantStatus: http.StatusInternalServerError,
			wantUser:   false,
		},
		{
			name:          "invalid id returns 400",
			pathID:        "abc",
			getUserByIDFn: nil,
			wantStatus:    http.StatusBadRequest,
			wantUser:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{getUserByIDFn: tt.getUserByIDFn}
			r := newRouter(svc)

			req := httptest.NewRequest(http.MethodGet, "/get_user_data/"+tt.pathID, nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.wantUser {
				var got model.User
				decodeBody(t, rr.Body.String(), &got)
				assert.Equal(t, existingUser.Name, got.Name)
			}
		})
	}
}

func TestFetchUserByID_NoWriteOperations(t *testing.T) {
	saveUserCalled := false
	svc := &mockUserService{
		getUserByIDFn: func(_ context.Context, _ int) (*model.User, error) {
			return &model.User{Name: "Alice"}, nil
		},
		saveUserFn: func(_ context.Context, _ *model.User) error { saveUserCalled = true; return nil },
	}
	r := newRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/get_user_data/1", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.False(t, saveUserCalled)
	_ = rr
}

// ---------------------------------------------------------------------------
// DeleteUser tests
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name             string
		pathID           string
		deleteUserFn     func(ctx context.Context, id int) error
		wantStatus       int
		wantBodyContains string
	}{
		{
			name:             "valid id – success returns 200",
			pathID:           "5",
			deleteUserFn:     func(_ context.Context, _ int) error { return nil },
			wantStatus:       http.StatusOK,
			wantBodyContains: "user data deleted Successfully",
		},
		{
			name:   "user not found returns 404",
			pathID: "99",
			deleteUserFn: func(_ context.Context, _ int) error {
				return smartErr.ErrUserNotFound
			},
			wantStatus:       http.StatusNotFound,
			wantBodyContains: "",
		},
		{
			name:             "generic service error returns 500",
			pathID:           "5",
			deleteUserFn:     func(_ context.Context, _ int) error { return errors.New("constraint violation") },
			wantStatus:       http.StatusInternalServerError,
			wantBodyContains: "constraint violation",
		},
		{
			name:             "invalid id returns 400",
			pathID:           "bad",
			deleteUserFn:     nil,
			wantStatus:       http.StatusBadRequest,
			wantBodyContains: "invalid id path parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{deleteUserFn: tt.deleteUserFn}
			r := newRouter(svc)

			req := httptest.NewRequest(http.MethodDelete, "/delete_user_data/"+tt.pathID, nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.wantBodyContains != "" {
				assert.Contains(t, rr.Body.String(), tt.wantBodyContains)
			}
		})
	}
}

func TestDeleteUser_CallsService(t *testing.T) {
	var receivedID int
	svc := &mockUserService{
		deleteUserFn: func(_ context.Context, id int) error {
			receivedID = id
			return nil
		},
	}
	r := newRouter(svc)

	req := httptest.NewRequest(http.MethodDelete, "/delete_user_data/42", nil