// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package host

import (
	"bytes"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	func_e "github.com/tetratelabs/func-e"
	"github.com/tetratelabs/func-e/api"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/crypto"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/utils/file"
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
		HomeDir:         homeDir,
		Logger:          logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo),
		EnvoyGateway:    cfg.EnvoyGateway,
		proxyContextMap: make(map[string]*proxyContext),
		sdsConfigPath:   proxyDir,
	}
	return infra
}

func TestInfraCreateProxy(t *testing.T) {
	cfg, err := config.New(os.Stdout)
	require.NoError(t, err)
	infra := newMockInfra(t, cfg)

	// TODO: add more tests once it supports configurable homeDir and runDir.
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

func TestInfra_runEnvoy_stopEnvoy(t *testing.T) {
	tmpdir := t.TempDir()
	// Ensures that all the required binaries are available.
	err := func_e.Run(t.Context(), []string{"--version"}, api.HomeDir(tmpdir))
	require.NoError(t, err)

	i := &Infra{proxyContextMap: make(map[string]*proxyContext), HomeDir: tmpdir}
	// Ensures that run -> stop will successfully stop the envoy and we can
	// run it again without any issues.
	for range 5 {
		args := []string{
			"--config-yaml",
			"admin: {address: {socket_address: {address: '127.0.0.1', port_value: 9901}}}",
		}
		out := &bytes.Buffer{}
		i.runEnvoy(t.Context(), out, "", "test", args)
		require.Len(t, i.proxyContextMap, 1)
		i.stopEnvoy("test")
		require.Empty(t, i.proxyContextMap)
		// If the cleanup didn't work, the error due to "address already in use" will be tried to be written to the nil logger,
		// which will panic.
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
