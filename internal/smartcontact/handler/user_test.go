package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"migrated-app/internal/apperr"
	"migrated-app/internal/model"
)

// ---------------------------------------------------------------------------
// Stub service
// ---------------------------------------------------------------------------

type stubUserService struct {
	saveErr      error
