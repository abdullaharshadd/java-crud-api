// Package httpapi contains the HTTP transport layer: handlers, middleware and
// the centralized mapping from domain/framework errors to HTTP responses.
//
// This file replaces the source's @ControllerAdvice
// RestResponseEntityExceptionHandling (which extends Spring's
// ResponseEntityExceptionHandler) plus the Spring Boot BasicErrorController
// fallback that renders unhandled failures as Boot error JSON.
package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"migrated-app/internal/model"
	"migrated-app/internal/service"
)

// Sentinel errors standing in for the Spring MVC framework exceptions that
// ResponseEntityExceptionHandler handles by default. Handlers wrap these with
// fmt.Errorf("...: %w", ErrXxx) and pass the result to WriteError.
var (
	// ErrValidation corresponds to MethodArgumentNotValidException / BindException (400).
	ErrValidation = errors.New("httpapi: request validation failed")
	// ErrMalformedBody corresponds to HttpMessageNotReadableException (400).
	ErrMalformedBody = errors.New("httpapi: malformed request body")
	// ErrInvalidPathParam corresponds to TypeMismatchException on a path variable (400).
	ErrInvalidPathParam = errors.New("httpapi: invalid path parameter")
	// ErrMissingParam corresponds to MissingServletRequestParameterException /
	// MissingPathVariableException as a client error (400).
	ErrMissingParam = errors.New("httpapi: missing request parameter")
	// ErrUnsupportedMediaType corresponds to HttpMediaTypeNotSupportedException (415).
	ErrUnsupportedMediaType = errors.New("httpapi: unsupported media type")
	// ErrNotAcceptable corresponds to HttpMediaTypeNotAcceptableException (406).
	ErrNotAcceptable = errors.New("httpapi: not acceptable")
)

// bootTimestampLayout is the timestamp layout Spring Boot's DefaultErrorAttributes
// produces through Jackson (rendered in UTC).
const bootTimestampLayout = "2006-01-02T15:04:05.000-07:00"

// errorMessageBody is the wire form of model.ErrorMessage. Message is a pointer
// so that an exception without a detail message serializes as null, exactly
// like Jackson rendering a null String field.
type errorMessageBody struct {
	Status  string  `json:"status"`
	Message *string `json:"message"`
}

// bootErrorBody mirrors Spring Boot's default error JSON (no message/trace,
// matching server.error.include-message=never defaults).
type bootErrorBody struct {
	Timestamp string `json:"timestamp"`
	Status    int    `json:"status"`
	Error     string `json:"error"`
	Path      string `json:"path"`
}

// WriteError maps err to the HTTP response the Spring application would have
// produced and writes it to w. It is the single place every handler routes its
// errors through.
//
//   - service.ErrUserNotFound (or *service.UserNotFoundError): 404 with an
//     ErrorMessage JSON body {"status":"NOT_FOUND","message":...}.
//   - Framework-equivalent errors: conventional status with an empty body
//     (Spring 5 ResponseEntityExceptionHandler behaviour).
//   - Anything else: 500 with Boot error JSON.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}
	if detail, ok := service.UserNotFoundDetail(err); ok {
		writeUserNotFound(w, detail)
		return
	}
	switch {
	case errors.Is(err, ErrValidation),
		errors.Is(err, ErrMalformedBody),
		errors.Is(err, ErrInvalidPathParam),
		errors.Is(err, ErrMissingParam):
		w.WriteHeader(http.StatusBadRequest)
	case errors.Is(err, ErrUnsupportedMediaType):
		w.WriteHeader(http.StatusUnsupportedMediaType)
	case errors.Is(err, ErrNotAcceptable):
		w.WriteHeader(http.StatusNotAcceptable)
	default:
		log.Error().Err(err).Str("method", r.Method).Str("path", r.URL.Path).Msg("unhandled request error")
		WriteBootError(w, r, http.StatusInternalServerError)
	}
}

