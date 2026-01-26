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

// TestRunnerGoroutineRace specifically reproduces the data race from CI by
// simulating the exact condition: goroutines logging while test cleanup closes t.Output()
//
// The race happens because runner.Close() is a no-op - it returns immediately
// without waiting for goroutines to finish. When the test ends, cleanup closes
// t.Output() while goroutines are still active.
//
// Run with: go test -race -run TestRunnerGoroutineRace -count=100 ./internal/cmd/
func TestRunnerGoroutineRace(t *testing.T) {
	// Skip if not running with race detector
	// This test is specifically designed to catch the race
	if !testing.Short() {
		t.Skip("Run with -race flag to detect the race")
	}

	configHome := t.TempDir()
	cfgFileContent := strings.ReplaceAll(fileProviderGatewayConfig, "[CONFIG_HOME_PLACE_HODLER]", configHome)
	configPath := path.Join(t.TempDir(), "envoy-gateway.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(cfgFileContent), 0o600))

	require.NoError(t, certGen(t.Context(), t.Output(), true, configHome))

	// Use a context WITHOUT defer cancel to keep goroutines alive longer
	ctx, cancel := context.WithCancel(context.Background())

	hook := func(c context.Context, cfg *config.Server) error {
		cfg.Logger = logging.DefaultLogger(t.Output(), egv1a1.LogLevelInfo)
		return startRunners(c, cfg, nil)
	}

	errCh := make(chan error, 1)

	go func() {
		errCh <- server(ctx, t.Output(), t.Output(), configPath, hook, nil)
	}()

	// Let runners start and become active
	time.Sleep(80 * time.Millisecond)

	// Cancel context - triggers shutdown but goroutines may still be logging
	cancel()

	// Don't wait for server to complete - this creates the race!
	// Test will end, cleanup will close t.Output(), while goroutines
	// are still running and trying to log
	select {
	case <-errCh:
		// Server finished
	case <-time.After(20 * time.Millisecond):
		// Timeout - goroutines likely still running
		// Test ends here, cleanup starts -> RACE!
	}

	// Test ends immediately - race window is NOW
	// Without fix: goroutines still running, trying to log to closed t.Output()
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

			require.NoError(t, certGen(t.Context(), t.Output(), true, configHome))

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			hook := func(c context.Context, cfg *config.Server) error {
				cfg.Logger = logging.DefaultLogger(t.Output(), egv1a1.LogLevelInfo)
				return startRunners(c, cfg, nil)
			}

			errCh := make(chan error, 1)

			go func() {
				errCh <- server(ctx, t.Output(), t.Output(), configPath, hook, nil)
			}()

			// Vary timing to hit different race windows
			sleepTime := 50 + (i * 30)
			time.Sleep(time.Duration(sleepTime) * time.Millisecond)

			cancel()

			err := <-errCh
			require.NoError(t, err)

			// Race window: goroutines may still be logging
		})
	}
}
