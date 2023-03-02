// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
)

var (
	inPath = "./testdata/decoder/in/"
)

func TestDecode(t *testing.T) {
	testCases := []struct {
		in     string
		out    *v1alpha1.EnvoyGateway
		expect bool
	}{
		{
			in: inPath + "kube-provider.yaml",
			out: &v1alpha1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.KindEnvoyGateway,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Provider: v1alpha1.DefaultProvider(),
				},
			},
			expect: true,
		},
		{
			in: inPath + "gateway-controller-name.yaml",
			out: &v1alpha1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.KindEnvoyGateway,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway: v1alpha1.DefaultGateway(),
				},
			},
			expect: true,
		},
		{
			in: inPath + "provider-with-gateway.yaml",
			out: &v1alpha1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.KindEnvoyGateway,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultProvider(),
				},
			},
			expect: true,
		},
		{
			in: inPath + "provider-mixing-gateway.yaml",
			out: &v1alpha1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.KindEnvoyGateway,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Provider: v1alpha1.DefaultProvider(),
				},
			},
			expect: true,
		},
		{
			in: inPath + "gateway-mixing-provider.yaml",
			out: &v1alpha1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.KindEnvoyGateway,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway: v1alpha1.DefaultGateway(),
				},
			},
			expect: true,
		},
		{
			in: inPath + "provider-mixing-gateway.yaml",
			out: &v1alpha1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.KindEnvoyGateway,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Provider: v1alpha1.DefaultProvider(),
					Gateway:  v1alpha1.DefaultGateway(),
				},
			},
			expect: false,
		},
		{
			in: inPath + "gateway-mixing-provider.yaml",
			out: &v1alpha1.EnvoyGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.KindEnvoyGateway,
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Provider: v1alpha1.DefaultProvider(),
					Gateway:  v1alpha1.DefaultGateway(),
				},
			},
			expect: false,
		},
		{
			in:     inPath + "no-api-version.yaml",
			expect: false,
		},
		{
			in:     inPath + "no-kind.yaml",
			expect: false,
		},
		{
			in:     "/non/existent/config.yaml",
			expect: false,
		},
		{
			in:     inPath + "invalid-gateway-group.yaml",
			expect: false,
		},
		{
			in:     inPath + "invalid-gateway-kind.yaml",
			expect: false,
		},
		{
			in:     inPath + "invalid-gateway-version.yaml",
			expect: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			eg, err := Decode(tc.in)
			if tc.expect {
				require.NoError(t, err)
				require.Equal(t, tc.out, eg)
			} else {
				require.Equal(t, !reflect.DeepEqual(tc.out, eg) || err != nil, true)
			}
		})
	}
}