// writeUserNotFound renders the body of the source's userNotFoundException
// handler: ErrorMessage(HttpStatus.NOT_FOUND, exception.getMessage()).
func writeUserNotFound(w http.ResponseWriter, detail string) {
	em := model.NewErrorMessage(http.StatusNotFound, detail)
	body := errorMessageBody{Status: em.GetStatus()}
	if msg := em.GetMessage(); msg != "" {
		body.Message = &msg
	}
	writeJSON(w, http.StatusNotFound, body)
}

// WriteBootError writes Spring Boot's default error JSON for the given status.
func WriteBootError(w http.ResponseWriter, r *http.Request, status int) {
	writeJSON(w, status, bootErrorBody{
		Timestamp: time.Now().UTC().Format(bootTimestampLayout),
		Status:    status,
		Error:     http.StatusText(status),
		Path:      r.URL.Path,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	payload, err := json.Marshal(v)
	if err != nil {
		log.Error().Err(err).Msg("encode error response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(payload); err != nil {
		log.Debug().Err(err).Msg("write error response")
	}
}

// statusWriter records whether the response header has been sent so the
// recover middleware knows if it can still emit a 500.
type statusWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (s *statusWriter) WriteHeader(code int) {
	s.wroteHeader = true
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusWriter) Write(b []byte) (int, error) {
	s.wroteHeader = true
	return s.ResponseWriter.Write(b)
}

// Unwrap exposes the underlying writer to http.ResponseController.
func (s *statusWriter) Unwrap() http.ResponseWriter { return s.ResponseWriter }

// Recover converts handler panics into a 500 Boot error response instead of a
// dropped connection, matching Spring's handling of uncaught exceptions.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w}
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			if rec == http.ErrAbortHandler {
				panic(rec)
			}
			log.Error().Str("panic", fmt.Sprint(rec)).Str("method", r.Method).Str("path", r.URL.Path).Msg("handler panic")
			if !sw.wroteHeader {
				WriteBootError(sw, r, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(sw, r)
	})
}

// captureWriter swallows the ServeMux's plain-text fallback responses so they
// can be re-rendered the Spring way.
type captureWriter struct {
	header http.Header
	status int
}

func (c *captureWriter) Header() http.Header { return c.header }

func (c *captureWriter) Write(b []byte) (int, error) {
	if c.status == 0 {
		c.status = http.StatusOK
	}
	return len(b), nil
}

func (c *captureWriter) WriteHeader(code int) {
	if c.status == 0 {
		c.status = code
	}
}

// SpringFallbacks reshapes the mux's own responses for unmatched requests:
// unknown paths become a 404 Boot error JSON, wrong methods become an empty
// 405 with an Allow header (Spring 5 HttpRequestMethodNotSupportedException).
func SpringFallbacks(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, pattern := mux.Handler(r)
		if pattern != "" {
			mux.ServeHTTP(w, r)
			return
		}
		cw := &captureWriter{header: make(http.Header)}
		h.ServeHTTP(cw, r)
		switch cw.status {
		case http.StatusMethodNotAllowed:
			if allow := cw.header.Values("Allow"); len(allow) > 0 {
				w.Header()["Allow"] = allow
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
		case 0, http.StatusNotFound:
			WriteBootError(w, r, http.StatusNotFound)
		default:
			// Redirects (path cleaning) and anything else: pass status + Location through.
			if loc := cw.header.Get("Location"); loc != "" {
				w.Header().Set("Location", loc)
			}
			w.WriteHeader(cw.status)
		}
	})
}

// StripTrailingSlash removes a single trailing slash from the request path so
// "/users/" routes like "/users", mirroring Spring MVC 5's default
// trailing-slash matching.
func StripTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if len(p) > 1 && strings.HasSuffix(p, "/") {
			r2 := r.Clone(r.Context())
			r2.URL.Path = strings.TrimSuffix(p, "/")
			r2.URL.RawPath = ""
			r = r2
		}
		next.ServeHTTP(w, r)
	})
}

// WrapMux applies the full global error-handling chain to mux; it is the Go
// counterpart of registering the @ControllerAdvice for all controllers.
func WrapMux(mux *http.ServeMux) http.Handler {
	return StripTrailingSlash(Recover(SpringFallbacks(mux)))
}
