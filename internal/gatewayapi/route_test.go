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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
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

func TestResolveBackendCluster(t *testing.T) {
	newIdentity := func() *BackendClusterKey {
		return &BackendClusterKey{Kind: "Service", Namespace: "default", Name: "service-1", Port: 8080}
	}
	serviceBackendRef := gwapiv1.BackendObjectReference{Name: "service-1"}
	emptyDS := &ir.DestinationSetting{}
	newParentRef := func(sectionName string, port int32) *RouteParentContext {
		pr := &RouteParentContext{ParentReference: &gwapiv1.ParentReference{}}
		if sectionName != "" {
			pr.SectionName = SectionNamePtr(sectionName)
		}
		if port != 0 {
			pr.Port = new(port)
		}
		return pr
	}

	t.Run("nil gatewayCtx never merges", func(t *testing.T) {
		tr := &Translator{MergeBackends: true, TranslatorContext: &TranslatorContext{}}
		cluster := tr.resolveBackendCluster("route-scoped-name", nil, newParentRef("", 0), nil, false, newIdentity(), serviceBackendRef, "default", emptyDS)
		require.False(t, cluster.Merge)
		require.Equal(t, "route-scoped-name", cluster.Name)
		require.Equal(t, &BackendClusterKey{Name: "route-scoped-name"}, cluster.Key)
	})

	t.Run("merge disabled falls back to gateway-scoped route name", func(t *testing.T) {
		tr := &Translator{MergeBackends: false, TranslatorContext: &TranslatorContext{}}
		gwCtx := &GatewayContext{Gateway: &gwapiv1.Gateway{}}
		cluster := tr.resolveBackendCluster("route-scoped-name", gwCtx, newParentRef("", 0), nil, false, newIdentity(), serviceBackendRef, "default", emptyDS)
		require.False(t, cluster.Merge)
		require.Equal(t, "route-scoped-name", cluster.Name)
		require.Equal(t, &BackendClusterKey{GatewayIRKey: tr.getIRKey(gwCtx.Gateway), Name: "route-scoped-name"}, cluster.Key)
	})

	t.Run("route-scoped key differs across gateways for the same rule (multi-parent route)", func(t *testing.T) {
		tr := &Translator{MergeBackends: false, TranslatorContext: &TranslatorContext{}}
		gwCtx1 := &GatewayContext{Gateway: &gwapiv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: "envoy-gateway", Name: "gateway-1"}}}
		gwCtx2 := &GatewayContext{Gateway: &gwapiv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: "envoy-gateway", Name: "gateway-2"}}}
		cluster1 := tr.resolveBackendCluster("httproute/default/httproute-1/rule/0", gwCtx1, newParentRef("", 0), nil, false, newIdentity(), serviceBackendRef, "default", emptyDS)
		cluster2 := tr.resolveBackendCluster("httproute/default/httproute-1/rule/0", gwCtx2, newParentRef("", 0), nil, false, newIdentity(), serviceBackendRef, "default", emptyDS)
		require.NotEqual(t, cluster1.Key, cluster2.Key, "the same route rule processed under two different parent gateways must not collide in BackendClusterMap")
	})

	t.Run("route-scoped key differs across parentRefs on the same gateway (multi-listener route)", func(t *testing.T) {
		tr := &Translator{MergeBackends: false, TranslatorContext: &TranslatorContext{}}
		gwCtx := &GatewayContext{Gateway: &gwapiv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: "envoy-gateway", Name: "gateway-1"}}}
		cluster1 := tr.resolveBackendCluster("httproute/default/httproute-1/rule/0", gwCtx, newParentRef("http-a", 0), nil, false, newIdentity(), serviceBackendRef, "default", emptyDS)
		cluster2 := tr.resolveBackendCluster("httproute/default/httproute-1/rule/0", gwCtx, newParentRef("http-b", 0), nil, false, newIdentity(), serviceBackendRef, "default", emptyDS)
		require.NotEqual(t, cluster1.Key, cluster2.Key, "the same rule attached to two listeners on one gateway must not collide in BackendClusterMap")
	})

	t.Run("route-scoped key across all combinations of sectionName/parentPort presence", func(t *testing.T) {
		tests := []struct {
			name      string
			section1  string
			port1     int32
			section2  string
			port2     int32
			wantEqual bool
		}{
			{"neither set, repeated", "", 0, "", 0, true},
			{"section only, repeated", "http-a", 0, "http-a", 0, true},
			{"port only, repeated", "", 8080, "", 8080, true},
			{"both set, repeated", "http-a", 8080, "http-a", 8080, true},
			{"neither vs section only", "", 0, "http-a", 0, false},
			{"neither vs port only", "", 0, "", 8080, false},
			{"neither vs both", "", 0, "http-a", 8080, false},
			{"section only vs port only", "http-a", 0, "", 8080, false},
			{"section only vs both (port differs)", "http-a", 0, "http-a", 8080, false},
			{"port only vs both (section differs)", "", 8080, "http-a", 8080, false},
			{"both vs both, different section", "http-a", 8080, "http-b", 8080, false},
			{"both vs both, different port", "http-a", 8080, "http-a", 9090, false},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				tr := &Translator{MergeBackends: false, TranslatorContext: &TranslatorContext{}}
				gwCtx := &GatewayContext{Gateway: &gwapiv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: "envoy-gateway", Name: "gateway-1"}}}
				cluster1 := tr.resolveBackendCluster("httproute/default/httproute-1/rule/0", gwCtx, newParentRef(tc.section1, tc.port1), nil, false, newIdentity(), serviceBackendRef, "default", emptyDS)
				cluster2 := tr.resolveBackendCluster("httproute/default/httproute-1/rule/0", gwCtx, newParentRef(tc.section2, tc.port2), nil, false, newIdentity(), serviceBackendRef, "default", emptyDS)
				if tc.wantEqual {
					require.Equal(t, cluster1.Key, cluster2.Key)
				} else {
					require.NotEqual(t, cluster1.Key, cluster2.Key)
				}
			})
		}
	})

	t.Run("merge enabled resolves to backend-identity name", func(t *testing.T) {
		tr := &Translator{MergeBackends: true, TranslatorContext: &TranslatorContext{}}
		gwCtx := &GatewayContext{Gateway: &gwapiv1.Gateway{}}
		identity := newIdentity()
		cluster := tr.resolveBackendCluster("route-scoped-name", gwCtx, newParentRef("", 0), nil, false, identity, serviceBackendRef, "default", emptyDS)
		require.True(t, cluster.Merge)
		require.Equal(t, "backend/service/default/service-1/8080", cluster.Name)
		require.Equal(t, identity.Kind, cluster.Key.Kind)
		require.Equal(t, identity.Name, cluster.Key.Name)
	})

	t.Run("merge-incompatible excludes even when routing type matches", func(t *testing.T) {
		tr := &Translator{MergeBackends: true, TranslatorContext: &TranslatorContext{}}
		gwCtx := &GatewayContext{Gateway: &gwapiv1.Gateway{}}
		cluster := tr.resolveBackendCluster("route-scoped-name", gwCtx, newParentRef("", 0), nil, true, newIdentity(), serviceBackendRef, "default", emptyDS)
		require.False(t, cluster.Merge)
		require.Equal(t, "route-scoped-name", cluster.Name)
		require.Equal(t, &BackendClusterKey{GatewayIRKey: tr.getIRKey(gwCtx.Gateway), Name: "route-scoped-name"}, cluster.Key)
	})

	t.Run("dynamic resolver backend never merges", func(t *testing.T) {
		dynamicResolverType := egv1a1.BackendTypeDynamicResolver
		dynamicBackendRef := gwapiv1.BackendObjectReference{
			Group: GroupPtr(egv1a1.GroupName),
			Kind:  KindPtr(egv1a1.KindBackend),
			Name:  "be-dynamic",
		}
		backendMap := map[types.NamespacedName]*egv1a1.Backend{
			{Namespace: "default", Name: "be-dynamic"}: {
				ObjectMeta: metav1.ObjectMeta{Name: "be-dynamic", Namespace: "default"},
				Spec:       egv1a1.BackendSpec{Type: &dynamicResolverType},
			},
		}
		tr := &Translator{MergeBackends: true, TranslatorContext: &TranslatorContext{BackendMap: backendMap}}
		gwCtx := &GatewayContext{Gateway: &gwapiv1.Gateway{}}
		cluster := tr.resolveBackendCluster("route-scoped-name", gwCtx, newParentRef("", 0), nil, false, newIdentity(), dynamicBackendRef, "default", emptyDS)
		require.False(t, cluster.Merge)
		require.Equal(t, "route-scoped-name", cluster.Name)
		require.Equal(t, &BackendClusterKey{GatewayIRKey: tr.getIRKey(gwCtx.Gateway), Name: "route-scoped-name"}, cluster.Key)
	})
}

