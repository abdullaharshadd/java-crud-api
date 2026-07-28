```go
package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func strPtr(s string) *string { return &s }

// ---------------------------------------------------------------------------
// TestUser_ZeroValue – "no-args constructor" spec
// ---------------------------------------------------------------------------

func TestUser_ZeroValue(t *testing.T) {
	tests := []struct {
		name string
		fn   func() User
		want User
	}{
		{
			name: "instantiate with no arguments – all fields at zero value",
			fn:   func() User { return User{} },
			want: User{ID: 0, Name: "", Email: "", Password: "", Role: "", About: ""},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.fn()
			assert.Equal(t, tc.want.ID, got.ID, "ID should be 0")
			assert.Equal(t, tc.want.Name, got.Name, "Name should be empty string")
			assert.Equal(t, tc.want.Email, got.Email, "Email should be empty string")
			assert.Equal(t, tc.want.Password, got.Password, "Password should be empty string")
			assert.Equal(t, tc.want.Role, got.Role, "Role should be empty string")
			assert.Equal(t, tc.want.About, got.About, "About should be empty string")
		})
	}
}

// ---------------------------------------------------------------------------
// TestNewUser – "all-args constructor" spec
// ---------------------------------------------------------------------------

func TestNewUser(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		uName    string
		email    string
		password string
		role     string
		about    string
	}{
		{
			name:     "all fields provided",
			id:       42,
			uName:    "Alice",
			email:    "alice@example.com",
			password: "s3cr3t",
			role:     "ADMIN",
			about:    "About Alice",
		},
		{
			name:     "zero id with non-empty strings",
			id:       0,
			uName:    "Bob",
			email:    "bob@example.com",
			password: "pass",
			role:     "USER",
			about:    "",
		},
		{
			name:     "empty strings",
			id:       0,
			uName:    "",
			email:    "",
			password: "",
			role:     "",
			about:    "",
		},
		{
			name:     "about at max length (500 chars)",
			id:       1,
			uName:    "Charlie",
			email:    "charlie@example.com",
			password: "pw",
			role:     "USER",
			about:    string(make([]byte, 500)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := NewUser(tc.id, tc.uName, tc.email, tc.password, tc.role, tc.about)
			assert.NotNil(t, u)
			assert.Equal(t, tc.id, u.ID)
			assert.Equal(t, tc.uName, u.Name)
			assert.Equal(t, tc.email, u.Email)
			assert.Equal(t, tc.password, u.Password)
			assert.Equal(t, tc.role, u.Role)
			assert.Equal(t, tc.about, u.About)
		})
	}
}

// ---------------------------------------------------------------------------
// TestUser_Builder – "builder" spec (via struct literal pattern)
// ---------------------------------------------------------------------------

func TestUser_Builder(t *testing.T) {
	tests := []struct {
		name string
		fn   func() User
		want User
	}{
		{
			name: "build with subset of fields – unset fields at defaults",
			fn: func() User {
				return User{Name: "Dave", Email: "dave@example.com"}
			},
			want: User{ID: 0, Name: "Dave", Email: "dave@example.com", Password: "", Role: "", About: ""},
		},
		{
			name: "build with all fields set – equivalent to all-args constructor",
			fn: func() User {
				return User{ID: 7, Name: "Eve", Email: "eve@example.com", Password: "pw", Role: "MOD", About: "Hi"}
			},
			want: User{ID: 7, Name: "Eve", Email: "eve@example.com", Password: "pw", Role: "MOD", About: "Hi"},
		},
		{
			name: "build with only ID set",
			fn:   func() User { return User{ID: 99} },
			want: User{ID: 99},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.fn()
			assert.Equal(t, tc.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// TestUser_GetSetID – getId / setId specs
// ---------------------------------------------------------------------------

func TestUser_GetSetID(t *testing.T) {
	tests := []struct {
		name    string
		initial int
		setTo   int
	}{
		{"set positive id", 0, 10},
		{"set zero id", 5, 0},
		{"set negative id", 0, -1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &User{ID: tc.initial}
			assert.Equal(t, tc.initial, u.ID, "initial ID")
			u.ID = tc.setTo
			assert.Equal(t, tc.setTo, u.ID, "after mutation")
		})
	}
}

// ---------------------------------------------------------------------------
// TestUser_GetSetName – getName / setName specs
// ---------------------------------------------------------------------------

func TestUser_GetSetName(t *testing.T) {
	tests := []struct {
		name    string
		initial string
		setTo   string
	}{
		{"set non-empty name", "", "Alice"},
		{"overwrite name", "Alice", "Bob"},
		{"set empty name", "Alice", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &User{Name: tc.initial}
			assert.Equal(t, tc.initial, u.Name)
			u.Name = tc.setTo
			assert.Equal(t, tc.setTo, u.Name)
		})
	}
}

// ---------------------------------------------------------------------------
// TestUser_GetSetEmail – getEmail / setEmail specs
// ---------------------------------------------------------------------------

func TestUser_GetSetEmail(t *testing.T) {
	tests := []struct {
		name    string
		initial string
		setTo   string
	}{
		{"set valid email", "", "user@example.com"},
		{"overwrite email", "old@example.com", "new@example.com"},
		{"set empty email", "user@example.com", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &User{Email: tc.initial}
			assert.Equal(t, tc.initial, u.Email)
			u.Email = tc.setTo
			assert.Equal(t, tc.setTo, u.Email)
		})
	}
}

// ---------------------------------------------------------------------------
// TestUser_GetSetPassword – getPassword / setPassword specs
// ---------------------------------------------------------------------------

func TestUser_GetSetPassword(t *testing.T) {
	tests := []struct {
		name    string
		initial string
		setTo   string
	}{
		{"set hashed password", "", "$2a$10$xyz"},
		{"overwrite password", "old", "new"},
		{"set empty password", "old", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &User{Password: tc.initial}
			assert.Equal(t, tc.initial, u.Password)
			u.Password = tc.setTo
			assert.Equal(t, tc.setTo, u.Password)
		})
	}
}

// ---------------------------------------------------------------------------
// TestUser_GetSetRole – getRole / setRole specs
// ---------------------------------------------------------------------------

func TestUser_GetSetRole(t *testing.T) {
	tests := []struct {
		name    string
		initial string
		setTo   string
	}{
		{"set ADMIN role", "", "ADMIN"},
		{"change role to USER", "ADMIN", "USER"},
		{"set empty role", "USER", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &User{Role: tc.initial}
			assert.Equal(t, tc.initial, u.Role)
			u.Role = tc.setTo
			assert.Equal(t, tc.setTo, u.Role)
		})
	}
}

// ---------------------------------------------------------------------------
// TestUser_GetSetAbout – getAbout / setAbout specs
// ---------------------------------------------------------------------------

func TestUser_GetSetAbout(t *testing.T) {
	longAbout := string(make([]byte, 500))
	tests := []struct {
		name    string
		initial string
		setTo   string
	}{
		{"set about text", "", "About me"},
		{"overwrite about", "Old about", "New about"},
		{"set empty about", "Some text", ""},
		{"set max-length about (500 chars)", "", longAbout},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &User{About: tc.initial}
			assert.Equal(t, tc.initial, u.About)
			u.About = tc.setTo
			assert.Equal(t, tc.setTo, u.About)
		})
	}
}

// ---------------------------------------------------------------------------
// TestUser_Equality – equals spec
// ---------------------------------------------------------------------------

func TestUser_Equality(t *testing.T) {
	base := User{ID: 1, Name: "Alice", Email: "a@a.com", Password: "pw", Role: "ADMIN", About: "hi"}

	tests := []struct {
		name   string
		a, b   User
		wantEq bool
	}{
		{
			name:   "identical users are equal",
			a:      base,
			b:      base,
			wantEq: true,
		},
		{
			name:   "reflexive equality",
			a:      base,
			b:      base,
			wantEq: true,
		},
		{
			name:   "differ in ID",
			a:      base,
			b:      User{ID: 2, Name: "Alice", Email: "a@a.com", Password: "pw", Role: "ADMIN", About: "hi"},
			wantEq: false,
		},
		{
			name:   "differ in Name",
			a:      base,
			b:      User{ID: 1, Name: "Bob", Email: "a@a.com", Password: "pw", Role: "ADMIN", About: "hi"},
			wantEq: false,
		},
		{
			name:   "differ in Email",
			a:      base,
			b:      User{ID: 1, Name: "Alice", Email: "b@b.com", Password: "pw", Role: "ADMIN", About: "hi"},
			wantEq: false,
		},
		{
			name:   "differ in Password",
			a:      base,
			b:      User{ID: 1, Name: "Alice", Email: "a@a.com", Password: "other", Role: "ADMIN", About: "hi"},
			wantEq: false,
		},
		{
			name:   "differ in Role",
			a:      base,
			b:      User{ID: 1, Name: "Alice", Email: "a@a.com", Password: "pw", Role: "USER", About: "hi"},
			wantEq: false,
		},
		{
			name:   "differ in About",
			a:      base,
			b:      User{ID: 1, Name: "Alice", Email: "a@a.com", Password: "pw", Role: "ADMIN", About: "bye"},
			wantEq: false,
		},
		{
			name:   "zero-value users are equal",
			a:      User{},
			b:      User{},
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

// ---------------------------------------------------------------------------
// TestUser_ToString – toString spec
// ---------------------------------------------------------------------------

func TestUser_ToString(t *testing.T) {
	tests := []struct {
		name     string
		u        User
		contains []string
	}{
		{
			name: "populated user – fmt.Sprintf contains all field values",
			u:    User{ID: 5, Name: "Alice", Email: "a@a.com", Password: "pw", Role: "ADMIN", About: "hi"},
			contains: []string{"5", "Alice", "a@a.com", "pw", "ADMIN", "hi"},
		},
		{
			name:     "zero-value user",
			u:        User{},
			contains: []string{"0"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := fmt.Sprintf("%+v", tc.u)
			for _, sub := range tc.contains {
				assert.Contains(t, s, sub)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestNsToPtr
// ---------------------------------------------------------------------------

func TestNsToPtr(t *testing.T) {
	tests := []struct {
		name    string
		input   sql.NullString
		wantNil bool
		wantVal string
	}{
		{
			name:    "valid NullString returns pointer to value",
			input:   sql.NullString{String: "hello", Valid: true},
			wantNil: false,
			wantVal: "hello",
		},
		{
			name:    "invalid NullString returns nil",
			input:   sql.NullString{String: "", Valid: false},
			wantNil: true,
		},
		{
			name:    "valid NullString with empty string returns pointer to empty string",
			input:   sql.NullString{String: "", Valid: true},
			wantNil: false,
			wantVal: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := nsToPtr(tc.input)
			if tc.wantNil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.Equal(t, tc.wantVal, *got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestUserRow_ToResponse
// ---------------------------------------------------------------------------

func TestUserRow_ToResponse(t *testing.T) {
	tests := []struct {
		name string
		row  userRow
		want UserResponse
	}{
		{
			name: "all fields valid",
			row: userRow{
				ID:       1,
				Name:     sql.NullString{String: "Alice", Valid: true},
				Email:    sql.NullString{String: "a@a.com", Valid: true},
				Password: sql.NullString{String: "secret", Valid: true},
				Role:     sql.NullString{String: "ADMIN", Valid: true},
				About:    sql.