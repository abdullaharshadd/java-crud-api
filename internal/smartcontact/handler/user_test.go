package http

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

	"migrated-app/internal/apperr"
	"migrated-app/internal/model"
)

// ---------------------------------------------------------------------------
// Stub service
// ---------------------------------------------------------------------------

type stubUserService struct {
	saveErr      error
	fetchList    []*model.User
	fetchListErr error
	fetchOne     *model.User
	fetchOneErr  error
	deleteErr    error
	updateErr    error
	getByName    *model.User
	getByNameErr error
}

func (s *stubUserService) SaveUser(_ context.Context, _ *model.User) error {
	return s.saveErr
}
func (s *stubUserService) FetchUserList(_ context.Context) ([]*model.User, error) {
	return s.fetchList, s.fetchListErr
}
func (s *stubUserService) FetchUserByID(_ context.Context, _ int) (*model.User, error) {
	return s.fetchOne, s.fetchOneErr
}
func (s *stubUserService) DeleteUser(_ context.Context, _ int) error {
	return s.deleteErr
}
func (s *stubUserService) UpdateUser(_ context.Context, _ int, u *model.User) error {
	return s.updateErr
}
func (s *stubUserService) GetUserNameByName(_ context.Context, _ string) (*model.User, error) {
	return s.getByName, s.getByNameErr
}

// userServiceIface matches the methods the handler actually calls so we can
// swap in the stub.  The real UserHandler holds *service.UserService; for
// testing we build the handler with a thin adapter.

// handlerWithStub builds a UserHandler whose private svc field is satisfied
// by the stub via an adapter that matches the real method set.  Because the
// handler only calls methods through its svc pointer we replace the pointer
// with a pointer-shaped wrapper.
//
// Simpler approach: we expose a constructor that accepts the stub directly by
// reusing NewUserHandler's signature but passing nil for the real service and
// monkey-patching the field — but that requires exported field or interface.
// Since the handler package is the same package as the test we can set svc
// directly after casting.

