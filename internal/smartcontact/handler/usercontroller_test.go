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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	smartErr "internal/smartcontact/error"
	"internal/smartcontact/handler"
	"internal/smartcontact/model"
)

// ---------------------------------------------------------------------------
// Mock implementation of service.UserService
// ---------------------------------------------------------------------------

type mockUserService struct {
	saveUserFn        func(ctx context.Context, user model.User) error
	fetchUserListFn   func(ctx context.Context) ([]model.User, error)
	fetchUserByIDFn   func(ctx context.Context, id int) (model.User, error)
	deleteUserFn      func(ctx context.Context, id int) error
	updateUserFn      func(ctx context.Context, id int, user model.User) error
	getUserByNameFn   func(ctx context.Context, name string) (model.User, error)
}

func (m *mockUserService) SaveUser(ctx context.Context, user model.User) error {
	if m.saveUserFn != nil {
		return m.saveUserFn(ctx, user)
	}
	return nil
}

func (m *mockUserService) FetchUserList(ctx context.Context) ([]model.User, error) {
	if m.fetchUserListFn != nil {
		return m.fetchUserListFn(ctx)
	}
	return nil, nil
}

func (m *mockUserService) FetchUserByID(ctx context.Context, id int) (model.User, error) {
	if m.fetchUserByIDFn != nil {
		return m.fetchUserByIDFn(ctx, id)
	}
	return model.User{}, nil
}

func (m *mockUserService) DeleteUser(ctx context.Context, id int) error {
	if m.deleteUserFn != nil {
		return m.deleteUserFn(ctx, id)
	}
	return nil
}

func (m *mockUserService) UpdateUser(ctx context.Context, id int, user model.User) error {
	if m.updateUserFn != nil {
		return m.updateUserFn(ctx, id, user)
	}
	return nil
}

func (m *mockUserService) GetUserByName(ctx context.Context, name string) (model.User, error) {
	if m.getUserByNameFn != nil {
		return m.getUserByNameFn(ctx, name)
	}
	return model.User{}, nil
}

// ---------------------------------------------------------------------------
// Helper: build a chi router wired to a UserController
// ---------------------------------------------------------------------------

func newTestRouter(svc *mockUserService) http.Handler {
	r := chi.NewRouter()
	c := handler.NewUserController(svc)
	c.RegisterRoutes(r)
	return r
}

// ---------------------------------------------------------------------------
// Helper: execute a request against the router and return the recorder
// ---------------------------------------------------------------------------

