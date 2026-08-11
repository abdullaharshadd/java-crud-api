package smartcontact

// MIGRATION_NOTE: The Java source was a default Spring Boot smoke test
// (SmartContactApplicationTests) annotated with @SpringBootTest. It contained
// no test methods; the mere presence of @SpringBootTest causes Spring to load
// the full application context, and the test passes if the context wires up
// without error. There is no business logic to preserve.
//
// The idiomatic Go equivalent is a smoke test that constructs the same
// dependency graph the application wires at startup (repository -> service ->
// handler), registers the HTTP routes, and asserts that the router responds on
// each declared route. This mirrors the intent of the source test: prove the
// object graph assembles and the transport layer is reachable, without hitting
// a real database.
//
// Because Go's context-loading has no DI container, we assemble the graph by
// hand and use an in-memory fake UserService so the test needs no database.
// This file lives beside the package it exercises; the actual assertions live
// in smartcontactapplicationtests_test.go (Go test files must end in _test.go
// and cannot be named arbitrarily). This file provides the reusable test
// harness and route inventory that the _test.go file drives.

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"

	"internal/smartcontact/handler"
	"internal/smartcontact/model"
	"internal/smartcontact/service"
)

// SmokeRoute describes a single HTTP route that the application must expose.
// It is used by the context-load smoke test to verify every route the handler
// registers is reachable (i.e. does not return 404 "no route").
type SmokeRoute struct {
	// Method is the HTTP verb, e.g. http.MethodGet.
	Method string
	// Path is the concrete request path to exercise (path variables already
	// substituted with sample values).
	Path string
}

// SmokeRoutes returns the full inventory of routes the UserController handler
// registers, with sample values substituted for path variables. Keeping this
// list in one place lets the smoke test assert route wiring without depending
// on the handler's internal routing details.
func SmokeRoutes() []SmokeRoute {
	return []SmokeRoute{
		{Method: http.MethodPost, Path: "/api/users"},
		{Method: http.MethodGet, Path: "/api/users"},
		{Method: http.MethodGet, Path: "/api/users/1"},
		{Method: http.MethodDelete, Path: "/api/users/1"},
		{Method: http.MethodPut, Path: "/api/users/1"},
		{Method: http.MethodGet, Path: "/api/users/name/alice"},
	}
}

// fakeUserService is an in-memory implementation of service.UserService used
// by the context-load smoke test so it can run without a real database. Every
// method returns a benign zero value, mirroring the fact that the Spring
// context-loading test never exercised real behavior.
//
// MIGRATION_NOTE: The method set here must match service.UserService. If the
// interface changes, this fake must be updated; the Go compiler enforces that
// via the interface assignment in NewSmokeHandler.
type fakeUserService struct{}

func (fakeUserService) SaveUser(_ context.Context, u *model.User) (*model.User, error) {
	return u, nil
}

func (fakeUserService) FetchUserList(_ context.Context) ([]model.User, error) {
	return []model.User{}, nil
}

func (fakeUserService) FetchUserByID(_ context.Context, id int64) (*model.User, error) {
	return model.NewUser(), nil
}

func (fakeUserService) DeleteUser(_ context.Context, id int64) error {
	return nil
}

func (fakeUserService) UpdateUser(_ context.Context, id int64, u *model.User) (*model.User, error) {
	return u, nil
}

func (fakeUserService) GetUserByName(_ context.Context, name string) (*model.User, error) {
	return model.NewUser(), nil
}

// NewSmokeRouter assembles the same handler/route graph the application wires
// at startup, backed by an in-memory fake service so no database is required.
// It is the Go analogue of Spring loading the application context: if the
// dependency graph cannot be constructed, this function's callers will fail.
//
// The compile-time interface assertion guarantees fakeUserService satisfies
// service.UserService, so a drift between the fake and the real interface is
// caught at build time rather than at test time.
func NewSmokeRouter() http.Handler {
	var svc service.UserService = fakeUserService{}

	h := handler.NewHandler(svc)

	r := chi.NewRouter()
	h.RegisterRoutes(r)
	return r
}

// SmokeTest verifies the application "context" loads: it builds the full route
// graph and issues a request against every declared route, returning an error
// if any route is unrouted (HTTP 404 from the router) or if the router itself
// fails to serve. It is the direct equivalent of the Spring @SpringBootTest
// context-loading smoke test.
func SmokeTest(ctx context.Context) error {
	router := NewSmokeRouter()

	for _, route := range SmokeRoutes() {
		req := httptest.NewRequest(route.Method, route.Path, http.NoBody).WithContext(ctx)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		// A 404 means the router has no handler for this method+path, i.e.
		// the route was never registered. Any other status (including 4xx/5xx
		// from the handler body) still proves the route is wired.
		if rec.Code == http.StatusNotFound {
			return &RouteNotWiredError{Method: route.Method, Path: route.Path}
		}
	}
	return nil
}

// RouteNotWiredError indicates that a declared route was not registered on the
// router, i.e. the request returned HTTP 404 with no matching handler.
type RouteNotWiredError struct {
	// Method is the HTTP verb of the missing route.
	Method string
	// Path is the path of the missing route.
	Path string
}

// Error implements the error interface.
func (e *RouteNotWiredError) Error() string {
	return "route not wired: " + e.Method + " " + e.Path
}