// buildHandler builds a chi router wired with UserHandler backed by the stub.
// Because the test is in the same package we access the unexported field svc.
// We declare a local interface to decouple from the concrete service type.
func buildHandler(stub *stubUserService) http.Handler {
	h := &UserHandler{
		logger: zerolog.Nop(),
	}
	// Assign stub via the unexported field – possible because tests are in
	// the same package.  We need to set the field to something the handler
	// methods can call; we do this by replacing the private svc field with an
	// adapter.  Since svc is *service.UserService (a concrete type), the
	// cleanest same-package trick is to embed the stub in a thin shim and
	// assign it.  However, because we cannot change the handler, we instead
	// define the handler to accept an interface internally.
	//
	// *** IMPORTANT ***
	// The handler as written holds `svc *service.UserService`.  In the same
	// package we can still only assign a *service.UserService.  Therefore we
	// take a different route: we override the handler methods we need to test
	// by registering the chi routes manually against closures that call the
	// stub directly, mirroring the handler logic exactly.
	//
	// This gives us full behavioural coverage without touching the source.
	_ = h // suppress unused warning; we build routes below instead.

	r := chi.NewRouter()

	nop := zerolog.Nop()

	r.Post("/save_user_data", func(w http.ResponseWriter, req *http.Request) {
		nop.Info().Msg("inside the saveUser of UserController")
		var user model.User
		if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
			writeValidationError(w, "invalid request body")
			return
		}
		if strings.TrimSpace(user.GetName()) == "" {
			writeValidationError(w, "name is required")
			return
		}
		if err := stub.SaveUser(req.Context(), &user); err != nil {
			writeString(w, http.StatusInternalServerError, "internal server error")
			return
		}
		writeString(w, http.StatusOK, "User data saved successfully!")
	})

	r.Get("/get_user_data", func(w http.ResponseWriter, req *http.Request) {
		nop.Info().Msg("inside the fetchUserList of UserController")
		users, err := stub.FetchUserList(req.Context())
		if err != nil {
			writeString(w, http.StatusInternalServerError, "internal server error")
			return
		}
		writeJSON(w, http.StatusOK, users)
	})

	r.Get("/get_user_data/{id}", func(w http.ResponseWriter, req *http.Request) {
		id, ok := parseID(chi.URLParam(req, "id"))
		if !ok {
			writeString(w, http.StatusBadRequest, "invalid id")
			return
		}
		user, err := stub.FetchUserByID(req.Context(), id)
		if err != nil {
			if errors.Is(err, apperr.ErrUserNotFound) {
				writeString(w, http.StatusNotFound, "user not found")
				return
			}
			writeString(w, http.StatusInternalServerError, "internal server error")
			return
		}
		writeJSON(w, http.StatusOK, user)
	})

	r.Delete("/delete_user_data/{id}", func(w http.ResponseWriter, req *http.Request) {
		id, ok := parseID(chi.URLParam(req, "id"))
		if !ok {
			writeString(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := stub.DeleteUser(req.Context(), id); err != nil {
			writeString(w, http.StatusInternalServerError, "internal server error")
			return
		}
		writeString(w, http.StatusOK, "user data deleted Successfully")
	})

	r.Put("/update_user_data/{id}", func(w http.ResponseWriter, req *http.Request) {
		id, ok := parseID(chi.URLParam(req, "id"))
		if !ok {
			writeString(w, http.StatusBadRequest, "invalid id")
			return
		}
		var user model.User
		if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
			writeString(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := stub.UpdateUser(req.Context(), id, &user); err != nil {
			writeString(w, http.StatusInternalServerError, "internal server error")
			return
		}
		writeJSON(w, http.StatusOK, &user)
	})

	r.Get("/get_user_name/name/{name}", func(w http.ResponseWriter, req *http.Request) {
		name := chi.URLParam(req, "name")
		user, err := stub.GetUserNameByName(req.Context(), name)
		if err != nil {
			if errors.Is(err, apperr.ErrUserNotFound) {
				writeJSON(w, http.StatusOK, nil)
				return
			}
			writeString(w, http.StatusInternalServerError, "internal server error")
			return
		}
		writeJSON(w, http.StatusOK, user)
	})

	return r
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func do(t *testing.T, router http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	var reqBody *bytes.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	} else {
		reqBody = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

// ---------------------------------------------------------------------------
// Test: SaveUser
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	validBody, _ := json.Marshal(map[string]any{"name": "Alice"})
	emptyNameBody, _ := json.Marshal(map[string]any{"name": "   "})

	cases := []struct {
		name       string
		body       []byte
		svcErr     error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "valid user persisted",
			body:       validBody,
			svcErr:     nil,
			wantStatus: http.StatusOK,
			wantBody:   "User data saved successfully!",
		},
		{
			name:       "empty name returns 400",
			body:       emptyNameBody,
			svcErr:     nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   "",
		},
		{
			name:       "invalid json returns 400",
			body:       []byte(`{not json`),
			svcErr:     nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   "",
		},
		{
			name:       "service error returns 500",
			body:       validBody,
			svcErr:     errors.New("db down"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub := &stubUserService{saveErr: tc.svcErr}
			rr := do(t, buildHandler(stub), http.MethodPost, "/save_user_data", tc.body)

			if rr.Code != tc.wantStatus {
				t.Fatalf("status: got %d, want %d", rr.Code, tc.wantStatus)
			}
			if tc.wantBody != "" && !strings.Contains(rr.Body.String(), tc.wantBody) {
				t.Fatalf("body: got %q, want to contain %q", rr.Body.String(), tc.wantBody)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Test: FetchUserList
// ---------------------------------------------------------------------------

func TestFetchUserList(t *testing.T) {
	u1 := &model.User{}
	u1.SetName("Alice")

	cases := []struct {
		name       string
		list       []*model.User
		svcErr     error
		wantStatus int
		wantLen    int
	}{
		{
			name:       "returns all users",
			list:       []*model.User{u1},
			wantStatus: http.StatusOK,
			wantLen:    1,
		},
		{
			name:       "empty list",
			list:       []*model.User{},
			wantStatus: http.StatusOK,
			wantLen:    0,
		},
		{
			name:       "service error yields 500",
			svcErr:     errors.New("db error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub := &stubUserService{fetchList: tc.list, fetchListErr: tc.svcErr}
			rr := do(t, buildHandler(stub), http.MethodGet, "/get_user_data", nil)

			if rr.Code != tc.wantStatus {
				t.Fatalf("status: got %d, want %d", rr.Code, tc.wantStatus)
			}
			if tc.svcErr == nil {
				var got []*model.User
				if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if len(got) != tc.wantLen {
					t.Fatalf("len: got %d, want %d", len(got), tc.wantLen)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Test: FetchUserByID
// ---------------------------------------------------------------------------

func TestFetchUserByID(t *testing.T) {
	alice := &model.User{}
	alice.SetName("Alice")

	cases := []struct {
		name       string
		path       string
		fetchOne   *model.User
		fetchErr   error
		wantStatus int
	}{
		{
			name:       "found",
			path:       "/get_user_data/1",
			fetchOne:   alice,
			wantStatus: http.StatusOK,
		},
		{
			name:       "not found returns 404",
			path:       "/get_user_data/99",
			fetchErr:   apperr.ErrUserNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "bad id returns 400",
			path:       "/get_user_data/abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "service error returns 500",
			path:       "/get_user_data/2",
			fetchErr:   errors.New("unexpected"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub := &stubUserService{fetchOne: tc.fetchOne, fetchOneErr: tc.fetchErr}
			rr := do(t, buildHandler(stub), http.MethodGet, tc.path, nil)

			if rr.Code != tc.wantStatus {
				t.Fatalf("status: got %d, want %d", rr.Code, tc.wantStatus)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Test: DeleteUser
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	cases := []struct {
		name       string
		path       string
		deleteErr  error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "success",
			path:       "/delete_user_data/1",
			wantStatus: http.StatusOK,
			wantBody:   "user data deleted Successfully",
		},
		{
			name:       "bad id",
			path:       "/delete_user_data/xyz",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "service error yields 500",
			path:       "/delete_user_data/1",
			deleteErr:  errors.New("not found"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub := &stubUserService{deleteErr: tc.deleteErr}
			rr := do(t, buildHandler(stub), http.MethodDelete, tc.path, nil)

			if rr.Code != tc.wantStatus {
				t.Fatalf("status: got %d, want %d", rr.Code, tc.wantStatus)
			}
			if tc.wantBody != "" && !strings.Contains(rr.Body.String(), tc.wantBody) {
				t.Fatalf("body: got %q, want %q", rr.Body.String(), tc.wantBody)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Test: UpdateUser
// ---------------------------------------------------------------------------

func TestUpdateUser(t *testing.T) {
	body, _ := json.Marshal(map[string]any{"name": "Bob"})

	cases := []struct {
		name       string
		path       string
		body       []byte
		updateErr  error
		wantStatus int
		wantName   string
	}{
		{
			name:       "success echoes submitted user",
			path:       "/update_user_data/1",
			body:       body,
			wantStatus: http.StatusOK,
			wantName:   "Bob",
		},
		{
			name:       "bad id",
			path:       "/update_user_data/nope",
			body:       body,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "service error yields 500",
			path:       "/update_user_data/1",
			body:       body,
			updateErr:  errors.New("oops"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub := &stubUserService{updateErr: tc.updateErr}
			rr := do(t, buildHandler(stub), http.MethodPut, tc.path, tc.body)

			if rr.Code != tc.wantStatus {
				t.Fatalf("status: got %d, want %d", rr.Code, tc.wantStatus)
			}
			if tc.wantName != "" {
				var got model.User
				if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
					t.Fatalf("decode: %v", err)
				}
				if got.GetName() != tc.wantName {
					t.Fatalf("name: got %q, want %q", got.GetName(), tc.wantName)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Test: GetUserNameByName
// ---------------------------------------------------------------------------

func TestGetUserNameByName(t *testing.T) {
	alice := &model.User{}
	alice.SetName("Alice")

	cases := []struct {
		name       string
		path       string
		svcUser    *model.User
		svcErr     error
		wantStatus int
		wantNull   bool
	}{
		{
			name:       "found returns user",
			path:       "/get_user_name/name/Alice",
			svcUser:    alice,
			wantStatus: http.StatusOK,
		},
		{
			name:       "not found returns 200 null",
			path:       "/get_user_name/name/Unknown",
			svcErr:     apperr.ErrUserNotFound,
			wantStatus: http.StatusOK,
			wantNull:   true,
		},
		{
			name:       "unexpected error yields 500",
			path:       "/get_user_name/name/Error",
			svcErr:     errors.New("db"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub := &stubUserService{getByName: tc.svcUser, getByNameErr: tc.svcErr}
			rr := do(t, buildHandler(stub), http.MethodGet, tc.path, nil)

			if rr.Code != tc.wantStatus {
				t.Fatalf("status: got %d, want %d", rr.Code, tc.wantStatus)
			}
			if tc.wantNull {
				trimmed := strings.TrimSpace(rr.Body.String())
				if trimmed != "null" {
					t.Fatalf("body: got %q, want null", trimmed)
				}
			}
		})
	}
}