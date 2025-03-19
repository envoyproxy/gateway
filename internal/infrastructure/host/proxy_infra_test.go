// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package host

import (
	"bytes"
	"context"
	funcE "github.com/tetratelabs/func-e/api"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

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
		Logger:          logging.DefaultLogger(egv1a1.LogLevelInfo),
		EnvoyGateway:    cfg.EnvoyGateway,
		proxyContextMap: make(map[string]*proxyContext),
		sdsConfigPath:   proxyDir,
	}
	return infra
}

func TestInfraCreateProxy(t *testing.T) {
	cfg, err := config.New()
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
			err = infra.CreateOrUpdateProxyInfra(context.Background(), tc.infra)
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
	err := funcE.Run(context.Background(), []string{"--version"}, funcE.HomeDir(tmpdir))
	require.NoError(t, err)

	i := &Infra{proxyContextMap: make(map[string]*proxyContext), HomeDir: tmpdir}
	// Ensures that run -> stop will successfully stop the envoy and we can
	// run it again without any issues.
	for range 10 {
		args := []string{
			"--config-yaml",
			"admin: {address: {socket_address: {address: '127.0.0.1', port_value: 9901}}}",
		}
		out := &bytes.Buffer{}
		i.runEnvoy(context.Background(), out, "test", args)
		require.Len(t, i.proxyContextMap, 1)
		i.stopEnvoy("test")
		require.Empty(t, i.proxyContextMap)
		require.NotContains(t, out.String(), "Address already in use")
	}
}
