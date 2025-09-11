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
    - name: https
      protocol: HTTPS
      port: 443
      tls:
        mode: Terminate
        certificateRefs:
          - name: tls-secret
    - name: grpc
      protocol: HTTP
      port: 81
    - name: udp
      protocol: UDP
      port: 82
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
	tlsSecretYAML = `---
apiVersion: v1
kind: Secret
metadata:
  name: tls-secret
  namespace: default
type: kubernetes.io/tls
data:
  tls.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURERENDQWZTZ0F3SUJBZ0lVRUZNaFA5ZUo5WEFCV3NRNVptNmJSazJjTE5Rd0RRWUpLb1pJaHZjTkFRRUwKQlFBd0ZqRVVNQklHQTFVRUF3d0xabTl2TG1KaGNpNWpiMjB3SGhjTk1qUXdNakk1TURrekNEWXdXaGNOTXpRdwpNakkyTUJUME1ERXdXakFXTVJRd0VnWURWUVFEREF0bWIyOHVZbUZ5TG1OdmJUQ0NBU0l3RFFZSktvWklodmNOCkFRRUJCUUFEZ2dFUEFEQ0NBUW9DZ2dFQkFKbEk2WXhFOVprQ1BzNnBDUXhickNtZWl4OVA1RGZ4OVJ1NUxENFQKSm1kVzdJS2R0UVYvd2ZMbXRzdTc2QithVGRDaldlMEJUZmVPT1JCYlIzY1BBRzZFbFFMaWNsUVVydW4zcStncwpKcEsrSTdjSStqNXc4STY4WEg1V1E3clZVdGJ3SHBxYncrY1ZuQnFJVU9MaUlhdGpJZjdLWDUxTTF1RjljZkVICkU0RG5jSDZyYnI1OS9SRlpCc2toeHM1T3p3Sklmb2hreXZGd2V1VHd4Sy9WcGpJKzdPYzQ4QUJDWHBOTzlEL3EKRWgrck9hdWpBTWNYZ0hRSVRrQ2lpVVRjVW82TFNIOXZMWlB0YXFmem9acTZuaE1xcFc2NUUxcEF3RjNqeVRUeAphNUk4SmNmU0Zqa2llWjIwTFVRTW43TThVNHhIamFvL2d2SDBDQWZkQjdSTFUyc0NBd0VBQWFOTE1GRXdIUVlEClZSME9CQllFRk9SQ0U4dS8xRERXN2loWnA3Y3g5dFNtUG02T01COEdBMVVkSXdRWU1CYUFGT1JDRTh1LzFERFcKN2loWnA3Y3g5dFNtUG02T01BOEdBMVVkRXdFQi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQgpBRnQ1M3pqc3FUYUg1YThFMmNodm1XQWdDcnhSSzhiVkxNeGl3TkdqYm1FUFJ6K3c2TngrazBBOEtFY0lEc0tjClNYY2k1OHU0b1didFZKQmx6YS9adWpIUjZQMUJuT3BsK2FveTc4NGJiZDRQMzl3VExvWGZNZmJCQ20xdmV2aDkKQUpLbncyWnRxcjRta2JMY3hFcWxxM3NCTEZBUzlzUUxuS05DZTJjR0xkVHAyYm9HK3FjZ3lRZ0NJTTZmOEVNdgpXUGlmQ01NR3V6Sy9HUkY0YlBPL1lGNDhld0R1M1VlaWgwWFhkVUFPRTlDdFVhOE5JaGMxVVBhT3pQcnRZVnFyClpPR2t2L0t1K0I3OGg4U0VzTzlYclFjdXdiT25KeDZLdFIrYWV5a3ZBcFhDUTNmWkMvYllLQUFSK1A4QUpvUVoKYndJVW1YaTRnajVtK2JLUGhlK2lyK0U9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
  tls.key: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUV2UUlCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktjd2dnU2pBZ0VBQW9JQkFRQ1pTT21NUlBXWkFqN08KcVFrTVc2d3Bub3NmVCtRMzhmVWJ1U3crRXlablZ1eUNuYlVGZjhIeTVyYkx1K2dObWszUW8xbnRBVTMzamprcwpXMGQzRHdCdWhKVUM0bkpVRks3cDk2dm9MQ2FTdmlPM0NQbytjUENPdkZ4K1ZrTzYxVkxXOEI2YW04UG5GWndhCmlGRGk0aUdyWXlIK3lsK2RUTmJoZlhIeEJ4T0E1M0IrcTI2K2ZmMFJXUWJKSWNiT1RzOENTSDZJWk1yeGNIcmsKOE1TdjFhWXlQdXpuT1BBQVFsNlRUdlEvNmhJZnF6bXJvd0RIRjRCMENFNUFvb2xFM0ZLT2kwaC9ieTJUN1dxbgo4NkdhdXA0VEtxVnV1Uk5hUU1CZDQ4azA4V3VTUENYSDBoWTVJbm1kdEMxRURKK3pQRk9NUjQycVA0THg5QWdICjNRZTBTMU5yQWdNQkFBRUNnZ0VBWTFGTUlLNDVXTkVNUHJ6RTZUY3NNdVV2RkdhQVZ4bVk5NW5SMEtwajdvb3IKY21CVys2ZXN0TTQ4S1AwaitPbXd3VFpMY29Cd3VoWGN0V1Bob1lXcDhteWUxRUlEdjNyaHRHMDdocEQ1NGg2dgpCZzh3ejdFYStzMk9sT0N6UnlKNzBSY281YlhjWDNGaGJjdnFlRWJwaFFyQnpOSEtLMjZ4cmZqNWZIT3p6T1FGCmJHdUZ3SDVic3JGdFhlajJXM3c4eW90N0ZQSDV3S3RpdnhvSWU5RjMyOXNnOU9EQnZqWnpiaG1LVTArckFTK1kKRGVield2bFJyaEUrbXVmQTN6M0N0QXhDOFJpNzNscFNoTDRQQWlvcG1SUXlxZXRXMjYzOFFxcnM0R3hnNzhwbApJUXJXTmNBc2s3Slg5d3RZenV6UFBXSXRWTTFscFJiQVRhNTJqdFl2NVFLQmdRRE5tMTFtZTRYam1ZSFV2cStZCmFTUzdwK2UybXZEMHVaOU9JeFluQnBWMGkrckNlYnFFMkE1Rm5hcDQ5Yld4QTgwUElldlVkeUpCL2pUUkoxcVMKRUpXQkpMWm1LVkg2K1QwdWw1ZUtOcWxFTFZHU0dCSXNpeE9SUXpDZHBoMkx0UmtBMHVjSVUzY3hiUmVMZkZCRQpiSkdZWENCdlNGcWd0VDlvZTFldVpMVmFOd0tCZ1FERWdENzJENk81eGIweEQ1NDQ1M0RPMUJhZmd6aThCWDRTCk1SaVd2LzFUQ0w5N05sRWtoeXovNmtQd1owbXJRcE5CMzZFdkpKZFVteHdkU2MyWDhrOGcxMC85NVlLQkdWQWoKL3d0YVZYbE9WeEFvK0ZSelpZeFpyQ29uWWFSMHVwUzFybDRtenN4REhlZU9mUVZUTUgwUjdZN0pnbTA5dXQ4SwplanAvSXZBb1F3S0JnQjNaRWlRUWhvMVYrWjBTMlpiOG5KS0plMy9zMmxJTXFHM0ZkaS9RS3Q0eWViQWx6OGY5ClBZVXBzRmZEQTg5Z3grSU1nSm5sZVptdTk2ZnRXSjZmdmJSenllN216TG5zZU05TXZua1lHbGFGWmJRWnZubXMKN3ZoRmtzY3dHRlh4d21GMlBJZmU1Z3pNMDRBeVdjeTFIaVhLS2dNOXM3cGsxWUdyZGowZzdacmRBb0dCQUtLNApDR3MrbkRmMEZTMFJYOWFEWVJrRTdBNy9YUFhtSG5YMkRnU1h5N0Q4NTRPaWdTTWNoUmtPNTErbVNJejNQbllvCk41T1FXM2lHVVl1M1YvYmhnc0VSUzM1V2xmRk9BdDBzRUR5bjF5SVdXcDF5dG93d3BUNkVvUXVuZ2NYZjA5RjMKS1NROXowd3M4VmsvRWkvSFVXcU5LOWFXbU51cmFaT0ZqL2REK1ZkOUFvR0FMWFN3dEE3K043RDRkN0VEMURSRQpHTWdZNVd3OHFvdDZSdUNlNkpUY0FnU3B1MkhNU3JVY2dXclpiQnJZb09FUnVNQjFoMVJydk5ybU1qQlM0VW9FClgyZC8vbGhpOG1wL2VESWN3UDNRa2puanBJRFJWMFN1eWxrUkVaZURKZjVZb3R6eDdFdkJhbzFIbkQrWEg4eUIKVUtmWGJTaHZKVUdhRmgxT3Q1Y3JoM1k9Ci0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0K
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
	securityPolicyYAML = `---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: security-policy
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: backend
  cors:
    allowOrigins:
      - "https://www.example.com"
    allowMethods:
      - GET
      - POST
    allowHeaders:
      - "Content-Type"
