// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestAppProtocolToIRAppProtocol(t *testing.T) {
	tests := []struct {
		name            string
		appProtocol     string
		defaultProtocol ir.AppProtocol
		want            ir.AppProtocol
		wantForceHTTP1  bool
	}{
		{
			name:            "h2c service convention",
			appProtocol:     "kubernetes.io/h2c",
			defaultProtocol: ir.HTTP,
			want:            ir.HTTP2,
		},
		{
			name:            "h2c backend convention",
			appProtocol:     "gateway.envoyproxy.io/h2c",
			defaultProtocol: ir.HTTP,
			want:            ir.HTTP2,
		},
		{
			name:            "ws service convention",
			appProtocol:     "kubernetes.io/ws",
			defaultProtocol: ir.HTTP,
			want:            ir.HTTP,
			wantForceHTTP1:  true,
		},
		{
			name:            "wss service convention",
			appProtocol:     "kubernetes.io/wss",
			defaultProtocol: ir.HTTP,
			want:            ir.HTTP,
			wantForceHTTP1:  true,
		},
		{
			name:            "ws backend convention",
			appProtocol:     "gateway.envoyproxy.io/ws",
			defaultProtocol: ir.HTTP,
			want:            ir.HTTP,
			wantForceHTTP1:  true,
		},
		{
			name:            "wss backend convention",
			appProtocol:     "gateway.envoyproxy.io/wss",
			defaultProtocol: ir.HTTP,
			want:            ir.HTTP,
			wantForceHTTP1:  true,
		},
		{
			name:            "grpc",
			appProtocol:     "grpc",
			defaultProtocol: ir.HTTP,
			want:            ir.GRPC,
		},
		{
			name:            "unknown",
			appProtocol:     "example.com/custom",
			defaultProtocol: ir.HTTP,
			want:            ir.HTTP,
		},
		{
			// appProtocol must not refine the protocol of non-HTTP (L4) routes.
			name:            "h2c ignored on non-HTTP route",
			appProtocol:     "kubernetes.io/h2c",
			defaultProtocol: ir.TCP,
			want:            ir.TCP,
		},
		{
			name:            "grpc ignored on non-HTTP route",
			appProtocol:     "grpc",
			defaultProtocol: ir.TCP,
			want:            ir.TCP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protocol := resolveBackendProtocol(tt.appProtocol, tt.defaultProtocol)
			require.Equal(t, tt.want, protocol)
			ap := tt.appProtocol
			require.Equal(t, tt.wantForceHTTP1, shouldForceHTTP1Upstream(protocol, &ap))
		})
	}
}

func TestStableClusterRouteDestinationReusesSameTLSBackend(t *testing.T) {
	weight := uint32(1)
	addressType := ir.IP
	sni := "backend.example.com"
	metadata := &ir.ResourceMetadata{
		Kind:      "Service",
		Name:      "backend",
		Namespace: "default",
	}
	settings := []*ir.DestinationSetting{
		{
			Name:        "httproute/default/route-a/rule/0/backend/0",
			Weight:      &weight,
			Protocol:    ir.HTTP,
			AddressType: &addressType,
			Endpoints: []*ir.DestinationEndpoint{
				ir.NewDestEndpoint(nil, "10.0.0.1", 8080, false, nil),
			},
			TLS: &ir.TLSUpstreamConfig{
				SNI: &sni,
				CACertificate: &ir.TLSCACertificate{
					Name:        "backend-ca",
					Certificate: []byte("ca-bundle"),
				},
			},
			Metadata: metadata,
		},
	}

	routeMetadata := &ir.ResourceMetadata{
		Kind:      "HTTPRoute",
		Name:      "route-a",
		Namespace: "default",
	}
	otherRouteMetadata := &ir.ResourceMetadata{
		Kind:      "HTTPRoute",
		Name:      "route-b",
		Namespace: "default",
	}
	first := stableClusterRouteDestination(resource.KindHTTPRoute, "httproute/default/route-a/rule/0", settings, nil, nil, routeMetadata, stableClusterRouteDestinationOptions{})
	second := stableClusterRouteDestination(resource.KindHTTPRoute, "httproute/default/route-b/rule/0", settings, nil, nil, otherRouteMetadata, stableClusterRouteDestinationOptions{})
	routeOnlyTraffic := &ir.TrafficFeatures{
		Retry: &ir.Retry{
			NumRetries: new(uint32),
		},
	}
	third := stableClusterRouteDestination(resource.KindHTTPRoute, "httproute/default/route-c/rule/0", settings, routeOnlyTraffic, nil, routeMetadata, stableClusterRouteDestinationOptions{})

	require.Equal(t, first.Name, second.Name)
	require.Equal(t, first.Name, third.Name)
	require.Equal(t, first.Settings[0].Name, second.Settings[0].Name)
	require.NotEqual(t, "httproute/default/route-a/rule/0", first.Name)
	require.Equal(t, settings[0].Name, "httproute/default/route-a/rule/0/backend/0")
}

