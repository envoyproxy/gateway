// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package deprecated

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

func parse(t *testing.T, doc string) *yaml.Node {
	t.Helper()
	var n yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte(doc), &n))
	return n.Content[0]
}

func TestDetect(t *testing.T) {
	tests := []struct {
		name string
		kind string
		spec string
		want []string
	}{
		{
			name: "ClientTrafficPolicy",
			kind: "ClientTrafficPolicy",
			spec: `
targetRef:
  name: eg
enableProxyProtocol: false
headers:
  preserveXRequestID: true
tls:
  clientValidation:
    optional: true
`,
			want: []string{
				"spec.enableProxyProtocol",
				"spec.headers.preserveXRequestID",
				"spec.targetRef",
				"spec.tls.clientValidation.optional",
			},
		},
		{
			name: "ClientTrafficPolicy without deprecated fields",
			kind: "ClientTrafficPolicy",
			spec: `
targetRefs:
- name: eg
enableProxyProtocol: null
tls:
  clientValidation:
    optional: false
`,
		},
		{
			name: "BackendTrafficPolicy",
			kind: "BackendTrafficPolicy",
			spec: `
targetRef:
  name: eg
compression:
- type: Gzip
loadBalancer:
  type: ConsistentHash
  consistentHash:
    type: Header
    header:
      name: x-user
`,
			want: []string{
				"spec.compression",
				"spec.loadBalancer.consistentHash.header",
				"spec.loadBalancer.consistentHash.type: Header",
				"spec.targetRef",
			},
		},
		{
			name: "BackendTrafficPolicy with empty compression",
			kind: "BackendTrafficPolicy",
			spec: `
compression: []
`,
		},
		{
			name: "SecurityPolicy",
			kind: "SecurityPolicy",
			spec: `
extAuth:
  grpc:
    backendRef:
      name: authz
  http:
    backendRefs:
    - name: authz
    backendSettings:
      loadBalancer:
        consistentHash:
          type: Header
jwt:
  providers:
  - name: a
  - name: b
    remoteJWKS:
      backendRef:
        name: jwks
oidc:
  provider:
    backendRef:
      name: idp
`,
			want: []string{
				"spec.extAuth.grpc.backendRef",
				"spec.extAuth.http.backendSettings.loadBalancer.consistentHash.type: Header",
				"spec.jwt.providers[1].remoteJWKS.backendRef",
				"spec.oidc.provider.backendRef",
			},
		},
		{
			name: "EnvoyExtensionPolicy",
			kind: "EnvoyExtensionPolicy",
			spec: `
targetRef:
  name: eg
extProc:
- backendRef:
    name: proc
backends:
- name: a
- name: b
  backendSettings:
    loadBalancer:
      consistentHash:
        type: Headers
        header:
          name: x-user
`,
			want: []string{
				"spec.backends[1].backendSettings.loadBalancer.consistentHash.header",
				"spec.extProc[0].backendRef",
				"spec.targetRef",
			},
		},
		{
			name: "EnvoyProxy",
			kind: "EnvoyProxy",
			spec: `
luaValidation: Strict
telemetry:
  tracing:
    provider:
      host: otel.monitoring.svc
      port: 4317
  metrics:
    sinks:
    - type: OpenTelemetry
      openTelemetry:
        port: 4318
        backendRefs:
        - name: otel
    - type: OpenTelemetry
      openTelemetry:
        port: 0
        backendRefs:
        - name: otel
    - type: OpenTelemetry
      openTelemetry:
        # The CRD default, set on every sink read from a cluster.
        port: 4317
        backendRefs:
        - name: otel
  accessLog:
    settings:
    - sinks:
      - type: ALS
        als:
          backendRef:
            name: als
      - type: OpenTelemetry
        openTelemetry:
          backendRef:
            name: otel
`,
			want: []string{
				"spec.luaValidation",
				"spec.telemetry.accessLog.settings[0].sinks[0].als.backendRef",
				"spec.telemetry.accessLog.settings[0].sinks[1].openTelemetry.backendRef",
				"spec.telemetry.metrics.sinks[0].openTelemetry.port",
				"spec.telemetry.tracing.provider.host",
			},
		},
		{
			name: "unknown kind",
			kind: "HTTPRoute",
			spec: `
targetRef:
  name: eg
`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := parse(t, "spec:\n"+indent(tc.spec))
			got := slices.Sorted(maps.Keys(detect(tc.kind, root)))
			require.Equal(t, tc.want, emptyToNil(got))
		})
	}
}

func TestDetectEnvoyGateway(t *testing.T) {
	eg := parse(t, `
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
extensionApis:
  disableLua: false
extensionManager:
  service:
    host: ext.default.svc
    port: 5005
extensionManagers:
- name: a
  service:
    fqdn:
      hostname: ext.example.com
      port: 443
- name: b
  service:
    fqdn:
      hostname: ext.example.com
      port: 443
    port: 443
provider:
  custom:
    infrastructure:
      remote:
        service:
          port: 0
envoyProxy:
  luaValidation: Disabled
  telemetry:
    tracing:
      provider:
        port: 4317
`)
	warnings := detectEnvoyGateway(eg)
	got := make([]string, 0, len(warnings))
	for _, w := range warnings {
		got = append(got, w.path)
	}
	require.Equal(t, []string{
		"extensionApis.disableLua",
		"extensionManager.service.host",
		"extensionManagers[1].service.port",
		"envoyProxy.luaValidation",
		"envoyProxy.telemetry.tracing.provider.port",
	}, got)
}

func indent(s string) string {
	return "  " + strings.ReplaceAll(strings.Trim(s, "\n"), "\n", "\n  ") + "\n"
}

func emptyToNil(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	return s
}