`
	backendTrafficPolicyYAML = `---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: backend-traffic-policy
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: backend
  circuitBreaker:
    maxConnections: 100
    maxPendingRequests: 50
  loadBalancer:
    type: RoundRobin
`
	clientTrafficPolicyYAML = `---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: client-traffic-policy
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  timeout:
    http:
      requestReceivedTimeout: 30s
`
)

// Helpers for benchmark policy generation.
func genSecurityPolicies(n int) string {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		sb.WriteString(fmt.Sprintf(`---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: security-policy-%d
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: backend-%d
  cors:
    allowOrigins:
      - "https://www.example-%d.com"
    allowMethods:
      - GET
      - POST
    allowHeaders:
      - "Content-Type"
`, i, i, i))
	}
	return sb.String()
}

func genBackendTrafficPolicies(n int) string {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		sb.WriteString(fmt.Sprintf(`---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: backend-traffic-policy-%d
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: backend-%d
  circuitBreaker:
    maxConnections: %d
    maxPendingRequests: %d
  loadBalancer:
    type: RoundRobin
`, i, i, 100+i*10, 50+i*5))
	}
	return sb.String()
}

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
	medium := baseYAML + backendYAML + tlsSecretYAML + clientTrafficPolicyYAML +
		genHTTPRoutes(100) +
		genGRPCRoutes(50) +
		genUDPRoutes(20) +
		genSecurityPolicies(100) +
		genBackendTrafficPolicies(100)
	large := baseYAML + backendYAML + tlsSecretYAML + clientTrafficPolicyYAML +
		genHTTPRoutes(1000) +
		genGRPCRoutes(500) +
		genUDPRoutes(100) +
		genSecurityPolicies(1000) +
		genBackendTrafficPolicies(1000)

	cases := []benchCase{
		{
			name: "small",
			yaml: baseYAML + httpRouteYAML + backendYAML + tlsSecretYAML + securityPolicyYAML + backendTrafficPolicyYAML + clientTrafficPolicyYAML,
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
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err = egctl.TranslateGatewayAPIToXds("default", "cluster.local", "all", rs)
				if err != nil && strings.Contains(err.Error(), "failed to translate xds") {
					b.Fatalf("%v", err)
				}
			}
		})
	}
}
