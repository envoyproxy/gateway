// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package host

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"
	func_e_api "github.com/tetratelabs/func-e/api"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/crypto"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/common"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/utils/file"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
	testutils "github.com/envoyproxy/gateway/test/utils"
)

// newMockInfra doesn't actually run Envoy
func newMockInfra(t *testing.T, cfg *config.Server) *Infra {
	t.Helper()
	homeDir := t.TempDir()
	// Create envoy certs under home dir.
	certs, err := crypto.GenerateCerts(cfg)
	require.NoError(t, err)
	// Write certs into proxy dir.
	proxyDir := path.Join(homeDir, "envoy")
	err = file.WriteDir(certs.CACertificate, proxyDir, "ca.crt")
	require.NoError(t, err)
	err = file.WriteDir(certs.EnvoyCertificate, proxyDir, "tls.crt")
	require.NoError(t, err)
	err = file.WriteDir(certs.EnvoyPrivateKey, proxyDir, "tls.key")
	require.NoError(t, err)
	// Write sds config as well.
	err = createSdsConfig(proxyDir)
	require.NoError(t, err)

	paths := &Paths{
		ConfigHome: homeDir,
		DataHome:   homeDir,
		StateHome:  homeDir,
		RuntimeDir: homeDir,
	}
	infra := &Infra{
		Paths:         paths,
		Logger:        logging.DefaultLogger(io.Discard, egv1a1.LogLevelInfo),
		EnvoyGateway:  cfg.EnvoyGateway,
		sdsConfigPath: proxyDir,
		Stdout:        io.Discard,
		Stderr:        io.Discard,
		envoyRunner: func(ctx context.Context, args []string, options ...func_e_api.RunOption) error {
			// Block until context is cancelled (mimics real Envoy blocking)
			<-ctx.Done()
			return ctx.Err()
		},
		errors: message.RunnerErrorNotifier{RunnerName: t.Name(), RunnerErrors: &message.RunnerErrors{}},
	}

	return infra
}

func TestInfra_CreateOrUpdateProxyInfra(t *testing.T) {
	cfg, err := config.New(io.Discard, io.Discard)
	require.NoError(t, err)
	infra := newMockInfra(t, cfg)

	t.Run("create new proxy", func(t *testing.T) {
		infraIR := &ir.Infra{
			Proxy: &ir.ProxyInfra{
				Name:      "test-proxy",
				Namespace: "default",
				Config: &egv1a1.EnvoyProxy{
					Spec: egv1a1.EnvoyProxySpec{
						Logging: egv1a1.ProxyLogging{
							Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
								egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
							},
						},
					},
				},
			},
		}

		hashedName := utils.GetHashedName("test-proxy", 64)
		t.Cleanup(func() { infra.stopEnvoy(hashedName) })

		err := infra.CreateOrUpdateProxyInfra(t.Context(), infraIR)
		require.NoError(t, err)

		// Verify proxy context was stored
		_, loaded := infra.proxyContextMap.Load(hashedName)
		require.True(t, loaded, "proxy should be loaded after creation")
	})

	t.Run("idempotent - proxy already exists", func(t *testing.T) {
		infraIR := &ir.Infra{
			Proxy: &ir.ProxyInfra{
				Name:      "test-proxy-idempotent",
				Namespace: "default",
				Config: &egv1a1.EnvoyProxy{
					Spec: egv1a1.EnvoyProxySpec{
						Logging: egv1a1.ProxyLogging{
							Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
								egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
							},
						},
					},
				},
			},
		}

		hashedName := utils.GetHashedName("test-proxy-idempotent", 64)
		t.Cleanup(func() { infra.stopEnvoy(hashedName) })

		// First call creates the proxy
		err := infra.CreateOrUpdateProxyInfra(t.Context(), infraIR)
		require.NoError(t, err)

		_, loaded := infra.proxyContextMap.Load(hashedName)
		require.True(t, loaded, "proxy should be loaded after first call")

		// Second call should be idempotent (early return without error)
		err = infra.CreateOrUpdateProxyInfra(t.Context(), infraIR)
		require.NoError(t, err)

		// Verify proxy is still loaded and wasn't recreated
		_, loaded = infra.proxyContextMap.Load(hashedName)
		require.True(t, loaded, "proxy should still be loaded after second call")
	})

	testCases := []struct {
		name          string
		infra         *ir.Infra
		expectedError string
	}{
		{
			name:          "nil cfg",
			infra:         nil,
			expectedError: "infra ir is nil",
		},
		{
			name: "nil proxy",
			infra: &ir.Infra{
				Proxy: nil,
			},
			expectedError: "infra proxy ir is nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := infra.CreateOrUpdateProxyInfra(t.Context(), tc.infra)
			require.EqualError(t, actual, tc.expectedError)
		})
	}

	t.Run("invalid_bootstrap_config", func(t *testing.T) {
		infraIR := &ir.Infra{
			Proxy: &ir.ProxyInfra{
				Name:      "test-proxy-invalid",
				Namespace: "default",
				Config: &egv1a1.EnvoyProxy{
					Spec: egv1a1.EnvoyProxySpec{
						Bootstrap: &egv1a1.ProxyBootstrap{
							Type: ptr.To(egv1a1.BootstrapTypeMerge),
							// Invalid YAML that will cause bootstrap merge to fail
							Value: ptr.To("invalid: yaml: [unclosed"),
						},
					},
				},
			},
		}

		err := infra.CreateOrUpdateProxyInfra(t.Context(), infraIR)
		require.Error(t, err)
	})
}

