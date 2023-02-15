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

func TestGetServerCommand(t *testing.T) {
	got := getServerCommand()
	assert.Equal(t, "server", got.Use)
}

func TestGetConfigValidate(t *testing.T) {
	file, err := os.CreateTemp("", "config")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	_, err = file.Write([]byte(`
kind: EnvoyGateway
apiVersion: config.gateway.envoyproxy.io/v1alpha1
gateway: {}
`))
	assert.NoError(t, err)

	_, err = getConfigByPath(file.Name())
	assert.Error(t, err)
}
