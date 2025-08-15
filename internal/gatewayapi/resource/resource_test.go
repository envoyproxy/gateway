// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestEqualXds(t *testing.T) {
	tests := []struct {
		desc  string
		a     *ControllerResources
		b     *ControllerResources
		equal bool
	}{
		{
			desc: "different resources",
			a: &ControllerResources{
				{
					GatewayClass: &gwapiv1.GatewayClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "foo",
						},
					},
				},
			},
			b: &ControllerResources{
				{
					GatewayClass: &gwapiv1.GatewayClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "bar",
						},
					},
				},
			},
			equal: false,
		},
		{
			desc: "same order resources are equal",
			a: &ControllerResources{
				{
					GatewayClass: &gwapiv1.GatewayClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "foo",
						},
					},
				},
				{
					GatewayClass: &gwapiv1.GatewayClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "bar",
						},
					},
				},
			},
			b: &ControllerResources{
				{
					GatewayClass: &gwapiv1.GatewayClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "foo",
						},
					},
				},
				{
					GatewayClass: &gwapiv1.GatewayClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "bar",
						},
					},
				},
			},
			equal: true,
		},
		{
			desc: "out of order resources are equal",
			a: &ControllerResources{
				{
					GatewayClass: &gwapiv1.GatewayClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "foo",
						},
					},
				},
				{
					GatewayClass: &gwapiv1.GatewayClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "bar",
						},
					},
				},
			},
			b: &ControllerResources{
				{
					GatewayClass: &gwapiv1.GatewayClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "bar",
						},
					},
				},
				{
					GatewayClass: &gwapiv1.GatewayClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "foo",
						},
					},
				},
			},
			equal: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			tc.a.Sort()
			tc.b.Sort()
			diff := cmp.Diff(tc.a, tc.b)
			got := diff == ""
			require.Equal(t, tc.equal, got)
		})
	}
}

func TestGetEndpointSlicesForBackendDualStack(t *testing.T) {
	// Test data setup
	dualStackService := &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dual-stack-service",
			Namespace: "default",
			Labels: map[string]string{
				discoveryv1.LabelServiceName: "my-dual-stack-service",
			},
		},
		AddressType: discoveryv1.AddressTypeIPv4,
		Endpoints: []discoveryv1.Endpoint{
			{
				Addresses: []string{"192.0.2.1"},
			},
			{
				Addresses: []string{"192.0.2.2"},
			},
		},
	}

	dualStackServiceIPv6 := &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dual-stack-service-ipv6",
			Namespace: "default",
			Labels: map[string]string{
				discoveryv1.LabelServiceName: "my-dual-stack-service",
			},
		},
		AddressType: discoveryv1.AddressTypeIPv6,
		Endpoints: []discoveryv1.Endpoint{
			{
				Addresses: []string{"2001:db8::1"},
			},
			{
				Addresses: []string{"2001:db8::2"},
			},
		},
	}

	resources := &Resources{
		EndpointSlices: []*discoveryv1.EndpointSlice{dualStackService, dualStackServiceIPv6},
	}

	t.Run("Dual Stack Service", func(t *testing.T) {
		result := resources.GetEndpointSlicesForBackend("default", "my-dual-stack-service", KindService)

		assert.Len(t, result, 2, "Expected 2 EndpointSlices for dual-stack service")

		var ipv4Slice, ipv6Slice *discoveryv1.EndpointSlice
		for _, slice := range result {
			switch slice.AddressType {
			case discoveryv1.AddressTypeIPv4:
				ipv4Slice = slice
			case discoveryv1.AddressTypeIPv6:
				ipv6Slice = slice
			}
		}

		assert.NotNil(t, ipv4Slice, "Expected to find an IPv4 EndpointSlice")
		assert.NotNil(t, ipv6Slice, "Expected to find an IPv6 EndpointSlice")

		if ipv4Slice != nil {
			assert.Len(t, ipv4Slice.Endpoints, 2, "Expected 2 IPv4 endpoints")
			assert.Equal(t, "192.0.2.1", ipv4Slice.Endpoints[0].Addresses[0], "Unexpected IPv4 address")
			assert.Equal(t, "192.0.2.2", ipv4Slice.Endpoints[1].Addresses[0], "Unexpected IPv4 address")
		}

		if ipv6Slice != nil {
			assert.Len(t, ipv6Slice.Endpoints, 2, "Expected 2 IPv6 endpoints")
			assert.Equal(t, "2001:db8::1", ipv6Slice.Endpoints[0].Addresses[0], "Unexpected IPv6 address")
			assert.Equal(t, "2001:db8::2", ipv6Slice.Endpoints[1].Addresses[0], "Unexpected IPv6 address")
		}
	})
}
