// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
)

// stubProvider stands in for a provider whose manager has exited, for example
// because the leader lease was lost.
type stubProvider struct {
	startErr error
}

func (s *stubProvider) Start(_ context.Context) error { return s.startErr }

func (s *stubProvider) Type() egv1a1.ProviderType { return egv1a1.ProviderTypeCustom }

func newTestRunner(t *testing.T, runnerErrors *message.RunnerErrors) *Runner {
	t.Helper()
	return New(&Config{
		Server: config.Server{
			Logger: logging.DefaultLogger(t.Output(), egv1a1.LogLevelInfo),
		},
		RunnerErrors: runnerErrors,
	})
}

func TestStartProviderReportsFailure(t *testing.T) {
	runnerErrors := new(message.RunnerErrors)
	r := newTestRunner(t, runnerErrors)
	notifier := message.RunnerErrorNotifier{RunnerName: r.Name(), RunnerErrors: runnerErrors}

	wantErr := errors.New("leader election lost")
	r.startProvider(t.Context(), &stubProvider{startErr: wantErr}, notifier)

	got, ok := runnerErrors.Load(r.Name())
	require.True(t, ok, "a provider failure must reach RunnerErrors, otherwise the process stays up with no provider and no health endpoint")
	assert.ErrorIs(t, got, wantErr)
}

func TestStartProviderCleanExitReportsNothing(t *testing.T) {
	runnerErrors := new(message.RunnerErrors)
	r := newTestRunner(t, runnerErrors)
	notifier := message.RunnerErrorNotifier{RunnerName: r.Name(), RunnerErrors: runnerErrors}

	r.startProvider(t.Context(), &stubProvider{startErr: nil}, notifier)

	_, ok := runnerErrors.Load(r.Name())
	assert.False(t, ok, "a clean shutdown must not be reported as a runner error")
}
