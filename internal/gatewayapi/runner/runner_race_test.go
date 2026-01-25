// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/telepresenceio/watchable"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
)

// TestRunnerDataRace reproduces the exact data race from the stack trace:
// - Goroutine from subscribeAndTranslate() tries to log (runner.go:401)
// - Main test goroutine completes and cleanup closes t.Output()
// - Race detector catches write to closed test output
//
// This test intentionally creates the race condition window by:
// 1. Starting the runner with logging to t.Output()
// 2. Triggering activity that causes logging
// 3. Calling Close() which returns immediately (without fix)
// 4. Test ends, cleanup starts
// 5. Goroutines still running try to log -> RACE
func TestRunnerDataRace(t *testing.T) {
	// Create minimal config for runner
	serverCfg := &config.Server{
		EnvoyGateway: &egv1a1.EnvoyGateway{
			EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
				Gateway: &egv1a1.Gateway{
					ControllerName: "test-controller",
				},
				Provider: &egv1a1.EnvoyGatewayProvider{
					Type: egv1a1.ProviderTypeCustom,
					Custom: &egv1a1.EnvoyGatewayCustomProvider{
						Infrastructure: &egv1a1.EnvoyGatewayInfrastructureProvider{
							Type: egv1a1.InfrastructureProviderTypeHost,
							Host: &egv1a1.EnvoyGatewayHostInfrastructureProvider{},
						},
					},
				},
			},
		},
		// This is the critical part - logger writes to t.Output()
		Logger: logging.DefaultLogger(t.Output(), egv1a1.LogLevelInfo),
	}

	// Create provider resources with watchable store
	providerResources := new(message.ProviderResources)
	providerResources.GatewayAPIResources = watchable.Map[string, *resource.ControllerResourcesContext]{}

	cfg := &Config{
		Server:            *serverCfg,
		ProviderResources: providerResources,
		XdsIR:             new(message.XdsIR),
		InfraIR:           new(message.InfraIR),
		RunnerErrors:      new(message.RunnerErrors),
	}

	runner := New(cfg)

	// Create context for the runner
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the runner - this spawns goroutines
	err := runner.Start(ctx)
	require.NoError(t, err)

	// Give goroutines time to start and get into their message loops
	time.Sleep(50 * time.Millisecond)

	// Trigger an update to cause logging in subscribeAndTranslate
	// This will cause the goroutine to call r.Logger.Info("received an update", ...)
	providerResources.GatewayAPIResources.Store("test-controller", &resource.ControllerResourcesContext{
		Context: ctx,
	})

	// Small delay to let the goroutine start processing the update
	time.Sleep(30 * time.Millisecond)

	// Cancel context to signal shutdown
	cancel()

	// Call Close() - WITHOUT THE FIX, this returns immediately
	// leaving goroutines still running
	err = runner.Close()
	require.NoError(t, err)

	// WITHOUT THE FIX: goroutines are still running here
	// Test function is about to return and cleanup will close t.Output()
	// Meanwhile, goroutines are trying to log -> RACE DETECTED
	//
	// WITH THE FIX: Close() waits for goroutines via WaitGroup,
	// so they're already stopped when we get here
}

// TestRunnerDataRaceAggressive runs multiple iterations with varying timing
// to maximize the chance of hitting the race condition
func TestRunnerDataRaceAggressive(t *testing.T) {
	for i := 0; i < 5; i++ {
		t.Run("iteration", func(t *testing.T) {
			serverCfg := &config.Server{
				EnvoyGateway: &egv1a1.EnvoyGateway{
					EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
						Gateway: &egv1a1.Gateway{
							ControllerName: "test-controller",
						},
						Provider: &egv1a1.EnvoyGatewayProvider{
							Type: egv1a1.ProviderTypeCustom,
							Custom: &egv1a1.EnvoyGatewayCustomProvider{
								Infrastructure: &egv1a1.EnvoyGatewayInfrastructureProvider{
									Type: egv1a1.InfrastructureProviderTypeHost,
									Host: &egv1a1.EnvoyGatewayHostInfrastructureProvider{},
								},
							},
						},
					},
				},
				Logger: logging.DefaultLogger(t.Output(), egv1a1.LogLevelInfo),
			}

			providerResources := new(message.ProviderResources)
			providerResources.GatewayAPIResources = watchable.Map[string, *resource.ControllerResourcesContext]{}
			cfg := &Config{
				Server:            *serverCfg,
				ProviderResources: providerResources,
				XdsIR:             new(message.XdsIR),
				InfraIR:           new(message.InfraIR),
				RunnerErrors:      new(message.RunnerErrors),
			}

			runner := New(cfg)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := runner.Start(ctx)
			require.NoError(t, err)

			// Continuously trigger updates in background to keep goroutines busy
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 3; j++ {
					select {
					case <-ctx.Done():
						return
					default:
						providerResources.GatewayAPIResources.Store("test-controller", &resource.ControllerResourcesContext{
							Context: ctx,
						})
						time.Sleep(10 * time.Millisecond)
					}
				}
			}()

			// Varying timing to hit different race windows
			time.Sleep(time.Duration(20+i*10) * time.Millisecond)

			cancel()
			wg.Wait() // Wait for our update goroutine

			// Close runner - race happens here if no WaitGroup
			err = runner.Close()
			require.NoError(t, err)

			// Race window: goroutines still logging after Close() returns
		})
	}
}

// TestRunnerDataRaceImmediate tests the most extreme case:
// Start and immediately close, goroutines barely started
func TestRunnerDataRaceImmediate(t *testing.T) {
	serverCfg := &config.Server{
		EnvoyGateway: &egv1a1.EnvoyGateway{
			EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
				Gateway: &egv1a1.Gateway{
					ControllerName: "test-controller",
				},
				Provider: &egv1a1.EnvoyGatewayProvider{
					Type: egv1a1.ProviderTypeCustom,
					Custom: &egv1a1.EnvoyGatewayCustomProvider{
						Infrastructure: &egv1a1.EnvoyGatewayInfrastructureProvider{
							Type: egv1a1.InfrastructureProviderTypeHost,
							Host: &egv1a1.EnvoyGatewayHostInfrastructureProvider{},
						},
					},
				},
			},
		},
		Logger: logging.DefaultLogger(t.Output(), egv1a1.LogLevelInfo),
	}

	providerResources := new(message.ProviderResources)
	providerResources.GatewayAPIResources = watchable.Map[string, *resource.ControllerResourcesContext]{}

	cfg := &Config{
		Server:            *serverCfg,
		ProviderResources: providerResources,
		XdsIR:             new(message.XdsIR),
		InfraIR:           new(message.InfraIR),
		RunnerErrors:      new(message.RunnerErrors),
	}

	runner := New(cfg)
	ctx, cancel := context.WithCancel(context.Background())

	_ = runner.Start(ctx)

	// Trigger update immediately
	providerResources.GatewayAPIResources.Store("test-controller", &resource.ControllerResourcesContext{
		Context: ctx,
	})

	// Cancel and close almost immediately
	time.Sleep(5 * time.Millisecond)
	cancel()

	// Close - goroutines still starting up
	_ = runner.Close()

	// Race window is maximized here
}