func TestShouldMergeBackend(t *testing.T) {
	gwNN := types.NamespacedName{Namespace: "envoy-gateway", Name: "gateway-1"}
	gwCtx := &GatewayContext{Gateway: &gwapiv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: gwNN.Namespace, Name: gwNN.Name}}}
	serviceRT := egv1a1.ServiceRoutingType
	endpointRT := egv1a1.EndpointRoutingType
	dynamicResolverType := egv1a1.BackendTypeDynamicResolver

	serviceBackendRef := gwapiv1.BackendObjectReference{Name: "service-1"}
	dynamicResolverBackendRef := gwapiv1.BackendObjectReference{
		Group: GroupPtr(egv1a1.GroupName),
		Kind:  KindPtr(egv1a1.KindBackend),
		Name:  "be-dynamic",
	}
	dynamicResolverBackend := &egv1a1.Backend{
		ObjectMeta: metav1.ObjectMeta{Name: "be-dynamic", Namespace: "default"},
		Spec:       egv1a1.BackendSpec{Type: &dynamicResolverType},
	}

	tests := []struct {
		name              string
		mergeEnabled      bool
		gatewayBaselineRT *egv1a1.RoutingType
		effectiveRT       *egv1a1.RoutingType
		mergeIncompatible bool
		backendRef        gwapiv1.BackendObjectReference
		backend           *egv1a1.Backend
		filters           *ir.DestinationFilters
		want              bool
	}{
		{
			name:         "disabled globally never merges",
			mergeEnabled: false,
			backendRef:   serviceBackendRef,
			want:         false,
		},
		{
			name:         "enabled, no routing type anywhere: baseline == effective (both Endpoint)",
			mergeEnabled: true,
			backendRef:   serviceBackendRef,
			want:         true,
		},
		{
			name:              "enabled, uniform gateway-level routing type: baseline == effective",
			mergeEnabled:      true,
			gatewayBaselineRT: &serviceRT,
			effectiveRT:       &serviceRT,
			backendRef:        serviceBackendRef,
			want:              true,
		},
		{
			name:              "enabled, route-rule overrides routing type away from gateway baseline: diverges",
			mergeEnabled:      true,
			gatewayBaselineRT: &endpointRT,
			effectiveRT:       &serviceRT,
			backendRef:        serviceBackendRef,
			want:              false,
		},
		{
			name:              "enabled, uniform routing but marked merge-incompatible (route-level cluster settings, session persistence, ConsistentHash, or fallback): excluded",
			mergeEnabled:      true,
			mergeIncompatible: true,
			backendRef:        serviceBackendRef,
			want:              false,
		},
		{
			name:         "dynamic resolver backend never merges even when otherwise eligible",
			mergeEnabled: true,
			backendRef:   dynamicResolverBackendRef,
			backend:      dynamicResolverBackend,
			want:         false,
		},
		{
			name:         "CredentialInjection-filtered backendRef never merges even when otherwise eligible",
			mergeEnabled: true,
			backendRef:   serviceBackendRef,
			filters:      &ir.DestinationFilters{CredentialInjection: &ir.CredentialInjection{}},
			want:         false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			backendMap := map[types.NamespacedName]*egv1a1.Backend{}
			if tc.backend != nil {
				backendMap[types.NamespacedName{Namespace: tc.backend.Namespace, Name: tc.backend.Name}] = tc.backend
			}
			tr := &Translator{
				MergeBackends: tc.mergeEnabled,
				TranslatorContext: &TranslatorContext{
					BackendMap: backendMap,
					BTPRoutingTypeIndex: &BTPRoutingTypeIndex{
						gatewayLevel: map[btpRoutingKey]*egv1a1.RoutingType{
							{Kind: "Gateway", Namespace: gwNN.Namespace, Name: gwNN.Name}: tc.gatewayBaselineRT,
						},
					},
				},
			}
			got := tr.shouldMergeBackend(gwCtx, tc.effectiveRT, tc.mergeIncompatible, tc.backendRef, "default", &ir.DestinationSetting{Filters: tc.filters})
			require.Equal(t, tc.want, got)
		})
	}
}

