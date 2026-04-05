// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"context"
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

func TestEqualControllerResourcesContext(t *testing.T) {
	c1 := context.Background()
	c2 := context.TODO()
	r1 := &ControllerResourcesContext{
		Resources: &ControllerResources{
			{
				GatewayClass: &gwapiv1.GatewayClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
				},
			},
		},
		Context: c1,
	}
	r2 := &ControllerResourcesContext{
		Resources: &ControllerResources{
			{
				GatewayClass: &gwapiv1.GatewayClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
				},
			},
		},
		Context: c2,
	}

	assert.True(t, r1.Equal(r2))
	assert.True(t, r2.Equal(r1))
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

func TestControllerResourcesContextDeepCopy(t *testing.T) {
	tests := []struct {
		name string
		ctx  *ControllerResourcesContext
	}{
		{
			name: "nil context",
			ctx:  nil,
		},
		{
			name: "empty context",
			ctx: &ControllerResourcesContext{
				Resources: &ControllerResources{},
				Context:   context.Background(),
			},
		},
		{
			name: "context with resources",
			ctx: &ControllerResourcesContext{
				Resources: &ControllerResources{
					{
						GatewayClass: &gwapiv1.GatewayClass{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-gateway-class",
							},
						},
					},
				},
				Context: context.Background(),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			copied := tc.ctx.DeepCopy()

			if tc.ctx == nil {
				assert.Nil(t, copied)
				return
			}

			// Verify the copy is not nil
			require.NotNil(t, copied)

			// Verify the copy is a different object
			assert.NotSame(t, tc.ctx, copied)

			// Verify Resources are deep copied
			if tc.ctx.Resources != nil {
				require.NotNil(t, copied.Resources)
				assert.NotSame(t, tc.ctx.Resources, copied.Resources)

				// Verify the contents are equal
				assert.Len(t, *copied.Resources, len(*tc.ctx.Resources))
			}

			// Verify Context is preserved (not deep copied, same reference)
			assert.Equal(t, tc.ctx.Context, copied.Context)
		})
	}
}

func TestControllerResourcesDeepCopy(t *testing.T) {
	tests := []struct {
		name      string
		resources *ControllerResources
	}{
		{
			name:      "nil resources",
			resources: nil,
		},
		{
			name:      "empty resources",
			resources: &ControllerResources{},
		},
		{
			name: "resources with gateway class",
			resources: &ControllerResources{
				{
					GatewayClass: &gwapiv1.GatewayClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-gateway-class",
						},
					},
				},
			},
		},
		{
			name: "multiple resources",
			resources: &ControllerResources{
				{
					GatewayClass: &gwapiv1.GatewayClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "gateway-class-1",
						},
					},
				},
				{
					GatewayClass: &gwapiv1.GatewayClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "gateway-class-2",
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			copied := tc.resources.DeepCopy()

			if tc.resources == nil {
				assert.Nil(t, copied)
				return
			}

			// Verify the copy is not nil
			require.NotNil(t, copied)

			// Verify the copy is a different object
			assert.NotSame(t, tc.resources, copied)

			// Verify the length is the same
			assert.Len(t, *copied, len(*tc.resources))

			// Verify each resource is deep copied
			for i := range *tc.resources {
				if (*tc.resources)[i] != nil {
					require.NotNil(t, (*copied)[i])
					assert.NotSame(t, (*tc.resources)[i], (*copied)[i])
				}
			}
		})
	}
}