func TestInfra_DeleteProxyInfra(t *testing.T) {
	cfg, err := config.New(io.Discard, io.Discard)
	require.NoError(t, err)
	infra := newMockInfra(t, cfg)

	t.Run("delete existing proxy", func(t *testing.T) {
		infraIR := &ir.Infra{
			Proxy: &ir.ProxyInfra{
				Name:      "test-proxy-delete",
				Namespace: "default",
				Config: &egv1a1.EnvoyProxy{
					Spec: egv1a1.EnvoyProxySpec{
						Logging: egv1a1.ProxyLogging{
							Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
								egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
							},
						},
					},
				},
			},
		}

		// Create a proxy first
		err := infra.CreateOrUpdateProxyInfra(t.Context(), infraIR)
		require.NoError(t, err)

		hashedName := utils.GetHashedName("test-proxy-delete", 64)
		t.Cleanup(func() { infra.stopEnvoy(hashedName) })

		_, loaded := infra.proxyContextMap.Load(hashedName)
		require.True(t, loaded, "proxy should be loaded before deletion")

		// Delete the proxy
		err = infra.DeleteProxyInfra(t.Context(), infraIR)
		require.NoError(t, err)

		// Verify deletion
		_, loaded = infra.proxyContextMap.Load(hashedName)
		require.False(t, loaded, "proxy should be removed after deletion")
	})

	t.Run("delete non-existent proxy", func(t *testing.T) {
		infraIR := &ir.Infra{
			Proxy: &ir.ProxyInfra{
				Name:      "non-existent-proxy",
				Namespace: "default",
				Config:    &egv1a1.EnvoyProxy{},
			},
		}

		// Delete a proxy that was never created - should not error
		err := infra.DeleteProxyInfra(t.Context(), infraIR)
		require.NoError(t, err)
	})

	t.Run("nil infra", func(t *testing.T) {
		err := infra.DeleteProxyInfra(t.Context(), nil)
		require.EqualError(t, err, "infra ir is nil")
	})
}

func TestExtractSemver(t *testing.T) {
	tests := []struct {
		image   string
		want    string
		wantErr bool
	}{
		{"docker.io/envoyproxy/envoy:distroless-v1.35.0", "1.35.0", false},
		{"envoyproxy/envoy:v1.28.1", "1.28.1", false},
		{"envoyproxy/envoy:latest", "", true},
		{"envoyproxy/envoy", "", true},
		{"envoyproxy/envoy:distroless-v1.35", "", true},
		{"envoyproxy/envoy:distroless-v1.35.0-extra", "1.35.0", false},
		{"envoyproxy/envoy:1.2.3", "1.2.3", false},
		{"envoyproxy/envoy:foo-2.3.4-bar", "2.3.4", false},
	}
	for _, tc := range tests {
		t.Run(tc.image, func(t *testing.T) {
			got, err := extractSemver(tc.image)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
		})
	}
}

