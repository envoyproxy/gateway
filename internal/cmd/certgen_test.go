// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/crypto"
)

func TestGetCertgenCommand(t *testing.T) {
	got := GetCertGenCommand()
	assert.Equal(t, "certgen", got.Use)
}

func TestOutputCertsForLocal(t *testing.T) {
	cfg, err := getConfig()
	require.NoError(t, err)

	certs, err := crypto.GenerateCerts(cfg)
	require.NoError(t, err)

	tmpDir := t.TempDir()
	err = outputCertsForLocal(tmpDir, certs)
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(tmpDir, "envoy-gateway", "ca.crt"))
	assert.FileExists(t, filepath.Join(tmpDir, "envoy-gateway", "tls.crt"))
	assert.FileExists(t, filepath.Join(tmpDir, "envoy-gateway", "tls.key"))
	assert.FileExists(t, filepath.Join(tmpDir, "envoy", "ca.crt"))
	assert.FileExists(t, filepath.Join(tmpDir, "envoy", "tls.crt"))
	assert.FileExists(t, filepath.Join(tmpDir, "envoy", "tls.key"))
	assert.FileExists(t, filepath.Join(tmpDir, "envoy-rate-limit", "ca.crt"))
	assert.FileExists(t, filepath.Join(tmpDir, "envoy-rate-limit", "tls.crt"))
	assert.FileExists(t, filepath.Join(tmpDir, "envoy-rate-limit", "tls.key"))
	assert.FileExists(t, filepath.Join(tmpDir, "envoy-oidc-hmac", "hmac-secret"))
}
