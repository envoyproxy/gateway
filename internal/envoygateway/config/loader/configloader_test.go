// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package loader

import (
	"context"
	_ "embed"
	"os"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
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

	ctx, cancel := context.WithCancel(context.Background())
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

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	type testResult struct {
		changed int32
		extMgr  *egv1a1.ExtensionManager
	}
	resultChannel := make(chan testResult, 1)

	var changed int32
	loader := New(cfgPath, s, func(_ context.Context, cfg *config.Server) error {
		c := atomic.AddInt32(&changed, 1)
		t.Logf("config changed %d times", c)
		if c > 1 {
			resultChannel <- testResult{changed: c, extMgr: cfg.EnvoyGateway.ExtensionManager}
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
	res := <-resultChannel // wait until second reload
	require.Equal(t, int32(2), res.changed)
	require.NotNil(t, res.extMgr)
	require.NotNil(t, res.extMgr.PolicyResources)
	require.Len(t, res.extMgr.PolicyResources, 1)
	require.Equal(t, "gateway.example.io", res.extMgr.PolicyResources[0].Group)
	require.Equal(t, "v1alpha1", res.extMgr.PolicyResources[0].Version)
	require.Equal(t, "ExampleExtPolicy", res.extMgr.PolicyResources[0].Kind)
}