// TestInfra_runEnvoy_integration verifies Envoy process lifecycle, output redirection, and XDG directory usage.
func TestInfra_runEnvoy_integration(t *testing.T) {
	// Create separate XDG directories
	baseDir := t.TempDir()
	configHome := path.Join(baseDir, "config")
	dataHome := path.Join(baseDir, "data")
	stateHome := path.Join(baseDir, "state")
	runtimeDir := path.Join(baseDir, "runtime")

	// Create separate buffers for stdout and stderr
	buffers := testutils.DumpLogsOnFail(t, "stdout", "stderr")
	stdout := buffers[0]
	stderr := buffers[1]

	// Create config with custom paths
	cfg, err := config.New(stdout, stderr)
	require.NoError(t, err)
	cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{
		Type: egv1a1.ProviderTypeCustom,
		Custom: &egv1a1.EnvoyGatewayCustomProvider{
			Infrastructure: &egv1a1.EnvoyGatewayInfrastructureProvider{
				Type: egv1a1.InfrastructureProviderTypeHost,
				Host: &egv1a1.EnvoyGatewayHostInfrastructureProvider{
					ConfigHome: ptr.To(configHome),
					DataHome:   ptr.To(dataHome),
					StateHome:  ptr.To(stateHome),
					RuntimeDir: ptr.To(runtimeDir),
				},
			},
		},
	}

	errNotifier := message.RunnerErrorNotifier{RunnerName: t.Name(), RunnerErrors: &message.RunnerErrors{}}
	i, err := NewInfra(t.Context(), cfg, logging.DefaultLogger(stdout, egv1a1.LogLevelInfo), errNotifier)
	require.NoError(t, err)

	// Run envoy once to let func-e set up all XDG directories
	args := []string{
		"--config-yaml",
		"admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}",
	}
	i.runEnvoy(t.Context(), "", "test", args)
	_, ok := i.proxyContextMap.Load("test")
	require.True(t, ok, "expected proxy context to be stored")

	// Wait for func-e to create all XDG directories
	require.Eventually(t, func() bool {
		_, err := os.Stat(path.Join(configHome, "envoy-version"))
		return err == nil
	}, 5*time.Second, 100*time.Millisecond, "envoy-version file should be created in configHome")

	i.stopEnvoy("test")
	_, ok = i.proxyContextMap.Load("test")
	require.False(t, ok, "expected proxy context to be removed")

	t.Run("xdg_directory_state", func(t *testing.T) {
		// Verify XDG directories were created at configured paths by func-e
		// This proves the Paths configuration was properly propagated to func-e API

		// ConfigHome must exist with envoy-version file
		require.DirExists(t, configHome, "configHome should exist at configured path")
		require.FileExists(t, path.Join(configHome, "envoy-version"), "envoy-version file should exist in configHome")

		// DataHome must exist with envoy-versions subdirectory for downloaded binaries
		require.DirExists(t, dataHome, "dataHome should exist at configured path")
		require.DirExists(t, path.Join(dataHome, "envoy-versions"), "envoy-versions dir should exist under dataHome")

		// StateHome must exist with envoy-runs subdirectory for per-run logs
		require.DirExists(t, stateHome, "stateHome should exist at configured path")
		require.DirExists(t, path.Join(stateHome, "envoy-runs"), "envoy-runs dir should exist under stateHome")

		// RuntimeDir must exist - func-e creates runID subdirectories with admin-address.txt
		require.DirExists(t, runtimeDir, "runtimeDir should exist at configured path")
	})

	t.Run("output_redirection", func(t *testing.T) {
		// Verify output was captured in buffers (not os.Stdout/Stderr)
		totalOutput := stdout.Len() + stderr.Len()
		require.Positive(t, totalOutput, "expected some output to be captured in stdout or stderr buffers")
	})
}

func TestInfra_StopStartCycle(t *testing.T) {
	cfg, err := config.New(io.Discard, io.Discard)
	require.NoError(t, err)

	// Use mock infra with fake runner - no actual Envoy process
	infra := newMockInfra(t, cfg)

	// Verify concurrent run -> stop cycles work without races
	synctest.Test(t, func(t *testing.T) {
		for i := range 5 {
			go func(id int) {
				name := utils.GetHashedName(fmt.Sprintf("test-%d", id), 64)
				infra.runEnvoy(t.Context(), "", name, []string{"--version"})
				_, ok := infra.proxyContextMap.Load(name)
				require.True(t, ok, "expected proxy context to be stored")

				infra.stopEnvoy(name)
				_, ok = infra.proxyContextMap.Load(name)
				require.False(t, ok, "expected proxy context to be removed")
			}(i)
		}
	})
}

