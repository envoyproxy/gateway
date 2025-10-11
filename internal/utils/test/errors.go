// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package test

import "testing"

// RunnerErrorsChan creates a channel to receive errors and starts a goroutine to
// consume the channel. It also registers a cleanup function to close the channel when
// the test ends.
func RunnerErrorsChan(t *testing.T, callbacks ...func(*testing.T, error)) chan<- error {
	t.Helper()

	errors := make(chan error)
	t.Cleanup(func() { close(errors) })

	go func() {
		for err := range errors {
			for _, callback := range callbacks {
				callback(t, err)
			}
		}
	}()

	return errors
}
