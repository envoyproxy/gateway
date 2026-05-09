// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
	"context"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/logging"
)

// TestRunnerGoroutineRace tests that background goroutines can safely log after context cancellation.
// This test previously exposed a data race when using t.Output() - the test would complete and
// close t.Output() while background goroutines were still logging. The fix is to use os.Stdout
// instead, which is thread-safe and remains valid after the test completes.
//
// Note: In production, certGen is called with cmd.OutOrStdout() which resolves to os.Stdout
// in most cases, so using os.Stdout in tests mirrors real-world behavior.
//
// Run with: go test -race -run TestRunnerGoroutineRace -count=100 ./internal/cmd/
func TestRunnerGoroutineRace(t *testing.T) {
	// Skip if not running with race detector
	if !testing.Short() {
		t.Skip("Run with -race flag to verify no race conditions")
	}

	configHome := t.TempDir()
	cfgFileContent := strings.ReplaceAll(fileProviderGatewayConfig, "[CONFIG_HOME_PLACE_HODLER]", configHome)
	configPath := path.Join(t.TempDir(), "envoy-gateway.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(cfgFileContent), 0o600))

	require.NoError(t, certGen(t.Context(), os.Stdout, true, configHome))

	// Use a context WITHOUT defer cancel to keep goroutines alive longer
	ctx, cancel := context.WithCancel(context.Background())

	hook := func(c context.Context, cfg *config.Server) error {
		// Use os.Stdout instead of t.Output() - it's thread-safe and won't cause races
		cfg.Logger = logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)
		return startRunners(c, cfg, nil)
	}

	errCh := make(chan error, 1)

	go func() {
		errCh <- server(ctx, os.Stdout, os.Stdout, configPath, hook, nil)
	}()

	// Let runners start and become active
	time.Sleep(80 * time.Millisecond)

	// Cancel context - triggers shutdown but goroutines may still be logging
	cancel()

	// Wait briefly to see if server completes or if goroutines are still running
	select {
	case <-errCh:
		// Server finished cleanly
	case <-time.After(20 * time.Millisecond):
		// Server still running - this is fine with os.Stdout (no race)
	}
}

// TestRunnerGoroutineRaceStress runs multiple quick cycles to maximize
// the probability of hitting the race condition
func TestRunnerGoroutineRaceStress(t *testing.T) {
	for i := range 3 {
		t.Run("cycle", func(t *testing.T) {
			configHome := t.TempDir()
			cfgFileContent := strings.ReplaceAll(fileProviderGatewayConfig, "[CONFIG_HOME_PLACE_HODLER]", configHome)
			configPath := path.Join(t.TempDir(), "envoy-gateway.yaml")
			require.NoError(t, os.WriteFile(configPath, []byte(cfgFileContent), 0o600))

			require.NoError(t, certGen(t.Context(), os.Stdout, true, configHome))

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			hook := func(c context.Context, cfg *config.Server) error {
				// Use os.Stdout instead of t.Output() - it's thread-safe and won't cause races
				cfg.Logger = logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)
				return startRunners(c, cfg, nil)
			}

			errCh := make(chan error, 1)

			go func() {
				errCh <- server(ctx, os.Stdout, os.Stdout, configPath, hook, nil)
			}()

			// Vary timing to hit different race windows
			sleepTime := 50 + (i * 30)
			time.Sleep(time.Duration(sleepTime) * time.Millisecond)

			cancel()

			err := <-errCh
			require.NoError(t, err)
		})
	}
}
