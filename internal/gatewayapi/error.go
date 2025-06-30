// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import "errors"

// PolicyTranslationError represents an error that occurred during policy translation.
// It wraps the original error and includes a FailOpen flag to control error handling behavior.
type PolicyTranslationError struct {
	Wrapped error
	// FailOpen indicates whether the policy should fail open or not.
	// If true, the targeted routes will skip the policy application and requests will be processed as if the policy was not present.
	// If false, the targeted routes will return a 500 error and requests will not be forwarded to the backend.
	FailOpen bool
}

// NewPolicyTranslationError creates a new PolicyTranslationError with the given wrapped error and failOpen flag.
func NewPolicyTranslationError(err error, failOpen bool) *PolicyTranslationError {
	if err == nil {
		return nil
	}
	return &PolicyTranslationError{
		Wrapped:  err,
		FailOpen: failOpen,
	}
}

// Error returns the error message from the wrapped error.
func (p *PolicyTranslationError) Error() string {
	if p.Wrapped == nil {
		return "policy translation error"
	}
	return p.Wrapped.Error()
}

// Unwrap returns the wrapped error for errors.Is and errors.As compatibility.
func (p *PolicyTranslationError) Unwrap() error {
	return p.Wrapped
}

// Is checks if the error is a PolicyTranslationError.
func (p *PolicyTranslationError) Is(target error) bool {
	_, ok := target.(*PolicyTranslationError)
	return ok
}

// ShouldFailOpen checks if the error is a PolicyTranslationError and returns its FailOpen status.
func ShouldFailOpen(err error) bool {
	if err == nil {
		return false
	}

	// Check if error implements interface{ Unwrap() []error }
	if unwrapper, ok := err.(interface{ Unwrap() []error }); ok {
		// FailOpen only if all errors in the slice are PolicyTranslationError with FailOpen set to true
		wrappedErrors := unwrapper.Unwrap()
		if len(wrappedErrors) == 0 {
			return false
		}
		for _, err := range wrappedErrors {
			if !shouldFailOpenForSingleError(err) {
				return false
			}
		}
		return true
	}

	return shouldFailOpenForSingleError(err)
}

func shouldFailOpenForSingleError(err error) bool {
	if err == nil {
		return false
	}

	var policyErr *PolicyTranslationError
	if errors.As(err, &policyErr) {
		return policyErr.FailOpen
	}

	// Check if error can be unwrapped and recurse
	if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
		if wrapped := unwrapper.Unwrap(); wrapped != nil {
			return shouldFailOpenForSingleError(wrapped)
		}
	}

	return false
}