func TestIsMergeableBackendKind(t *testing.T) {
	dynamicResolverType := egv1a1.BackendTypeDynamicResolver
	tests := []struct {
		name                string
		backendRef          gwapiv1.BackendObjectReference
		backend             *egv1a1.Backend
		extensionGroupKinds []schema.GroupKind
		want                bool
	}{
		{
			name:       "service is mergeable",
			backendRef: gwapiv1.BackendObjectReference{Name: "service-1"},
			want:       true,
		},
		{
			name: "backend CR is mergeable",
			backendRef: gwapiv1.BackendObjectReference{
				Group: GroupPtr(egv1a1.GroupName),
				Kind:  KindPtr(egv1a1.KindBackend),
				Name:  "be-1",
			},
			backend: &egv1a1.Backend{
				ObjectMeta: metav1.ObjectMeta{Name: "be-1", Namespace: "default"},
			},
			want: true,
		},
		{
			name: "dynamic resolver backend is never mergeable",
			backendRef: gwapiv1.BackendObjectReference{
				Group: GroupPtr(egv1a1.GroupName),
				Kind:  KindPtr(egv1a1.KindBackend),
				Name:  "be-dynamic",
			},
			backend: &egv1a1.Backend{
				ObjectMeta: metav1.ObjectMeta{Name: "be-dynamic", Namespace: "default"},
				Spec:       egv1a1.BackendSpec{Type: &dynamicResolverType},
			},
			want: false,
		},
		{
			name: "custom backend is never mergeable",
			backendRef: gwapiv1.BackendObjectReference{
				Group: GroupPtr("example.io"),
				Kind:  KindPtr("Foo"),
				Name:  "custom-1",
			},
			extensionGroupKinds: []schema.GroupKind{{Group: "example.io", Kind: "Foo"}},
			want:                false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tr := &Translator{ExtensionGroupKinds: tc.extensionGroupKinds}
			backendMap := map[types.NamespacedName]*egv1a1.Backend{}
			if tc.backend != nil {
				backendMap[types.NamespacedName{Namespace: tc.backend.Namespace, Name: tc.backend.Name}] = tc.backend
			}
			tr.TranslatorContext = &TranslatorContext{BackendMap: backendMap}
			require.Equal(t, tc.want, tr.isMergeableBackendKind(tc.backendRef, "default"))
		})
	}
}

