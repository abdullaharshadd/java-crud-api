// Package httpapi contains the HTTP-layer types and helpers shared by
// handlers: the JSON error payload returned to clients and the explicit
// replacement for the source's global exception handler
// (com.smartContact.error.RestResponseEntityExceptionHandling).
//
// MIGRATION_NOTE: Spring's @ControllerAdvice intercepted exceptions thrown by
// any controller. Go has no implicit interception, so handlers must pass
// failures to HandleError, which uses errors.Is/errors.As to choose the
// response. Wrap the mux with Recoverer to turn panics into 500s.
//
// The default handling inherited from ResponseEntityExceptionHandler is
// reproduced by these helpers:
//   - WriteEmpty: status code with an empty body. Use it for 400 (validation
//     failures, malformed JSON, unparsable path ids) and 415 (non-JSON
//     Content-Type on POST/PUT).
//   - WriteMethodNotAllowed: 405 with an Allow header and an empty body.
//   - WriteSpringError: Spring Boot 2.7's default /error JSON
//     {timestamp,status,error,path}. Use it for unhandled 500s and for
//     unknown-route 404s (see NotFoundHandler).
//
// REQUIRES MANUAL REVIEW: every handler in cmd/server/router.go must send
// errors through HandleError and use the helpers above for framework-level
// failures. Otherwise the wire format will differ from the Spring service.
package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"migrated-app/internal/user"
)

// ErrorMessage is the JSON error body returned by the API
// (source: com.smartContact.model.ErrorMessage).
//
// Status holds the Spring HttpStatus enum constant name (e.g. "NOT_FOUND").
// Jackson serialises the enum by name, so the wire format is
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
// message (source: Lombok @AllArgsConstructor). The status code is converted
// to its Spring HttpStatus enum name.
func NewErrorMessage(status int, message string) *ErrorMessage {
	return &ErrorMessage{Status: StatusName(status), Message: message}
}

// GetStatus returns the Spring HttpStatus name stored in the message.
func (e *ErrorMessage) GetStatus() string { return e.Status }

// SetStatus sets the Spring HttpStatus name stored in the message.
func (e *ErrorMessage) SetStatus(status string) { e.Status = status }

// GetMessage returns the human-readable error message.
func (e *ErrorMessage) GetMessage() string { return e.Message }

// SetMessage sets the human-readable error message.
func (e *ErrorMessage) SetMessage(message string) { e.Message = message }

// Equal reports whether e and other carry the same status and message
// (source: Lombok @Data equals). Two nil pointers are equal.
func (e *ErrorMessage) Equal(other *ErrorMessage) bool {
	if e == nil || other == nil {
		return e == other
	}
	return e.Status == other.Status && e.Message == other.Message
}

// String renders the message in Lombok's toString format.
func (e *ErrorMessage) String() string {
	if e == nil {
		return "ErrorMessage(status=null, message=null)"
	}
	return fmt.Sprintf("ErrorMessage(status=%s, message=%s)", e.Status, e.Message)
}

// WriteError writes an ErrorMessage JSON body with the given HTTP status.
func WriteError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, NewErrorMessage(status, message))
}

// StatusName converts an HTTP status code to the matching Spring HttpStatus
// enum constant name, e.g. 404 -> "NOT_FOUND". Codes Spring does not know
// produce the upper-snake form of net/http's status text. If that is empty
// too, the result is the decimal code.
func StatusName(code int) string {
	if name, ok := springStatusNames[code]; ok {
		return name
	}
	text := http.StatusText(code)
	if text == "" {
		return fmt.Sprintf("%d", code)
	}
	r := strings.NewReplacer(" ", "_", "-", "_", "'", "")
	return strings.ToUpper(r.Replace(text))
}

// springStatusNames lists the codes whose Spring enum name differs from a
// mechanical conversion of net/http's StatusText, plus the common codes.
var springStatusNames = map[int]string{
	http.StatusOK:                    "OK",
	http.StatusCreated:               "CREATED",
	http.StatusNoContent:             "NO_CONTENT",
	http.StatusBadRequest:            "BAD_REQUEST",
	http.StatusUnauthorized:          "UNAUTHORIZED",
	http.StatusForbidden:             "FORBIDDEN",
	http.StatusNotFound:              "NOT_FOUND",
	http.StatusMethodNotAllowed:      "METHOD_NOT_ALLOWED",
	http.StatusNotAcceptable:         "NOT_ACCEPTABLE",
	http.StatusConflict:              "CONFLICT",
	http.StatusRequestEntityTooLarge: "PAYLOAD_TOO_LARGE",
	http.StatusRequestURITooLong:     "URI_TOO_LONG",
	http.StatusUnsupportedMediaType:  "UNSUPPORTED_MEDIA_TYPE",
	http.StatusTeapot:                "I_AM_A_TEAPOT",
	http.StatusInternalServerError:   "INTERNAL_SERVER_ERROR",
	http.StatusServiceUnavailable:    "SERVICE_UNAVAILABLE",
}

