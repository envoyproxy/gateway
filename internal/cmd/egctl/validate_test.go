// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"bytes"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunValidate(t *testing.T) {
	testCases := []struct {
		name   string
		output string
	}{
		{
			name: "invalid-resources",
			output: `apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: eg1
  namespace: default
spec:
...
local validation error: Gateway.gateway.networking.k8s.io "eg1" is invalid: spec.listeners[0].port: Invalid value: 88888: spec.listeners[0].port in body should be less than or equal to 65535

apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: eg2
  namespace: default
spec:
...
local validation error: Gateway.gateway.networking.k8s.io "eg2" is invalid: [spec.listeners[1]: Duplicate value: map[string]interface {}{"name":"tcp"}, spec.listeners: Invalid value: "array": Listener name must be unique within the Gateway, spec.listeners: Invalid value: "array": Combination of port, protocol and hostname must be unique for each listener]

apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend
  namespace: default
spec:
...
local validation error: HTTPRoute.gateway.networking.k8s.io "backend" is invalid: spec.hostnames[0]: Invalid value: ".;'.';[]": spec.hostnames[0] in body should match '^(\*\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$'

`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer

			err := runValidate(&out, path.Join("testdata", "validate", tc.name+".yaml"))
			require.NoError(t, err)
			require.Equal(t, tc.output, out.String())
		})
	}
}