func TestIsFallbackBackend(t *testing.T) {
	fallbackTrue := true
	tests := []struct {
		name       string
		backendRef gwapiv1.BackendObjectReference
		backend    *egv1a1.Backend
		want       bool
	}{
		{
			name:       "service backendRef is never a fallback backend",
			backendRef: gwapiv1.BackendObjectReference{Name: "service-1"},
			want:       false,
		},
		{
			name: "backend CR with Fallback true",
			backendRef: gwapiv1.BackendObjectReference{
				Group: GroupPtr(egv1a1.GroupName),
				Kind:  KindPtr(egv1a1.KindBackend),
				Name:  "be-fallback",
			},
			backend: &egv1a1.Backend{
				ObjectMeta: metav1.ObjectMeta{Name: "be-fallback", Namespace: "default"},
				Spec:       egv1a1.BackendSpec{Fallback: &fallbackTrue},
			},
			want: true,
		},
		{
			name: "backend CR without Fallback set",
			backendRef: gwapiv1.BackendObjectReference{
				Group: GroupPtr(egv1a1.GroupName),
				Kind:  KindPtr(egv1a1.KindBackend),
				Name:  "be-plain",
			},
			backend: &egv1a1.Backend{
				ObjectMeta: metav1.ObjectMeta{Name: "be-plain", Namespace: "default"},
			},
			want: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			backendMap := map[types.NamespacedName]*egv1a1.Backend{}
			if tc.backend != nil {
				backendMap[types.NamespacedName{Namespace: tc.backend.Namespace, Name: tc.backend.Name}] = tc.backend
			}
			tr := &Translator{TranslatorContext: &TranslatorContext{BackendMap: backendMap}}
			got := tr.isFallbackBackend(tc.backendRef, "default")
			require.Equal(t, tc.want, got)
		})
	}
}