// HandleError converts an error returned by a handler's service call into an
// HTTP response. It replaces the @ControllerAdvice/@ExceptionHandler pair:
//
//   - user not found (user.ErrNotFound / *user.NotFoundError) -> 404 with an
//     ErrorMessage body {"status":"NOT_FOUND","message":<exception message>}
//   - anything else -> 500 with Spring Boot's default error JSON
//
// A nil err writes nothing.
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	var nf *user.NotFoundError
	switch {
	case errors.As(err, &nf):
		// Use the exception's own message, as exception.getMessage() did,
		// and not any fmt.Errorf wrapping prefixes added on the way up.
		WriteError(w, http.StatusNotFound, nf.Error())
	case errors.Is(err, user.ErrNotFound):
		WriteError(w, http.StatusNotFound, user.NotFoundMessage)
	default:
		log.Error().Err(err).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("unhandled error while processing request")
		WriteSpringError(w, r, http.StatusInternalServerError)
	}
}

// springErrorBody mirrors Spring Boot 2.7's DefaultErrorAttributes output.
// Since 2.3 the "message" attribute is excluded by default.
type springErrorBody struct {
	Timestamp string `json:"timestamp"`
	Status    int    `json:"status"`
	Error     string `json:"error"`
	Path      string `json:"path"`
}

// springTimestampLayout matches Jackson's default java.util.Date format used
// by Spring Boot, e.g. "2024-01-02T03:04:05.678+00:00".
const springTimestampLayout = "2006-01-02T15:04:05.000-07:00"

// WriteSpringError writes Spring Boot's default error JSON
// {timestamp,status,error,path} with the given status code.
func WriteSpringError(w http.ResponseWriter, r *http.Request, status int) {
	path := ""
	if r != nil && r.URL != nil {
		path = r.URL.Path
	}
	writeJSON(w, status, springErrorBody{
		Timestamp: time.Now().UTC().Format(springTimestampLayout),
		Status:    status,
		Error:     http.StatusText(status),
		Path:      path,
	})
}

// WriteEmpty writes only the status code with an empty body. This is what
// ResponseEntityExceptionHandler returns for standard Spring MVC exceptions
// such as MethodArgumentNotValidException, HttpMessageNotReadableException,
// MethodArgumentTypeMismatchException (400) and
// HttpMediaTypeNotSupportedException (415).
func WriteEmpty(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
}

// WriteMethodNotAllowed writes a 405 with an empty body and an Allow header
// listing the permitted methods, matching Spring's handling of
// HttpRequestMethodNotSupportedException.
func WriteMethodNotAllowed(w http.ResponseWriter, allowed ...string) {
	if len(allowed) > 0 {
		w.Header().Set("Allow", strings.Join(allowed, ", "))
	}
	WriteEmpty(w, http.StatusMethodNotAllowed)
}

// NotFoundHandler returns a handler for unknown routes. It responds with
// Spring Boot's default 404 error JSON.
func NotFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteSpringError(w, r, http.StatusNotFound)
	})
}

// Recoverer is middleware that turns a panic in a downstream handler into a
// 500 response with Spring's default error JSON. This mirrors how an uncaught
// RuntimeException surfaced in the source. http.ErrAbortHandler is re-panicked
// so net/http keeps its abort semantics.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			if rec == http.ErrAbortHandler {
				panic(rec)
			}
			log.Error().
				Interface("panic", rec).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("recovered panic in HTTP handler")
			WriteSpringError(w, r, http.StatusInternalServerError)
		}()
		next.ServeHTTP(w, r)
	})
}

// writeJSON encodes v as the JSON response body with the given status.
// Encoding failures are logged because the status line has already been sent.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error().Err(err).Int("status", status).Msg("failed to encode JSON error response")
	}
}
