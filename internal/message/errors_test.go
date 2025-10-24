// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunnerErrors(t *testing.T) {
	t.Run("notify nil", func(t *testing.T) {
		errChan := make(chan error)
		close(errChan)
		notifier := RunnerErrorNotifier(t.Name(), errChan)
		require.NotPanics(t, func() { notifier.Notify(nil) }) // would panic if notifying
	})

	t.Run("notify error", func(t *testing.T) {
		errChan := make(chan error)
		t.Cleanup(func() { close(errChan) })

		notifier := RunnerErrorNotifier(t.Name(), errChan)
		go func() { notifier.Notify(errors.New("test error")) }()

		err := <-errChan
		require.Error(t, err)

		var re RunnerError
		require.ErrorAs(t, err, &re)
		require.Equal(t, t.Name(), re.Runner())
		require.Contains(t, re.Error(), "test error")
	})
}
