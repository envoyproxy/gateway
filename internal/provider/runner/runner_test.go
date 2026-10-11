// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/logging"
)

// stubProvider is a provider whose Start returns err as soon as it is called
// if stopsWhileRunning, and once ctx is done otherwise.
type stubProvider struct {
	err               error
	stopsWhileRunning bool
}

func (p *stubProvider) Start(ctx context.Context) error {
	if !p.stopsWhileRunning {
		<-ctx.Done()
	}
	return p.err
}

func (p *stubProvider) Type() egv1a1.ProviderType { return egv1a1.ProviderTypeCustom }

// TestWait checks that Wait returns the error of a provider that stops while
// the runner's context is live, e.g. because it lost its leader lease, and
// otherwise returns nil once the context is done: a provider that stops
// because of the cancellation, with or without an error, is not reported.
func TestWait(t *testing.T) {
	stopErr := errors.New("leader election lost")
	testCases := []struct {
		name     string
		provider *stubProvider
		wantErr  error
	}{
		{name: "error while running", provider: &stubProvider{err: stopErr, stopsWhileRunning: true}, wantErr: stopErr},
		{name: "error after cancel", provider: &stubProvider{err: stopErr}},
		{name: "clean stop after cancel", provider: &stubProvider{}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			r := New(&Config{Server: config.Server{Logger: logging.DefaultLogger(t.Output(), egv1a1.LogLevelInfo)}})

			// As Start does once it has created the provider.
			r.wg.Add(1)
			go r.run(ctx, tc.provider)

			if tc.wantErr == nil {
				cancel()
			}
			require.Equal(t, tc.wantErr, r.Wait(ctx))
			cancel()
			require.NoError(t, r.Close())
			require.Empty(t, r.stopped, "the provider's stop was reported although the context was done")
		})
	}
}
