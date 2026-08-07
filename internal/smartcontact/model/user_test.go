```go
package model_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"internal/smartcontact/model"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeUser(id int64, name, email, password, role, about string) *model.User {
	u := model.NewUser(name, email, password, role, about)
	u.ID = id
	return u
}

// ---------------------------------------------------------------------------
// No-arg construction (zero-value struct)
// ---------------------------------------------------------------------------

func TestUser_NoArgConstruction(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "instantiate with no arguments yields zero values"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var u model.User

			assert.Equal(t, int64(0), u.ID, "id must be 0 for a zero-value User")
			assert.Equal(t, "", u.Name, "name must be empty string")
			assert.Equal(t, "", u.Email, "email must be empty string")
			assert.Equal(t, "", u.Password, "password must be empty string")
			assert.Equal(t, "", u.Role, "role must be empty string")
			assert.Equal(t, "", u.About, "about must be empty string")
		})
	}
}

// ---------------------------------------------------------------------------
// All-args construction via NewUser + manual ID set
// ---------------------------------------------------------------------------

func TestUser_AllArgsConstruction(t *testing.T) {
	tests := []struct {
		name     string
		id       int64
		uname    string
		email    string
		password string
		role     string
		about    string
	}{
		{
			name:     "all fields set – typical user",
			id:       42,
			uname:    "Alice",
			email:    "alice@example.com",
			password: "s3cr3t",
			role:     "ADMIN",
			about:    "Lead developer",
		},
		{
			name:     "all fields set – another user",
			id:       1,
			uname:    "Bob",
			email:    "bob@example.com",
			password: "hunter2",
			role:     "USER",
			about:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := makeUser(tc.id, tc.uname, tc.email, tc.password, tc.role, tc.about)

			assert.Equal(t, tc.id, u.ID)
			assert.Equal(t, tc.uname, u.Name)
			assert.Equal(t, tc.email, u.Email)
			assert.Equal(t, tc.password, u.Password)
			assert.Equal(t, tc.role, u.Role)
			assert.Equal(t, tc.about, u.About)
		})
	}
}

// ---------------------------------------------------------------------------
// "Builder" path – partial construction using NewUser then field assignment
// ---------------------------------------------------------------------------

func TestUser_BuilderStyleConstruction(t *testing.T) {
	tests := []struct {
		name        string
		setName     bool
		nameVal     string
		setEmail    bool
		emailVal    string
		setPassword bool
		passwordVal string
		setRole     bool
		roleVal     string
		setAbout    bool
		aboutVal    string
		wantID      int64
	}{
		{
			name:    "build with no fields – all defaults",
			wantID:  0,
		},
		{
			name:     "build with only name set",
			setName:  true,
			nameVal:  "Charlie",
			wantID:   0,
		},
		{
			name:        "build with a subset of fields",
			setName:     true,
			nameVal:     "Dana",
			setEmail:    true,
			emailVal:    "dana@example.com",
			setPassword: true,
			passwordVal: "pw123",
			wantID:      0,
		},
		{
			name:     "build with all fields explicitly",
			setName:  true,
			nameVal:  "Eve",
			setEmail: true,
			emailVal: "eve@example.com",
			setRole:  true,
			roleVal:  "MOD",
			setAbout: true,
			aboutVal: "Moderator",
			wantID:   0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var u model.User // start from defaults

			if tc.setName {
				u.Name = tc.nameVal
			}
			if tc.setEmail {
				u.Email = tc.emailVal
			}
			if tc.setPassword {
				u.Password = tc.passwordVal
			}
			if tc.setRole {
				u.Role = tc.roleVal
			}
			if tc.setAbout {
				u.About = tc.aboutVal
			}

			// ID is always zero unless explicitly set
			assert.Equal(t, tc.wantID, u.ID)

			if tc.setName {
				assert.Equal(t, tc.nameVal, u.Name)
			} else {
				assert.Equal(t, "", u.Name)
			}
			if tc.setEmail {
				assert.Equal(t, tc.emailVal, u.Email)
			} else {
				assert.Equal(t, "", u.Email)
			}
			if tc.setPassword {
				assert.Equal(t, tc.passwordVal, u.Password)
			} else {
				assert.Equal(t, "", u.Password)
			}
			if tc.setRole {
				assert.Equal(t, tc.roleVal, u.Role)
			} else {
				assert.Equal(t, "", u.Role)
			}
			if tc.setAbout {
				assert.Equal(t, tc.aboutVal, u.About)
			} else {
				assert.Equal(t, "", u.About)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// getId / setId
// ---------------------------------------------------------------------------

func TestUser_IDGetterSetter(t *testing.T) {
	tests := []struct {
		name  string
		setID int64
	}{
		{name: "set id = 0", setID: 0},
		{name: "set id = 1", setID: 1},
		{name: "set id = 9999", setID: 9999},
		{name: "set id = max int64", setID: 9223372036854775807},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var u model.User
			u.ID = tc.setID
			assert.Equal(t, tc.setID, u.ID)
		})
	}
}

// ---------------------------------------------------------------------------
// getName / setName + validation
// ---------------------------------------------------------------------------

func TestUser_NameGetterSetter(t *testing.T) {
	tests := []struct {
		name    string
		nameVal string
		want    string
	}{
		{name: "ordinary name", nameVal: "Alice", want: "Alice"},
		{name: "name with spaces", nameVal: "John Doe", want: "John Doe"},
		{name: "unicode name", nameVal: "日本語", want: "日本語"},
		{name: "empty string round-trip", nameVal: "", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var u model.User
			u.Name = tc.nameVal
			assert.Equal(t, tc.want, u.Name)
		})
	}
}

func TestUser_NameValidation(t *testing.T) {
	tests := []struct {
		name        string
		nameVal     string
		wantErr     bool
		errContains string
	}{
		{name: "valid name", nameVal: "Alice", wantErr: false},
		{name: "empty name fails validation", nameVal: "", wantErr: true, errContains: "name"},
		{name: "whitespace-only is non-empty string (no trim in Go)", nameVal: "   ", wantErr: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &model.User{Name: tc.nameVal}
			err := u.Validate()
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errContains != "" {
					assert.ErrorIs(t, err, model.ErrNameRequired)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// getEmail / setEmail
// ---------------------------------------------------------------------------

func TestUser_EmailGetterSetter(t *testing.T) {
	tests := []struct {
		name     string
		emailVal string
	}{
		{name: "typical email", emailVal: "user@example.com"},
		{name: "email with subdomain", emailVal: "user@mail.example.org"},
		{name: "empty email", emailVal: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var u model.User
			u.Email = tc.emailVal
			assert.Equal(t, tc.emailVal, u.Email)
		})
	}
}

// ---------------------------------------------------------------------------
// getPassword / setPassword
// ---------------------------------------------------------------------------

func TestUser_PasswordGetterSetter(t *testing.T) {
	tests := []struct {
		name        string
		passwordVal string
	}{
		{name: "plain password", passwordVal: "s3cr3t"},
		{name: "bcrypt hash", passwordVal: "$2a$10$abcdefghijklmnopqrstuuVGNy3Mc4AKJrD2m4s91zHN.jLdOcLhe"},
		{name: "empty password", passwordVal: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var u model.User
			u.Password = tc.passwordVal
			assert.Equal(t, tc.passwordVal, u.Password)
		})
	}
}

// ---------------------------------------------------------------------------
// getRole / setRole
// ---------------------------------------------------------------------------

func TestUser_RoleGetterSetter(t *testing.T) {
	tests := []struct {
		name    string
		roleVal string
	}{
		{name: "ADMIN role", roleVal: "ADMIN"},
		{name: "USER role", roleVal: "USER"},
		{name: "empty role", roleVal: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var u model.User
			u.Role = tc.roleVal
			assert.Equal(t, tc.roleVal, u.Role)
		})
	}
}

// ---------------------------------------------------------------------------
// getAbout / setAbout
// ---------------------------------------------------------------------------

func TestUser_AboutGetterSetter(t *testing.T) {
	tests := []struct {
		name     string
		aboutVal string
		wantLen  int
	}{
		{name: "short about", aboutVal: "Go developer", wantLen: 12},
		{name: "empty about", aboutVal: "", wantLen: 0},
		{name: "exactly 500 chars", aboutVal: strings.Repeat("x", 500), wantLen: 500},
		// Note: >500 chars is allowed in-memory; DB constraint is enforced at persistence layer.
		{name: "501 chars stored in memory", aboutVal: strings.Repeat("y", 501), wantLen: 501},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var u model.User
			u.About = tc.aboutVal
			assert.Equal(t, tc.aboutVal, u.About)
			assert.Len(t, u.About, tc.wantLen)
		})
	}
}

// ---------------------------------------------------------------------------
// equals (struct comparison)
// ---------------------------------------------------------------------------

func TestUser_Equals(t *testing.T) {
	tests := []struct {
		name     string
		a        model.User
		b        model.User
		wantEq   bool
	}{
		{
			name:   "identical Users are equal",
			a:      model.User{ID: 1, Name: "Alice", Email: "a@b.com", Password: "pw", Role: "ADMIN", About: "x"},
			b:      model.User{ID: 1, Name: "Alice", Email: "a@b.com", Password: "pw", Role: "ADMIN", About: "x"},
			wantEq: true,
		},
		{
			name:   "differ in ID",
			a:      model.User{ID: 1, Name: "Alice"},
			b:      model.User{ID: 2, Name: "Alice"},
			wantEq: false,
		},
		{
			name:   "differ in Name",
			a:      model.User{ID: 1, Name: "Alice"},
			b:      model.User{ID: 1, Name: "Bob"},
			wantEq: false,
		},
		{
			name:   "differ in Email",
			a:      model.User{ID: 1, Email: "a@b.com"},
			b:      model.User{ID: 1, Email: "c@d.com"},
			wantEq: false,
		},
		{
			name:   "differ in Password",
			a:      model.User{ID: 1, Password: "pw1"},
			b:      model.User{ID: 1, Password: "pw2"},
			wantEq: false,
		},
		{
			name:   "differ in Role",
			a:      model.User{ID: 1, Role: "ADMIN"},
			b:      model.User{ID: 1, Role: "USER"},
			wantEq: false,
		},
		{
			name:   "differ in About",
			a:      model.User{ID: 1, About: "hello"},
			b:      model.User{ID: 1, About: "world"},
			wantEq: false,
		},
		{
			name:   "zero-value Users are equal",
			a:      model.User{},
			b:      model.User{},
			wantEq: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantEq {
				assert.Equal(t, tc.a, tc.b)
			} else {
				assert.NotEqual(t, tc.a, tc.b)
			}
		})
	}
}

// Reflexive, symmetric, transitive checks
func TestUser_EqualsReflexiveSymmetricTransitive(t *testing.T) {
	a := model.User{ID: 1, Name: "Alice", Email: "a@example.com", Password: "pw", Role: "ADMIN", About: "dev"}
	b := model.User{ID: 1, Name: "Alice", Email: "a@example.com", Password: "pw", Role: "ADMIN", About: "dev"}
	c := model.User{ID: 1, Name: "Alice", Email: "a@example.com", Password: "pw", Role: "ADMIN", About: "dev"}

	// Reflexive
	assert.Equal(t, a, a, "reflexive: a == a")

	// Symmetric
	assert.Equal(t, a, b, "symmetric: a == b")
	assert.Equal(t, b, a, "symmetric: b == a")

	// Transitive
	assert.Equal(t, a, b)
	assert.Equal(t, b, c)
	assert.