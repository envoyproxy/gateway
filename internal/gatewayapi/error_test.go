// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"
	"testing"
)

func TestFailOpen(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "regular error",
			err:      errors.New("regular error"),
			expected: false,
		},
		{
			name: "PolicyTranslationError with FailOpen true",
			err: &PolicyTranslationError{
				Wrapped:  errors.New("policy error"),
				FailOpen: true,
			},
			expected: true,
		},
		{
			name: "PolicyTranslationError with FailOpen false",
			err: &PolicyTranslationError{
				Wrapped:  errors.New("policy error"),
				FailOpen: false,
			},
			expected: false,
		},
		{
			name:     "wrapped PolicyTranslationError with FailOpen true",
			err:      fmt.Errorf("wrapped: %w", &PolicyTranslationError{Wrapped: errors.New("policy error"), FailOpen: true}),
			expected: true,
		},
		{
			name:     "wrapped PolicyTranslationError with FailOpen false",
			err:      fmt.Errorf("wrapped: %w", &PolicyTranslationError{Wrapped: errors.New("policy error"), FailOpen: false}),
			expected: false,
		},
		{
			name:     "wrapped regular error",
			err:      fmt.Errorf("wrapped: %w", errors.New("regular error")),
			expected: false,
		},
		{
			name: "multiple errors all PolicyTranslationError with FailOpen true",
			err: &multiError{
				errors: []error{
					&PolicyTranslationError{Wrapped: errors.New("error1"), FailOpen: true},
					&PolicyTranslationError{Wrapped: errors.New("error2"), FailOpen: true},
				},
			},
			expected: true,
		},
		{
			name: "multiple errors with one PolicyTranslationError FailOpen false",
			err: &multiError{
				errors: []error{
					&PolicyTranslationError{Wrapped: errors.New("error1"), FailOpen: true},
					&PolicyTranslationError{Wrapped: errors.New("error2"), FailOpen: false},
				},
			},
			expected: false,
		},
		{
			name: "multiple errors with one regular error",
			err: &multiError{
				errors: []error{
					&PolicyTranslationError{Wrapped: errors.New("error1"), FailOpen: true},
					errors.New("regular error"),
				},
			},
			expected: false,
		},
		{
			name: "multiple errors all regular errors",
			err: &multiError{
				errors: []error{
					errors.New("error1"),
					errors.New("error2"),
				},
			},
			expected: false,
		},
		{
			name: "empty multiple errors",
			err: &multiError{
				errors: []error{},
			},
			expected: false,
		},
		{
			name: "single PolicyTranslationError in slice with FailOpen true",
			err: &multiError{
				errors: []error{
					&PolicyTranslationError{Wrapped: errors.New("error1"), FailOpen: true},
				},
			},
			expected: true,
		},
		{
			name: "single PolicyTranslationError in slice with FailOpen false",
			err: &multiError{
				errors: []error{
					&PolicyTranslationError{Wrapped: errors.New("error1"), FailOpen: false},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldFailOpen(tt.err)
			if result != tt.expected {
				t.Errorf("FailOpen(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestPolicyTranslationError_Error(t *testing.T) {
	originalErr := errors.New("original error message")
	policyErr := &PolicyTranslationError{
		Wrapped:  originalErr,
		FailOpen: true,
	}

	if policyErr.Error() != originalErr.Error() {
		t.Errorf("PolicyTranslationError.Error() = %q, expected %q", policyErr.Error(), originalErr.Error())
	}
}

// multiError is a test helper that implements the interface{ Unwrap() []error } interface
type multiError struct {
	errors []error
}

func (m *multiError) Error() string {
	return "multiple errors"
}

func (m *multiError) Unwrap() []error {
	return m.errors
}
