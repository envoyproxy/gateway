// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"bytes"
	"context"
	"io"
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
local validation error: Gateway.gateway.networking.k8s.io "eg2" is invalid: [spec.listeners[1]: Duplicate value: {"name":"tcp"}, spec.listeners: Invalid value: "array": Listener name must be unique within the Gateway, spec.listeners: Invalid value: "array": Combination of port, protocol and hostname must be unique for each listener]

apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend
  namespace: default
spec:
...
local validation error: HTTPRoute.gateway.networking.k8s.io "backend" is invalid: spec.hostnames[0]: Invalid value: ".;'.';[]": spec.hostnames[0] in body should match '^(\*\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$'

apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: backend-1
  namespace: default
spec:
...
local validation error: Backend.gateway.envoyproxy.io "backend-1" is invalid: spec.endpoints[0].ip.address: Invalid value: "a.b.c.d": spec.endpoints[0].ip.address in body should match '^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$|^(([0-9a-fA-F]{1,4}:){1,7}[0-9a-fA-F]{1,4}|::|(([0-9a-fA-F]{1,4}:){0,5})?(:[0-9a-fA-F]{1,4}){1,2})$'

apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: backend-2
  namespace: default
spec:
...
local validation error: Backend.gateway.envoyproxy.io "backend-2" is invalid: spec.endpoints: Invalid value: "array": fqdn addresses cannot be mixed with other address types

`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := bytes.NewBufferString("")
			root := newValidateCommand()
			root.SetOut(b)
			root.SetErr(b)
			args := []string{
				"--file",
				path.Join("testdata", "validate", tc.name+".yaml"),
			}

			root.SetArgs(args)
			err := root.ExecuteContext(context.Background())
			require.NoError(t, err)

			out, err := io.ReadAll(b)
			require.NoError(t, err)
			require.Equal(t, tc.output, string(out))
		})
	}
}
