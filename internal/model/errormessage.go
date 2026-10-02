package model

import "fmt"

// ErrorMessage is a data transfer object representing an API error response.
// It bundles an HTTP status with a descriptive message.
//
// MIGRATION_NOTE: The source used Spring's HttpStatus enum for the Status
// field. Jackson serializes that enum by its NAME (e.g. "NOT_FOUND",
// "INTERNAL_SERVER_ERROR", "BAD_REQUEST"), not by numeric code, so Status is
// modeled here as a string holding the enum name to preserve wire format.
type ErrorMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// NewErrorMessage creates an ErrorMessage initialized with the given HTTP
// status (enum name) and message. Mirrors the Lombok @AllArgsConstructor.
func NewErrorMessage(status, message string) *ErrorMessage {
	return &ErrorMessage{
		Status:  status,
		Message: message,
	}
}

// NewEmptyErrorMessage creates an ErrorMessage with all fields unset
// (zero values). Mirrors the Lombok @NoArgsConstructor.
func NewEmptyErrorMessage() *ErrorMessage {
	return &ErrorMessage{}
}

// GetStatus returns the HTTP status currently stored.
func (e *ErrorMessage) GetStatus() string {
	return e.Status
}

// SetStatus sets the HTTP status.
func (e *ErrorMessage) SetStatus(status string) {
	e.Status = status
}

// GetMessage returns the message string currently stored.
func (e *ErrorMessage) GetMessage() string {
	return e.Message
}

// SetMessage sets the descriptive message.
func (e *ErrorMessage) SetMessage(message string) {
	e.Message = message
}

// EqualsErrorMessage reports whether two ErrorMessage instances are equal
// by value, comparing their status and message fields.
//
// MIGRATION_NOTE: Named EqualsErrorMessage rather than Equals to avoid
// colliding with the Equals method already defined in this package
// (internal/model/user.go).
func (e *ErrorMessage) EqualsErrorMessage(other *ErrorMessage) bool {
	if e == nil || other == nil {
		return e == other
	}
	return e.Status == other.Status && e.Message == other.Message
}

// HashCode produces a hash code derived from the status and message fields,
// mirroring Lombok's generated hashCode.
func (e *ErrorMessage) HashCode() int {
	const prime = 59
	result := 1
	result = result*prime + hashString(e.Status)
	result = result*prime + hashString(e.Message)
	return result
}

func hashString(s string) int {
	h := 0
	for _, r := range s {
		h = 31*h + int(r)
	}
	return h
}

// StringErrorMessage returns a human-readable representation including the
// status and message field values.
//
// MIGRATION_NOTE: Named StringErrorMessage rather than String to avoid
// colliding with the String method already defined in this package
// (internal/model/user.go).
func (e *ErrorMessage) StringErrorMessage() string {
	return fmt.Sprintf("ErrorMessage(status=%s, message=%s)", e.Status, e.Message)
}
