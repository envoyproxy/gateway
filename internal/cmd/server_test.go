// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
)

func TestGetServerCommand(t *testing.T) {
	got := GetServerCommand()
	assert.Equal(t, "server", got.Use)
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

			_, err = getConfigByPath(io.Discard, io.Discard, file.Name())
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

// TestServerCommand_OutputRedirection verifies that the server command respects output redirection.
func TestServerCommand_OutputRedirection(t *testing.T) {
	file, err := os.CreateTemp("", "config")
	require.NoError(t, err)
	defer os.Remove(file.Name())

	_, err = file.WriteString(validGatewayConfig)
	require.NoError(t, err)

	// Create separate buffers for stdout and stderr
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	// Test that getConfigByPath uses the provided writers
	cfg, err := getConfigByPath(stdout, stderr, file.Name())
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify the config has the writers set
	require.Equal(t, stdout, cfg.Stdout)
	require.Equal(t, stderr, cfg.Stderr)
}
