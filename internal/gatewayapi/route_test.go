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
	mcsapiv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

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
			endpoints, addrType := getIREndpointsFromEndpointSlices(tt.endpointSlices, tt.portName, tt.portProtocol, nil)

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

func TestServiceEndpointHostname(t *testing.T) {
	t.Run("nil setting returns nil", func(t *testing.T) {
		translator := &Translator{}
		service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "service-1", Namespace: "default"}}

		hostname := translator.serviceEndpointHostname(service, nil)

		require.Nil(t, hostname)
	})

	t.Run("none type returns nil", func(t *testing.T) {
		translator := &Translator{}
		service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "service-1", Namespace: "default"}}
		setting := &egv1a1.BackendEndpointHostname{
			Type: egv1a1.BackendEndpointHostnameTypeNone,
		}

		hostname := translator.serviceEndpointHostname(service, setting)

		require.Nil(t, hostname)
	})

	t.Run("kubernetes service uses default cluster domain", func(t *testing.T) {
		translator := &Translator{}
		service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "service-1", Namespace: "default"}}
		setting := &egv1a1.BackendEndpointHostname{
			Type: egv1a1.BackendEndpointHostnameTypeKubernetesService,
		}

		hostname := translator.serviceEndpointHostname(service, setting)

		require.Equal(t, new("service-1.default.svc.cluster.local"), hostname)
	})

	t.Run("kubernetes service uses configured dns domain", func(t *testing.T) {
		translator := &Translator{DNSDomain: "example.internal"}
		service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "service-1", Namespace: "default"}}
		setting := &egv1a1.BackendEndpointHostname{
			Type: egv1a1.BackendEndpointHostnameTypeKubernetesService,
		}

		hostname := translator.serviceEndpointHostname(service, setting)

		require.Equal(t, new("service-1.default.svc.example.internal"), hostname)
	})

	t.Run("kubernetes service ignores missing service name", func(t *testing.T) {
		translator := &Translator{}
		service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}}
		setting := &egv1a1.BackendEndpointHostname{
			Type: egv1a1.BackendEndpointHostnameTypeKubernetesService,
		}

		hostname := translator.serviceEndpointHostname(service, setting)

		require.Nil(t, hostname)
	})

	t.Run("kubernetes service ignores missing service namespace", func(t *testing.T) {
		translator := &Translator{}
		service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "service-1"}}
		setting := &egv1a1.BackendEndpointHostname{
			Type: egv1a1.BackendEndpointHostnameTypeKubernetesService,
		}

		hostname := translator.serviceEndpointHostname(service, setting)

		require.Nil(t, hostname)
	})

	t.Run("endpoint slices use resolved hostname", func(t *testing.T) {
		endpointSlices := []*discoveryv1.EndpointSlice{{
			AddressType: discoveryv1.AddressTypeIPv4,
			Endpoints: []discoveryv1.Endpoint{{
				Addresses: []string{"10.0.0.1"},
				Conditions: discoveryv1.EndpointConditions{
					Ready: new(true),
				},
			}},
			Ports: []discoveryv1.EndpointPort{{
				Name:     new("http"),
				Protocol: new(corev1.ProtocolTCP),
				Port:     new(int32(8080)),
			}},
		}}

		endpoints, _ := getIREndpointsFromEndpointSlices(endpointSlices, "http", corev1.ProtocolTCP, new("service-1.default.svc.cluster.local"))

		require.Len(t, endpoints, 1)
		require.Equal(t, new("service-1.default.svc.cluster.local"), endpoints[0].Hostname)
	})

	t.Run("cluster ip endpoint uses resolved hostname", func(t *testing.T) {
		translator := &Translator{}
		port := int32(8080)
		portNum := port
		backendRef := gwapiv1.BackendObjectReference{
			Name: "service-1",
			Port: &portNum,
		}
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "service-1", Namespace: "default"},
			Spec: corev1.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports: []corev1.ServicePort{{
					Name: "http",
					Port: port,
				}},
			},
		}
		translator.TranslatorContext = &TranslatorContext{
			ServiceMap: map[types.NamespacedName]*corev1.Service{
				{Namespace: "default", Name: "service-1"}: service,
			},
		}
		setting := &egv1a1.BackendEndpointHostname{
			Type: egv1a1.BackendEndpointHostnameTypeKubernetesService,
		}
		serviceRouting := egv1a1.ServiceRoutingType

		ds, err := translator.processServiceDestinationSetting("test", backendRef, "default", ir.HTTP, nil, &serviceRouting, setting)

		require.NoError(t, err)
		require.Len(t, ds.Endpoints, 1)
		require.Equal(t, new("service-1.default.svc.cluster.local"), ds.Endpoints[0].Hostname)
	})

	t.Run("static type returns specified hostname", func(t *testing.T) {
		translator := &Translator{}
		service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "service-1", Namespace: "default"}}
		setting := &egv1a1.BackendEndpointHostname{
			Type:     egv1a1.BackendEndpointHostnameTypeStatic,
			Hostname: new("custom-static.example.com"),
		}

		hostname := translator.serviceEndpointHostname(service, setting)

		require.Equal(t, new("custom-static.example.com"), hostname)
	})

	t.Run("static type with nil hostname returns nil", func(t *testing.T) {
		translator := &Translator{}
		service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "service-1", Namespace: "default"}}
		setting := &egv1a1.BackendEndpointHostname{
			Type:     egv1a1.BackendEndpointHostnameTypeStatic,
			Hostname: nil,
		}

		hostname := translator.serviceEndpointHostname(service, setting)

		require.Nil(t, hostname)
	})

	t.Run("static type ignores nil service", func(t *testing.T) {
		translator := &Translator{}
		setting := &egv1a1.BackendEndpointHostname{
			Type:     egv1a1.BackendEndpointHostnameTypeStatic,
			Hostname: new("custom-static.example.com"),
		}

		hostname := translator.serviceEndpointHostname(nil, setting)

		require.Equal(t, new("custom-static.example.com"), hostname)
	})
}

