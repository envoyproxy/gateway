// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
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
			require.Equal(t, tc.equal, cmp.Equal(tc.a, tc.b))
		})
	}
}
