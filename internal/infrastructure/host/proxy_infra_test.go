// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package host

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	func_e "github.com/tetratelabs/func-e"
	"github.com/tetratelabs/func-e/api"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/crypto"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/common"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/utils/file"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
	testutils "github.com/envoyproxy/gateway/test/utils"
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

	infra := &Infra{
		HomeDir:       homeDir,
		Logger:        logging.DefaultLogger(io.Discard, egv1a1.LogLevelInfo),
		EnvoyGateway:  cfg.EnvoyGateway,
		sdsConfigPath: proxyDir,
		Stdout:        io.Discard,
		Stderr:        io.Discard,
	}
	return infra
}

func TestInfraCreateProxy(t *testing.T) {
	cfg, err := config.New(io.Discard, io.Discard)
	require.NoError(t, err)
	infra := newMockInfra(t, cfg)

	// TODO: add more tests once it supports configurable homeDir and runDir.
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
}

func TestInfra_CreateOrUpdateProxyInfra_Success(t *testing.T) {
	tmpdir := t.TempDir()
	// Ensures that all the required binaries are available.
	err := func_e.Run(t.Context(), []string{"--version"}, api.HomeDir(tmpdir))
	require.NoError(t, err)

	cfg, err := config.New(io.Discard, io.Discard)
	require.NoError(t, err)
	infra := newMockInfra(t, cfg)

	testCases := []struct {
		name              string
		proxyName         string
		expectProxyLoaded bool
	}{
		{
			name:              "create new proxy",
			proxyName:         "test-proxy",
			expectProxyLoaded: true,
		},
		{
			name:              "idempotent - proxy already exists",
			proxyName:         "test-proxy-idempotent",
			expectProxyLoaded: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			infraIR := &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Name:      tc.proxyName,
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

			hashedName := utils.GetHashedName(tc.proxyName, 64)
			t.Cleanup(func() { infra.stopEnvoy(hashedName) })

			// First call should create the proxy
			actual := infra.CreateOrUpdateProxyInfra(t.Context(), infraIR)
			require.NoError(t, actual)

			// Verify proxy context was stored
			_, loaded := infra.proxyContextMap.Load(hashedName)
			require.Equal(t, tc.expectProxyLoaded, loaded)

			// Second call should be idempotent (early return)
			actual = infra.CreateOrUpdateProxyInfra(t.Context(), infraIR)
			require.NoError(t, actual)

			// Verify proxy is still loaded
			_, loaded = infra.proxyContextMap.Load(hashedName)
			require.Equal(t, tc.expectProxyLoaded, loaded)
		})
	}
}

func TestInfra_DeleteProxyInfra(t *testing.T) {
	tmpdir := t.TempDir()
	// Ensures that all the required binaries are available.
	err := func_e.Run(t.Context(), []string{"--version"}, api.HomeDir(tmpdir))
	require.NoError(t, err)

	cfg, err := config.New(io.Discard, io.Discard)
	require.NoError(t, err)
	infra := newMockInfra(t, cfg)

	testCases := []struct {
		name          string
		setupProxy    bool
		proxyName     string
		infraIR       *ir.Infra
		expectedError string
		expectRemoved bool
	}{
		{
			name:       "delete existing proxy",
			setupProxy: true,
			proxyName:  "test-proxy-delete",
			infraIR: &ir.Infra{
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
			},
			expectedError: "",
			expectRemoved: true,
		},
		{
			name:       "delete non-existent proxy",
			setupProxy: false,
			proxyName:  "non-existent-proxy",
			infraIR: &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Name:      "non-existent-proxy",
					Namespace: "default",
					Config:    &egv1a1.EnvoyProxy{},
				},
			},
			expectedError: "",
			expectRemoved: false,
		},
		{
			name:          "nil infra",
			setupProxy:    false,
			infraIR:       nil,
			expectedError: "infra ir is nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var hashedName string
			if tc.setupProxy {
				// Create a proxy first
				actual := infra.CreateOrUpdateProxyInfra(t.Context(), tc.infraIR)
				require.NoError(t, actual)

				hashedName = utils.GetHashedName(tc.proxyName, 64)
				t.Cleanup(func() { infra.stopEnvoy(hashedName) })

				_, loaded := infra.proxyContextMap.Load(hashedName)
				require.True(t, loaded, "proxy should be loaded before deletion")
			}

			// Delete the proxy
			actual := infra.DeleteProxyInfra(t.Context(), tc.infraIR)
			if tc.expectedError != "" {
				require.EqualError(t, actual, tc.expectedError)
			} else {
				require.NoError(t, actual)
			}

			// Verify deletion
			if tc.expectRemoved {
				_, loaded := infra.proxyContextMap.Load(hashedName)
				require.False(t, loaded, "proxy should be removed after deletion")
			}
		})
	}
}

func TestInfra_runEnvoy_stopEnvoy(t *testing.T) {
	tmpdir := t.TempDir()
	// Ensures that all the required binaries are available.
	err := func_e.Run(t.Context(), []string{"--version"}, api.HomeDir(tmpdir))
	require.NoError(t, err)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	i := &Infra{
		HomeDir: tmpdir,
		Logger:  logging.DefaultLogger(stdout, egv1a1.LogLevelInfo),
		Stdout:  stdout,
		Stderr:  stderr,
	}
	// Ensures that run -> stop will successfully stop the envoy and we can
	// run it again without any issues.
	for range 5 {
		args := []string{
			"--config-yaml",
			"admin: {address: {socket_address: {address: '127.0.0.1', port_value: 9901}}}",
		}
		i.runEnvoy(t.Context(), "", "test", args)
		_, ok := i.proxyContextMap.Load("test")
		require.True(t, ok, "expected proxy context to be stored")
		i.stopEnvoy("test")
		_, ok = i.proxyContextMap.Load("test")
		require.False(t, ok, "expected proxy context to be removed")
		// If the cleanup didn't work, the error due to "address already in use" will be tried to be written to the nil logger,
		// which will panic.
	}
}

