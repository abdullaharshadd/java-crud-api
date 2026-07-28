```go
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperr "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/handler"
	"github.com/smartContact/internal/smartcontact/model"
)

// ---------------------------------------------------------------------------
// Mock service
// ---------------------------------------------------------------------------

type mockUserService struct {
	saveUserFn      func(ctx context.Context, user model.User) error
	fetchUserListFn func(ctx context.Context) ([]model.UserResponse, error)
	fetchUserByIDFn func(ctx context.Context, id int) (model.UserResponse, error)
	deleteUserFn    func(ctx context.Context, id int) error
	updateUserFn    func(ctx context.Context, id int, user model.User) error
	getUserByNameFn func(ctx context.Context, name string) (model.UserResponse, error)
}

func (m *mockUserService) SaveUser(ctx context.Context, user model.User) error {
	if m.saveUserFn != nil {
		return m.saveUserFn(ctx, user)
	}
	return nil
}

func (m *mockUserService) FetchUserList(ctx context.Context) ([]model.UserResponse, error) {
	if m.fetchUserListFn != nil {
		return m.fetchUserListFn(ctx)
	}
	return nil, nil
}

func (m *mockUserService) FetchUserByID(ctx context.Context, id int) (model.UserResponse, error) {
	if m.fetchUserByIDFn != nil {
		return m.fetchUserByIDFn(ctx, id)
	}
	return model.UserResponse{}, nil
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

func (m *mockUserService) GetUserByName(ctx context.Context, name string) (model.UserResponse, error) {
	if m.getUserByNameFn != nil {
		return m.getUserByNameFn(ctx, name)
	}
	return model.UserResponse{}, nil
}

// ---------------------------------------------------------------------------
// Helper: build a chi router wired to the controller
// ---------------------------------------------------------------------------

func newRouter(svc handler.UserService) chi.Router {
	r := chi.NewRouter()
	ctrl := handler.NewUserController(svc, nil)
	ctrl.RegisterRoutes(r)
	return r
}

// ---------------------------------------------------------------------------
// Helper: execute a request against the router and return the recorder
// ---------------------------------------------------------------------------

func doRequest(r chi.Router, method, path string, body []byte) *httptest.ResponseRecorder {
	var bodyReader *bytes.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}
	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

// ---------------------------------------------------------------------------
// POST /save_user_data
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	validUser := model.User{Name: "Alice"}
	validBody, _ := json.Marshal(validUser)

	noNameUser := model.User{Name: ""}
	noNameBody, _ := json.Marshal(noNameUser)

	var savedUser model.User
	saveCalled := false

	tests := []struct {
		name           string
		body           []byte
		svc            *mockUserService
		wantStatus     int
		wantBodyText   string
		wantSaveCalled bool
	}{
		{
			name: "valid user persists and returns 200 success message",
			body: validBody,
			svc: &mockUserService{
				saveUserFn: func(ctx context.Context, user model.User) error {
					saveCalled = true
					savedUser = user
					return nil
				},
			},
			wantStatus:     http.StatusOK,
			wantBodyText:   "User data saved successfully!",
			wantSaveCalled: true,
		},
		{
			name: "empty name returns 400 and does not persist",
			body: noNameBody,
			svc: &mockUserService{
				saveUserFn: func(ctx context.Context, user model.User) error {
					saveCalled = true
					return nil
				},
			},
			wantStatus:     http.StatusBadRequest,
			wantSaveCalled: false,
		},
		{
			name:           "nil body returns 400",
			body:           nil,
			svc:            &mockUserService{},
			wantStatus:     http.StatusBadRequest,
			wantSaveCalled: false,
		},
		{
			name:           "malformed JSON returns 400",
			body:           []byte("{bad json"),
			svc:            &mockUserService{},
			wantStatus:     http.StatusBadRequest,
			wantSaveCalled: false,
		},
		{
			name: "service error returns 500",
			body: validBody,
			svc: &mockUserService{
				saveUserFn: func(ctx context.Context, user model.User) error {
					return fmt.Errorf("db error")
				},
			},
			wantStatus:     http.StatusInternalServerError,
			wantSaveCalled: false, // we don't track this variant
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			saveCalled = false
			savedUser = model.User{}

			r := newRouter(tc.svc)

			var rr *httptest.ResponseRecorder
			if tc.body == nil {
				// send request with nil body
				req := httptest.NewRequest(http.MethodPost, "/save_user_data", http.NoBody)
				rr = httptest.NewRecorder()
				r.ServeHTTP(rr, req)
			} else {
				rr = doRequest(r, http.MethodPost, "/save_user_data", tc.body)
			}

			assert.Equal(t, tc.wantStatus, rr.Code)

			if tc.wantBodyText != "" {
				assert.Contains(t, rr.Body.String(), tc.wantBodyText)
			}

			if tc.wantSaveCalled {
				assert.True(t, saveCalled, "expected SaveUser to be called")
				assert.Equal(t, validUser.Name, savedUser.Name)
			} else if tc.name != "service error returns 500" {
				assert.False(t, saveCalled, "expected SaveUser NOT to be called")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GET /get_user_data
// ---------------------------------------------------------------------------

func TestFetchUserList(t *testing.T) {
	user1 := model.UserResponse{Name: "Alice"}
	user2 := model.UserResponse{Name: "Bob"}

	tests := []struct {
		name           string
		svc            *mockUserService
		wantStatus     int
		wantCount      int
		wantEmptyArray bool
	}{
		{
			name: "users exist returns 200 with list",
			svc: &mockUserService{
				fetchUserListFn: func(ctx context.Context) ([]model.UserResponse, error) {
					return []model.UserResponse{user1, user2}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name: "no users returns 200 with empty array",
			svc: &mockUserService{
				fetchUserListFn: func(ctx context.Context) ([]model.UserResponse, error) {
					return []model.UserResponse{}, nil
				},
			},
			wantStatus:     http.StatusOK,
			wantEmptyArray: true,
		},
		{
			name: "nil slice from service returns 200 with empty array",
			svc: &mockUserService{
				fetchUserListFn: func(ctx context.Context) ([]model.UserResponse, error) {
					return nil, nil
				},
			},
			wantStatus:     http.StatusOK,
			wantEmptyArray: true,
		},
		{
			name: "service error returns 500",
			svc: &mockUserService{
				fetchUserListFn: func(ctx context.Context) ([]model.UserResponse, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newRouter(tc.svc)
			rr := doRequest(r, http.MethodGet, "/get_user_data", nil)

			assert.Equal(t, tc.wantStatus, rr.Code)

			if tc.wantStatus == http.StatusOK {
				var users []model.UserResponse
				err := json.Unmarshal(rr.Body.Bytes(), &users)
				require.NoError(t, err)

				if tc.wantEmptyArray {
					assert.Empty(t, users)
				} else {
					assert.Len(t, users, tc.wantCount)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GET /get_user_data/{id}
// ---------------------------------------------------------------------------

func TestFetchUserByID(t *testing.T) {
	foundUser := model.UserResponse{Name: "Alice"}

	tests := []struct {
		name       string
		pathID     string
		svc        *mockUserService
		wantStatus int
		wantUser   *model.UserResponse
	}{
		{
			name:   "existing id returns 200 with user",
			pathID: "1",
			svc: &mockUserService{
				fetchUserByIDFn: func(ctx context.Context, id int) (model.UserResponse, error) {
					assert.Equal(t, 1, id)
					return foundUser, nil
				},
			},
			wantStatus: http.StatusOK,
			wantUser:   &foundUser,
		},
		{
			name:   "non-existing id returns 404 via UserNotFoundError",
			pathID: "999",
			svc: &mockUserService{
				fetchUserByIDFn: func(ctx context.Context, id int) (model.UserResponse, error) {
					return model.UserResponse{}, &apperr.UserNotFoundError{ID: id}
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "non-integer id returns 400",
			pathID: "abc",
			svc:    &mockUserService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "service generic error returns 500",
			pathID: "2",
			svc: &mockUserService{
				fetchUserByIDFn: func(ctx context.Context, id int) (model.UserResponse, error) {
					return model.UserResponse{}, fmt.Errorf("unexpected error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newRouter(tc.svc)
			rr := doRequest(r, http.MethodGet, "/get_user_data/"+tc.pathID, nil)

			assert.Equal(t, tc.wantStatus, rr.Code)

			if tc.wantUser != nil {
				var got model.UserResponse
				err := json.Unmarshal(rr.Body.Bytes(), &got)
				require.NoError(t, err)
				assert.Equal(t, tc.wantUser.Name, got.Name)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DELETE /delete_user_data/{id}
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	deleteCalled := false
	deletedID := 0

	tests := []struct {
		name         string
		pathID       string
		svc          *mockUserService
		wantStatus   int
		wantBodyText string
	}{
		{
			name:   "valid id deletes and returns 200 success message",
			pathID: "42",
			svc: &mockUserService{
				deleteUserFn: func(ctx context.Context, id int) error {
					deleteCalled = true
					deletedID = id
					return nil
				},
			},
			wantStatus:   http.StatusOK,
			wantBodyText: "user data deleted Successfully",
		},
		{
			name:       "non-integer id returns 400",
			pathID:     "xyz",
			svc:        &mockUserService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "service error returns 500",
			pathID: "1",
			svc: &mockUserService{
				deleteUserFn: func(ctx context.Context, id int) error {
					return fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:   "UserNotFoundError from service returns 404",
			pathID: "5",
			svc: &mockUserService{
				deleteUserFn: func(ctx context.Context, id int) error {
					return &apperr.UserNotFoundError{ID: id}
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			deleteCalled = false
			deletedID = 0

			r := newRouter(tc.svc)
			rr := doRequest(r, http.MethodDelete, "/delete_user_data/"+tc.pathID, nil)

			assert.Equal(t, tc.wantStatus, rr.Code)

			if tc.wantBodyText != "" {
				assert.Contains(t, rr.Body.String(), tc.wantBodyText)
			}

			if tc.name == "valid id deletes and returns 200 success message" {
				assert.True(t, deleteCalled)
				assert.Equal(t, 42, deletedID)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// PUT /update_user_data/{id}
// ---------------------------------------------------------------------------

func TestUpdateUser(t *testing.T) {
	updateCalled := false
	updatedID := 0

	userBody := model.User{Name: "UpdatedAlice"}
	userBodyBytes, _ := json.Marshal(userBody)

	tests := []struct {
		name         string
		pathID       string
		body         []byte
		svc          *mockUserService
		wantStatus   int
		wantBodyName string
	}{
		{
			name:   "valid id and body updates and returns 200 with echoed user",
			pathID: "10",
			body:   userBodyBytes,
			svc: &mockUserService{
				updateUserFn: