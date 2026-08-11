```go
package service

import (
	"context"
	"testing"

	"github.com/smartContact/internal/smartcontact/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserByName_TableDriven(t *testing.T) {
	hemraj := model.User{
		ID:       3,
		Name:     "hemraj",
		Email:    "hemrajmalhi1234@gmail.com",
		About:    "Sr",
		Password: "root",
		Role:     "java developer",
	}

	alice := model.User{
		ID:       7,
		Name:     "alice",
		Email:    "alice@example.com",
		About:    "Engineer",
		Password: "secret",
		Role:     "go developer",
	}

	tests := []struct {
		name          string
		seededUsers   []model.User
		lookupName    string
		wantErr       bool
		wantName      string
		wantEmail     string
		wantAbout     string
		wantPassword  string
		wantRole      string
		wantID        int
	}{
		{
			name:         "valid name 'hemraj' returns matching user with correct fields",
			seededUsers:  []model.User{hemraj},
			lookupName:   "hemraj",
			wantErr:      false,
			wantName:     "hemraj",
			wantEmail:    "hemrajmalhi1234@gmail.com",
			wantAbout:    "Sr",
			wantPassword: "root",
			wantRole:     "java developer",
			wantID:       3,
		},
		{
			name:        "valid name 'alice' returns matching user with correct fields",
			seededUsers: []model.User{alice},
			lookupName:  "alice",
			wantErr:     false,
			wantName:    "alice",
			wantEmail:   "alice@example.com",
			wantAbout:   "Engineer",
			wantPassword: "secret",
			wantRole:    "go developer",
			wantID:      7,
		},
		{
			name:        "multiple users seeded, lookup finds correct one by name",
			seededUsers: []model.User{hemraj, alice},
			lookupName:  "hemraj",
			wantErr:     false,
			wantName:    "hemraj",
			wantEmail:   "hemrajmalhi1234@gmail.com",
			wantAbout:   "Sr",
			wantPassword: "root",
			wantRole:    "java developer",
			wantID:      3,
		},
		{
			name:        "multiple users seeded, lookup finds alice specifically",
			seededUsers: []model.User{hemraj, alice},
			lookupName:  "alice",
			wantErr:     false,
			wantName:    "alice",
			wantEmail:   "alice@example.com",
			wantAbout:   "Engineer",
			wantPassword: "secret",
			wantRole:    "go developer",
			wantID:      7,
		},
		{
			name:        "unknown name returns error (user not found)",
			seededUsers: []model.User{hemraj},
			lookupName:  "nobody",
			wantErr:     true,
		},
		{
			name:        "empty name returns error (user not found)",
			seededUsers: []model.User{hemraj},
			lookupName:  "",
			wantErr:     true,
		},
		{
			name:        "empty repository returns error for any lookup",
			seededUsers: []model.User{},
			lookupName:  "hemraj",
			wantErr:     true,
		},
		{
			name:        "case-sensitive lookup does not match different case",
			seededUsers: []model.User{hemraj},
			lookupName:  "Hemraj",
			wantErr:     true,
		},
		{
			name:        "lookup with trailing space does not match stored name",
			seededUsers: []model.User{hemraj},
			lookupName:  "hemraj ",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeUserRepository(tt.seededUsers...)
			svc := NewUserService(repo)

			found, err := svc.GetUserByName(context.Background(), tt.lookupName)

			if tt.wantErr {
				assert.Error(t, err, "expected an error for lookup %q but got none", tt.lookupName)
				return
			}

			require.NoError(t, err, "unexpected error for lookup %q: %v", tt.lookupName, err)

			// Invariant: returned user's name must equal the requested name.
			assert.Equal(t, tt.wantName, found.Name,
				"returned user name should equal the requested name")

			// Global invariant: consistent field values as stored.
			assert.Equal(t, tt.wantEmail, found.Email,
				"returned user email should match stored email")
			assert.Equal(t, tt.wantAbout, found.About,
				"returned user about should match stored about")
			assert.Equal(t, tt.wantPassword, found.Password,
				"returned user password should match stored password")
			assert.Equal(t, tt.wantRole, found.Role,
				"returned user role should match stored role")
			assert.Equal(t, tt.wantID, found.ID,
				"returned user ID should match stored ID")
		})
	}
}

func TestGetUserByName_ReturnedUserNameMatchesRequest(t *testing.T) {
	// This test specifically validates the invariant:
	// "When a user is found, the returned User's name must equal the requested name."
	users := []model.User{
		{ID: 1, Name: "alice", Email: "alice@example.com", About: "dev", Password: "pass1", Role: "engineer"},
		{ID: 2, Name: "bob", Email: "bob@example.com", About: "qa", Password: "pass2", Role: "tester"},
		{ID: 3, Name: "hemraj", Email: "hemrajmalhi1234@gmail.com", About: "Sr", Password: "root", Role: "java developer"},
	}

	names := []string{"alice", "bob", "hemraj"}

	repo := newFakeUserRepository(users...)
	svc := NewUserService(repo)

	for _, name := range names {
		name := name
		t.Run("invariant_name_matches_"+name, func(t *testing.T) {
			found, err := svc.GetUserByName(context.Background(), name)
			require.NoError(t, err)
			assert.Equal(t, name, found.Name,
				"invariant violation: returned user name %q does not equal requested name %q",
				found.Name, name)
		})
	}
}

func TestGetUserByName_NotFoundReturnsNullEquivalent(t *testing.T) {
	// Validates: "may return null when user not found" error case from spec.
	tests := []struct {
		name       string
		lookupName string
	}{
		{name: "lookup nonexistent user 'ghost'", lookupName: "ghost"},
		{name: "lookup nonexistent user 'unknown'", lookupName: "unknown"},
		{name: "lookup empty string", lookupName: ""},
	}

	hemraj := model.User{
		ID:       3,
		Name:     "hemraj",
		Email:    "hemrajmalhi1234@gmail.com",
		About:    "Sr",
		Password: "root",
		Role:     "java developer",
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeUserRepository(hemraj)
			svc := NewUserService(repo)

			_, err := svc.GetUserByName(context.Background(), tt.lookupName)
			assert.Error(t, err,
				"GetUserByName(%q) should return an error when user does not exist", tt.lookupName)
		})
	}
}

func TestGetUserByName_ServiceLookupCapability(t *testing.T) {
	// Validates global invariant: "The service must be able to look up users by their name."
	hemraj := model.User{
		ID:       3,
		Name:     "hemraj",
		Email:    "hemrajmalhi1234@gmail.com",
		About:    "Sr",
		Password: "root",
		Role:     "java developer",
	}

	repo := newFakeUserRepository(hemraj)
	svc := NewUserService(repo)

	found, err := svc.GetUserByName(context.Background(), "hemraj")
	require.NoError(t, err, "service must be able to look up users by name")
	assert.NotEmpty(t, found.Name, "returned user must have a name")
}

func TestFakeUserRepository_FindByName(t *testing.T) {
	hemraj := model.User{
		ID:       3,
		Name:     "hemraj",
		Email:    "hemrajmalhi1234@gmail.com",
		About:    "Sr",
		Password: "root",
		Role:     "java developer",
	}

	tests := []struct {
		name       string
		seedUsers  []model.User
		lookupName string
		wantFound  bool
		wantUser   model.User
	}{
		{
			name:       "finds seeded user by name",
			seedUsers:  []model.User{hemraj},
			lookupName: "hemraj",
			wantFound:  true,
			wantUser:   hemraj,
		},
		{
			name:       "returns not found for missing user",
			seedUsers:  []model.User{hemraj},
			lookupName: "missing",
			wantFound:  false,
		},
		{
			name:       "empty repo returns not found",
			seedUsers:  []model.User{},
			lookupName: "hemraj",
			wantFound:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeUserRepository(tt.seedUsers...)
			u, ok, err := repo.FindByName(context.Background(), tt.lookupName)
			require.NoError(t, err)
			assert.Equal(t, tt.wantFound, ok)
			if tt.wantFound {
				assert.Equal(t, tt.wantUser, u)
			}
		})
	}
}

func TestFakeUserRepository_Save(t *testing.T) {
	newUser := model.User{
		ID:       10,
		Name:     "newuser",
		Email:    "new@example.com",
		About:    "new",
		Password: "newpass",
		Role:     "developer",
	}

	repo := newFakeUserRepository()
	saved, err := repo.Save(context.Background(), newUser)
	require.NoError(t, err)
	assert.Equal(t, newUser, saved)

	found, ok, err := repo.FindByName(context.Background(), "newuser")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, newUser, found)
}

func TestFakeUserRepository_DeleteByID(t *testing.T) {
	hemraj := model.User{
		ID:       3,
		Name:     "hemraj",
		Email:    "hemrajmalhi1234@gmail.com",
		About:    "Sr",
		Password: "root",
		Role:     "java developer",
	}

	repo := newFakeUserRepository(hemraj)
	err := repo.DeleteByID(context.Background(), 3)
	require.NoError(t, err)

	_, ok, err := repo.FindByName(context.Background(), "hemraj")
	require.NoError(t, err)
	assert.False(t, ok, "user should be deleted")
}

func TestFakeUserRepository_FindAll(t *testing.T) {
	users := []model.User{
		{ID: 1, Name: "alice"},
		{ID: 2, Name: "bob"},
	}

	repo := newFakeUserRepository(users...)
	all, err := repo.FindAll(context.Background())
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestFakeUserRepository_FindByID(t *testing.T) {
	hemraj := model.User{
		ID:       3,
		Name:     "hemraj",
		Email:    "hemrajmalhi1234@gmail.com",
		About:    "Sr",
		Password: "root",
		Role:     "java developer",
	}

	tests := []struct {
		name      string
		lookupID  int
		wantFound bool
		wantUser  model.User
	}{
		{
			name:      "finds user by existing ID",
			lookupID:  3,
			wantFound: true,
			wantUser:  hemraj,
		},
		{
			name:      "returns not found for missing ID",
			lookupID:  999,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeUserRepository(hemraj)
			u, ok, err := repo.FindByID(context.Background(), tt.lookupID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantFound, ok)
			if tt.wantFound {
				assert.Equal(t, tt.wantUser, u)
			}
		})
	}
}
```