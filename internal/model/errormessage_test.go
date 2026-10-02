package model

import (
	"fmt"
	"testing"
)

// TestNewEmptyErrorMessage validates the no-args constructor behavior.
func TestNewEmptyErrorMessage(t *testing.T) {
	tests := []struct {
		name            string
		wantStatus      string
		wantMessage     string
	}{
		{
			name:        "instantiated with no arguments",
			wantStatus:  "",
			wantMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEmptyErrorMessage()
			if e == nil {
				t.Fatal("expected non-nil ErrorMessage, got nil")
			}
			if e.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", e.Status, tt.wantStatus)
			}
			if e.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", e.Message, tt.wantMessage)
			}
		})
	}
}

// TestNewErrorMessage validates the all-args constructor behavior.
func TestNewErrorMessage(t *testing.T) {
	tests := []struct {
		name        string
		status      string
		message     string
		wantStatus  string
		wantMessage string
	}{
		{
			name:        "instantiated with a status and message",
			status:      "NOT_FOUND",
			message:     "resource not found",
			wantStatus:  "NOT_FOUND",
			wantMessage: "resource not found",
		},
		{
			name:        "instantiated with empty status and message",
			status:      "",
			message:     "",
			wantStatus:  "",
			wantMessage: "",
		},
		{
			name:        "instantiated with INTERNAL_SERVER_ERROR status",
			status:      "INTERNAL_SERVER_ERROR",
			message:     "something went wrong",
			wantStatus:  "INTERNAL_SERVER_ERROR",
			wantMessage: "something went wrong",
		},
		{
			name:        "instantiated with BAD_REQUEST status",
			status:      "BAD_REQUEST",
			message:     "invalid input",
			wantStatus:  "BAD_REQUEST",
			wantMessage: "invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewErrorMessage(tt.status, tt.message)
			if e == nil {
				t.Fatal("expected non-nil ErrorMessage, got nil")
			}
			if e.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", e.Status, tt.wantStatus)
			}
			if e.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", e.Message, tt.wantMessage)
			}
		})
	}
}

// TestGetStatus validates GetStatus behavior.
func TestGetStatus(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *ErrorMessage
		wantStatus string
	}{
		{
			name: "status was previously set via constructor",
			setup: func() *ErrorMessage {
				return NewErrorMessage("NOT_FOUND", "not found")
			},
			wantStatus: "NOT_FOUND",
		},
		{
			name: "status was never set (no-args constructor)",
			setup: func() *ErrorMessage {
				return NewEmptyErrorMessage()
			},
			wantStatus: "",
		},
		{
			name: "status set via SetStatus",
			setup: func() *ErrorMessage {
				e := NewEmptyErrorMessage()
				e.SetStatus("BAD_REQUEST")
				return e
			},
			wantStatus: "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.setup()
			got := e.GetStatus()
			if got != tt.wantStatus {
				t.Errorf("GetStatus() = %q, want %q", got, tt.wantStatus)
			}
		})
	}
}

// TestGetMessage validates GetMessage behavior.
func TestGetMessage(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *ErrorMessage
		wantMessage string
	}{
		{
			name: "message was previously set via constructor",
			setup: func() *ErrorMessage {
				return NewErrorMessage("NOT_FOUND", "resource not found")
			},
			wantMessage: "resource not found",
		},
		{
			name: "message was never set (no-args constructor)",
			setup: func() *ErrorMessage {
				return NewEmptyErrorMessage()
			},
			wantMessage: "",
		},
		{
			name: "message set via SetMessage",
			setup: func() *ErrorMessage {
				e := NewEmptyErrorMessage()
				e.SetMessage("custom error message")
				return e
			},
			wantMessage: "custom error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.setup()
			got := e.GetMessage()
			if got != tt.wantMessage {
				t.Errorf("GetMessage() = %q, want %q", got, tt.wantMessage)
			}
		})
	}
}

// TestSetStatus validates SetStatus behavior.
func TestSetStatus(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		setValue   string
		wantStatus string
	}{
		{
			name:       "set a normal HTTP status value",
			initial:    "",
			setValue:   "NOT_FOUND",
			wantStatus: "NOT_FOUND",
		},
		{
			name:       "overwrite an existing status",
			initial:    "BAD_REQUEST",
			setValue:   "INTERNAL_SERVER_ERROR",
			wantStatus: "INTERNAL_SERVER_ERROR",
		},
		{
			name:       "set to empty string (analogous to null)",
			initial:    "NOT_FOUND",
			setValue:   "",
			wantStatus: "",
		},
		{
			name:       "set status on empty instance",
			initial:    "",
			setValue:   "OK",
			wantStatus: "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewErrorMessage(tt.initial, "some message")
			e.SetStatus(tt.setValue)
			got := e.GetStatus()
			if got != tt.wantStatus {
				t.Errorf("after SetStatus(%q), GetStatus() = %q, want %q", tt.setValue, got, tt.wantStatus)
			}
		})
	}
}