func TestStableClusterRouteDestinationUsesXDSClusterInputs(t *testing.T) {
	weight := uint32(1)
	addressType := ir.IP
	settings := []*ir.DestinationSetting{
		{
			Name:        "httproute/default/route-a/rule/0/backend/0",
			Weight:      &weight,
			Protocol:    ir.HTTP,
			AddressType: &addressType,
			Endpoints: []*ir.DestinationEndpoint{
				ir.NewDestEndpoint(nil, "10.0.0.1", 8080, false, nil),
			},
			TLS: &ir.TLSUpstreamConfig{
				CACertificate: &ir.TLSCACertificate{
					Name:        "backend-ca",
					Certificate: []byte("ca-bundle"),
				},
			},
			Metadata: &ir.ResourceMetadata{
				Kind:      "Service",
				Name:      "backend",
				Namespace: "default",
			},
		},
	}

	routeAMetadata := &ir.ResourceMetadata{Kind: "HTTPRoute", Name: "route-a", Namespace: "default"}
	routeBMetadata := &ir.ResourceMetadata{Kind: "HTTPRoute", Name: "route-b", Namespace: "default"}
	routeAStatName := "route-a-cluster"
	routeBStatName := "route-b-cluster"

	routeMetadataOnly := stableClusterRouteDestination(resource.KindHTTPRoute, "httproute/default/route-b/rule/0", settings, nil, nil, routeBMetadata, stableClusterRouteDestinationOptions{})
	sameRouteMetadataOnly := stableClusterRouteDestination(resource.KindHTTPRoute, "httproute/default/route-a/rule/0", settings, nil, nil, routeAMetadata, stableClusterRouteDestinationOptions{})
	first := stableClusterRouteDestination(resource.KindHTTPRoute, "httproute/default/route-a/rule/0", settings, nil, nil, routeAMetadata, stableClusterRouteDestinationOptions{
		StatName: &routeAStatName,
	})
	second := stableClusterRouteDestination(resource.KindHTTPRoute, "httproute/default/route-b/rule/0", settings, nil, nil, routeBMetadata, stableClusterRouteDestinationOptions{
		StatName: &routeBStatName,
	})
	withExtensionRef := stableClusterRouteDestination(resource.KindHTTPRoute, "httproute/default/route-c/rule/0", settings, nil, nil, routeAMetadata, stableClusterRouteDestinationOptions{
		HasExtensionRefs: true,
	})

	require.Equal(t, sameRouteMetadataOnly.Name, routeMetadataOnly.Name)
	require.NotEqual(t, first.Name, second.Name)
	require.Equal(t, "httproute/default/route-c/rule/0", withExtensionRef.Name)
	require.Equal(t, "httproute/default/route-a/rule/0/backend/0", settings[0].Name)

	healthCheckTraffic := &ir.TrafficFeatures{
		HealthCheck: &ir.HealthCheck{
			Active: &ir.ActiveHealthCheck{
				HTTP: &ir.HTTPHealthChecker{},
			},
		},
	}
	routeHostA := stableClusterRouteDestination(resource.KindHTTPRoute, "httproute/default/route-a/rule/0", settings, healthCheckTraffic, nil, routeAMetadata, stableClusterRouteDestinationOptions{
		RouteHostname: "route-a.example.com",
	})
	routeHostB := stableClusterRouteDestination(resource.KindHTTPRoute, "httproute/default/route-a/rule/0", settings, healthCheckTraffic, nil, routeAMetadata, stableClusterRouteDestinationOptions{
		RouteHostname: "route-b.example.com",
	})
	require.NotEqual(t, routeHostA.Name, routeHostB.Name)
}

