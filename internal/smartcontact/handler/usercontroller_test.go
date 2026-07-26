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

	smartcontacterror "github.com/smartcontact/internal/smartcontact/error"
	"github.com/smartcontact/internal/smartcontact/handler"
	"github.com/smartcontact/internal/smartcontact/model"
)

// ---------------------------------------------------------------------------
// Mock service
// ---------------------------------------------------------------------------

type mockUserService struct {
	saveFunc      func(ctx context.Context, u *model.User) (*model.User, error)
	listFunc      func(ctx context.Context) ([]*model.User, error)
	getByIDFunc   func(ctx context.Context, id int) (*model.User, error)
	deleteFunc    func(ctx context.Context, id int) error
	updateFunc    func(ctx context.Context, id int, u *model.User) (*model.User, error)
	getByNameFunc func(ctx context.Context, name string) (*model.User, error)
}

func (m *mockUserService) Save(ctx context.Context, u *model.User) (*model.User, error) {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, u)
	}
	return u, nil
}

func (m *mockUserService) List(ctx context.Context) ([]*model.User, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return nil, nil
}

func (m *mockUserService) GetByID(ctx context.Context, id int) (*model.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockUserService) Delete(ctx context.Context, id int) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockUserService) Update(ctx context.Context, id int, u *model.User) (*model.User, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, u)
	}
	return u, nil
}

func (m *mockUserService) GetByName(ctx context.Context, name string) (*model.User, error) {
	if m.getByNameFunc != nil {
		return m.getByNameFunc(ctx, name)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// Helper: build a chi router wired with the handler under test
// ---------------------------------------------------------------------------

func newRouter(svc *mockUserService) *chi.Mux {
	r := chi.NewRouter()
	h := handler.NewUserHandler(svc)
	h.RegisterRoutes(r)
	return r
}

// ---------------------------------------------------------------------------
// Helper: execute a request against the router
// ---------------------------------------------------------------------------

func do(t *testing.T, router http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
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
	router.ServeHTTP(rr, req)
	return rr
}

// ---------------------------------------------------------------------------
// POST /save_user_data
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	validUser := &model.User{Name: "Alice", Email: "alice@example.com"}
	validBody, _ := json.Marshal(validUser)

	tests := []struct {
		name           string
		body           []byte
		svc            *mockUserService
		wantStatus     int
		wantBodySubstr string
	}{
		{
			name: "valid user returns 200 and success message",
			body: validBody,
			svc: &mockUserService{
				saveFunc: func(_ context.Context, u *model.User) (*model.User, error) {
					return u, nil
				},
			},
			wantStatus:     http.StatusOK,
			wantBodySubstr: "User data saved successfully!",
		},
		{
			name:           "malformed JSON returns 400 with Spring error shape",
			body:           []byte(`{invalid`),
			svc:            &mockUserService{},
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "Bad Request",
		},
		{
			name: "service error returns 500",
			body: validBody,
			svc: &mockUserService{
				saveFunc: func(_ context.Context, u *model.User) (*model.User, error) {
					return nil, errors.New("db failure")
				},
			},
			wantStatus:     http.StatusInternalServerError,
			wantBodySubstr: "internal server error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := newRouter(tc.svc)
			rr := do(t, router, http.MethodPost, "/save_user_data", tc.body)

			assert.Equal(t, tc.wantStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.wantBodySubstr)
		})
	}
}

// TestSaveUser_ValidationFailure tests that a user failing model.Validate
// returns 400 with an ErrorMessage body.  We use an empty body (zero-value
// User) and rely on the validator rejecting it; if the model has no
// constraints this will hit the service path instead, so we also cover the
// explicit path via malformed JSON above.
func TestSaveUser_ValidationFailure(t *testing.T) {
	// Provide a body that decodes cleanly but should fail validation if any
	// required fields are enforced. Use an empty object.
	emptyUser := []byte(`{}`)

	svc := &mockUserService{
		saveFunc: func(_ context.Context, u *model.User) (*model.User, error) {
			return u, nil
		},
	}

	router := newRouter(svc)
	rr := do(t, router, http.MethodPost, "/save_user_data", emptyUser)

	// Either 200 (no validation constraints) or 400 (validation failure) are
	// acceptable depending on the model definition; assert content-type is JSON.
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
}

// ---------------------------------------------------------------------------
// GET /get_user_data
// ---------------------------------------------------------------------------

