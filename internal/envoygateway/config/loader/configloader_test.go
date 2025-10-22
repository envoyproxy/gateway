// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package loader

import (
	"context"
	_ "embed"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

var (
	//go:embed testdata/default.yaml
	defaultConfig string
	//go:embed testdata/enable-redis.yaml
	redisConfig string
	//go:embed testdata/standalone.yaml
	standaloneConfig string
	//go:embed testdata/standalone-enable-extension-server.yaml
	standaloneConfigWithExtensionServer string
)

func TestConfigLoader(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "envoy-gateway-configloader-test")
	require.NoError(t, err)
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tmpDir)

	cfgPath := tmpDir + "/config.yaml"
	require.NoError(t, os.WriteFile(cfgPath, []byte(defaultConfig), 0o600))
	s, err := config.New(os.Stdout, os.Stderr)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.TODO())
	defer func() {
		cancel()
	}()

	changed := 0
	loader := New(cfgPath, s, func(_ context.Context, cfg *config.Server) error {
		changed++
		t.Logf("config changed %d times", changed)
		if changed > 1 {
			cancel()
		}
		return nil
	})

	require.NoError(t, loader.Start(ctx, os.Stdout))
	go func() {
		_ = os.WriteFile(cfgPath, []byte(redisConfig), 0o600)
	}()

	<-ctx.Done()
}

func TestConfigLoaderStandaloneExtensionServerAndCustomResource(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "envoy-gateway-configloader-test")
	require.NoError(t, err)
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tmpDir)

	cfgPath := tmpDir + "/config.yaml"
	require.NoError(t, os.WriteFile(cfgPath, []byte(standaloneConfig), 0o600))
	s, err := config.New(os.Stdout, os.Stderr)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.TODO())
	defer func() {
		cancel()
	}()

	changed := 0
	loader := New(cfgPath, s, func(_ context.Context, cfg *config.Server) error {
		changed++
		t.Logf("config changed %d times", changed)
		if changed > 1 {
			cancel()
		}
		return nil
	})

	require.NoError(t, loader.Start(ctx, os.Stdout))
	require.NotNil(t, loader.cfg.EnvoyGateway)
	require.Nil(t, loader.cfg.EnvoyGateway.ExtensionManager)

	go func() {
		_ = os.WriteFile(cfgPath, []byte(standaloneConfigWithExtensionServer), 0o600)
	}()

	<-ctx.Done()
	require.Equal(t, 2, changed)
	require.NotNil(t, loader.cfg.EnvoyGateway.ExtensionManager)
	require.NotNil(t, loader.cfg.EnvoyGateway.ExtensionManager.PolicyResources)
	require.Equal(t, "gateway.example.io", loader.cfg.EnvoyGateway.ExtensionManager.PolicyResources[0].Group)
	require.Equal(t, "v1alpha1", loader.cfg.EnvoyGateway.ExtensionManager.PolicyResources[0].Version)
	require.Equal(t, "ExampleExtPolicy", loader.cfg.EnvoyGateway.ExtensionManager.PolicyResources[0].Kind)
}
