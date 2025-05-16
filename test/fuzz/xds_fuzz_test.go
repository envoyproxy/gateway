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
	// Add seed corpus for golang native fuzzing. OSS-Fuzz will not take these corpus into consideration.
	// grpc route
	f.Add([]byte("apiVersion: gateway.networking.k8s.io/v1\nkind: GatewayClass\nmetadata:\n  name: eg\nspec:\n  controllerName: gateway.envoyproxy.io/gatewayclass-controller\n---\napiVersion: gateway.networking.k8s.io/v1\nkind: Gateway\nmetadata:\n  name: eg\n  namespace: default\nspec:\n  gatewayClassName: eg\n  listeners:\n    - name: http\n      protocol: HTTP\n      port: 80\n---\napiVersion: gateway.networking.k8s.io/v1\nkind: GRPCRoute\nmetadata:\n  name: backend\n  namespace: default\nspec:\n  parentRefs:\n    - name: eg\n      sectionName: grpc\n  hostnames:\n    - \"www.grpc-example.com\"\n  rules:\n    - matches:\n        - method:\n            service: com.example.Things\n            method: DoThing\n          headers:\n            - name: com.example.Header\n              value: foobar\n      backendRefs:\n        - name: provided-backend\n          port: 9000\n---\napiVersion: gateway.envoyproxy.io/v1alpha1\nkind: Backend\nmetadata:\n  name: provided-backend\n  namespace: default\nspec:\n  endpoints:\n    - ip:\n        address: 0.0.0.0\n        port: 8000\n"))
	// http route
	f.Add([]byte("apiVersion: gateway.networking.k8s.io/v1\nkind: GatewayClass\nmetadata:\n  name: eg\nspec:\n  controllerName: gateway.envoyproxy.io/gatewayclass-controller\n---\napiVersion: gateway.networking.k8s.io/v1\nkind: Gateway\nmetadata:\n  name: eg\n  namespace: default\nspec:\n  gatewayClassName: eg\n  listeners:\n    - name: http\n      protocol: HTTP\n      port: 80\n---\napiVersion: gateway.networking.k8s.io/v1\nkind: HTTPRoute\nmetadata:\n  name: backend\n  namespace: default\nspec:\n  parentRefs:\n    - name: eg\n  hostnames:\n    - \"www.example.com\"\n  rules:\n    - backendRefs:\n        - name: provided-backend\n          port: 8000\n---\napiVersion: gateway.envoyproxy.io/v1alpha1\nkind: Backend\nmetadata:\n  name: provided-backend\n  namespace: default\nspec:\n  endpoints:\n    - ip:\n        address: 0.0.0.0\n        port: 8000\n"))
	// udp route
	f.Add([]byte("apiVersion: gateway.networking.k8s.io/v1\nkind: GatewayClass\nmetadata:\n  name: eg\nspec:\n  controllerName: gateway.envoyproxy.io/gatewayclass-controller\n---\napiVersion: gateway.networking.k8s.io/v1\nkind: Gateway\nmetadata:\n  name: eg\n  namespace: default\nspec:\n  gatewayClassName: eg\n  listeners:\n    - name: http\n      protocol: HTTP\n      port: 80\n---\napiVersion: gateway.networking.k8s.io/v1alpha2\nkind: UDPRoute\nmetadata:\n  name: backend\n  namespace: default\nspec:\n  parentRefs:\n    - name: eg\n      sectionName: udp\n  rules:\n    - backendRefs:\n        - name: backend\n          port: 3000\n---\napiVersion: gateway.envoyproxy.io/v1alpha1\nkind: Backend\nmetadata:\n  name: provided-backend\n  namespace: default\nspec:\n  endpoints:\n    - ip:\n        address: 0.0.0.0\n        port: 8000\n"))
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
