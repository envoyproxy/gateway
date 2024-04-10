// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
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
	got := getServerCommand()
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
		test := test
		t.Run(test.name, func(t *testing.T) {
			file, err := os.CreateTemp("", "config")
			require.NoError(t, err)
			defer os.Remove(file.Name())

			_, err = file.Write([]byte(test.input))
			require.NoError(t, err)

			_, err = getConfigByPath(file.Name())
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