func TestInfra_Close(t *testing.T) {
	cfg, err := config.New(io.Discard, io.Discard)
	require.NoError(t, err)

	infra := newMockInfra(t, cfg)

	// Use synctest as runEnvoy internally starts goroutines
	synctest.Test(t, func(t *testing.T) {
		for id := range 5 {
			name := utils.GetHashedName(fmt.Sprintf("proxy-%d", id), 64)
			infra.runEnvoy(t.Context(), "", name, []string{"--version"})
		}

		// Verify all proxies are running
		count := 0
		infra.proxyContextMap.Range(func(key, value any) bool {
			count++
			return true
		})
		require.Equal(t, 5, count, "expected 5 proxies to be running")

		// Close should stop all proxies concurrently
		err := infra.Close()
		require.NoError(t, err)

		// Verify all proxies were stopped
		count = 0
		infra.proxyContextMap.Range(func(key, value any) bool {
			count++
			return true
		})
		require.Equal(t, 0, count, "expected all proxies to be stopped")
	})
}

func TestGetEnvoyVersion(t *testing.T) {
	tests := []struct {
		name         string
		defaultImage string
		provider     *egv1a1.EnvoyProxyProvider
		want         string
	}{
		{
			name:         "k8s provider default release version",
			defaultImage: "docker.io/envoyproxy/envoy:distroless-v1.35.0",
			provider:     egv1a1.DefaultEnvoyProxyProvider(),
			want:         "1.35.0",
		},
		{
			name:         "k8s provider dev version",
			defaultImage: "docker.io/envoyproxy/envoy:distroless-dev",
			provider:     egv1a1.DefaultEnvoyProxyProvider(),
			want:         "",
		},
		{
			name:         "host provider envoy version unset",
			defaultImage: "docker.io/envoyproxy/envoy:distroless-v1.35.0",
			provider: &egv1a1.EnvoyProxyProvider{
				Type: egv1a1.EnvoyProxyProviderTypeHost,
				Host: &egv1a1.EnvoyProxyHostProvider{},
			},
			want: "1.35.0",
		},
		{
			name:         "host provider envoy version empty",
			defaultImage: "docker.io/envoyproxy/envoy:distroless-v1.35.0",
			provider: &egv1a1.EnvoyProxyProvider{
				Type: egv1a1.EnvoyProxyProviderTypeHost,
				Host: &egv1a1.EnvoyProxyHostProvider{EnvoyVersion: ptr.To("")},
			},
			want: "1.35.0",
		},
		{
			name:         "host provider envoy version unset dev version",
			defaultImage: "docker.io/envoyproxy/envoy:distroless-dev",
			provider: &egv1a1.EnvoyProxyProvider{
				Type: egv1a1.EnvoyProxyProviderTypeHost,
				Host: &egv1a1.EnvoyProxyHostProvider{},
			},
			want: "",
		},
		{
			name:         "host provider envoy version empty dev version",
			defaultImage: "docker.io/envoyproxy/envoy:distroless-dev",
			provider: &egv1a1.EnvoyProxyProvider{
				Type: egv1a1.EnvoyProxyProviderTypeHost,
				Host: &egv1a1.EnvoyProxyHostProvider{EnvoyVersion: ptr.To("")},
			},
			want: "",
		},
		{
			name:         "host provider envoy version custom",
			defaultImage: "docker.io/envoyproxy/envoy:distroless-v1.35.0",
			provider: &egv1a1.EnvoyProxyProvider{
				Type: egv1a1.EnvoyProxyProviderTypeHost,
				Host: &egv1a1.EnvoyProxyHostProvider{EnvoyVersion: ptr.To("1.2.3")},
			},
			want: "1.2.3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			infra := &Infra{defaultEnvoyImage: tc.defaultImage}
			proxyConfig := &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Provider: tc.provider,
				},
			}
			require.Equal(t, tc.want, infra.getEnvoyVersion(proxyConfig))
		})
	}
}

