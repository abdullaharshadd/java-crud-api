// Package httpapi contains the HTTP-layer types shared by handlers, such as
// the JSON error payload returned to clients.
package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorMessage is the JSON error body returned by the API
// (source: com.smartContact.model.ErrorMessage).
//
// Status holds the Spring HttpStatus enum constant name (e.g. "NOT_FOUND"),
// because Jackson serialises the enum by name, so the wire format is
// {"status":"NOT_FOUND","message":"..."}.
type ErrorMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// NewEmptyErrorMessage returns an ErrorMessage with all fields unset
// (source: Lombok @NoArgsConstructor).
func NewEmptyErrorMessage() *ErrorMessage {
	return &ErrorMessage{}
}

// NewErrorMessage returns an ErrorMessage for the given HTTP status code and
// message, in source declaration order: status first, then message
// (source: Lombok @AllArgsConstructor). The code is converted to the Spring
// HttpStatus enum name.
func NewErrorMessage(status int, message string) *ErrorMessage {
	return &ErrorMessage{Status: StatusName(status), Message: message}
}

// GetStatus returns the status enum name.
func (e *ErrorMessage) GetStatus() string { return e.Status }

// SetStatus sets the status from an HTTP status code.
func (e *ErrorMessage) SetStatus(status int) { e.Status = StatusName(status) }

// GetMessage returns the human-readable message.
func (e *ErrorMessage) GetMessage() string { return e.Message }

// SetMessage sets the human-readable message.
func (e *ErrorMessage) SetMessage(message string) { e.Message = message }

// Equal reports whether e and other have the same status and message
// (source: Lombok @Data equals). Two nil values are equal.
func (e *ErrorMessage) Equal(other *ErrorMessage) bool {
	if e == nil || other == nil {
		return e == other
	}
	return e.Status == other.Status && e.Message == other.Message
}

// String renders the value in Lombok @Data toString format, e.g.
// "ErrorMessage(status=NOT_FOUND, message=User are not available)".
// Unset fields render as "null", matching Java.
func (e *ErrorMessage) String() string {
	if e == nil {
		return "null"
	}
	return fmt.Sprintf("ErrorMessage(status=%s, message=%s)", orNull(e.Status), orNull(e.Message))
}

func orNull(s string) string {
	if s == "" {
		return "null"
	}
	return s
}

// WriteError writes an ErrorMessage as JSON with the given HTTP status code
// (source: RestResponseEntityExceptionHandling building a ResponseEntity).
func WriteError(w http.ResponseWriter, status int, message string) error {
	body, err := json.Marshal(NewErrorMessage(status, message))
	if err != nil {
		return fmt.Errorf("httpapi: marshal error message: %w", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		return fmt.Errorf("httpapi: write error message: %w", err)
	}
	return nil
}

// statusNames maps HTTP status codes to Spring's HttpStatus enum names.
var statusNames = map[int]string{
	100: "CONTINUE",
	101: "SWITCHING_PROTOCOLS",
	102: "PROCESSING",
	103: "CHECKPOINT",
	200: "OK",
	201: "CREATED",
	202: "ACCEPTED",
	203: "NON_AUTHORITATIVE_INFORMATION",
	204: "NO_CONTENT",
	205: "RESET_CONTENT",
	206: "PARTIAL_CONTENT",
	207: "MULTI_STATUS",
	208: "ALREADY_REPORTED",
	226: "IM_USED",
	300: "MULTIPLE_CHOICES",
	301: "MOVED_PERMANENTLY",
	302: "FOUND",
	303: "SEE_OTHER",
	304: "NOT_MODIFIED",
	307: "TEMPORARY_REDIRECT",
	308: "PERMANENT_REDIRECT",
	400: "BAD_REQUEST",
	401: "UNAUTHORIZED",
	402: "PAYMENT_REQUIRED",
	403: "FORBIDDEN",
	404: "NOT_FOUND",
	405: "METHOD_NOT_ALLOWED",
	406: "NOT_ACCEPTABLE",
	407: "PROXY_AUTHENTICATION_REQUIRED",
	408: "REQUEST_TIMEOUT",
	409: "CONFLICT",
	410: "GONE",
	411: "LENGTH_REQUIRED",
	412: "PRECONDITION_FAILED",
	413: "PAYLOAD_TOO_LARGE",
	414: "URI_TOO_LONG",
	415: "UNSUPPORTED_MEDIA_TYPE",
	416: "REQUESTED_RANGE_NOT_SATISFIABLE",
	417: "EXPECTATION_FAILED",
	418: "I_AM_A_TEAPOT",
	422: "UNPROCESSABLE_ENTITY",
	423: "LOCKED",
	424: "FAILED_DEPENDENCY",
	425: "TOO_EARLY",
	426: "UPGRADE_REQUIRED",
	428: "PRECONDITION_REQUIRED",
	429: "TOO_MANY_REQUESTS",
	431: "REQUEST_HEADER_FIELDS_TOO_LARGE",
	451: "UNAVAILABLE_FOR_LEGAL_REASONS",
	500: "INTERNAL_SERVER_ERROR",
	501: "NOT_IMPLEMENTED",
	502: "BAD_GATEWAY",
	503: "SERVICE_UNAVAILABLE",
	504: "GATEWAY_TIMEOUT",
	505: "HTTP_VERSION_NOT_SUPPORTED",
	506: "VARIANT_ALSO_NEGOTIATES",
	507: "INSUFFICIENT_STORAGE",
	508: "LOOP_DETECTED",
	509: "BANDWIDTH_LIMIT_EXCEEDED",
	510: "NOT_EXTENDED",
	511: "NETWORK_AUTHENTICATION_REQUIRED",
}

// StatusName returns the Spring HttpStatus enum name for code. Codes Spring
// does not define yield an empty string (Java has no enum constant for them).
func StatusName(code int) string {
	return statusNames[code]
}