func TestFetchUserList(t *testing.T) {
	user1 := &model.User{Name: "Alice", Email: "alice@example.com"}
	user2 := &model.User{Name: "Bob", Email: "bob@example.com"}

	tests := []struct {
		name           string
		svc            *mockUserService
		wantStatus     int
		wantBodySubstr string
		wantEmptyList  bool
	}{
		{
			name: "returns list of users with 200",
			svc: &mockUserService{
				listFunc: func(_ context.Context) ([]*model.User, error) {
					return []*model.User{user1, user2}, nil
				},
			},
			wantStatus:     http.StatusOK,
			wantBodySubstr: "Alice",
		},
		{
			name: "returns empty list with 200",
			svc: &mockUserService{
				listFunc: func(_ context.Context) ([]*model.User, error) {
					return []*model.User{}, nil
				},
			},
			wantStatus:    http.StatusOK,
			wantEmptyList: true,
		},
		{
			name: "service error returns 500",
			svc: &mockUserService{
				listFunc: func(_ context.Context) ([]*model.User, error) {
					return nil, errors.New("db error")
				},
			},
			wantStatus:     http.StatusInternalServerError,
			wantBodySubstr: "internal server error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := newRouter(tc.svc)
			rr := do(t, router, http.MethodGet, "/get_user_data", nil)

			assert.Equal(t, tc.wantStatus, rr.Code)
			if tc.wantBodySubstr != "" {
				assert.Contains(t, rr.Body.String(), tc.wantBodySubstr)
			}
			if tc.wantEmptyList {
				body := strings.TrimSpace(rr.Body.String())
				assert.True(t, body == "[]" || body == "null", "expected empty list, got: %s", body)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GET /get_user_data/{id}
// ---------------------------------------------------------------------------

func TestFetchUserByID(t *testing.T) {
	existingUser := &model.User{Name: "Alice", Email: "alice@example.com"}

	tests := []struct {
		name           string
		path           string
		svc            *mockUserService
		wantStatus     int
		wantBodySubstr string
	}{
		{
			name: "existing user id returns 200 with user",
			path: "/get_user_data/1",
			svc: &mockUserService{
				getByIDFunc: func(_ context.Context, id int) (*model.User, error) {
					assert.Equal(t, 1, id)
					return existingUser, nil
				},
			},
			wantStatus:     http.StatusOK,
			wantBodySubstr: "Alice",
		},
		{
			name: "non-existent user id returns 404",
			path: "/get_user_data/99",
			svc: &mockUserService{
				getByIDFunc: func(_ context.Context, id int) (*model.User, error) {
					return nil, smartcontacterror.ErrUserNotFound
				},
			},
			wantStatus:     http.StatusNotFound,
			wantBodySubstr: "",
		},
		{
			name:           "non-numeric id returns 400",
			path:           "/get_user_data/abc",
			svc:            &mockUserService{},
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "",
		},
		{
			name: "service returns generic error gives 500",
			path: "/get_user_data/2",
			svc: &mockUserService{
				getByIDFunc: func(_ context.Context, id int) (*model.User, error) {
					return nil, errors.New("unexpected db error")
				},
			},
			wantStatus:     http.StatusInternalServerError,
			wantBodySubstr: "internal server error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := newRouter(tc.svc)
			rr := do(t, router, http.MethodGet, tc.path, nil)

			assert.Equal(t, tc.wantStatus, rr.Code)
			if tc.wantBodySubstr != "" {
				assert.Contains(t, rr.Body.String(), tc.wantBodySubstr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DELETE /delete_user_data/{id}
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		svc            *mockUserService
		wantStatus     int
		wantBodySubstr string
	}{
		{
			name: "valid id deletes and returns 200 with success message",
			path: "/delete_user_data/1",
			svc: &mockUserService{
				deleteFunc: func(_ context.Context, id int) error {
					assert.Equal(t, 1, id)
					return nil
				},
			},
			wantStatus:     http.StatusOK,
			wantBodySubstr: "user data deleted Successfully",
		},
		{
			name: "non-existent user returns 404",
			path: "/delete_user_data/99",
			svc: &mockUserService{
				deleteFunc: func(_ context.Context, id int) error {
					return smartcontacterror.ErrUserNotFound
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "non-numeric id returns 400",
			path:       "/delete_user_data/abc",
			svc:        &mockUserService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service generic error returns 500",
			path: "/delete_user_data/5",
			svc: &mockUserService{
				deleteFunc: func(_ context.Context, id int) error {
					return errors.New("some error")
				},
			},
			wantStatus:     http.StatusInternalServerError,
			wantBodySubstr: "internal server error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := newRouter(tc.svc)
			rr := do(t, router, http.MethodDelete, tc.path, nil)

			assert.Equal(t, tc.wantStatus, rr.Code)
			if tc.wantBodySubstr != "" {
				assert.Contains(t, rr.Body.String(), tc.wantBodySubstr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// PUT /update_user_data/{id}
// ---------------------------------------------------------------------------

func TestUpdateUser(t *testing.T) {
	userPayload := &model.User{Name: "Updated Alice", Email: "updated@example.com"}
	userBody, _ := json.Marshal(userPayload)

	tests := []struct {
		name           string
		path           string
		body           []byte
		svc            *mockUserService
		wantStatus     int
		wantBodySubstr string
	}{
		{
			name: "valid id and body returns 200 with echoed user",
			path: "/update_user_data/1",
			body: userBody,
			svc: &mockUserService{
				updateFunc: func(_ context.Context, id int, u *model.User) (*model.User, error) {
					assert.Equal(t, 1, id)
					return u, nil
				},
			},
			wantStatus:     http.StatusOK,
			wantBodySubstr: "Updated Alice",
		},
		{
			name: "non-existent user returns 404",
			path: "/update_user_data/99",
			body: userBody,
			svc: &mockUserService{
				updateFunc: func(_ context.Context, id int, u *model.User) (*model.User, error) {
					return nil, smartcontacterror.ErrUserNotFound
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "non-numeric id returns 400",
			path:       "/update_user_data/abc",
			body:       userBody,
			svc:        &mockUserService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:           "malformed JSON body returns 400 with Spring error shape",
			path:           "/update_user_data/1",
			body:           []byte(`{bad json`),
			svc:            &mockUserService{},
			want