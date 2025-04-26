// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"testing"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/ir"
)

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
						{Name: ptr.To("http"), Port: ptr.To(int32(80)), Protocol: ptr.To(corev1.ProtocolTCP)},
					},
				},
				{
					ObjectMeta:  metav1.ObjectMeta{Name: "slice2"},
					AddressType: discoveryv1.AddressTypeIPv6,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"2001:db8::1"}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: ptr.To("http"), Port: ptr.To(int32(80)), Protocol: ptr.To(corev1.ProtocolTCP)},
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
						{Name: ptr.To("http"), Port: ptr.To(int32(80)), Protocol: ptr.To(corev1.ProtocolTCP)},
					},
				},
				{
					ObjectMeta:  metav1.ObjectMeta{Name: "slice2"},
					AddressType: discoveryv1.AddressTypeFQDN,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"example.com"}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: ptr.To("http"), Port: ptr.To(int32(80)), Protocol: ptr.To(corev1.ProtocolTCP)},
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
						{Name: ptr.To("http"), Port: ptr.To(int32(80)), Protocol: ptr.To(corev1.ProtocolTCP)},
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
						{Name: ptr.To("http"), Port: ptr.To(int32(80)), Protocol: ptr.To(corev1.ProtocolTCP)},
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
						{Name: ptr.To("http"), Port: ptr.To(int32(80)), Protocol: ptr.To(corev1.ProtocolTCP)},
					},
				},
				{
					ObjectMeta:  metav1.ObjectMeta{Name: "slice2-ipv6"},
					AddressType: discoveryv1.AddressTypeIPv6,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"2001:db8::1"}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: ptr.To("http"), Port: ptr.To(int32(80)), Protocol: ptr.To(corev1.ProtocolTCP)},
					},
				},
				{
					ObjectMeta:  metav1.ObjectMeta{Name: "slice3-fqdn"},
					AddressType: discoveryv1.AddressTypeFQDN,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"example.com"}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: ptr.To("http"), Port: ptr.To(int32(80)), Protocol: ptr.To(corev1.ProtocolTCP)},
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
			name: "Keep serving and terminating as draining",
			endpointSlices: []*discoveryv1.EndpointSlice{
				{
					ObjectMeta:  metav1.ObjectMeta{Name: "slice1"},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{Addresses: []string{"192.0.2.1"}, Conditions: discoveryv1.EndpointConditions{
							Ready: ptr.To(false), Serving: ptr.To(true), Terminating: ptr.To(true),
						}},
						{Addresses: []string{"192.0.2.2"}, Conditions: discoveryv1.EndpointConditions{
							Ready: ptr.To(false), Serving: ptr.To(false), Terminating: ptr.To(true),
						}},
						{Addresses: []string{"192.0.2.3"}, Conditions: discoveryv1.EndpointConditions{
							Ready: ptr.To(false),
						}},
					},
					Ports: []discoveryv1.EndpointPort{
						{Name: ptr.To("http"), Port: ptr.To(int32(80)), Protocol: ptr.To(corev1.ProtocolTCP)},
					},
				},
			},
			portName:     "http",
			portProtocol: corev1.ProtocolTCP,
			expectedEndpoints: []*ir.DestinationEndpoint{
				{Host: "192.0.2.1", Port: 80, Draining: true},
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

func TestValidateDestinationSettings(t *testing.T) {
	svcKind := gwapiv1.Kind(resource.KindService)

	tests := []struct {
		name                    string
		ds                      *ir.DestinationSetting
		endpointRoutingDisabled bool
		kind                    *gwapiv1.Kind
		wantErr                 bool
		wantReason              gwapiv1.RouteConditionReason
	}{
		{
			name: "headless service rejected when endpointRoutingDisabled=true",
			ds: &ir.DestinationSetting{
				Name:      "headless",
				Endpoints: []*ir.DestinationEndpoint{{Host: "None"}},
			},
			endpointRoutingDisabled: true,
			kind:                    &svcKind,
			wantErr:                 true,
			wantReason:              status.RouteReasonUnsupportedSetting,
		},
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
			name: "mixed address type rejected when EndpointSlice routing",
			ds: &ir.DestinationSetting{
				Name:        "mixed",
				Endpoints:   []*ir.DestinationEndpoint{{Host: "10.0.0.1"}},
				AddressType: ptr.To(ir.MIXED),
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
func TestIsHeadlessService(t *testing.T) {
	tests := []struct {
		name      string
		endpoints []*ir.DestinationEndpoint
		want      bool
	}{
		{
			name: "headless Service",
			endpoints: []*ir.DestinationEndpoint{
				{Host: "None"},
			},
			want: true,
		},
		{
			name: "non headless Service with valid ClusterIP",
			endpoints: []*ir.DestinationEndpoint{
				{Host: "10.0.0.1"},
			},
			want: false,
		},
		{
			name:      "empty slice",
			endpoints: nil,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &ir.DestinationSetting{Endpoints: tt.endpoints}
			got := isHeadlessService(ds)
			require.Equal(t, tt.want, got)
		})
	}
}
