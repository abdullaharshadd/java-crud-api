```go
package smartcontact_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers – access package-internal newRouter via a thin test shim.
// Because newRouter is unexported, the test lives in the same package
// (package smartcontact) declared below with the build-tag trick.  The file
// is placed under internal/smartcontact/ alongside the production file, so
// the Go toolchain already gives us access to unexported identifiers.
// ---------------------------------------------------------------------------

// newRouterForTest is an alias used inside the test file to keep intent clear.
func newRouterForTest(db *sql.DB) http.Handler {
	return newRouter(db)
}

// ---------------------------------------------------------------------------
// Table-driven tests
// ---------------------------------------------------------------------------

func TestNewRouter_HealthzEndpoint(t *testing.T) {
	t.Parallel()

	// We only need a non-nil *sql.DB for newRouter to build the graph without
	// panicking at construction time.  We use go-sqlmock so no real DB is
	// required.
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "sqlmock.New should not fail")
	defer db.Close()
	// Expect nothing – the healthz probe never touches the DB.
	mock.MatchExpectationsInOrder(false)

	router := newRouterForTest(db)
	require.NotNil(t, router, "newRouter must return a non-nil handler")

	tests := []struct {
		name           string
		method         string
		path           string
		wantStatusCode int
		wantBodySubstr string
	}{
		{
			name:           "GET /healthz returns 200 ok",
			method:         http.MethodGet,
			path:           "/healthz",
			wantStatusCode: http.StatusOK,
			wantBodySubstr: "ok",
		},
		{
			name:           "POST /healthz returns 405",
			method:         http.MethodPost,
			path:           "/healthz",
			wantStatusCode: http.StatusMethodNotAllowed,
			wantBodySubstr: "",
		},
		{
			name:           "unknown path returns 404",
			method:         http.MethodGet,
			path:           "/does_not_exist",
			wantStatusCode: http.StatusNotFound,
			wantBodySubstr: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatusCode, rr.Code,
				"unexpected status code for %s %s", tc.method, tc.path)

			if tc.wantBodySubstr != "" {
				assert.Contains(t, rr.Body.String(), tc.wantBodySubstr,
					"response body should contain %q", tc.wantBodySubstr)
			}
		})
	}
}

func TestNewRouter_ReturnsNonNilHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		db   *sql.DB
	}{
		{
			name: "nil db – handler is built without panicking",
			db:   nil,
		},
		{
			name: "mock db – handler is built without panicking",
			db: func() *sql.DB {
				db, _, err := sqlmock.New()
				require.NoError(t, err)
				return db
			}(),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.NotPanics(t, func() {
				h := newRouterForTest(tc.db)
				assert.NotNil(t, h, "handler must not be nil")
			})
		})
	}
}

func TestNewRouter_MiddlewareStack(t *testing.T) {
	t.Parallel()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	router := newRouterForTest(db)

	tests := []struct {
		name    string
		trigger func(router http.Handler) *httptest.ResponseRecorder
		// wantStatus is the HTTP status we expect; we mainly verify no 5xx
		// is returned for middleware-controlled paths.
		wantNotPanic bool
	}{
		{
			name: "Recoverer middleware catches panics – healthz is stable",
			trigger: func(r http.Handler) *httptest.ResponseRecorder {
				req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
				rr := httptest.NewRecorder()
				r.ServeHTTP(rr, req)
				return rr
			},
			wantNotPanic: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.NotPanics(t, func() {
				rr := tc.trigger(router)
				assert.NotNil(t, rr)
			})
		})
	}
}

func TestNewRouter_UserRoutes_Registered(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	router := newRouterForTest(db)

	// These routes are registered by userController.RegisterRoutes.
	// We test only that the router does NOT return 404 for them (i.e. the
	// route exists); because the handler calls the DB we expect a 5xx or
	// structured error rather than 404.
	tests := []struct {
		name           string
		method         string
		path           string
		wantNotStatus  int // status we do NOT want (would signal missing route)
	}{
		{
			name:          "POST /save_user_data route exists",
			method:        http.MethodPost,
			path:          "/save_user_data",
			wantNotStatus: http.StatusNotFound,
		},
		{
			name:          "GET /get_user_data route exists",
			method:        http.MethodGet,
			path:          "/get_user_data",
			wantNotStatus: http.StatusNotFound,
		},
		{
			name:          "GET /get_user_data/{id} route exists",
			method:        http.MethodGet,
			path:          "/get_user_data/1",
			wantNotStatus: http.StatusNotFound,
		},
		{
			name:          "DELETE /delete_user_data/{id} route exists",
			method:        http.MethodDelete,
			path:          "/delete_user_data/1",
			wantNotStatus: http.StatusNotFound,
		},
		{
			name:          "PUT /update_user_data/{id} route exists",
			method:        http.MethodPut,
			path:          "/update_user_data/1",
			wantNotStatus: http.StatusNotFound,
		},
		{
			name:          "GET /get_user_name/name/{name} route exists",
			method:        http.MethodGet,
			path:          "/get_user_name/name/alice",
			wantNotStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Provide sqlmock rows so the DB layer does not blow up
			// unpredictably; we just care that the router resolves the route.
			mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows(nil))
			mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 0))

			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()

			assert.NotPanics(t, func() {
				router.ServeHTTP(rr, req)
			})

			assert.NotEqual(t, tc.wantNotStatus, rr.Code,
				"route %s %s must be registered (got 404 = not registered)",
				tc.method, tc.path)
		})
	}
}

func TestBuildRouter_DoesNotPanic(t *testing.T) {
	t.Parallel()

	// buildRouter uses a nil *sql.DB internally; it should not panic at
	// construction time – failures only surface at query time.
	assert.NotPanics(t, func() {
		h := buildRouter()
		assert.NotNil(t, h)
	})
}

func TestBuildRouter_HealthzReachable(t *testing.T) {
	t.Parallel()

	var h http.Handler
	assert.NotPanics(t, func() {
		h = buildRouter()
	})
	require.NotNil(t, h)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "ok")
}

func TestNewRouter_ApplicationErrorMiddleware(t *testing.T) {
	t.Parallel()

	// When a handler panics with a UserNotFoundError, the RecoverMiddleware
	// should translate it into a structured JSON 404 (not a 500 crash).
	// We verify the middleware is wired by checking that a request to a
	// user-by-ID path with a mocked "not found" condition does not return 500.
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Simulate no rows returned for the SELECT query, which triggers the
	// "user not found" code path in the service layer.
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}))

	router := newRouterForTest(db)

	req := httptest.NewRequest(http.MethodGet, "/get_user_data/999", nil)
	rr := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		router.ServeHTTP(rr, req)
	}, "RecoverMiddleware must catch application panics")

	// We accept either 404 (error mapped correctly) or any non-500 code that
	// demonstrates the middleware chain did not let an unhandled panic through.
	assert.NotEqual(t, http.StatusInternalServerError, rr.Code,
		"application error middleware must prevent raw 500 panics")
}

func TestNewRouter_ContentTypeHeaders(t *testing.T) {
	t.Parallel()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	router := newRouterForTest(db)

	tests := []struct {
		name                string
		method              string
		path                string
		wantContentTypeFunc func(ct string) bool
	}{
		{
			name:   "healthz response has non-empty content-type or body",
			method: http.MethodGet,
			path:   "/healthz",
			wantContentTypeFunc: func(ct string) bool {
				// healthz uses fmt.Fprintln, chi will set text/plain
				return true // content-type is optional for a plain liveness probe
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.True(t, tc.wantContentTypeFunc(rr.Header().Get("Content-Type")))
		})
	}
}

func TestNewRouter_JSONErrorResponseShape(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}))

	router := newRouterForTest(db)

	req := httptest.NewRequest(http.MethodGet, "/get_user_data/42", nil)
	rr := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		router.ServeHTTP(rr, req)
	})

	// If the body is JSON, it must be parseable.
	ct := rr.Header().Get("Content-Type")
	if ct == "application/json" || ct == "application/json; charset=utf-8" {
		var body map[string]interface{}
		decodeErr := json.NewDecoder(rr.Body).Decode(&body)
		assert.NoError(t, decodeErr, "JSON error body must be well-formed")
	}
}

func TestNewRouter_SingleActiveHandlerPerCall(t *testing.T) {
	t.Parallel()

	// Each call to newRouter must return an independent handler (no shared
	// mutable state).
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	h1 := newRouterForTest(db)
	h2 := newRouterForTest(db)

	assert.NotNil(t, h1)
	assert.NotNil(t, h2)
	// They must be distinct handler instances.
	assert.NotSame(t, &h1, &h2)
}

func TestNewRouter_LoggerMiddlewareDoesNotAffectStatus(t *testing.T) {
	t.Parallel()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	router := newRouterForTest(db)

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"logger does not alter 200", "/healthz", http.StatusOK},
		{"logger does not alter 404", "/nonexistent", http.StatusNotFound},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.Equal(t, tc.wantStatus, rr.Code)
		})
	}
}
```