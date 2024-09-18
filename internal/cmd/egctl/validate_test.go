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
  name: eg
  namespace: default
spec:
  gatewayClassName: eg
  listeners:
    - name: http
      protocol: HTTP
      port: 88888
------
spec.listeners[0].port: Invalid value: 88888: spec.listeners[0].port in body should be less than or equal to 65535
spec.listeners: Invalid value: "array": invalid data, expected int, got float64 evaluating rule: Combination of port, protocol and hostname must be unique for each listener

apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend
  namespace: default
spec:
  parentRefs:
    - name: eg
  hostnames:
    - ".;'.';[]"
  rules:
    - backendRefs:
        - group: ""
          kind: Service
          name: backend
          port: 3000
          weight: 1
      matches:
        - path:
            type: PathPrefix
            value: /
------
spec.hostnames[0]: Invalid value: ".;'.';[]": spec.hostnames[0] in body should match '^(\*\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$'
spec.parentRefs: Invalid value: "array": no such key: group evaluating rule: sectionName or port must be specified when parentRefs includes 2 or more references to the same parent
spec.parentRefs: Invalid value: "array": no such key: group evaluating rule: sectionName or port must be unique when parentRefs includes 2 or more references to the same parent
spec.rules[0].backendRefs[0]: Invalid value: "object": invalid data, expected int, got float64 evaluating rule: Must have port for Service reference

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
