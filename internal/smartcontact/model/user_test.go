```go
package model

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newValidUser() *User {
	return &User{
		ID:       1,
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "secret",
		Role:     "admin",
		About:    "A short bio",
	}
}

// ---------------------------------------------------------------------------
// Zero-value / no-args constructor
// ---------------------------------------------------------------------------

func TestUser_ZeroValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   func(t *testing.T, u *User)
	}{
		{
			name: "id defaults to 0",
			fn:   func(t *testing.T, u *User) { assert.Equal(t, 0, u.ID) },
		},
		{
			name: "name defaults to empty string",
			fn:   func(t *testing.T, u *User) { assert.Equal(t, "", u.Name) },
		},
		{
			name: "email defaults to empty string",
			fn:   func(t *testing.T, u *User) { assert.Equal(t, "", u.Email) },
		},
		{
			name: "password defaults to empty string",
			fn:   func(t *testing.T, u *User) { assert.Equal(t, "", u.Password) },
		},
		{
			name: "role defaults to empty string",
			fn:   func(t *testing.T, u *User) { assert.Equal(t, "", u.Role) },
		},
		{
			name: "about defaults to empty string",
			fn:   func(t *testing.T, u *User) { assert.Equal(t, "", u.About) },
		},
		{
			name: "zero-value User is not nil",
			fn:   func(t *testing.T, u *User) { assert.NotNil(t, u) },
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var u User
			tc.fn(t, &u)
		})
	}
}

// ---------------------------------------------------------------------------
// Field get / set round-trips (replaces Java getter/setter specs)
// ---------------------------------------------------------------------------

func TestUser_FieldRoundTrips(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "ID: setId then getId returns assigned value",
			fn: func(t *testing.T) {
				u := &User{}
				u.ID = 42
				assert.Equal(t, 42, u.ID)
			},
		},
		{
			name: "ID: id=0 denotes unpersisted user",
			fn: func(t *testing.T) {
				u := &User{}
				assert.Equal(t, 0, u.ID)
			},
		},
		{
			name: "Name: setName then getName returns assigned value",
			fn: func(t *testing.T) {
				u := &User{}
				u.Name = "Bob"
				assert.Equal(t, "Bob", u.Name)
			},
		},
		{
			name: "Name: unset name is empty string",
			fn: func(t *testing.T) {
				u := &User{}
				assert.Equal(t, "", u.Name)
			},
		},
		{
			name: "Email: setEmail then getEmail returns assigned value",
			fn: func(t *testing.T) {
				u := &User{}
				u.Email = "bob@example.com"
				assert.Equal(t, "bob@example.com", u.Email)
			},
		},
		{
			name: "Email: unset email is empty string",
			fn: func(t *testing.T) {
				u := &User{}
				assert.Equal(t, "", u.Email)
			},
		},
		{
			name: "Password: setPassword then getPassword returns assigned value",
			fn: func(t *testing.T) {
				u := &User{}
				u.Password = "p@ssw0rd"
				assert.Equal(t, "p@ssw0rd", u.Password)
			},
		},
		{
			name: "Password: unset password is empty string",
			fn: func(t *testing.T) {
				u := &User{}
				assert.Equal(t, "", u.Password)
			},
		},
		{
			name: "Role: setRole then getRole returns assigned value",
			fn: func(t *testing.T) {
				u := &User{}
				u.Role = "viewer"
				assert.Equal(t, "viewer", u.Role)
			},
		},
		{
			name: "Role: unset role is empty string",
			fn: func(t *testing.T) {
				u := &User{}
				assert.Equal(t, "", u.Role)
			},
		},
		{
			name: "About: setAbout then getAbout returns assigned value",
			fn: func(t *testing.T) {
				u := &User{}
				u.About = "Some bio"
				assert.Equal(t, "Some bio", u.About)
			},
		},
		{
			name: "About: unset about is empty string",
			fn: func(t *testing.T) {
				u := &User{}
				assert.Equal(t, "", u.About)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.fn(t)
		})
	}
}

// ---------------------------------------------------------------------------
// NewUser (all-args constructor)
// ---------------------------------------------------------------------------

func TestNewUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		inName   string
		inEmail  string
		inPwd    string
		inRole   string
		inAbout  string
		wantID   int
		wantName string
		wantEmail string
		wantPwd  string
		wantRole string
		wantAbout string
	}{
		{
			name:      "all fields populated",
			inName:    "Alice",
			inEmail:   "alice@example.com",
			inPwd:     "secret",
			inRole:    "admin",
			inAbout:   "I am Alice",
			wantID:    0,
			wantName:  "Alice",
			wantEmail: "alice@example.com",
			wantPwd:   "secret",
			wantRole:  "admin",
			wantAbout: "I am Alice",
		},
		{
			name:      "empty strings for optional fields",
			inName:    "Bob",
			inEmail:   "",
			inPwd:     "",
			inRole:    "",
			inAbout:   "",
			wantID:    0,
			wantName:  "Bob",
			wantEmail: "",
			wantPwd:   "",
			wantRole:  "",
			wantAbout: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := NewUser(tc.inName, tc.inEmail, tc.inPwd, tc.inRole, tc.inAbout)
			require.NotNil(t, got, "NewUser must never return nil")
			assert.Equal(t, tc.wantID, got.ID, "ID must default to 0")
			assert.Equal(t, tc.wantName, got.Name)
			assert.Equal(t, tc.wantEmail, got.Email)
			assert.Equal(t, tc.wantPwd, got.Password)
			assert.Equal(t, tc.wantRole, got.Role)
			assert.Equal(t, tc.wantAbout, got.About)
		})
	}
}

// ---------------------------------------------------------------------------
// Struct-literal "builder" equivalence
// ---------------------------------------------------------------------------

func TestUser_StructLiteralBuilder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   User
		want User
	}{
		{
			name: "all fields set via struct literal",
			in: User{
				ID:       7,
				Name:     "Charlie",
				Email:    "charlie@example.com",
				Password: "pw",
				Role:     "user",
				About:    "bio",
			},
			want: User{
				ID:       7,
				Name:     "Charlie",
				Email:    "charlie@example.com",
				Password: "pw",
				Role:     "user",
				About:    "bio",
			},
		},
		{
			name: "partial struct literal — unset fields default",
			in: User{
				Name: "Dave",
			},
			want: User{
				ID:       0,
				Name:     "Dave",
				Email:    "",
				Password: "",
				Role:     "",
				About:    "",
			},
		},
		{
			name: "empty struct literal",
			in:   User{},
			want: User{},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.in)
		})
	}
}

// ---------------------------------------------------------------------------
// Equality (replaces Java equals / hashCode specs)
// ---------------------------------------------------------------------------

func TestUser_Equality(t *testing.T) {
	t.Parallel()

	base := User{ID: 1, Name: "Alice", Email: "a@b.com", Password: "pw", Role: "admin", About: "bio"}

	tests := []struct {
		name      string
		a         User
		b         User
		wantEqual bool
	}{
		{
			name:      "identical fields are equal",
			a:         base,
			b:         base,
			wantEqual: true,
		},
		{
			name:      "reflexive equality",
			a:         base,
			b:         base,
			wantEqual: true,
		},
		{
			name:      "different ID makes users not equal",
			a:         base,
			b:         User{ID: 2, Name: "Alice", Email: "a@b.com", Password: "pw", Role: "admin", About: "bio"},
			wantEqual: false,
		},
		{
			name:      "different Name makes users not equal",
			a:         base,
			b:         User{ID: 1, Name: "Bob", Email: "a@b.com", Password: "pw", Role: "admin", About: "bio"},
			wantEqual: false,
		},
		{
			name:      "different Email makes users not equal",
			a:         base,
			b:         User{ID: 1, Name: "Alice", Email: "x@y.com", Password: "pw", Role: "admin", About: "bio"},
			wantEqual: false,
		},
		{
			name:      "different Password makes users not equal",
			a:         base,
			b:         User{ID: 1, Name: "Alice", Email: "a@b.com", Password: "other", Role: "admin", About: "bio"},
			wantEqual: false,
		},
		{
			name:      "different Role makes users not equal",
			a:         base,
			b:         User{ID: 1, Name: "Alice", Email: "a@b.com", Password: "pw", Role: "viewer", About: "bio"},
			wantEqual: false,
		},
		{
			name:      "different About makes users not equal",
			a:         base,
			b:         User{ID: 1, Name: "Alice", Email: "a@b.com", Password: "pw", Role: "admin", About: "different"},
			wantEqual: false,
		},
		{
			name:      "zero-value users are equal",
			a:         User{},
			b:         User{},
			wantEqual: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.wantEqual {
				assert.Equal(t, tc.a, tc.b)
			} else {
				assert.NotEqual(t, tc.a, tc.b)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// toString equivalent — fmt.Sprintf reflects all fields
// ---------------------------------------------------------------------------

func TestUser_StringRepresentation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		user           User
		mustContain    []string
		mustNotBeEmpty bool
	}{
		{
			name: "string contains all field values",
			user: User{
				ID:       5,
				Name:     "Eve",
				Email:    "eve@example.com",
				Password: "s3cr3t",
				Role:     "mod",
				About:    "about eve",
			},
			mustContain:    []string{"5", "Eve", "eve@example.com", "s3cr3t", "mod", "about eve"},
			mustNotBeEmpty: true,
		},
		{
			name:           "zero-value user string is not empty",
			user:           User{},
			mustContain:    []string{},
			mustNotBeEmpty: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := fmt.Sprintf("%+v", tc.user)
			if tc.mustNotBeEmpty {
				assert.NotEmpty(t, s)
			}
			for _, substr := range tc.mustContain {
				assert.Contains(t, s, substr, "string representation should contain %q", substr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Validate
// ---------------------------------------------------------------------------

func TestUser_Validate(t *testing.T) {
	t.Parallel()

	v := validator.New()

	tests := []struct {
		name      string
		user      User
		validator *validator.Validate
		wantErr   bool
		errTag    string // expected failing tag, if wantErr
	}{
		{
			name:      "valid user passes validation",
			user:      User{Name: "Alice", Email: "a@b.com", Password: "pw", Role: "admin", About: "bio"},
			validator: v,
			wantErr:   false,
		},
		{
			name:      "empty name fails required validation",
			user:      User{Name: "", Email: "a@b.com"},
			validator: v,
			wantErr:   true,
			errTag:    "required",
		},
		{
			name:      "about within 500 chars passes",
			user:      User{Name: "Alice", About: strings.Repeat("a", 500)},
			validator: v,
			wantErr:   false,
		},
		{
			name:      "about exceeding 500 chars fails max validation",
			user:      User{Name: "Alice", About: strings.Repeat("a", 501)},
			validator: v,
			wantErr:   true,
			errTag:    "max",
		},
		{
			name:      "nil validator uses internally created validator",
			user:      User{Name: "Alice"},
			validator: nil,
			wantErr:   false,
		},
		{
			name:      "nil validator catches required name violation