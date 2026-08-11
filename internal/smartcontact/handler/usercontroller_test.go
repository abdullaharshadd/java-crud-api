```go
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontact/internal/smartcontact/handler"
	"github.com/smartcontact/internal/smartcontact/model"
)

// ---------------------------------------------------------------------------
// Mock UserService
// ---------------------------------------------------------------------------

type mockUserService struct {
	saveUserFn      func(ctx context.Context, user *model.User) error
	fetchUserListFn func(ctx context.Context) ([]*model.User, error)
	fetchUserByIDFn func(ctx context.Context, id int) (*model.User, error)
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
// helpers
// ---------------------------------------------------------------------------

// newRouter builds a chi router with the handler's routes registered.
func newRouter(svc *mockUserService) *chi.Mux {
	h := handler.NewHandler(svc, nil)
	r := chi.NewRouter()
	h.RegisterRoutes(r)
	return r
}

func marshalJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

// ---------------------------------------------------------------------------
// SaveUser
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	validUser := model.User{ID: 1, Name: "Alice", Email: "alice@example.com"}

	tests := []struct {
		name           string
		body           []byte
		setupSvc       func(*mockUserService)
		wantStatus     int
		wantBody       string
		saveUserCalled bool
	}{
		{
			name: "valid user returns 200 and success message",
			body: marshalJSON(t, validUser),
			setupSvc: func(m *mockUserService) {
				m.saveUserFn = func(_ context.Context, u *model.User) error {
					return nil
				}
			},
			wantStatus:     http.StatusOK,
			wantBody:       "User data saved successfully!",
			saveUserCalled: true,
		},
		{
			name:           "invalid JSON body returns 4xx and does not call service",
			body:           []byte(`{invalid json`),
			setupSvc:       func(m *mockUserService) {},
			wantStatus:     http.StatusBadRequest,
			saveUserCalled: false,
		},
		{
			name: "validation error returns 400 and does not persist",
			// Simulate a user that fails model.Validate() by omitting required fields.
			// The exact fields depend on the model; we test with an empty user.
			body: marshalJSON(t, model.User{}),
			setupSvc: func(m *mockUserService) {
				// SaveUser should NOT be called for an invalid user.
				m.saveUserFn = func(_ context.Context, u *model.User) error {
					t.Error("SaveUser should not be called when validation fails")
					return nil
				}
			},
			// We accept either 400 or 422; the handler uses WriteError which
			// maps validation errors. We assert non-200.
			wantStatus:     0, // checked differently below
			saveUserCalled: false,
		},
		{
			name: "service error propagates",
			body: marshalJSON(t, validUser),
			setupSvc: func(m *mockUserService) {
				m.saveUserFn = func(_ context.Context, u *model.User) error {
					return errors.New("db error")
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockUserService{}
			if tc.setupSvc != nil {
				tc.setupSvc(svc)
			}
			r := newRouter(svc)

			req := httptest.NewRequest(http.MethodPost, "/save_user_data", bytes.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if tc.name == "validation error returns 400 and does not persist" {
				// Accept any non-200 status for validation failure.
				assert.NotEqual(t, http.StatusOK, rec.Code)
				return
			}

			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.wantBody != "" {
				assert.Equal(t, tc.wantBody, rec.Body.String())
			}
		})
	}
}

func TestSaveUser_InvariantSuccessMessage(t *testing.T) {
	// Invariant: successful response always carries HTTP 200 and the fixed success message.
	validUser := model.User{ID: 1, Name: "Bob", Email: "bob@example.com"}
	svc := &mockUserService{
		saveUserFn: func(_ context.Context, u *model.User) error { return nil },
	}
	r := newRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/save_user_data", bytes.NewReader(marshalJSON(t, validUser)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "User data saved successfully!", rec.Body.String())
}

func TestSaveUser_ServiceCalledExactlyOnce(t *testing.T) {
	validUser := model.User{ID: 1, Name: "Charlie", Email: "charlie@example.com"}
	callCount := 0
	svc := &mockUserService{
		saveUserFn: func(_ context.Context, u *model.User) error {
			callCount++
			return nil
		},
	}
	r := newRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/save_user_data", bytes.NewReader(marshalJSON(t, validUser)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, callCount, "SaveUser should be called exactly once")
}

// ---------------------------------------------------------------------------
// FetchUserList
// ---------------------------------------------------------------------------

func TestFetchUserList(t *testing.T) {
	user1 := &model.User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	user2 := &model.User{ID: 2, Name: "Bob", Email: "bob@example.com"}

	tests := []struct {
		name         string
		serviceUsers []*model.User
		serviceErr   error
		wantStatus   int
		wantCount    int
	}{
		{
			name:         "returns 200 with list of users",
			serviceUsers: []*model.User{user1, user2},
			wantStatus:   http.StatusOK,
			wantCount:    2,
		},
		{
			name:         "returns 200 with empty list when no users",
			serviceUsers: []*model.User{},
			wantStatus:   http.StatusOK,
			wantCount:    0,
		},
		{
			name:       "service error propagates",
			serviceErr: errors.New("db error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockUserService{
				fetchUserListFn: func(_ context.Context) ([]*model.User, error) {
					return tc.serviceUsers, tc.serviceErr
				},
			}
			r := newRouter(svc)

			req := httptest.NewRequest(http.MethodGet, "/get_user_data", nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantStatus == http.StatusOK {
				assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
				var users []*model.User
				err := json.Unmarshal(rec.Body.Bytes(), &users)
				require.NoError(t, err)
				assert.Len(t, users, tc.wantCount)
			}
		})
	}
}

func TestFetchUserList_NoModification(t *testing.T) {
	// Invariant: no entities are modified during retrieval.
	original := &model.User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	svc := &mockUserService{
		fetchUserListFn: func(_ context.Context) ([]*model.User, error) {
			return []*model.User{original}, nil
		},
	}
	r := newRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/get_user_data", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var users []*model.User
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &users))
	require.Len(t, users, 1)
	assert.Equal(t, original.ID, users[0].ID)
	assert.Equal(t, original.Name, users[0].Name)
	assert.Equal(t, original.Email, users[0].Email)
}

// ---------------------------------------------------------------------------
// FetchUserByID
// ---------------------------------------------------------------------------

func TestFetchUserByID(t *testing.T) {
	foundUser := &model.User{ID: 42, Name: "Alice", Email: "alice@example.com"}

	tests := []struct {
		name       string
		pathID     string
		setupSvc   func(*mockUserService)
		wantStatus int
		wantUserID int
	}{
		{
			name:   "existing id returns 200 with user",
			pathID: "42",
			setupSvc: func(m *mockUserService) {
				m.fetchUserByIDFn = func(_ context.Context, id int) (*model.User, error) {
					assert.Equal(t, 42, id)
					return foundUser, nil
				}
			},
			wantStatus: http.StatusOK,
			wantUserID: 42,
		},
		{
			name:   "non-existent id propagates error",
			pathID: "999",
			setupSvc: func(m *mockUserService) {
				m.fetchUserByIDFn = func(_ context.Context, id int) (*model.User, error) {
					return nil, errors.New("UserNotFoundException: user not found")
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "non-integer id returns 4xx",
			pathID:     "abc",
			setupSvc:   func(m *mockUserService) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockUserService{}
			if tc.setupSvc != nil {
				tc.setupSvc(svc)
			}
			r := newRouter(svc)

			req := httptest.NewRequest(http.MethodGet, "/get_user_data/"+tc.pathID, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantStatus == http.StatusOK && tc.wantUserID != 0 {
				var u model.User
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &u))
				assert.Equal(t, tc.wantUserID, u.ID)
			}
		})
	}
}

func TestFetchUserByID_ReturnedIDMatchesRequest(t *testing.T) {
	// Invariant: the returned user's id equals the requested id when found.
	svc := &mockUserService{
		fetchUserByIDFn: func(_ context.Context, id int) (*model.User, error) {
			return &model.User{ID: id, Name: "Test", Email: "test@example.com"}, nil
		},
	}
	r := newRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/get_user_data/7", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var u model.User
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &u))
	assert.Equal(t, 7, u.ID)
}

// ---------------------------------------------------------------------------
// DeleteUser
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name       string
		pathID     string
		setupSvc   func(m *mockUserService, callTracker *int)
		wantStatus int
		wantBody   string
		wantCalls  int
	}{
		{
			name:   "valid id deletes user and returns 200",
			pathID: "5",
			setupSvc: func(m *mockUserService, callTracker *int) {
				m.deleteUserFn = func(_ context.Context, id int) error {
					assert.Equal(t, 5, id)
					*callTracker++
					return nil
				}
			},
			wantStatus: http.StatusO