func TestGetIREndpointsFromEndpointSlices(t *testing.T) {
	tests := []struct {
		name              string
		endpointSlices    []*discoveryv1.EndpointSlice
		portName          string
		portProtocol      corev1.Protocol
		expectedEndpoints []*ir.DestinationEndpoint
		expectedAddrType  ir.DestinationAddressType
	}{
		{
			name: "All IP endpoints",
			endpointSlices: []*discoveryv1.EndpointSlice{
				{
					ObjectMeta:  metav1.ObjectMeta{Name: "slice1"},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"192.0.2.1"}},
						{Addresses: []string{"192.0.2.2"}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: new("http"), Port: new(int32(80)), Protocol: new(corev1.ProtocolTCP)},
					},
				},
				{
					ObjectMeta:  metav1.ObjectMeta{Name: "slice2"},
					AddressType: discoveryv1.AddressTypeIPv6,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"2001:db8::1"}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: new("http"), Port: new(int32(80)), Protocol: new(corev1.ProtocolTCP)},
					},
				},
			},
			portName:     "http",
			portProtocol: corev1.ProtocolTCP,
			expectedEndpoints: []*ir.DestinationEndpoint{
				{Host: "192.0.2.1", Port: 80, Draining: false},
				{Host: "192.0.2.2", Port: 80, Draining: false},
				{Host: "2001:db8::1", Port: 80, Draining: false},
			},
			expectedAddrType: ir.IP,
		},
		{
			name: "Mixed IP and FQDN endpoints",
			endpointSlices: []*discoveryv1.EndpointSlice{
				{
					ObjectMeta:  metav1.ObjectMeta{Name: "slice1"},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"192.0.2.1"}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: new("http"), Port: new(int32(80)), Protocol: new(corev1.ProtocolTCP)},
					},
				},
				{
					ObjectMeta:  metav1.ObjectMeta{Name: "slice2"},
					AddressType: discoveryv1.AddressTypeFQDN,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"example.com"}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: new("http"), Port: new(int32(80)), Protocol: new(corev1.ProtocolTCP)},
					},
				},
			},
			portName:     "http",
			portProtocol: corev1.ProtocolTCP,
			expectedEndpoints: []*ir.DestinationEndpoint{
				{Host: "192.0.2.1", Port: 80, Draining: false},
				{Host: "example.com", Port: 80, Draining: false},
			},
			expectedAddrType: ir.MIXED,
		},
		{
			name: "Dual-stack IP endpoints",
			endpointSlices: []*discoveryv1.EndpointSlice{
				{
					ObjectMeta:  metav1.ObjectMeta{Name: "slice1-ipv4"},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"192.0.2.1"}},
						{Addresses: []string{"192.0.2.2"}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: new("http"), Port: new(int32(80)), Protocol: new(corev1.ProtocolTCP)},
					},
				},
				{
					ObjectMeta:  metav1.ObjectMeta{Name: "slice2-ipv6"},
					AddressType: discoveryv1.AddressTypeIPv6,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"2001:db8::1"}},
						{Addresses: []string{"2001:db8::2"}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: new("http"), Port: new(int32(80)), Protocol: new(corev1.ProtocolTCP)},
					},
				},
			},
			portName:     "http",
			portProtocol: corev1.ProtocolTCP,
			expectedEndpoints: []*ir.DestinationEndpoint{
				{Host: "192.0.2.1", Port: 80, Draining: false},
				{Host: "192.0.2.2", Port: 80, Draining: false},
				{Host: "2001:db8::1", Port: 80, Draining: false},
				{Host: "2001:db8::2", Port: 80, Draining: false},
			},
			expectedAddrType: ir.IP,
		},
		{
			name: "Dual-stack with FQDN",
			endpointSlices: []*discoveryv1.EndpointSlice{
				{
					ObjectMeta:  metav1.ObjectMeta{Name: "slice1-ipv4"},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"192.0.2.1"}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: new("http"), Port: new(int32(80)), Protocol: new(corev1.ProtocolTCP)},
					},
				},
				{
					ObjectMeta:  metav1.ObjectMeta{Name: "slice2-ipv6"},
					AddressType: discoveryv1.AddressTypeIPv6,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"2001:db8::1"}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: new("http"), Port: new(int32(80)), Protocol: new(corev1.ProtocolTCP)},
					},
				},
				{
					ObjectMeta:  metav1.ObjectMeta{Name: "slice3-fqdn"},
					AddressType: discoveryv1.AddressTypeFQDN,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"example.com"}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: new("http"), Port: new(int32(80)), Protocol: new(corev1.ProtocolTCP)},
					},
				},
			},
			portName:     "http",
			portProtocol: corev1.ProtocolTCP,
			expectedEndpoints: []*ir.DestinationEndpoint{
				{Host: "192.0.2.1", Port: 80, Draining: false},
				{Host: "2001:db8::1", Port: 80, Draining: false},
				{Host: "example.com", Port: 80, Draining: false},
			},
			expectedAddrType: ir.MIXED,
		},
		{
			name: "Keep non-serving or terminating as draining",
			endpointSlices: []*discoveryv1.EndpointSlice{
				{
					ObjectMeta:  metav1.ObjectMeta{Name: "slice1"},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"192.0.2.1"}, Conditions: discoveryv1.EndpointConditions{
							Ready: new(false), Serving: new(true), Terminating: new(true),
						}},
						{Addresses: []string{"192.0.2.2"}, Conditions: discoveryv1.EndpointConditions{
							Ready: new(false), Serving: new(false), Terminating: new(true),
						}},
						{Addresses: []string{"192.0.2.3"}, Conditions: discoveryv1.EndpointConditions{
							Ready: new(false),
						}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: new("http"), Port: new(int32(80)), Protocol: new(corev1.ProtocolTCP)},
					},
				},
			},
			portName:     "http",
			portProtocol: corev1.ProtocolTCP,
			expectedEndpoints: []*ir.DestinationEndpoint{
				{Host: "192.0.2.1", Port: 80, Draining: true},
				{Host: "192.0.2.2", Port: 80, Draining: true},
				{Host: "192.0.2.3", Port: 80, Draining: true},
			},
			expectedAddrType: ir.IP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoints, addrType := getIREndpointsFromEndpointSlices(tt.endpointSlices, tt.portName, tt.portProtocol)

			fmt.Printf("Test case: %s\n", tt.name)
			fmt.Printf("Number of endpoints: %d\n", len(endpoints))
			fmt.Printf("Address type: %v\n", *addrType)

			fmt.Println("Actual endpoints:")
			for i, endpoint := range endpoints {
				fmt.Printf("  Endpoint %d:\n", i+1)
				fmt.Printf("    Address: %s\n", endpoint.Host)
				fmt.Printf("    Port: %d\n", endpoint.Port)
				fmt.Printf("    Draining: %t\n", endpoint.Draining)

			}

			fmt.Println()
			require.Equal(t, tt.expectedEndpoints, endpoints)
			require.Equal(t, tt.expectedAddrType, *addrType)
		})
	}
}

