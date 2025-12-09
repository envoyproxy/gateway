// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import "github.com/telepresenceio/watchable"

// RunnerErrorsMessageName is the name of the RunnerErrors message
const RunnerErrorsMessageName MessageName = "runner-errors"

// WatchableError is an error that can captured in a watchable.Map
type WatchableError interface {
	error
	watchable.DeepCopier[WatchableError]
}

// RunnerErrors is a map of runner name to runnerError
type RunnerErrors = watchable.Map[string, WatchableError]

var _ WatchableError = (*watchableError)(nil)

// watchableError is an error that can captured in a watchable.Map
type watchableError struct {
	err error
}

// NewWatchableError creates a new WatchableError
func NewWatchableError(err error) WatchableError {
	return &watchableError{err: err}
}

// Unwrap returns the underlying error.
func (r *watchableError) Unwrap() error { return r.err }

// Error returns the error message.
func (r *watchableError) Error() string { return r.err.Error() }

// DeepCopy creates a copy of the watchableError.
// // The err field is preserved (not deep copied) since errors are meant to be bubbled up.
func (r *watchableError) DeepCopy() WatchableError { return &watchableError{err: r.err} }

// RunnerErrorNotifier is a helper to notify errors with a specific runner name.
type RunnerErrorNotifier struct {
	RunnerName   string
	RunnerErrors *RunnerErrors
}

// Store the error.
func (n *RunnerErrorNotifier) Store(err error) {
	n.RunnerErrors.Store(n.RunnerName, NewWatchableError(err))
}