func TestMergeIncompatibleForRule(t *testing.T) {
	fallbackTrue := true
	fallbackBackend := &egv1a1.Backend{
		ObjectMeta: metav1.ObjectMeta{Name: "be-fallback", Namespace: "default"},
		Spec:       egv1a1.BackendSpec{Fallback: &fallbackTrue},
	}
	fallbackRef := gwapiv1.BackendObjectReference{
		Group: GroupPtr(egv1a1.GroupName),
		Kind:  KindPtr(egv1a1.KindBackend),
		Name:  "be-fallback",
	}
	serviceRef1 := gwapiv1.BackendObjectReference{Name: "service-1"}
	serviceRef2 := gwapiv1.BackendObjectReference{Name: "service-2"}

	route := &HTTPRouteContext{HTTPRoute: &gwapiv1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "route-1"}}}
	parentRef := &RouteParentContext{ParentReference: &gwapiv1.ParentReference{}}
	ruleName := SectionNamePtr("rule-1")
	gatewayCtx := &GatewayContext{Gateway: &gwapiv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: "envoy-gateway", Name: "gateway-1"}}}

	// consistentHashIdx forces IsConsistentHash to return true for route-1/rule-1, regardless of gatewayCtx.
	consistentHashIdx := &BTPLoadBalancerIndex{
		routeRuleLevel: map[btpRoutingKey]*egv1a1.LoadBalancer{
			{Kind: "HTTPRoute", Namespace: "default", Name: "route-1", SectionName: "rule-1"}: {Type: egv1a1.ConsistentHashLoadBalancerType},
		},
	}

	tests := []struct {
		name                         string
		backendRefs                  []gwapiv1.BackendObjectReference
		hasRouteLevelClusterSettings bool
		sessionPersistent            bool
		gatewayCtx                   *GatewayContext
		lbIndex                      *BTPLoadBalancerIndex
		want                         bool
	}{
		{
			name:                         "route-level cluster settings short-circuits regardless of backendRefs",
			backendRefs:                  []gwapiv1.BackendObjectReference{serviceRef1},
			hasRouteLevelClusterSettings: true,
			want:                         true,
		},
		{
			name:        "single backendRef is always compatible",
			backendRefs: []gwapiv1.BackendObjectReference{fallbackRef},
			want:        false,
		},
		{
			name:              "multiple backendRefs with session persistence",
			backendRefs:       []gwapiv1.BackendObjectReference{serviceRef1, serviceRef2},
			sessionPersistent: true,
			want:              true,
		},
		{
			name:        "multiple backendRefs with a fallback backend",
			backendRefs: []gwapiv1.BackendObjectReference{serviceRef1, fallbackRef},
			want:        true,
		},
		{
			name:        "multiple plain backendRefs with ConsistentHash",
			backendRefs: []gwapiv1.BackendObjectReference{serviceRef1, serviceRef2},
			gatewayCtx:  gatewayCtx,
			lbIndex:     consistentHashIdx,
			want:        true,
		},
		{
			name:        "multiple plain backendRefs with ConsistentHash but nil gatewayCtx",
			backendRefs: []gwapiv1.BackendObjectReference{serviceRef1, serviceRef2},
			gatewayCtx:  nil,
			lbIndex:     consistentHashIdx,
			want:        false,
		},
		{
			name:        "multiple plain backendRefs, no incompatibility",
			backendRefs: []gwapiv1.BackendObjectReference{serviceRef1, serviceRef2},
			gatewayCtx:  gatewayCtx,
			want:        false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tr := &Translator{TranslatorContext: &TranslatorContext{
				BackendMap:           map[types.NamespacedName]*egv1a1.Backend{{Namespace: "default", Name: "be-fallback"}: fallbackBackend},
				BTPLoadBalancerIndex: tc.lbIndex,
			}}
			got := tr.mergeIncompatibleForRule(route, parentRef, ruleName, tc.backendRefs, tc.hasRouteLevelClusterSettings, tc.sessionPersistent, tc.gatewayCtx)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestGetOrCreateBackendCluster(t *testing.T) {
	key := BackendClusterKey{Kind: "Service", Namespace: "default", Name: "service-1", Port: 8080}
	ds1 := &ir.DestinationSetting{Name: "ds-1"}
	ds2 := &ir.DestinationSetting{Name: "ds-2"}

	t.Run("cache miss creates and registers into gwIR.Backends", func(t *testing.T) {
		tr := &Translator{TranslatorContext: &TranslatorContext{BackendClusterMap: map[BackendClusterKey]*ir.BackendCluster{}}}
		gwIR := &ir.Xds{}
		bc := tr.getOrCreateBackendCluster(gwIR, &key, "backend/service/default/service-1/8080", true, ds1, nil)
		require.Len(t, gwIR.Backends, 1)
		require.Same(t, bc, gwIR.Backends[0])
		require.Equal(t, []*ir.DestinationSetting{ds1}, bc.Settings)
	})

	t.Run("cache hit while merge=true does not append the new setting", func(t *testing.T) {
		tr := &Translator{TranslatorContext: &TranslatorContext{BackendClusterMap: map[BackendClusterKey]*ir.BackendCluster{}}}
		gwIR := &ir.Xds{}
		first := tr.getOrCreateBackendCluster(gwIR, &key, "backend/service/default/service-1/8080", true, ds1, nil)
		second := tr.getOrCreateBackendCluster(gwIR, &key, "backend/service/default/service-1/8080", true, ds2, nil)
		require.Same(t, first, second)
		require.Equal(t, []*ir.DestinationSetting{ds1}, second.Settings)
		require.Len(t, gwIR.Backends, 1)
	})

	t.Run("cache hit while merge=false appends the new setting", func(t *testing.T) {
		tr := &Translator{TranslatorContext: &TranslatorContext{BackendClusterMap: map[BackendClusterKey]*ir.BackendCluster{}}}
		gwIR := &ir.Xds{}
		routeScopedKey := BackendClusterKey{Name: "route-scoped-name"}
		first := tr.getOrCreateBackendCluster(gwIR, &routeScopedKey, "route-scoped-name", false, ds1, nil)
		second := tr.getOrCreateBackendCluster(gwIR, &routeScopedKey, "route-scoped-name", false, ds2, nil)
		require.Same(t, first, second)
		require.Equal(t, []*ir.DestinationSetting{ds1, ds2}, second.Settings)
	})
}

func TestBackendClusterKeyProtocolDivergence(t *testing.T) {
	tr := &Translator{MergeBackends: true, TranslatorContext: &TranslatorContext{}}
	gwCtx := &GatewayContext{Gateway: &gwapiv1.Gateway{}}
	parentRef := &RouteParentContext{ParentReference: &gwapiv1.ParentReference{}}
	serviceBackendRef := gwapiv1.BackendObjectReference{Name: "service-1"}
	identity := &BackendClusterKey{Kind: "Service", Namespace: "default", Name: "service-1", Port: 8080}

	cluster1 := tr.resolveBackendCluster("httproute-dest", gwCtx, parentRef, nil, false, identity, serviceBackendRef, "default", &ir.DestinationSetting{Protocol: ir.HTTP})
	require.True(t, cluster1.Merge)

	identity2 := &BackendClusterKey{Kind: "Service", Namespace: "default", Name: "service-1", Port: 8080}
	cluster2 := tr.resolveBackendCluster("grpcroute-dest", gwCtx, parentRef, nil, false, identity2, serviceBackendRef, "default", &ir.DestinationSetting{Protocol: ir.GRPC})
	require.True(t, cluster2.Merge)

	require.NotEqual(t, *cluster1.Key, *cluster2.Key, "an HTTPRoute and a GRPCRoute targeting the same backend must not share a BackendClusterKey")
	require.NotEqual(t, cluster1.Name, cluster2.Name, "an HTTPRoute and a GRPCRoute targeting the same backend must not resolve to the same cluster name")
}