func TestBuildRouteMatchCombinations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		ruleMatches   []gwapiv1.HTTPRouteMatch
		filterMatches []egv1a1.HTTPRouteMatchFilter
		expected      []routeMatchCombination
	}{
		{
			name:     "no rule or filter matches",
			expected: nil,
		},
		{
			name: "filter matches only",
			filterMatches: []egv1a1.HTTPRouteMatchFilter{
				{Cookies: []egv1a1.HTTPCookieMatch{{Name: "a", Value: "1"}}},
				{Cookies: []egv1a1.HTTPCookieMatch{{Name: "b", Value: "2"}}},
			},
			expected: []routeMatchCombination{
				{
					cookies: []egv1a1.HTTPCookieMatch{{Name: "a", Value: "1"}},
				},
				{
					cookies: []egv1a1.HTTPCookieMatch{{Name: "b", Value: "2"}},
				},
			},
		},
		{
			name: "rule matches only",
			ruleMatches: []gwapiv1.HTTPRouteMatch{
				{Path: &gwapiv1.HTTPPathMatch{Value: new("/foo")}},
				{Path: &gwapiv1.HTTPPathMatch{Value: new("/bar")}},
			},
			expected: []routeMatchCombination{
				{HTTPRouteMatch: gwapiv1.HTTPRouteMatch{Path: &gwapiv1.HTTPPathMatch{Value: new("/foo")}}},
				{HTTPRouteMatch: gwapiv1.HTTPRouteMatch{Path: &gwapiv1.HTTPPathMatch{Value: new("/bar")}}},
			},
		},
		{
			name: "rule and filter matches",
			ruleMatches: []gwapiv1.HTTPRouteMatch{
				{Path: &gwapiv1.HTTPPathMatch{Value: new("/foo")}},
				{
					Path: &gwapiv1.HTTPPathMatch{Value: new("/bar")},
					Headers: []gwapiv1.HTTPHeaderMatch{
						{Name: "a", Value: "1"},
						{Name: "b", Value: "2"},
						{Name: "c", Value: "3"},
					},
				},
			},
			filterMatches: []egv1a1.HTTPRouteMatchFilter{
				{Cookies: []egv1a1.HTTPCookieMatch{{Name: "a", Value: "1"}}},
				{Cookies: []egv1a1.HTTPCookieMatch{{Name: "b", Value: "2"}, {Name: "c", Value: "3"}}},
			},
			expected: []routeMatchCombination{
				{
					HTTPRouteMatch: gwapiv1.HTTPRouteMatch{Path: &gwapiv1.HTTPPathMatch{Value: new("/foo")}},
					cookies:        []egv1a1.HTTPCookieMatch{{Name: "a", Value: "1"}},
				},
				{
					HTTPRouteMatch: gwapiv1.HTTPRouteMatch{Path: &gwapiv1.HTTPPathMatch{Value: new("/foo")}},
					cookies:        []egv1a1.HTTPCookieMatch{{Name: "b", Value: "2"}, {Name: "c", Value: "3"}},
				},
				{
					HTTPRouteMatch: gwapiv1.HTTPRouteMatch{
						Path: &gwapiv1.HTTPPathMatch{Value: new("/bar")},
						Headers: []gwapiv1.HTTPHeaderMatch{
							{Name: "a", Value: "1"},
							{Name: "b", Value: "2"},
							{Name: "c", Value: "3"},
						},
					},
					cookies: []egv1a1.HTTPCookieMatch{{Name: "a", Value: "1"}},
				},
				{
					HTTPRouteMatch: gwapiv1.HTTPRouteMatch{
						Path: &gwapiv1.HTTPPathMatch{Value: new("/bar")},
						Headers: []gwapiv1.HTTPHeaderMatch{
							{Name: "a", Value: "1"},
							{Name: "b", Value: "2"},
							{Name: "c", Value: "3"},
						},
					},
					cookies: []egv1a1.HTTPCookieMatch{{Name: "b", Value: "2"}, {Name: "c", Value: "3"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			combos := buildRouteMatchCombinations(tt.ruleMatches, tt.filterMatches)
			require.Equal(t, tt.expected, combos)
		})
	}
}

func TestValidateDestinationSettings(t *testing.T) {
	svcKind := gwapiv1.Kind(resource.KindService)
	hostname := "www.gateway-test.com"

	tests := []struct {
		name                    string
		ds                      *ir.DestinationSetting
		endpointRoutingDisabled bool
		kind                    *gwapiv1.Kind
		wantErr                 bool
		wantReason              gwapiv1.RouteConditionReason
	}{
		{
			name: "normal service allowed with ClusterIP routing",
			ds: &ir.DestinationSetting{
				Name:      "normal",
				Endpoints: []*ir.DestinationEndpoint{{Host: "10.0.0.1"}},
			},
			endpointRoutingDisabled: true,
			kind:                    &svcKind,
			wantErr:                 false,
		},
		{
			name: "normal service allowed with hostname",
			ds: &ir.DestinationSetting{
				Name:      "normal with hostname",
				Endpoints: []*ir.DestinationEndpoint{{Hostname: &hostname, Host: "10.0.0.1"}},
			},
			endpointRoutingDisabled: true,
			kind:                    &svcKind,
			wantErr:                 false,
		},
		{
			name: "mixed address type rejected when EndpointSlice routing",
			ds: &ir.DestinationSetting{
				Name:        "mixed",
				Endpoints:   []*ir.DestinationEndpoint{{Host: "10.0.0.1"}},
				AddressType: new(ir.MIXED),
			},
			endpointRoutingDisabled: false,
			kind:                    &svcKind,
			wantErr:                 true,
			wantReason:              status.RouteReasonUnsupportedAddressType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDestinationSettings(tt.ds, tt.endpointRoutingDisabled, tt.kind)
			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.wantReason, err.Reason())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIsServiceHeadless(t *testing.T) {
	tests := []struct {
		name    string
		service *corev1.Service
		want    bool
	}{
		{
			name: "headless service with ClusterIP None",
			service: &corev1.Service{
				Spec: corev1.ServiceSpec{
					ClusterIP: "None",
				},
			},
			want: true,
		},
		{
			name: "normal service with ClusterIP",
			service: &corev1.Service{
				Spec: corev1.ServiceSpec{
					ClusterIP: "10.0.0.1",
				},
			},
			want: false,
		},
		{
			name: "dual-stack headless service",
			service: &corev1.Service{
				Spec: corev1.ServiceSpec{
					ClusterIP:  "None",
					ClusterIPs: []string{"None", "None"},
				},
			},
			want: true,
		},
		{
			name: "dual-stack service with valid IPs",
			service: &corev1.Service{
				Spec: corev1.ServiceSpec{
					ClusterIP:  "10.0.0.1",
					ClusterIPs: []string{"10.0.0.1", "2001:db8::1"},
				},
			},
			want: false,
		},
		{
			name:    "nil service",
			service: nil,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isServiceHeadless(tt.service)
			require.Equal(t, tt.want, got)
		})
	}
}
