package model

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errorMessageTestCases = []struct {
	name          string
	status        int
	message       string
	expectedError bool
}{
	{"no params", 0, "", false},
	{"with params", http.StatusBadRequest, "Bad request", false},
	{"invalid status", -1, "Invalid status", true},
}

func TestNewErrorMessage(t *testing.T) {
	for _, tc := range errorMessageTestCases {
		t.Run(tc.name, func(t *testing.T) {
			errMsg := NewErrorMessage(tc.status, tc.message)
			if tc.expectedError {
				assert.Nil(t, errMsg)
			} else {
				assert.NotNil(t, errMsg)
				assert.Equal(t, tc.status, errMsg.Status)
				assert.Equal(t, tc.message, errMsg.Message)
			}
		})
	}
}

func TestFromHTTPError(t *testing.T) {
	for _, tc := range errorMessageTestCases {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			if tc.message != "" {
				err = assert.AnError
			}
			errMsg := FromHTTPError(tc.status, err)
			if tc.expectedError || err == nil {
				assert.Nil(t, errMsg)
			} else {
				assert.NotNil(t, errMsg)
				assert.Equal(t, tc.status, errMsg.Status)
				assert.Equal(t, tc.message, errMsg.Message)
			}
		})
	}
}

func TestToHTTPError(t *testing.T) {
	testErr := assert.AnError
	testErrMsg := NewErrorMessage(http.StatusInternalServerError, testErr.Error())
	httpErr := testErrMsg.ToHTTPError()
	assert.Equal(t, testErr.Error(), httpErr.Error())
}

func TestHttpError_Error(t *testing.T) {
	testErrMsg := NewErrorMessage(http.StatusInternalServerError, "Server error")
	httpErr := &httpError{errorMessage: testErrMsg}
	assert.Equal(t, testErrMsg.Message, httpErr.Error())
}