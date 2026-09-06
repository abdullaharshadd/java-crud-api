package model

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errorMessageTests = []struct {
	name     string
	status   int
	message  string
	expected *ErrorMessage
}{
	{"Valid Error Message", http.StatusBadRequest, "Invalid request", &ErrorMessage{Status: http.StatusBadRequest, Message: "Invalid request"}},
	{"Empty Message", http.StatusInternalServerError, "", &ErrorMessage{Status: http.StatusInternalServerError, Message: ""}},
}

func TestNewErrorMessage(t *testing.T) {
	for _, tt := range errorMessageTests {
		t.Run(tt.name, func(t *testing.T) {
			actual := NewErrorMessage(tt.status, tt.message)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestGetStatus(t *testing.T) {
	for _, tt := range errorMessageTests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := NewErrorMessage(tt.status, tt.message)
			actual := errMsg.GetStatus()
			assert.Equal(t, tt.expected.Status, actual)
		})
	}
}

func TestGetMessage(t *testing.T) {
	for _, tt := range errorMessageTests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := NewErrorMessage(tt.status, tt.message)
			actual := errMsg.GetMessage()
			assert.Equal(t, tt.expected.Message, actual)
		})
	}
}

func TestErrorMessage_GlobalInvariants(t *testing.T) {
	for _, tt := range errorMessageTests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := NewErrorMessage(tt.status, tt.message)
			assert.NotZero(t, errMsg.GetStatus(), "HTTP status should not be zero")
			assert.NotNil(t, errMsg.GetMessage(), "Error message should not be nil")
		})
	}
}