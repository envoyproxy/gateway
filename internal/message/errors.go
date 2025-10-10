// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import "fmt"

// RunnerError is an interface that represents an error that occurred during the
// execution of a runner.
type RunnerError interface {
	error
	// Runner returns the name of the runner that produced the error.
	Runner() string
}

// RunnerError represents an error that occurred during the execution of a runner.
type runnerError struct {
	runner string
	err    error
}

// Runner returns the name of the runner that produced the error.
func (r *runnerError) Runner() string { return r.runner }

// Unwrap returns the underlying error.
func (r *runnerError) Unwrap() error { return r.err }

// Error returns the error message.
func (r *runnerError) Error() string {
	return fmt.Sprintf("%s runner error: %v", r.runner, r.err.Error())
}

// ErrorNotifier is an interface that represents a notifier for runner errors.
// It is used to notify the control loop of errors that occur during the execution of a runner
// so that the control loop can handle them appropriately.
type ErrorNotifier interface {
	// Notify sends the error to the control loop if it is not nil.
	Notify(err error)
}

// RunnerErrorNotifier creates a new ErrorNotifier for the given runner and error channel.
func RunnerErrorNotifier(runner string, errChan chan<- error) ErrorNotifier {
	return &runnerErrorNotifier{runner: runner, errChan: errChan}
}

// runnerErrorNotifier is a helper to send runner errors to a channel where the
// control loop can pick them up and handle them appropriately.
type runnerErrorNotifier struct {
	runner  string
	errChan chan<- error
}

// Notify sends the error to the error channel if it is not nil.
func (r *runnerErrorNotifier) Notify(err error) {
	if err != nil {
		r.errChan <- &runnerError{r.runner, err}
	}
}
