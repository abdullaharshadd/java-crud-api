package model

import (
	"fmt"
	"hash/fnv"
	"net/http"
)

// httpStatusNames maps numeric HTTP status codes to the constant names of
// Spring's org.springframework.http.HttpStatus enum, which is how Jackson
// serializes the status field (e.g. "NOT_FOUND").
var httpStatusNames = map[int]string{
	http.StatusContinue:                      "CONTINUE",
	http.StatusSwitchingProtocols:            "SWITCHING_PROTOCOLS",
	http.StatusOK:                            "OK",
	http.StatusCreated:                       "CREATED",
	http.StatusAccepted:                      "ACCEPTED",
	http.StatusNonAuthoritativeInfo:          "NON_AUTHORITATIVE_INFORMATION",
	http.StatusNoContent:                     "NO_CONTENT",
	http.StatusResetContent:                  "RESET_CONTENT",
	http.StatusPartialContent:                "PARTIAL_CONTENT",
	http.StatusMultipleChoices:               "MULTIPLE_CHOICES",
	http.StatusMovedPermanently:              "MOVED_PERMANENTLY",
	http.StatusFound:                         "FOUND",
	http.StatusSeeOther:                      "SEE_OTHER",
	http.StatusNotModified:                   "NOT_MODIFIED",
	http.StatusTemporaryRedirect:             "TEMPORARY_REDIRECT",
	http.StatusPermanentRedirect:             "PERMANENT_REDIRECT",
	http.StatusBadRequest:                    "BAD_REQUEST",
	http.StatusUnauthorized:                  "UNAUTHORIZED",
	http.StatusPaymentRequired:               "PAYMENT_REQUIRED",
	http.StatusForbidden:                     "FORBIDDEN",
	http.StatusNotFound:                      "NOT_FOUND",
	http.StatusMethodNotAllowed:              "METHOD_NOT_ALLOWED",
	http.StatusNotAcceptable:                 "NOT_ACCEPTABLE",
	http.StatusProxyAuthRequired:             "PROXY_AUTHENTICATION_REQUIRED",
	http.StatusRequestTimeout:                "REQUEST_TIMEOUT",
	http.StatusConflict:                      "CONFLICT",
	http.StatusGone:                          "GONE",
	http.StatusLengthRequired:                "LENGTH_REQUIRED",
	http.StatusPreconditionFailed:            "PRECONDITION_FAILED",
	http.StatusRequestEntityTooLarge:         "PAYLOAD_TOO_LARGE",
	http.StatusRequestURITooLong:             "URI_TOO_LONG",
	http.StatusUnsupportedMediaType:          "UNSUPPORTED_MEDIA_TYPE",
	http.StatusRequestedRangeNotSatisfiable:  "REQUESTED_RANGE_NOT_SATISFIABLE",
	http.StatusExpectationFailed:             "EXPECTATION_FAILED",
	http.StatusTeapot:                        "I_AM_A_TEAPOT",
	http.StatusUnprocessableEntity:           "UNPROCESSABLE_ENTITY",
	http.StatusLocked:                        "LOCKED",
	http.StatusFailedDependency:              "FAILED_DEPENDENCY",
	http.StatusTooEarly:                      "TOO_EARLY",
	http.StatusUpgradeRequired:               "UPGRADE_REQUIRED",
	http.StatusPreconditionRequired:          "PRECONDITION_REQUIRED",
	http.StatusTooManyRequests:               "TOO_MANY_REQUESTS",
	http.StatusRequestHeaderFieldsTooLarge:   "REQUEST_HEADER_FIELDS_TOO_LARGE",
	http.StatusUnavailableForLegalReasons:    "UNAVAILABLE_FOR_LEGAL_REASONS",
	http.StatusInternalServerError:           "INTERNAL_SERVER_ERROR",
	http.StatusNotImplemented:                "NOT_IMPLEMENTED",
	http.StatusBadGateway:                    "BAD_GATEWAY",
	http.StatusServiceUnavailable:            "SERVICE_UNAVAILABLE",
	http.StatusGatewayTimeout:                "GATEWAY_TIMEOUT",
	http.StatusHTTPVersionNotSupported:       "HTTP_VERSION_NOT_SUPPORTED",
	http.StatusVariantAlsoNegotiates:         "VARIANT_ALSO_NEGOTIATES",
	http.StatusInsufficientStorage:           "INSUFFICIENT_STORAGE",
	http.StatusLoopDetected:                  "LOOP_DETECTED",
	http.StatusNotExtended:                   "NOT_EXTENDED",
	http.StatusNetworkAuthenticationRequired: "NETWORK_AUTHENTICATION_REQUIRED",
}

// HTTPStatusName returns the Spring HttpStatus enum constant name for the
// given numeric status code (e.g. 404 -> "NOT_FOUND"). The boolean is false
// when the code has no corresponding Spring constant.
func HTTPStatusName(code int) (string, bool) {
	name, ok := httpStatusNames[code]
	return name, ok
}

// ErrorMessage is the JSON error payload returned to clients when a request
// fails (e.g. a user lookup that finds nothing). Status carries the Spring
// HttpStatus enum name such as "NOT_FOUND", matching how Jackson serialized
// the original enum field; an empty Status means "unset" (Java null).
type ErrorMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// NewEmptyErrorMessage returns an ErrorMessage with every field unset
// (equivalent of the Lombok @NoArgsConstructor).
func NewEmptyErrorMessage() *ErrorMessage {
	return &ErrorMessage{}
}

// NewErrorMessage returns an ErrorMessage for the given numeric HTTP status
// and message (equivalent of the Lombok @AllArgsConstructor). The status is
// stored as its Spring enum name; unknown codes leave Status unset because
// Spring's HttpStatus enum cannot represent them either.
func NewErrorMessage(status int, message string) *ErrorMessage {
	name, _ := HTTPStatusName(status)
	return &ErrorMessage{Status: name, Message: message}
}

// GetStatus returns the HTTP status enum name.
func (e *ErrorMessage) GetStatus() string { return e.Status }

// SetStatus sets the HTTP status from a numeric code, storing its Spring
// enum name. It returns false (leaving Status unchanged) for unknown codes.
func (e *ErrorMessage) SetStatus(status int) bool {
	name, ok := HTTPStatusName(status)
	if !ok {
		return false
	}
	e.Status = name
	return true
}

// GetMessage returns the human-readable error message.
func (e *ErrorMessage) GetMessage() string { return e.Message }

// SetMessage sets the human-readable error message.
func (e *ErrorMessage) SetMessage(message string) { e.Message = message }

// Equal reports whether e and other carry the same status and message.
// Two nil pointers are equal; a nil and a non-nil pointer are not.
func (e *ErrorMessage) Equal(other *ErrorMessage) bool {
	if e == nil || other == nil {
		return e == other
	}
	return e.Status == other.Status && e.Message == other.Message
}

// HashCode returns a hash derived from the status and message fields; equal
// values always produce equal hashes.
func (e *ErrorMessage) HashCode() uint64 {
	if e == nil {
		return 0
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(e.Status))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(e.Message))
	return h.Sum64()
}

// String renders the value in Lombok's toString format, e.g.
// "ErrorMessage(status=NOT_FOUND, message=User not found)". Unset fields are
// rendered as "null" like the Java original.
func (e *ErrorMessage) String() string {
	if e == nil {
		return "null"
	}
	return fmt.Sprintf("ErrorMessage(status=%s, message=%s)",
		errorMessageFieldOrNull(e.Status), errorMessageFieldOrNull(e.Message))
}

func errorMessageFieldOrNull(s string) string {
	if s == "" {
		return "null"
	}
	return s
}
