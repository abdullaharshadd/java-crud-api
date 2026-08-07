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

	smartcontacterror "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/handler"
	"github.com/smartContact/internal/smartcontact/model"
)

// ---------------------------------------------------------------------------
// Mock service
// ---------------------------------------------------------------------------

type mockUserService struct {
	saveUserFn      func(ctx context.Context, user *model.User) (*model.User, error)
	fetchUserListFn func(ctx context.Context) ([]*model.User, error)
	fetchUserByIDFn func(ctx context.Context, id int) (*model.User, error)
	deleteUserFn    func(ctx context.Context, id int) error
	updateUserFn    func(ctx context.Context, id int, user *model.User) (*model.User, error)
	getUserByNameFn func(ctx context.Context, name string) (*model.User, error)
}

func (m *mockUserService) SaveUser(ctx context.Context, user *model.User) (*model.User, error) {
	if m.saveUserFn != nil {
		return m.saveUserFn(ctx, user)
	}
	return user, nil
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

func (m *mockUserService) UpdateUser(ctx context.Context, id int, user *model.User) (*model.User, error) {
	if m.updateUserFn != nil {
		return m.updateUserFn(ctx, id, user)
	}
	return user, nil
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

func newTestMux(svc *mockUserService) *http.ServeMux {
	mux := http.NewServeMux()
	ctrl := handler.NewUserController(svc)
	ctrl.RegisterRoutes(mux)
	return mux
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

// ---------------------------------------------------------------------------
// SaveUser  POST /save_user_data
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	validUser := &model.User{
		Name:  "Alice",
		Email: "alice@example.com",
	}

	tests := []struct {
		name           string
		body           string
		svc            *mockUserService
		wantStatus     int
		wantBodyContains string
		wantContentType  string
	}{
		{
			name: "valid user – returns 200 plain-text success message",
			body: string(mustMarshal(t, validUser)),
			svc: &mockUserService{
				saveUserFn: func(_ context.Context, u *model.User) (*model.User, error) {
					return u, nil
				},
			},
			wantStatus:       http.StatusOK,
			wantBodyContains: "User data saved successfully!",
			wantContentType:  "text/plain; charset=utf-8",
		},
		{
			name: "invalid JSON body – returns 400",
			body: `{bad json`,
			svc:  &mockUserService{},
			wantStatus:       http.StatusBadRequest,
			wantBodyContains: "",
			wantContentType:  "",
		},
		{
			name: "validation failure – returns 400",
			// Assuming an empty Name triggers Validate()
			body: `{"email":"no-name@example.com"}`,
			svc:  &mockUserService{},
			wantStatus:       http.StatusBadRequest,
			wantBodyContains: "",
			wantContentType:  "",
		},
		{
			name: "service returns domain error – non-200",
			body: string(mustMarshal(t, validUser)),
			svc: &mockUserService{
				saveUserFn: func(_ context.Context, u *model.User) (*model.User, error) {
					return nil, smartcontacterror.ErrUserNotFound
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mux := newTestMux(tc.svc)
			req := httptest.NewRequest(http.MethodPost, "/save_user_data",
				strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatus, rr.Code)
			if tc.wantBodyContains != "" {
				assert.Contains(t, rr.Body.String(), tc.wantBodyContains)
			}
			if tc.wantContentType != "" {
				assert.Equal(t, tc.wantContentType, rr.Header().Get("Content-Type"))
			}
		})
	}
}

// SaveUser invariant: response body NEVER contains serialised user data.
func TestSaveUser_NeverReturnsUserData(t *testing.T) {
	user := &model.User{Name: "Alice", Email: "alice@example.com"}
	svc := &mockUserService{
		saveUserFn: func(_ context.Context, u *model.User) (*model.User, error) {
			return u, nil
		},
	}
	mux := newTestMux(svc)
	req := httptest.NewRequest(http.MethodPost, "/save_user_data",
		bytes.NewReader(mustMarshal(t, user)))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	body := rr.Body.String()
	assert.NotContains(t, body, "Alice")
	assert.NotContains(t, body, "alice@example.com")
}

// ---------------------------------------------------------------------------
// FetchUserList  GET /get_user_data
// ---------------------------------------------------------------------------

func TestFetchUserList(t *testing.T) {
	users := []*model.User{
		{ID: 1, Name: "Alice", Email: "alice@example.com"},
		{ID: 2, Name: "Bob", Email: "bob@example.com"},
	}

	tests := []struct {
		name       string
		svc        *mockUserService
		wantStatus int
		wantLen    int
	}{
		{
			name: "users exist – returns 200 JSON array",
			svc: &mockUserService{
				fetchUserListFn: func(_ context.Context) ([]*model.User, error) {
					return users, nil
				},
			},
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
		{
			name: "no users – returns 200 empty JSON array",
			svc: &mockUserService{
				fetchUserListFn: func(_ context.Context) ([]*model.User, error) {
					return []*model.User{}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantLen:    0,
		},
		{
			name: "service error – non-200",
			svc: &mockUserService{
				fetchUserListFn: func(_ context.Context) ([]*model.User, error) {
					return nil, errors.New("internal error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mux := newTestMux(tc.svc)
			req := httptest.NewRequest(http.MethodGet, "/get_user_data", nil)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatus, rr.Code)
			if tc.wantStatus == http.StatusOK {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
				var got []*model.User
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &got))
				assert.Len(t, got, tc.wantLen)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserByID  GET /get_user_data/{id}
// ---------------------------------------------------------------------------

func TestFetchUserByID(t *testing.T) {
	alice := &model.User{ID: 1, Name: "Alice", Email: "alice@example.com"}

	tests := []struct {
		name       string
		pathID     string
		svc        *mockUserService
		wantStatus int
		wantUserID int
	}{
		{
			name:   "existing user – returns 200 with user",
			pathID: "1",
			svc: &mockUserService{
				fetchUserByIDFn: func(_ context.Context, id int) (*model.User, error) {
					if id == 1 {
						return alice, nil
					}
					return nil, smartcontacterror.ErrUserNotFound
				},
			},
			wantStatus: http.StatusOK,
			wantUserID: 1,
		},
		{
			name:   "user not found – 404",
			pathID: "99",
			svc: &mockUserService{
				fetchUserByIDFn: func(_ context.Context, id int) (*model.User, error) {
					return nil, smartcontacterror.ErrUserNotFound
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "non-numeric id – 400",
			pathID:     "abc",
			svc:        &mockUserService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mux := newTestMux(tc.svc)
			req := httptest.NewRequest(http.MethodGet,
				"/get_user_data/"+tc.pathID, nil)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatus, rr.Code)
			if tc.wantStatus == http.StatusOK {
				var got model.User
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &got))
				assert.Equal(t, tc.wantUserID, got.ID)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteUser  DELETE /delete_user_data/{id}
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name             string
		pathID           string
		svc              *mockUserService
		wantStatus       int
		wantBodyContains string
		wantContentType  string
	}{
		{
			name:   "valid id – returns 200 plain-text success message",
			pathID: "5",
			svc: &mockUserService{
				deleteUserFn: func(_ context.Context, id int) error {
					return nil
				},
			},
			wantStatus:       http.StatusOK,
			wantBodyContains: "user data deleted Successfully",
			wantContentType:  "text/plain; charset=utf-8",
		},
		{
			name:       "non-numeric id – 400",
			pathID:     "xyz",
			svc:        &mockUserService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "service error – non-200",
			pathID: "5",
			svc: &mockUserService{
				deleteUserFn: func(_ context.Context, id int) error {
					return smartcontacterror.ErrUserNotFound
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mux := newTestMux(tc.svc)
			req := httptest.NewRequest(http.MethodDelete,
				"/delete_user_data/"+tc.pathID, nil)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatus, rr.Code)
			if tc.wantBodyContains != "" {
				assert.Contains(t, rr.Body.String(), tc.wantBodyContains)
			}
			if tc.wantContentType != "" {
				assert.Equal(t, tc.wantContentType, rr.Header().Get("Content-Type"))
			}
		})
	}
}

// DeleteUser invariant: always 200 + fixed message when service succeeds.
func TestDeleteUser_AlwaysFixedMessage(t *testing.T) {
	for _, id := range []string{"1", "42", "9999"} {
		id := id
		t.Run("id="+id, func(t *testing.T) {
			svc := &mockUserService{
				deleteUserFn: func(_ context.Context, _ int) error { return nil },
			}
			mux := newTestMux(svc)
			req := httptest.NewRequest(http.MethodDelete, "/delete_user_data/"+id, nil)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, "user data deleted Successfully", rr.Body.String())
		})
	}
}

// ---------------------------------------------------------------------------
// UpdateUser  PUT /update_user_data/{id}
// ---------------------------------------------------------------------------

func TestUpdateUser(t *testing.T) {
	inputUser := &model.User{Name: "Updated Alice", Email: "updated@example.com"}

	tests := []struct {
		name       string
		pathID     string
		body       string
		svc        *mockUserService
		wantStatus int
		checkBody  func(t *testing.T, body []byte)
	}{
		{
			name:   "valid update – returns 200 with request body echoed back",
			pathID: "1",
			body:   string(mustMarshal(t, inputUser)),
			svc: &mockUserService{
				updateUserFn: func(_ context.Context, id int, u *model.User) (*model.User, error) {
					// Simulate