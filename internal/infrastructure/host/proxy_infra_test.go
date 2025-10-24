// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package host

import (
	"io"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/crypto"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/common"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/utils/file"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
	"github.com/envoyproxy/gateway/test/utils"
)

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
		Paths:           paths,
		Logger:          logging.DefaultLogger(io.Discard, egv1a1.LogLevelInfo),
		EnvoyGateway:    cfg.EnvoyGateway,
		proxyContextMap: make(map[string]*proxyContext),
		sdsConfigPath:   proxyDir,
		Stdout:          io.Discard,
		Stderr:          io.Discard,
	}
	return infra
}

func TestInfraCreateProxy(t *testing.T) {
	cfg, err := config.New(io.Discard, io.Discard)
	require.NoError(t, err)
	infra := newMockInfra(t, cfg)

	testCases := []struct {
		name   string
		expect bool
		infra  *ir.Infra
	}{
		{
			name:   "nil cfg",
			expect: false,
			infra:  nil,
		},
		{
			name:   "nil proxy",
			expect: false,
			infra: &ir.Infra{
				Proxy: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err = infra.CreateOrUpdateProxyInfra(t.Context(), tc.infra)
			if tc.expect {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
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

// TestInfra_runEnvoy verifies Envoy process lifecycle, output redirection, and XDG directory usage.
func TestInfra_runEnvoy(t *testing.T) {
	// Create separate XDG directories
	baseDir := t.TempDir()
	configHome := path.Join(baseDir, "config")
	dataHome := path.Join(baseDir, "data")
	stateHome := path.Join(baseDir, "state")
	runtimeDir := path.Join(baseDir, "runtime")

	// Create separate buffers for stdout and stderr
	buffers := utils.DumpLogsOnFail(t, "stdout", "stderr")
	stdout := buffers[0]
	stderr := buffers[1]

	paths := &Paths{
		ConfigHome: configHome,
		DataHome:   dataHome,
		StateHome:  stateHome,
		RuntimeDir: runtimeDir,
	}
	i := &Infra{
		proxyContextMap: make(map[string]*proxyContext),
		Paths:           paths,
		Logger:          logging.DefaultLogger(stdout, egv1a1.LogLevelInfo),
		Stdout:          stdout,
		Stderr:          stderr,
	}

	// Run envoy once to let func-e set up all XDG directories
	args := []string{
		"--config-yaml",
		"admin: {address: {socket_address: {address: '127.0.0.1', port_value: 9901}}}",
	}
	i.runEnvoy(t.Context(), "", "test", args)
	require.Len(t, i.proxyContextMap, 1)

	// Wait for func-e to create all XDG directories
	require.Eventually(t, func() bool {
		_, err := os.Stat(path.Join(configHome, "envoy-version"))
		return err == nil
	}, 5*time.Second, 100*time.Millisecond, "envoy-version file should be created in configHome")

	i.stopEnvoy("test")
	require.Empty(t, i.proxyContextMap)

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

		// Verify each XDG directory is separate (not the same path)
		require.NotEqual(t, configHome, dataHome, "configHome and dataHome must be different")
		require.NotEqual(t, dataHome, stateHome, "dataHome and stateHome must be different")
		require.NotEqual(t, stateHome, runtimeDir, "stateHome and runtimeDir must be different")
	})

	t.Run("output_redirection", func(t *testing.T) {
		// Verify output was captured in buffers (not os.Stdout/Stderr)
		totalOutput := stdout.Len() + stderr.Len()
		require.Positive(t, totalOutput, "expected some output to be captured in stdout or stderr buffers")
	})

	t.Run("stop_start_cycle", func(t *testing.T) {
		// Ensures that run -> stop cycle works multiple times without issues
		for range 5 {
			args := []string{
				"--config-yaml",
				"admin: {address: {socket_address: {address: '127.0.0.1', port_value: 9901}}}",
			}
			i.runEnvoy(t.Context(), "", "test", args)
			require.Len(t, i.proxyContextMap, 1)
			i.stopEnvoy("test")
			require.Empty(t, i.proxyContextMap)
			// If the cleanup didn't work, the error due to "address already in use" will be
			// tried to be written to the nil logger, which will panic.
		}
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
