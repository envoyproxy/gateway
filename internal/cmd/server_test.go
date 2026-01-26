// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
	"context"
	"os"
	"path"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

var (
	validGatewayConfig = `
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
provider:
  type: Kubernetes
gateway:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
`
	invalidGatewayConfig = `
kind: EnvoyGateway
apiVersion: gateway.envoyproxy.io/v1alpha1
gateway: {}
`

	fileProviderGatewayConfig = `
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
gateway:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
provider:
  type: Custom
  custom:
    resource:
      type: File
      file:
        paths: ["/tmp/envoy-gateway-test"]
    infrastructure:
      type: Host
      host:
        configHome: [CONFIG_HOME_PLACE_HODLER]
`

	fileProviderGatewayConfigChanged = `
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
gateway:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
provider:
  type: Custom
  custom:
    resource:
      type: File
      file:
        paths: ["/tmp/envoy-gateway-test2"]
    infrastructure:
      type: Host
      host:
        configHome: [CONFIG_HOME_PLACE_HODLER]
`
)

func TestGetServerCommand(t *testing.T) {
	got := GetServerCommand()
	require.Equal(t, "server", got.Use)
}

func testHook(c context.Context, cfg *config.Server) error {
	if err := startRunners(c, cfg); err != nil {
		return err
	}
	return nil
}

func testCustomProvider(t *testing.T, genCert bool) (string, string) {
	// Use Custom provider to avoid take too much to discovery CRDs
	configHome := t.TempDir()
	cfgFileContent := strings.ReplaceAll(fileProviderGatewayConfig, "[CONFIG_HOME_PLACE_HODLER]", configHome)
	configPath := path.Join(t.TempDir(), "envoy-gateway.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(cfgFileContent), 0o600))

	if genCert {
		require.NoError(t, certGen(t.Context(), os.Stdout, true))
	}

	return configHome, configPath
}

func TestCustomProviderCancelWhenStarting(t *testing.T) {
	_, configPath := testCustomProvider(t, false)
	errCh := make(chan error)
	ctx, cancel := context.WithCancel(t.Context())
	go func() {
		errCh <- server(ctx, os.Stdout, os.Stdout, configPath, testHook, nil)
	}()
	go func() {
		cancel()
	}()

	err := <-errCh
	require.NoError(t, err)
}

func TestCustomProviderFailedToStart(t *testing.T) {
	_, configPath := testCustomProvider(t, false)

	errCh := make(chan error)
	ctx, cancel := context.WithCancel(t.Context())
	go func() {
		errCh <- server(ctx, os.Stdout, os.Stdout, configPath, testHook, nil)
	}()

	err := <-errCh
	cancel()
	require.Error(t, err, "failed to load TLS config")
}

func TestCustomProviderCancelWhenConfigReload(t *testing.T) {
	configHome, configPath := testCustomProvider(t, true)

	errCh := make(chan error)
	ctx, cancel := context.WithCancel(t.Context())
	count := atomic.Int32{}
	hook := func(c context.Context, cfg *config.Server) error {
		if count.Add(1) >= 2 {
			t.Logf("Config reload triggered, cancelling context")
			go cancel()
		}
		if err := startRunners(c, cfg); err != nil {
			return err
		}
		return nil
	}

	startedCallback := func() {
		t.Logf("Trigger config reload")
		go func() {
			cfgFileContentChanged := strings.ReplaceAll(fileProviderGatewayConfigChanged, "[CONFIG_HOME_PLACE_HODLER]", configHome)
			require.NoError(t, os.WriteFile(configPath, []byte(cfgFileContentChanged), 0o600))
		}()
	}

	go func() {
		errCh <- server(ctx, os.Stdout, os.Stdout, configPath, hook, startedCallback)
	}()

	err := <-errCh
	cancel()
	require.NoError(t, err)
}

func TestGetConfigValidate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		errors []string
	}{
		{
			name:   "valid gateway",
			input:  validGatewayConfig,
			errors: nil,
		},
		{
			name:   "invalid gateway",
			input:  invalidGatewayConfig,
			errors: []string{"is unspecified"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file, err := os.CreateTemp("", "config")
			require.NoError(t, err)
			defer os.Remove(file.Name())

			_, err = file.WriteString(test.input)
			require.NoError(t, err)

			_, err = getConfigByPath(os.Stderr, file.Name())
			if test.errors == nil {
				require.NoError(t, err)
			} else {
				for _, e := range test.errors {
					require.ErrorContains(t, err, e)
				}
			}
		})
	}
}