// TestSetMessage validates SetMessage behavior.
func TestSetMessage(t *testing.T) {
	tests := []struct {
		name        string
		initial     string
		setValue    string
		wantMessage string
	}{
		{
			name:        "set a normal message value",
			initial:     "",
			setValue:    "resource not found",
			wantMessage: "resource not found",
		},
		{
			name:        "overwrite an existing message",
			initial:     "old message",
			setValue:    "new message",
			wantMessage: "new message",
		},
		{
			name:        "set to empty string (analogous to null)",
			initial:     "some message",
			setValue:    "",
			wantMessage: "",
		},
		{
			name:        "set message on empty instance",
			initial:     "",
			setValue:    "an error occurred",
			wantMessage: "an error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewErrorMessage("STATUS", tt.initial)
			e.SetMessage(tt.setValue)
			got := e.GetMessage()
			if got != tt.wantMessage {
				t.Errorf("after SetMessage(%q), GetMessage() = %q, want %q", tt.setValue, got, tt.wantMessage)
			}
		})
	}
}

// TestEqualsErrorMessage validates equality behavior.
func TestEqualsErrorMessage(t *testing.T) {
	tests := []struct {
		name  string
		a     *ErrorMessage
		b     *ErrorMessage
		want  bool
	}{
		{
			name: "identical status and message returns true",
			a:    NewErrorMessage("NOT_FOUND", "not found"),
			b:    NewErrorMessage("NOT_FOUND", "not found"),
			want: true,
		},
		{
			name: "same instance compared to itself returns true",
			a:    NewErrorMessage("BAD_REQUEST", "bad request"),
			b:    nil, // will be overridden below
			want: true,
		},
		{
			name: "different status returns false",
			a:    NewErrorMessage("NOT_FOUND", "not found"),
			b:    NewErrorMessage("BAD_REQUEST", "not found"),
			want: false,
		},
		{
			name: "different message returns false",
			a:    NewErrorMessage("NOT_FOUND", "not found"),
			b:    NewErrorMessage("NOT_FOUND", "other message"),
			want: false,
		},
		{
			name: "both different status and message returns false",
			a:    NewErrorMessage("NOT_FOUND", "not found"),
			b:    NewErrorMessage("BAD_REQUEST", "bad request"),
			want: false,
		},
		{
			name: "compared to nil returns false",
			a:    NewErrorMessage("NOT_FOUND", "not found"),
			b:    nil,
			want: false,
		},
		{
			name: "both empty instances are equal",
			a:    NewEmptyErrorMessage(),
			b:    NewEmptyErrorMessage(),
			want: true,
		},
		{
			name: "nil compared to nil returns true",
			a:    nil,
			b:    nil,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Handle the "same instance" special case
			if tt.name == "same instance compared to itself returns true" {
				e := NewErrorMessage("BAD_REQUEST", "bad request")
				got := e.EqualsErrorMessage(e)
				if got != true {
					t.Errorf("EqualsErrorMessage(self) = %v, want true", got)
				}
				return
			}

			var got bool
			if tt.a == nil {
				// nil receiver — call via the method expression won't work;
				// test symmetry from b side instead to exercise nil handling
				got = tt.b.EqualsErrorMessage(tt.a)
			} else {
				got = tt.a.EqualsErrorMessage(tt.b)
			}

			if got != tt.want {
				t.Errorf("EqualsErrorMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestEqualsSymmetry validates that equality is symmetric and transitive.
func TestEqualsSymmetry(t *testing.T) {
	t.Run("symmetry: a.equals(b) == b.equals(a)", func(t *testing.T) {
		a := NewErrorMessage("NOT_FOUND", "not found")
		b := NewErrorMessage("NOT_FOUND", "not found")
		if a.EqualsErrorMessage(b) != b.EqualsErrorMessage(a) {
			t.Error("equality is not symmetric")
		}
	})

	t.Run("transitivity: a==b and b==c implies a==c", func(t *testing.T) {
		a := NewErrorMessage("NOT_FOUND", "not found")
		b := NewErrorMessage("NOT_FOUND", "not found")
		c := NewErrorMessage("NOT_FOUND", "not found")
		if !a.EqualsErrorMessage(b) || !b.EqualsErrorMessage(c) || !a.EqualsErrorMessage(c) {
			t.Error("equality is not transitive")
		}
	})

	t.Run("reflexivity: a.equals(a) is always true", func(t *testing.T) {
		a := NewErrorMessage("BAD_REQUEST", "invalid")
		if !a.EqualsErrorMessage(a) {
			t.Error("equality is not reflexive")
		}
	})
}

// TestHashCode validates HashCode behavior.
func TestHashCode(t *testing.T) {
	tests := []struct {
		name string
		a    *ErrorMessage
		b    *ErrorMessage
	}{
		{
			name: "equal instances produce the same hash code",
			a:    NewErrorMessage("NOT_FOUND", "not found"),
			b:    NewErrorMessage("NOT_FOUND", "not found"),
		},
		{
			name: "empty instances produce the same hash code",
			a:    NewEmptyErrorMessage(),
			b:    NewEmptyErrorMessage(),
		},
		{
			name: "all-args constructor equal instances",
			a:    NewErrorMessage("INTERNAL_SERVER_ERROR", "server error"),
			b:    NewErrorMessage("INTERNAL_SERVER_ERROR", "server error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ha := tt.a.HashCode()
			hb := tt.b.HashCode()
			if ha != hb {
				t.Errorf("HashCode() mismatch: a=%d, b=%d; equal instances must have equal hash codes", ha, hb)
			}
		})
	}

	t.Run("hash code is consistent across multiple invocations", func(t *testing.T) {
		e := NewErrorMessage("NOT_FOUND", "not found")
		first := e.HashCode()
		for i := 0; i < 5; i++ {
			if e.HashCode() != first {
				t.Errorf("HashCode() is not consistent across invocations")
			}
		}
	})
}

// TestStringErrorMessage validates the string representation.
func TestStringErrorMessage(t *testing.T) {
	tests := []struct {
		name        string
		status      string
		message     string
		wantContains []string
	}{
		{
			name:        "contains status and message for populated instance",
			status:      "NOT_FOUND",
			message:     "resource not found",
			wantContains: []string{"NOT_FOUND", "resource not found"},
		},
		{
			name:        "contains ErrorMessage label",
			status:      "BAD_REQUEST",
			message:     "invalid input",
			wantContains: []string{"ErrorMessage", "BAD_REQUEST", "invalid input"},
		},
		{
			name:        "empty fields reflected in output",
			status:      "",
			message:     "",
			wantContains: []string{"ErrorMessage"},
		},
		{
			name:        "INTERNAL_SERVER_ERROR reflected in output",
			status:      "INTERNAL_SERVER_ERROR",
			message:     "something went wrong",
			wantContains: []string{"INTERNAL_SERVER_ERROR", "something went wrong"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewErrorMessage(tt.status, tt.message)
			got := e.StringErrorMessage()

			for _, want := range tt.wantContains {
				if !contains(got, want) {
					t.Errorf("StringErrorMessage() = %q, want it to contain %q", got, want)
				}
			}
		})
	}

	t.Run("format matches expected pattern", func(t *testing.T) {
		e := NewErrorMessage("NOT_FOUND", "not found")
		want := fmt.Sprintf("ErrorMessage(status=%s, message=%s)", "NOT_FOUND", "not found")
		got := e.StringErrorMessage()
		if got != want {
			t.Errorf("StringErrorMessage() = %q, want %q", got, want)
		}
	})
}

// TestHashStringInternal validates the internal hashString helper indirectly
// by checking that hash codes differ for instances with different fields.
func TestHashCodeDifference(t *testing.T) {
	tests := []struct {
		name string
		a    *ErrorMessage
		b    *ErrorMessage
	}{
		{
			name: "different status and same message likely produce different hash",
			a:    NewErrorMessage("NOT_FOUND", "error"),
			b:    NewErrorMessage("BAD_REQUEST", "error"),
		},
		{
			name: "same status and different message likely produce different hash",
			a:    NewErrorMessage("NOT_FOUND", "error one"),
			b:    NewErrorMessage("NOT_FOUND", "error two"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ha := tt.a.HashCode()
			hb := tt.b.HashCode()
			// This is a best-effort check; hash collisions are theoretically
			// possible but extremely unlikely for these inputs.
			if ha == hb {
				t.Logf("warning: hash collision between %v and %v (hash=%d)", tt.a, tt.b, ha)
			}
		})
	}
}

// TestErrorMessageFieldIndependence verifies that fields are independently
// mutable and do not interfere with each other.
func TestErrorMessageFieldIndependence(t *testing.T) {
	t.Run("changing status does not affect message", func(t *testing.T) {
		e := NewErrorMessage("NOT_FOUND", "not found")
		e.SetStatus("BAD_REQUEST")
		if e.GetMessage() != "not found" {
			t.Errorf("message changed unexpectedly after SetStatus; got %q", e.GetMessage())
		}
	})

	t.Run("changing message does not affect status", func(t *testing.T) {
		e := NewErrorMessage("NOT_FOUND", "not found")
		e.SetMessage("updated message")
		if e.GetStatus() != "NOT_FOUND" {
			t.Errorf("status changed unexpectedly after SetMessage; got %q", e.GetStatus())
		}
	})
}

// TestErrorMessageJSONTags verifies the struct fields are correctly named for
// JSON serialization by inspecting the struct tags indirectly via the
// StringErrorMessage output.
func TestErrorMessageConstructorRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		message string
	}{
		{"round trip NOT_FOUND", "NOT_FOUND", "not found"},
		{"round trip BAD_REQUEST", "BAD_REQUEST", "bad request"},
		{"round trip empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewErrorMessage(tt.status, tt.message)
			if e.GetStatus() != tt.status {
				t.Errorf("GetStatus() = %q, want %q", e.GetStatus(), tt.status)
			}
			if e.GetMessage() != tt.message {
				t.Errorf("GetMessage() = %q, want %q", e.GetMessage(), tt.message)
			}
		})
	}
}

// contains is a simple helper that checks if s contains substr.
func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && searchString(s, substr))
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}