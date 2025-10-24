// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package host

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestGetPaths_Defaults(t *testing.T) {
	// Test with nil config - should return XDG defaults
	paths, err := GetPaths(nil)
	require.NoError(t, err)
	require.NotNil(t, paths)

	u, _ := user.Current()

	// Verify XDG defaults to home and temp dirs
	require.Equal(t, filepath.Join(u.HomeDir, ".config", "envoy-gateway"), paths.ConfigHome)
	require.Equal(t, filepath.Join(u.HomeDir, ".local", "share", "envoy-gateway"), paths.DataHome)
	require.Equal(t, filepath.Join(u.HomeDir, ".local", "state", "envoy-gateway"), paths.StateHome)
	require.Equal(t, filepath.Join(os.TempDir(), "envoy-gateway-"+u.Uid), paths.RuntimeDir)
}

func TestGetPaths_CustomConfig(t *testing.T) {
	// Test with custom configuration
	customConfigHome := "/custom/config"
	customDataHome := "/custom/data"
	customStateHome := "/custom/state"
	customRuntimeDir := "/custom/runtime"

	cfg := &egv1a1.EnvoyGatewayHostInfrastructureProvider{
		ConfigHome: &customConfigHome,
		DataHome:   &customDataHome,
		StateHome:  &customStateHome,
		RuntimeDir: &customRuntimeDir,
	}

	paths, err := GetPaths(cfg)
	require.NoError(t, err)
	require.NotNil(t, paths)

	// Verify custom paths are used
	require.Equal(t, customConfigHome, paths.ConfigHome)
	require.Equal(t, customDataHome, paths.DataHome)
	require.Equal(t, customStateHome, paths.StateHome)
	require.Equal(t, customRuntimeDir, paths.RuntimeDir)
}

func TestGetPaths_PartialConfig(t *testing.T) {
	// Test with only some fields configured
	customDataHome := "/custom/data"

	cfg := &egv1a1.EnvoyGatewayHostInfrastructureProvider{
		DataHome: &customDataHome,
	}

	paths, err := GetPaths(cfg)
	require.NoError(t, err)
	require.NotNil(t, paths)

	u, _ := user.Current()

	// Verify custom dataHome but defaults for others
	require.Equal(t, filepath.Join(u.HomeDir, ".config", "envoy-gateway"), paths.ConfigHome)
	require.Equal(t, customDataHome, paths.DataHome)
	require.Equal(t, filepath.Join(u.HomeDir, ".local", "state", "envoy-gateway"), paths.StateHome)
	require.Equal(t, filepath.Join(os.TempDir(), "envoy-gateway-"+u.Uid), paths.RuntimeDir)
}

func TestPaths_CertDir(t *testing.T) {
	// Test CertDir helper - certs are stored under ConfigHome
	paths := &Paths{
		ConfigHome: "/test/config",
	}

	certDir := paths.CertDir("envoy")
	require.Equal(t, filepath.Join("/test/config", "certs", "envoy"), certDir)

	certDir = paths.CertDir("envoy-gateway")
	require.Equal(t, filepath.Join("/test/config", "certs", "envoy-gateway"), certDir)
}

func TestGetPaths_RuntimeDirUID(t *testing.T) {
	// Verify UID is included in runtime directory
	paths, err := GetPaths(nil)
	require.NoError(t, err)

	u, _ := user.Current()
	expectedPrefix := filepath.Join(os.TempDir(), "envoy-gateway-"+u.Uid)

	require.Equal(t, expectedPrefix, paths.RuntimeDir)
}

func TestGetPaths_EmptyConfig(t *testing.T) {
	// Test with empty (but non-nil) config - should still use defaults
	cfg := &egv1a1.EnvoyGatewayHostInfrastructureProvider{}

	paths, err := GetPaths(cfg)
	require.NoError(t, err)
	require.NotNil(t, paths)

	u, _ := user.Current()

	// Should use defaults when all fields are nil
	require.Equal(t, filepath.Join(u.HomeDir, ".config", "envoy-gateway"), paths.ConfigHome)
	require.Equal(t, filepath.Join(u.HomeDir, ".local", "share", "envoy-gateway"), paths.DataHome)
	require.Equal(t, filepath.Join(u.HomeDir, ".local", "state", "envoy-gateway"), paths.StateHome)
	require.Equal(t, filepath.Join(os.TempDir(), "envoy-gateway-"+u.Uid), paths.RuntimeDir)
}