func TestInfra_Close(t *testing.T) {
	tmpdir := t.TempDir()
	// Ensures that all the required binaries are available.
	err := func_e.Run(t.Context(), []string{"--version"}, api.HomeDir(tmpdir))
	require.NoError(t, err)

	i := &Infra{
		HomeDir: tmpdir,
		Logger:  logging.DefaultLogger(io.Discard, egv1a1.LogLevelInfo),
		Stdout:  io.Discard,
		Stderr:  io.Discard,
	}

	// Start multiple proxies
	for idx := range 3 {
		args := []string{
			"--config-yaml",
			"admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}",
		}
		i.runEnvoy(t.Context(), "", fmt.Sprintf("test-%d", idx), args)
	}

	// Verify all proxies are running
	count := 0
	i.proxyContextMap.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	require.Equal(t, 3, count, "expected 3 proxies to be running")

	// Close should stop all proxies
	err = i.Close()
	require.NoError(t, err)

	// Verify all proxies are stopped
	found := false
	i.proxyContextMap.Range(func(key, value any) bool {
		found = true
		return false
	})
	require.False(t, found, "expected all proxies to be stopped")
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

// TestInfra_runEnvoy_OutputRedirection verifies that Envoy output goes to configured writers, not os.Stdout/Stderr.
func TestInfra_runEnvoy_OutputRedirection(t *testing.T) {
	tmpdir := t.TempDir()
	// Ensures that all the required binaries are available.
	err := func_e.Run(t.Context(), []string{"--version"}, api.HomeDir(tmpdir))
	require.NoError(t, err)

	// Create separate buffers for stdout and stderr
	buffers := testutils.DumpLogsOnFail(t, "stdout", "stderr")
	stdout := buffers[0]
	stderr := buffers[1]

	i := &Infra{
		HomeDir: tmpdir,
		Logger:  logging.DefaultLogger(stdout, egv1a1.LogLevelInfo),
		Stdout:  stdout,
		Stderr:  stderr,
	}

	// Run envoy with an invalid config to force it to write to stderr and exit quickly
	args := []string{
		"--config-yaml",
		"invalid: yaml: syntax",
	}

	i.runEnvoy(t.Context(), "", "test", args)
	_, ok := i.proxyContextMap.Load("test")
	require.True(t, ok, "expected proxy context to be stored")

	// Wait a bit for envoy to fail
	require.Eventually(t, func() bool {
		return stderr.Len() > 0 || stdout.Len() > 0
	}, 5*time.Second, 100*time.Millisecond, "expected output to be written to buffers")

	i.stopEnvoy("test")
	_, ok = i.proxyContextMap.Load("test")
	require.False(t, ok, "expected proxy context to be removed")

	// Verify that output was captured in buffers (either stdout or stderr should have content)
	totalOutput := stdout.Len() + stderr.Len()
	require.Positive(t, totalOutput, "expected some output to be captured in stdout or stderr buffers")
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
	// This test verifies successful creation of Infra using defaults.
	// NewInfra uses hardcoded paths like defaultHomeDir and defaultLocalCertPathDir,
	// so we can only test the success case where those directories exist.
	cfg, err := config.New(io.Discard, io.Discard)
	require.NoError(t, err)

	actual, err := NewInfra(t.Context(), cfg, logging.DefaultLogger(io.Discard, egv1a1.LogLevelInfo))
	require.NoError(t, err)
	require.NotNil(t, actual)
	require.Equal(t, defaultHomeDir, actual.HomeDir)
	require.Equal(t, defaultLocalCertPathDir, actual.sdsConfigPath)
	require.NotNil(t, actual.Logger)
	require.NotNil(t, actual.EnvoyGateway)
	require.Equal(t, egv1a1.DefaultEnvoyProxyImage, actual.defaultEnvoyImage)
	require.NotNil(t, actual.Stdout)
	require.NotNil(t, actual.Stderr)
}

func TestCreateSdsConfig(t *testing.T) {
	dir := t.TempDir()
	// Create required cert files
	require.NoError(t, file.WriteDir([]byte("test ca"), dir, XdsTLSCaFilename))
	require.NoError(t, file.WriteDir([]byte("test cert"), dir, XdsTLSCertFilename))
	require.NoError(t, file.WriteDir([]byte("test key"), dir, XdsTLSKeyFilename))

	actual := createSdsConfig(dir)
	require.NoError(t, actual)

	// Verify CA config was created
	caConfigPath := filepath.Join(dir, common.SdsCAFilename)
	actualCAConfig, err := os.ReadFile(caConfigPath)
	require.NoError(t, err)
	require.NotEmpty(t, actualCAConfig)

	// Verify cert config was created
	certConfigPath := filepath.Join(dir, common.SdsCertFilename)
	actualCertConfig, err := os.ReadFile(certConfigPath)
	require.NoError(t, err)
	require.NotEmpty(t, actualCertConfig)
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
