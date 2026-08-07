package model

// ErrorMessage is a data transfer object representing an error response
// returned to API clients. It carries an HTTP status and a descriptive
// human-readable message.
//
// MIGRATION_NOTE: The source Java type used Spring's HttpStatus enum for the
// status field. Go has no direct equivalent enum, and the standard library
// represents HTTP status codes as plain ints (see net/http). Since this DTO is
// serialized into a JSON error response (used by writeError), Status is
// modeled as a string so it can carry either the numeric code as text or a
// canonical status phrase (e.g. "Not Found"). Callers may use
// http.StatusText(code) or strconv.Itoa(code) to populate it. If a numeric
// status is preferred, change the field type to int and adjust callers.
type ErrorMessage struct {
	// Status is the HTTP status associated with the error.
	Status string `json:"status"`
	// Message is the descriptive error message.
	Message string `json:"message"`
}

// NewErrorMessage constructs an ErrorMessage with the given status and message.
func NewErrorMessage(status, message string) ErrorMessage {
	return ErrorMessage{
		Status:  status,
		Message: message,
	}
}