// TestBackendClusterKeyConstruction covers the key-building logic now inlined in processBackendRef
// (newBackendClusterKey + backendClusterKeyForGateway). Eligibility itself is covered separately and
// exhaustively by TestShouldMergeBackend.
func TestBackendClusterKeyConstruction(t *testing.T) {
	serviceBackendRef := gwapiv1.BackendObjectReference{Name: "service-1", Port: PortNumPtr(8080)}
	tr := &Translator{}

	t.Run("resolves to the backend's own identity", func(t *testing.T) {
		key := newBackendClusterKey(serviceBackendRef, "default")
		require.Equal(t, "Service", key.Kind)
		require.Equal(t, "service-1", key.Name)
	})

	t.Run("never collides across gateways", func(t *testing.T) {
		gwCtx1 := &GatewayContext{Gateway: &gwapiv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: "envoy-gateway", Name: "gateway-1"}}}
		gwCtx2 := &GatewayContext{Gateway: &gwapiv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: "envoy-gateway", Name: "gateway-2"}}}

		baseKey := newBackendClusterKey(serviceBackendRef, "default")
		key1 := tr.backendClusterKeyForGateway(&baseKey, gwCtx1, ir.HTTP)
		key2 := tr.backendClusterKeyForGateway(&baseKey, gwCtx2, ir.HTTP)

		require.NotEqual(t, key1, key2, "the same backend under two different gateways must not collide in BackendClusterMap")
	})

	t.Run("HTTPRoute and GRPCRoute targeting the same backend never collide", func(t *testing.T) {
		gwCtx := &GatewayContext{Gateway: &gwapiv1.Gateway{}}
		baseKey := newBackendClusterKey(serviceBackendRef, "default")
		key1 := tr.backendClusterKeyForGateway(&baseKey, gwCtx, ir.HTTP)
		key2 := tr.backendClusterKeyForGateway(&baseKey, gwCtx, ir.GRPC)

		require.NotEqual(t, key1, key2, "an HTTPRoute and a GRPCRoute targeting the same backend must not share a BackendClusterKey")
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
		mergeSelector     *metav1.LabelSelector
		gatewayEnvoyProxy *egv1a1.EnvoyProxy
		gatewayBaselineRT *egv1a1.RoutingType
		effectiveRT       *egv1a1.RoutingType
		mergeIncompatible bool
		backendRef        gwapiv1.BackendObjectReference
		backend           *egv1a1.Backend
		service           *corev1.Service
		serviceImport     *mcsapiv1a1.ServiceImport
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
			name:         "disabled globally, but Gateway-level EnvoyProxy enables it",
			mergeEnabled: false,
			gatewayEnvoyProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{MergeBackends: &egv1a1.MergeBackendsConfig{}},
			},
			backendRef: serviceBackendRef,
			want:       true,
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
		{
			name:         "any other per-backendRef filter (e.g. header modification) never merges either",
			mergeEnabled: true,
			backendRef:   serviceBackendRef,
			filters:      &ir.DestinationFilters{AddRequestHeaders: []ir.AddHeader{{Name: "x-foo", Value: []string{"bar"}}}},
			want:         false,
		},
		{
			name:         "nil default-level selector: matches everything",
			mergeEnabled: true,
			backendRef:   serviceBackendRef,
			want:         true,
		},
		{
			name:          "default-level selector matches the Service's labels: merges",
			mergeEnabled:  true,
			mergeSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"tier": "dedup"}},
			backendRef:    serviceBackendRef,
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: "service-1", Namespace: "default", Labels: map[string]string{"tier": "dedup"}},
			},
			want: true,
		},
		{
			name:          "default-level selector does not match the Service's labels: excluded",
			mergeEnabled:  true,
			mergeSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"tier": "dedup"}},
			backendRef:    serviceBackendRef,
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: "service-1", Namespace: "default", Labels: map[string]string{"tier": "other"}},
			},
			want: false,
		},
		{
			name:          "selector configured but target Service not found: excluded (fail-closed)",
			mergeEnabled:  true,
			mergeSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"tier": "dedup"}},
			backendRef:    serviceBackendRef,
			want:          false,
		},
		{
			name:         "selector fails to parse (In operator with no values): excluded (fail-closed)",
			mergeEnabled: true,
			mergeSelector: &metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{
				{Key: "tier", Operator: metav1.LabelSelectorOpIn, Values: nil},
			}},
			backendRef: serviceBackendRef,
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: "service-1", Namespace: "default", Labels: map[string]string{"tier": "dedup"}},
			},
			want: false,
		},
		{
			name:          "selector matches a Backend CR's labels, not just Service",
			mergeEnabled:  true,
			mergeSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"tier": "dedup"}},
			backendRef:    dynamicResolverBackendRef,
			backend: &egv1a1.Backend{
				ObjectMeta: metav1.ObjectMeta{Name: "be-dynamic", Namespace: "default", Labels: map[string]string{"tier": "dedup"}},
				// deliberately NOT a dynamic-resolver type here, so isMergeableBackendKind doesn't
				// exclude it first and this case actually exercises the selector path; reuse
				// dynamicResolverBackendRef only for its Kind=Backend, with a plain Spec.
				Spec: egv1a1.BackendSpec{},
			},
			want: true,
		},
		{
			name:          "selector matches a ServiceImport's labels",
			mergeEnabled:  true,
			mergeSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"tier": "dedup"}},
			backendRef: gwapiv1.BackendObjectReference{
				Group: GroupPtr(mcsapiv1a1.GroupName),
				Kind:  KindPtr(resource.KindServiceImport),
				Name:  "service-import-1",
			},
			serviceImport: &mcsapiv1a1.ServiceImport{
				ObjectMeta: metav1.ObjectMeta{Name: "service-import-1", Namespace: "default", Labels: map[string]string{"tier": "dedup"}},
			},
			want: true,
		},
		{
			name:          "Gateway-level EnvoyProxy selector overrides the default-level selector entirely",
			mergeEnabled:  true,
			mergeSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"tier": "dedup"}},
			gatewayEnvoyProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{MergeBackends: &egv1a1.MergeBackendsConfig{
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"tier": "other"}},
				}},
			},
			backendRef: serviceBackendRef,
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: "service-1", Namespace: "default", Labels: map[string]string{"tier": "dedup"}},
			},
			want: false, // matches the DEFAULT-level selector's labels, not the Gateway-level override's — must be excluded, proving the override, not the default, was actually applied
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			backendMap := map[types.NamespacedName]*egv1a1.Backend{}
			if tc.backend != nil {
				backendMap[types.NamespacedName{Namespace: tc.backend.Namespace, Name: tc.backend.Name}] = tc.backend
			}
			serviceMap := map[types.NamespacedName]*corev1.Service{}
			if tc.service != nil {
				serviceMap[types.NamespacedName{Namespace: tc.service.Namespace, Name: tc.service.Name}] = tc.service
			}
			serviceImportMap := map[types.NamespacedName]*mcsapiv1a1.ServiceImport{}
			if tc.serviceImport != nil {
				serviceImportMap[types.NamespacedName{Namespace: tc.serviceImport.Namespace, Name: tc.serviceImport.Name}] = tc.serviceImport
			}
			var mergeBackends *MergeBackendsConfig
			if tc.mergeEnabled {
				mergeBackends = &MergeBackendsConfig{Selector: tc.mergeSelector}
			}
			tr := &Translator{
				MergeBackends: mergeBackends,
				TranslatorContext: &TranslatorContext{
					BackendMap:       backendMap,
					ServiceMap:       serviceMap,
					ServiceImportMap: serviceImportMap,
					BTPRoutingTypeIndex: func() *BTPRoutingTypeIndex {
						idx := newBTPRoutingTypeIndex()
						idx.setGatewayLevel(gwNN, tc.gatewayBaselineRT)
						return idx
					}(),
				},
			}
			testGwCtx := gwCtx
			if tc.gatewayEnvoyProxy != nil {
				testGwCtx = &GatewayContext{Gateway: gwCtx.Gateway, envoyProxy: tc.gatewayEnvoyProxy}
			}
			got := tr.shouldMergeBackend(testGwCtx, tc.effectiveRT, tc.mergeIncompatible, tc.backendRef, "default", &ir.DestinationSetting{Filters: tc.filters})
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