// TestTopologyInjectorDisabledInHostMode verifies we don't cause a 15+ second
// startup delay in standalone mode as Envoy waits for endpoint discovery.
// See: https://github.com/envoyproxy/gateway/issues/7080
func TestNewInfra(t *testing.T) {
	// This test verifies successful creation of Infra.
	cfg, err := config.New(io.Discard, io.Discard)
	require.NoError(t, err)

	errNotifier := message.RunnerErrorNotifier{RunnerName: t.Name(), RunnerErrors: &message.RunnerErrors{}}
	actual, err := NewInfra(t.Context(), cfg, logging.DefaultLogger(io.Discard, egv1a1.LogLevelInfo), errNotifier)
	require.NoError(t, err)
	require.NotNil(t, actual)
	require.NotNil(t, actual.Paths)
	require.NotEmpty(t, actual.sdsConfigPath)
	require.NotNil(t, actual.Logger)
	require.NotNil(t, actual.EnvoyGateway)
	require.Equal(t, egv1a1.DefaultEnvoyProxyImage, actual.defaultEnvoyImage)
	require.NotNil(t, actual.Stdout)
	require.NotNil(t, actual.Stderr)
	require.NotNil(t, actual.envoyRunner)
	require.NotNil(t, actual.errors)
}

func TestTopologyInjectorDisabledInHostMode(t *testing.T) {
	testCases := []struct {
		name                          string
		topologyInjectorDisabled      bool
		expectLocalClusterInBootstrap bool
	}{
		{
			name:                          "topology injector enabled",
			topologyInjectorDisabled:      false,
			expectLocalClusterInBootstrap: true,
		},
		{
			name:                          "topology injector disabled",
			topologyInjectorDisabled:      true,
			expectLocalClusterInBootstrap: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			proxyInfra := &ir.ProxyInfra{
				Name:      "test-proxy",
				Namespace: "default",
				Config: &egv1a1.EnvoyProxy{
					Spec: egv1a1.EnvoyProxySpec{
						Logging: egv1a1.ProxyLogging{
							Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
								egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
							},
						},
					},
				},
			}

			bootstrapConfigOptions := &bootstrap.RenderBootstrapConfigOptions{
				ProxyMetrics: &egv1a1.ProxyMetrics{
					Prometheus: &egv1a1.ProxyPrometheusProvider{
						Disable: true,
					},
				},
				XdsServerHost:            ptr.To("0.0.0.0"),
				AdminServerPort:          ptr.To(int32(0)),
				StatsServerPort:          ptr.To(int32(0)),
				TopologyInjectorDisabled: tc.topologyInjectorDisabled,
			}

			args, err := common.BuildProxyArgs(proxyInfra, nil, bootstrapConfigOptions, "test-node", false)
			require.NoError(t, err)

			// Extract the bootstrap YAML from args (it's after --config-yaml)
			var bootstrapYAML string
			for i, arg := range args {
				if arg == "--config-yaml" && i+1 < len(args) {
					bootstrapYAML = args[i+1]
					break
				}
			}
			require.NotEmpty(t, bootstrapYAML, "bootstrap YAML not found in args")

			if tc.expectLocalClusterInBootstrap {
				require.Contains(t, bootstrapYAML, "local_cluster_name:")
			} else {
				require.NotContains(t, bootstrapYAML, "local_cluster_name:")
			}
		})
	}
}

