package smartcontact

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Assuming UserNotFoundException is defined as follows:
type UserNotFoundException struct {
	Message       string
	Cause         error
	EnableSuppression bool
	WritableStackTrace bool
}

func NewUserNotFoundException(message string, cause error, enableSuppression bool, writableStackTrace bool) *UserNotFoundException {
	return &UserNotFoundException{
		Message:           message,
		Cause:             cause,
		EnableSuppression: enableSuppression,
		WritableStackTrace: writableStackTrace,
	}
}

func TestUserNotFoundException(t *testing.T) {
	type testCase struct {
		name            string
		message         string
		cause           error
		enableSuppression bool
		writableStackTrace bool
		expectedMessage string
		expectedCause   error
	}

	testCases := []testCase{
		{
			name:            "No parameters provided",
			expectedMessage: "",
			expectedCause:   nil,
		},
		{
			name:            "A string message parameter is provided",
			message:         "User not found",
			expectedMessage: "User not found",
			expectedCause:   nil,
		},
		{
			name:            "A string message and a Throwable cause parameter are provided",
			message:         "User not found",
			cause:           errors.New("database error"),
			expectedMessage: "User not found",
			expectedCause:   errors.New("database error"),
		},
		{
			name:            "Only a Throwable cause parameter is provided",
			cause:           errors.New("database error"),
			expectedMessage: "",
			expectedCause:   errors.New("database error"),
		},
		{
			name:            "All parameters (message, cause, enableSuppression, writableStackTrace) are provided",
			message:         "User not found",
			cause:           errors.New("database error"),
			enableSuppression: true,
			writableStackTrace: true,
			expectedMessage: "User not found",
			expectedCause:   errors.New("database error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := NewUserNotFoundException(tc.message, tc.cause, tc.enableSuppression, tc.writableStackTrace)
			assert.Equal(t, tc.expectedMessage, err.Message)
			assert.Equal(t, tc.expectedCause, err.Cause)
		})
	}
}