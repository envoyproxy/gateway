// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	invalidGatewayConfig = `
kind: EnvoyGateway
apiVersion: config.gateway.envoyproxy.io/v1alpha1
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
			name:   "invalid gateway",
			input:  invalidGatewayConfig,
			errors: []string{"is unspecified"},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			file, err := os.CreateTemp("", "config")
			assert.NoError(t, err)
			defer os.Remove(file.Name())

			_, err = file.Write([]byte(test.input))
			assert.NoError(t, err)

			_, err = getConfigByPath(file.Name())
			for _, e := range test.errors {
				assert.ErrorContains(t, err, e)
			}
		})
	}

}