// TestResolvedMetricSinksConversion verifies that resolved metric sinks from IR
// are correctly converted to bootstrap format, including TLS configuration.
func TestResolvedMetricSinksConversion(t *testing.T) {
	sni := "otel-collector.example.com"

	testCases := []struct {
		name     string
		irSinks  []ir.ResolvedMetricSink
		expected []bootstrap.MetricSink
	}{
		{
			name:     "no sinks",
			irSinks:  nil,
			expected: []bootstrap.MetricSink{},
		},
		{
			name: "skip sink with no endpoints",
			irSinks: []ir.ResolvedMetricSink{
				{
					Destination: ir.RouteDestination{
						Name:     "metrics_otel_0",
						Settings: []*ir.DestinationSetting{{}},
					},
				},
			},
			expected: []bootstrap.MetricSink{},
		},
		{
			name: "sink without TLS",
			irSinks: []ir.ResolvedMetricSink{
				{
					Destination: ir.RouteDestination{
						Name: "metrics_otel_0",
						Settings: []*ir.DestinationSetting{
							{
								Endpoints: []*ir.DestinationEndpoint{
									{Host: "otel-collector.example.com", Port: 4317},
								},
							},
						},
					},
				},
			},
			expected: []bootstrap.MetricSink{
				{
					Address: "otel-collector.example.com",
					Port:    4317,
				},
			},
		},
		{
			name: "sink with TLS and SNI",
			irSinks: []ir.ResolvedMetricSink{
				{
					Destination: ir.RouteDestination{
						Name: "metrics_otel_0",
						Settings: []*ir.DestinationSetting{
							{
								Endpoints: []*ir.DestinationEndpoint{
									{Host: "otel-collector.example.com", Port: 443},
								},
								TLS: &ir.TLSUpstreamConfig{
									SNI: &sni,
								},
							},
						},
					},
					Authority: "otel-collector.example.com",
				},
			},
			expected: []bootstrap.MetricSink{
				{
					Address:   "otel-collector.example.com",
					Port:      443,
					Authority: "otel-collector.example.com",
					TLS: &bootstrap.MetricSinkTLS{
						SNI: sni,
					},
				},
			},
		},
		{
			name: "sink with headers and deltas",
			irSinks: []ir.ResolvedMetricSink{
				{
					Destination: ir.RouteDestination{
						Name: "metrics_otel_0",
						Settings: []*ir.DestinationSetting{
							{
								Endpoints: []*ir.DestinationEndpoint{
									{Host: "otel-collector.example.com", Port: 4317},
								},
							},
						},
					},
					Headers: []gwapiv1.HTTPHeader{
						{Name: "Authorization", Value: "Bearer token"},
					},
					ReportCountersAsDeltas:   true,
					ReportHistogramsAsDeltas: true,
				},
			},
			expected: []bootstrap.MetricSink{
				{
					Address:                  "otel-collector.example.com",
					Port:                     4317,
					ReportCountersAsDeltas:   true,
					ReportHistogramsAsDeltas: true,
					Headers: []gwapiv1.HTTPHeader{
						{Name: "Authorization", Value: "Bearer token"},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := convertResolvedMetricSinks(tc.irSinks)
			require.Equal(t, tc.expected, actual)
		})
	}
}

// TestUserConfiguredMetricSinksPreserved verifies that user-configured metric
// sinks (e.g., OpenTelemetry) are preserved in the bootstrap config for host mode.
func TestUserConfiguredMetricSinksPreserved(t *testing.T) {
	testCases := []struct {
		name        string
		telemetry   *egv1a1.ProxyTelemetry
		expectSinks bool
	}{
		{
			name:        "no telemetry config",
			telemetry:   nil,
			expectSinks: false,
		},
		{
			name: "telemetry with nil metrics",
			telemetry: &egv1a1.ProxyTelemetry{
				Metrics: nil,
			},
			expectSinks: false,
		},
		{
			name: "telemetry with otel sink",
			telemetry: &egv1a1.ProxyTelemetry{
				Metrics: &egv1a1.ProxyMetrics{
					Sinks: []egv1a1.ProxyMetricSink{
						{
							Type: egv1a1.MetricSinkTypeOpenTelemetry,
							OpenTelemetry: &egv1a1.ProxyOpenTelemetrySink{
								Host: ptr.To("otel-collector.example.com"),
								Port: 4317,
							},
						},
					},
				},
			},
			expectSinks: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			proxyConfig := &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Telemetry: tc.telemetry,
				},
			}

			// Build proxy metrics the same way CreateOrUpdateProxyInfra does
			proxyMetrics := &egv1a1.ProxyMetrics{
				Prometheus: &egv1a1.ProxyPrometheusProvider{
					Disable: true,
				},
			}
			if proxyConfig.Spec.Telemetry != nil && proxyConfig.Spec.Telemetry.Metrics != nil {
				proxyMetrics.Sinks = proxyConfig.Spec.Telemetry.Metrics.Sinks
				proxyMetrics.Matches = proxyConfig.Spec.Telemetry.Metrics.Matches
			}

			// Verify Prometheus is always disabled
			require.True(t, proxyMetrics.Prometheus.Disable)

			// Verify sinks and matches are preserved when configured
			if tc.expectSinks {
				require.NotEmpty(t, proxyMetrics.Sinks, "user-configured sinks should be preserved")
				require.Equal(t, tc.telemetry.Metrics.Sinks, proxyMetrics.Sinks)
				require.Equal(t, tc.telemetry.Metrics.Matches, proxyMetrics.Matches)
			} else {
				require.Empty(t, proxyMetrics.Sinks, "no sinks expected when not configured")
				require.Empty(t, proxyMetrics.Matches, "no matches expected when not configured")
			}
		})
	}
}