func TestMergeIncompatibleForWeightedRule(t *testing.T) {
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
	gatewayCtx := &GatewayContext{Gateway: &gwapiv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: "envoy-gateway", Name: "gateway-1"}}}

	// consistentHashIdx forces IsConsistentHash to return true for gatewayCtx's gateway.
	consistentHashIdx := func() *BTPLoadBalancerIndex {
		idx := newBTPLoadBalancerIndex()
		idx.setGatewayLevel(types.NamespacedName{Namespace: "envoy-gateway", Name: "gateway-1"}, true)
		return idx
	}()

	tests := []struct {
		name              string
		backendRefs       []gwapiv1.BackendObjectReference
		sessionPersistent bool
		gatewayCtx        *GatewayContext
		lbIndex           *BTPLoadBalancerIndex
		want              bool
	}{
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
			got := tr.mergeIncompatibleForWeightedRule(tc.gatewayCtx, route, tc.backendRefs, tc.sessionPersistent)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestGetOrCreateBackendCluster(t *testing.T) {
	key := BackendClusterKey{Kind: "Service", Namespace: "default", Name: "service-1", Port: 8080}
	ds1 := &ir.DestinationSetting{Name: "ds-1"}
	ds2 := &ir.DestinationSetting{Name: "ds-2"}

	t.Run("cache miss creates and registers into gwIR.BackendClusters", func(t *testing.T) {
		tr := &Translator{TranslatorContext: &TranslatorContext{BackendClusterMap: map[BackendClusterKey]*ir.BackendCluster{}}}
		gwIR := &ir.Xds{}
		bc := tr.getOrCreateBackendCluster(gwIR, &key, ds1)
		require.Len(t, gwIR.BackendClusters, 1)
		require.Same(t, bc, gwIR.BackendClusters[0])
		require.Equal(t, "service/default/service-1/8080", bc.Name)
		require.Equal(t, bc.Name, bc.Setting.Name, "the shared Setting's Name must match the BackendCluster's own, not whichever route-scoped ds.Name it was built from")
	})

	t.Run("cache hit returns the existing cluster without replacing its setting", func(t *testing.T) {
		tr := &Translator{TranslatorContext: &TranslatorContext{BackendClusterMap: map[BackendClusterKey]*ir.BackendCluster{}}}
		gwIR := &ir.Xds{}
		first := tr.getOrCreateBackendCluster(gwIR, &key, ds1)
		second := tr.getOrCreateBackendCluster(gwIR, &key, ds2)
		require.Same(t, first, second)
		require.Equal(t, first.Name, second.Setting.Name)
		require.Len(t, gwIR.BackendClusters, 1)
	})
}
