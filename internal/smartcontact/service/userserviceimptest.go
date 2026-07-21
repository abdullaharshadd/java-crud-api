// Package service contains the SmartContact service's business-logic layer.
// It is the Go equivalent of the Java package com.smartContact.service.
package service

import (
	"context"
	"testing"

	"internal/smartcontact/model"
	"internal/smartcontact/repository"
)

// MIGRATION_NOTE: The source UserServiceImpTest.java was a Spring Boot
// integration test (@SpringBootTest) that was fundamentally broken. It
// attempted to stub a real @Autowired bean using Mockito.when(...):
//
//	@Autowired private UserServiceImp userServiceImp;
//	@BeforeAll void setUp() {
//	    User user = User.builder().name("hemraj")...build();
//	    Mockito.when(userServiceImp.getUserNameByName("hemraj")).thenReturn(user);
//	}
//
// Because userServiceImp was the concrete, container-managed bean (not a
// Mockito mock), the Mockito.when(...) call actually INVOKED the real method
// against the real database before attempting to stub it. Stubbing a
// non-mock is a Mockito misuse and the returned stub had no effect. The
// @BeforeAll method was also non-static without @TestInstance(PER_CLASS),
// which JUnit 5 would reject outright.
//
// The idiomatic Go migration abandons the broken integration/mocking hybrid
// and instead exercises the real business logic (GetUserByName) against a
// fake in-memory repository.UserDao. This preserves the ORIGINAL INTENT of
// the test -- "given a user named 'hemraj' exists, GetUserByName('hemraj')
// returns a user whose name is 'hemraj'" -- without depending on a database
// or a mocking framework.
//
// MANUAL REVIEW: If a true end-to-end integration test against a live
// database is desired, it should be written separately using a real
// repository.UserDao backed by a test database (e.g. via testcontainers).

// fakeUserDao is a minimal in-memory implementation of repository.UserDao
// used purely for testing the service layer in isolation.
//
// MIGRATION_NOTE: This replaces Mockito's dynamic proxy mock. Go favours
// small hand-written fakes over reflection-based mocking libraries, so we
// define an explicit test double that satisfies the UserDao interface.
type fakeUserDao struct {
	usersByName map[string]*model.User
}

// FindByName returns the user stored under the given name, or nil if absent.
func (f *fakeUserDao) FindByName(ctx context.Context, name string) (*model.User, error) {
	if u, ok := f.usersByName[name]; ok {
		return u, nil
	}
	return nil, nil
}

// FindByID is not exercised by these tests.
func (f *fakeUserDao) FindByID(ctx context.Context, id int) (*model.User, error) {
	return nil, nil
}

// FindAll is not exercised by these tests.
func (f *fakeUserDao) FindAll(ctx context.Context) ([]*model.User, error) {
	return nil, nil
}

// Save is not exercised by these tests.
func (f *fakeUserDao) Save(ctx context.Context, user *model.User) (*model.User, error) {
	return user, nil
}

// DeleteByID is not exercised by these tests.
func (f *fakeUserDao) DeleteByID(ctx context.Context, id int) error {
	return nil
}

// Compile-time assertion that fakeUserDao satisfies the UserDao interface.
var _ repository.UserDao = (*fakeUserDao)(nil)

// TestUserServiceImp_GetUserByName verifies that, given a persisted user with
// a known name, UserService.GetUserByName returns a user whose name matches.
//
// MIGRATION_NOTE: This is a table-driven rewrite of the single JUnit test
// WhenValidDepartmentName_ThenUserShouldBeFound. The original only covered
// the happy path; we add explicit not-found handling to exercise the error
// path the Go implementation returns.
func TestUserServiceImp_GetUserByName(t *testing.T) {
	hemraj := model.NewUser(3, "hemraj", "hemrajmalhi1234@gmail.com", "root", "Sr", "java developer")

	tests := []struct {
		name      string
		lookup    string
		wantName  string
		wantFound bool
	}{
		{
			name:      "valid name returns matching user",
			lookup:    "hemraj",
			wantName:  "hemraj",
			wantFound: true,
		},
		{
			name:      "unknown name returns error",
			lookup:    "nobody",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dao := &fakeUserDao{
				usersByName: map[string]*model.User{
					"hemraj": hemraj,
				},
			}
			svc := NewUserService(dao)

			got, err := svc.GetUserByName(context.Background(), tt.lookup)

			if tt.wantFound {
				if err != nil {
					t.Fatalf("GetUserByName(%q) returned unexpected error: %v", tt.lookup, err)
				}
				if got == nil {
					t.Fatalf("GetUserByName(%q) returned nil user, want name %q", tt.lookup, tt.wantName)
				}
				if got.Name != tt.wantName {
					t.Errorf("GetUserByName(%q).Name = %q, want %q", tt.lookup, got.Name, tt.wantName)
				}
				return
			}

			if err == nil {
				t.Errorf("GetUserByName(%q) expected an error, got user %+v", tt.lookup, got)
			}
		})
	}
}