func doRequest(router http.Handler, method, target string, body []byte) *httptest.ResponseRecorder {
	var reqBody *bytes.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	} else {
		reqBody = bytes.NewReader([]byte{})
	}
	req := httptest.NewRequest(method, target, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// ---------------------------------------------------------------------------
// POST /save_user_data
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	validUser := model.User{
		Name:  "Alice",
		Email: "alice@example.com",
	}
	validBody, _ := json.Marshal(validUser)

	tests := []struct {
		name           string
		requestBody    []byte
		setupService   func(*mockUserService)
		wantStatus     int
		wantBodyContains string
	}{
		{
			name:        "valid user saves successfully",
			requestBody: validBody,
			setupService: func(m *mockUserService) {
				m.saveUserFn = func(_ context.Context, _ model.User) error { return nil }
			},
			wantStatus:       http.StatusOK,
			wantBodyContains: "User data saved successfully!",
		},
		{
			name:        "invalid JSON body returns 400",
			requestBody: []byte(`{not valid json`),
			setupService: func(m *mockUserService) {},
			wantStatus:       http.StatusBadRequest,
			wantBodyContains: "",
		},
		{
			name:        "validation failure returns 400",
			requestBody: func() []byte {
				// Empty name/email should fail model.Validate
				u := model.User{}
				b, _ := json.Marshal(u)
				return b
			}(),
			setupService: func(m *mockUserService) {},
			wantStatus:       http.StatusBadRequest,
			wantBodyContains: "",
		},
		{
			name:        "service error returns 500",
			requestBody: validBody,
			setupService: func(m *mockUserService) {
				m.saveUserFn = func(_ context.Context, _ model.User) error {
					return errors.New("db error")
				}
			},
			wantStatus:       http.StatusInternalServerError,
			wantBodyContains: "",
		},
		{
			name:        "service returns ErrUserNotFound yields 404",
			requestBody: validBody,
			setupService: func(m *mockUserService) {
				m.saveUserFn = func(_ context.Context, _ model.User) error {
					return smartErr.ErrUserNotFound
				}
			},
			wantStatus:       http.StatusNotFound,
			wantBodyContains: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockUserService{}
			tc.setupService(svc)
			router := newTestRouter(svc)

			rec := doRequest(router, http.MethodPost, "/save_user_data", tc.requestBody)

			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.wantBodyContains != "" {
				assert.Contains(t, rec.Body.String(), tc.wantBodyContains)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GET /get_user_data
// ---------------------------------------------------------------------------

func TestFetchUserList(t *testing.T) {
	alice := model.User{Name: "Alice", Email: "alice@example.com"}
	bob := model.User{Name: "Bob", Email: "bob@example.com"}

	tests := []struct {
		name         string
		setupService func(*mockUserService)
		wantStatus   int
		wantUsers    []model.User
		wantArray    bool
	}{
		{
			name: "returns list of users",
			setupService: func(m *mockUserService) {
				m.fetchUserListFn = func(_ context.Context) ([]model.User, error) {
					return []model.User{alice, bob}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantUsers:  []model.User{alice, bob},
			wantArray:  true,
		},
		{
			name: "returns empty array when no users exist",
			setupService: func(m *mockUserService) {
				m.fetchUserListFn = func(_ context.Context) ([]model.User, error) {
					return []model.User{}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantUsers:  []model.User{},
			wantArray:  true,
		},
		{
			name: "service returns nil slice yields empty JSON array",
			setupService: func(m *mockUserService) {
				m.fetchUserListFn = func(_ context.Context) ([]model.User, error) {
					return nil, nil
				}
			},
			wantStatus: http.StatusOK,
			wantUsers:  []model.User{},
			wantArray:  true,
		},
		{
			name: "service error returns 500",
			setupService: func(m *mockUserService) {
				m.fetchUserListFn = func(_ context.Context) ([]model.User, error) {
					return nil, errors.New("db failure")
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantArray:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockUserService{}
			tc.setupService(svc)
			router := newTestRouter(svc)

			rec := doRequest(router, http.MethodGet, "/get_user_data", nil)

			assert.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantArray {
				assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

				var got []model.User
				err := json.NewDecoder(rec.Body).Decode(&got)
				require.NoError(t, err)

				if tc.wantUsers != nil {
					assert.Equal(t, tc.wantUsers, got)
				} else {
					assert.NotNil(t, got)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GET /get_user_data/{id}
// ---------------------------------------------------------------------------

func TestFetchUserByID(t *testing.T) {
	alice := model.User{Name: "Alice", Email: "alice@example.com"}

	tests := []struct {
		name         string
		pathID       string
		setupService func(*mockUserService)
		wantStatus   int
		wantUser     *model.User
	}{
		{
			name:   "returns existing user by id",
			pathID: "1",
			setupService: func(m *mockUserService) {
				m.fetchUserByIDFn = func(_ context.Context, id int) (model.User, error) {
					assert.Equal(t, 1, id)
					return alice, nil
				}
			},
			wantStatus: http.StatusOK,
			wantUser:   &alice,
		},
		{
			name:   "user not found returns 404",
			pathID: "99",
			setupService: func(m *mockUserService) {
				m.fetchUserByIDFn = func(_ context.Context, id int) (model.User, error) {
					return model.User{}, smartErr.ErrUserNotFound
				}
			},
			wantStatus: http.StatusNotFound,
			wantUser:   nil,
		},
		{
			name:         "non-integer id returns 400",
			pathID:       "abc",
			setupService: func(m *mockUserService) {},
			wantStatus:   http.StatusBadRequest,
			wantUser:     nil,
		},
		{
			name:   "service error returns 500",
			pathID: "2",
			setupService: func(m *mockUserService) {
				m.fetchUserByIDFn = func(_ context.Context, id int) (model.User, error) {
					return model.User{}, errors.New("internal error")
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantUser:   nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockUserService{}
			tc.setupService(svc)
			router := newTestRouter(svc)

			rec := doRequest(router, http.MethodGet, "/get_user_data/"+tc.pathID, nil)

			assert.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantUser != nil {
				var got model.User
				err := json.NewDecoder(rec.Body).Decode(&got)
				require.NoError(t, err)
				assert.Equal(t, *tc.wantUser, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DELETE /delete_user_data/{id}
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name             string
		pathID           string
		setupService     func(*mockUserService)
		wantStatus       int
		wantBodyContains string
	}{
		{
			name:   "deletes user and returns success message",
			pathID: "5",
			setupService: func(m *mockUserService) {
				m.deleteUserFn = func(_ context.Context, id int) error {
					assert.Equal(t, 5, id)
					return nil
				}
			},
			wantStatus:       http.StatusOK,
			wantBodyContains: "user data deleted Successfully",
		},
		{
			name:         "non-integer id returns 400",
			pathID:       "xyz",
			setupService: func(m *mockUserService) {},
			wantStatus:   http.StatusBadRequest,
		},
		{
			name:   "user not found returns 404",
			pathID: "42",
			setupService: func(m *mockUserService) {
				m.deleteUserFn = func(_ context.Context, id int) error {
					return smartErr.ErrUserNotFound
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "service error returns 500",
			pathID: "3",
			setupService: func(m *mockUserService) {
				m.deleteUserFn = func(_ context.Context, id int) error {
					return errors.New("deletion failed")
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockUserService{}
			tc.setupService(svc)
			router := newTestRouter(svc)

			rec := doRequest(router, http.MethodDelete, "/delete_user_data/"+tc.pathID, nil)

			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.wantBodyContains != "" {
				assert.Contains(t, rec.Body.String(), tc.wantBodyContains)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// PUT /update_user_data/{id}
// ---------------------------------------------------------------------------

func TestUpdateUser(t *testing.T) {
	submittedUser := model.User{Name: "Updated Alice", Email: "updated@example.com"}
	submittedBody, _ := json.Marshal(submittedUser)

	tests := []struct {
		name         string
		pathID       string
		requestBody  []byte
		setupService func(*mockUserService)
		wantStatus   int
		wantUser     *model.User
	}{
		{
			name:        "updates user and returns submitted object",
			pathID:      "7",
			requestBody: submittedBody,
			setupService: func(m *mockUserService) {
				m.updateUserFn = func(_ context.Context, id int, user model.User) error {
					assert.Equal(t, 7, id)
					return nil
				}
			},
			wantStatus: http.StatusOK,
			wantUser:   &submittedUser,
		},
		{
			name:         "non-integer id returns 400",
			pathID:       "notanumber",
			requestBody:  submittedBody,
			setupService: func(m *mockUserService) {},
			wantStatus:   http.StatusBadRequest,
			wantUser:     nil,