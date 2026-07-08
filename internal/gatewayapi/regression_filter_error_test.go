// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
package gatewayapi

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
)

// Repro for regression: any filter error (invalid header value) now injects a 500 DirectResponse
// even when the rule already configured a RequestRedirect.
func TestRequestRedirectClobberedByFilterError(t *testing.T) {
	t.Helper()

	resourcesYAML := `
gateways:
- apiVersion: gateway.networking.k8s.io/v1
  kind: Gateway
  metadata:
    namespace: envoy-gateway
    name: gateway-1
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
    - name: http
      protocol: HTTP
      port: 80
      hostname: "*.envoyproxy.io"
      allowedRoutes:
        namespaces:
          from: All
httpRoutes:
- apiVersion: gateway.networking.k8s.io/v1
  kind: HTTPRoute
  metadata:
    namespace: default
    name: redirect-with-invalid-header-filter
  spec:
    hostnames:
    - gateway.envoyproxy.io
    parentRefs:
    - namespace: envoy-gateway
      name: gateway-1
      sectionName: http
    rules:
    - matches:
      - path:
          type: PathPrefix
          value: /
      backendRefs:
      - name: service-1
        port: 8080
      filters:
      - type: RequestRedirect
        requestRedirect:
          scheme: https
          statusCode: 302
      - type: ResponseHeaderModifier
        responseHeaderModifier:
          set:
          - name: bad-header
            value: |
              invalid
              multiline
`

resources := &resource.Resources{}
require.NoError(t, yaml.UnmarshalStrict([]byte(resourcesYAML), resources))

	// Namespaces referenced by Gateway/Route/Services.
	resources.Namespaces = []*corev1.Namespace{
		{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "envoy-gateway"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "envoy-gateway-system"}},
	}

	resources.GatewayClass = &gwapiv1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "envoy-gateway-class",
		},
		Spec: gwapiv1.GatewayClassSpec{
			ControllerName: gwapiv1.GatewayController(egv1a1.GatewayControllerName),
		},
	}

	// Add minimal envoy TLS secret expected by translator.
	resources.Secrets = append(resources.Secrets, &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "envoy-gateway-system",
			Name:      "envoy",
		},
	})

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "service-1",
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "1.1.1.1",
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
	resources.Services = append(resources.Services, svc)

	endptSlice := &discoveryv1.EndpointSlice{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "discovery.k8s.io/v1",
			Kind:       "EndpointSlice",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "endpointslice-1",
			Namespace: "default",
			Labels: map[string]string{
				discoveryv1.LabelServiceName: svc.Name,
			},
		},
		AddressType: discoveryv1.AddressTypeIPv4,
		Ports: []discoveryv1.EndpointPort{
			{
				Name:     ptr.To("http"),
				Port:     ptr.To[int32](8080),
				Protocol: ptr.To(corev1.ProtocolTCP),
			},
		},
		Endpoints: []discoveryv1.Endpoint{
			{
				Addresses: []string{"7.7.7.7"},
				Conditions: discoveryv1.EndpointConditions{
					Ready: ptr.To(true),
				},
			},
		},
	}
	resources.EndpointSlices = append(resources.EndpointSlices, endptSlice)

	translator := &Translator{
		GatewayControllerName:   egv1a1.GatewayControllerName,
		GatewayClassName:        "envoy-gateway-class",
		GlobalRateLimitEnabled:  true,
		EnvoyPatchPolicyEnabled: true,
		BackendEnabled:          true,
		ControllerNamespace:     "envoy-gateway-system",
		MergeGateways:           IsMergeGatewaysEnabled(resources),
		GatewayNamespaceMode:    false,
		WasmCache:               &mockWasmCache{},
		Logger:                  logging.DefaultLogger(io.Discard, egv1a1.LogLevelInfo),
	}

	accepted, failed := translator.GetRelevantGateways(resources)
	require.Len(t, failed, 0)
	require.NotEmpty(t, accepted, "no accepted gateways")

	result, err := translator.Translate(resources)
	require.NoError(t, err)

	// Ensure the IR route still contains the request redirect but now also has a 500 direct response injected.
	var (
		target     *ir.HTTPRoute
		routes     []*ir.HTTPRoute
		routeNames []string
	)
	for _, xds := range result.XdsIR {
		for _, httpListener := range xds.HTTP {
			for _, route := range httpListener.Routes {
				routes = append(routes, route)
				routeNames = append(routeNames, route.Name)
				if strings.Contains(route.Name, "redirect-with-invalid-header-filter") {
					target = route
				}
			}
		}
	}
	if target == nil && len(routes) > 0 {
		target = routes[0]
	}
	require.NotNilf(t, target, "route not found (routes seen: %v)", routeNames)
	require.NotNil(t, target.Redirect, "redirect should be configured")
	require.NotNil(t, target.DirectResponse, "unexpected 500 direct response now present due to filter error")
	require.Equal(t, uint32(500), *target.DirectResponse.StatusCode, "direct response status")
}
