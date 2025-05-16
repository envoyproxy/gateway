// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package fuzz

import (
	"strings"
	"testing"

	"github.com/envoyproxy/gateway/internal/cmd/egctl"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func FuzzGatewayAPIToXDS(f *testing.F) {
	baseYAML := `apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: eg
  namespace: default
spec:
  gatewayClassName: eg
  listeners:
    - name: http
      protocol: HTTP
      port: 80
`

	backendYAML := `---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: provided-backend
  namespace: default
spec:
  endpoints:
    - ip:
        address: 0.0.0.0
        port: 8000
`

	grpcRoute := `---
apiVersion: gateway.networking.k8s.io/v1
kind: GRPCRoute
metadata:
  name: backend
  namespace: default
spec:
  parentRefs:
    - name: eg
      sectionName: grpc
  hostnames:
    - "www.grpc-example.com"
  rules:
    - matches:
        - method:
            service: com.example.Things
            method: DoThing
          headers:
            - name: com.example.Header
              value: foobar
      backendRefs:
        - name: provided-backend
          port: 9000
`

	httpRoute := `---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend
  namespace: default
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - backendRefs:
        - name: provided-backend
          port: 8000
`

	udpRoute := `---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: UDPRoute
metadata:
  name: backend
  namespace: default
spec:
  parentRefs:
    - name: eg
      sectionName: udp
  rules:
    - backendRefs:
        - name: provided-backend
          port: 3000
`

	// Add seed corpus for golang native fuzzing. OSS-Fuzz will not take these corpus into consideration.
	f.Add([]byte(baseYAML + grpcRoute + backendYAML))
	f.Add([]byte(baseYAML + httpRoute + backendYAML))
	f.Add([]byte(baseYAML + udpRoute + backendYAML))

	f.Fuzz(func(t *testing.T, b []byte) {
		rs, err := resource.LoadResourcesFromYAMLBytes(b, true)
		if err != nil {
			return
		}

		_, err = egctl.TranslateGatewayAPIToXds("default", "cluster.local", "all", rs)
		if err != nil && strings.Contains(err.Error(), "failed to translate xds") {
			t.Fatalf("%v", err)
		}
	})
}
