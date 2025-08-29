// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package fuzz

import (
	"fmt"
	"strings"
	"testing"

	"github.com/envoyproxy/gateway/internal/cmd/egctl"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

// Reused YAML snippets.
const (
	baseYAML = `apiVersion: gateway.networking.k8s.io/v1
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
	backendYAML = `---
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
	grpcRouteYAML = `---
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
	httpRouteYAML = `---
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
	udpRouteYAML = `---
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
)

// Helpers for benchmark route generation.
func genHTTPRoutes(n int) string {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		sb.WriteString(fmt.Sprintf(`---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend-%d
  namespace: default
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example-%d.com"
  rules:
    - backendRefs:
        - name: provided-backend
          port: 8000
`, i, i))
	}
	return sb.String()
}

func genGRPCRoutes(n int) string {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		sb.WriteString(fmt.Sprintf(`---
apiVersion: gateway.networking.k8s.io/v1
kind: GRPCRoute
metadata:
  name: backend-grpc-%d
  namespace: default
spec:
  parentRefs:
    - name: eg
      sectionName: grpc
  hostnames:
    - "www.grpc-%d.example.com"
  rules:
    - matches:
        - method:
            service: com.example.Service%d
            method: Call
      backendRefs:
        - name: provided-backend
          port: 9000
`, i, i, i))
	}
	return sb.String()
}

func genUDPRoutes(n int) string {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		sb.WriteString(fmt.Sprintf(`---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: UDPRoute
metadata:
  name: backend-udp-%d
  namespace: default
spec:
  parentRefs:
    - name: eg
      sectionName: udp
  rules:
    - backendRefs:
        - name: provided-backend
          port: %d
`, i, 3000+i))
	}
	return sb.String()
}

// Benchmark cases: small / medium / large.
func BenchmarkGatewayAPItoXDS(b *testing.B) {
	type benchCase struct {
		name string
		yaml string
	}
	medium := baseYAML + backendYAML +
		genHTTPRoutes(10) +
		genGRPCRoutes(5) +
		genUDPRoutes(2)
	large := baseYAML + backendYAML +
		genHTTPRoutes(100) +
		genGRPCRoutes(50) +
		genUDPRoutes(10)

	cases := []benchCase{
		{
			name: "small",
			yaml: baseYAML + httpRouteYAML + backendYAML,
		},
		{
			name: "medium",
			yaml: medium,
		},
		{
			name: "large",
			yaml: large,
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			rs, err := resource.LoadResourcesFromYAMLBytes([]byte(tc.yaml), true)
			if err != nil {
				b.Fatalf("load: %v", err)
			}
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, err = egctl.TranslateGatewayAPIToXds("default", "cluster.local", "all", rs)
				if err != nil && strings.Contains(err.Error(), "failed to translate xds") {
					b.Fatalf("%v", err)
				}
			}
		})
	}
}
