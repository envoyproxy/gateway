// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// This file contains code derived from Contour,
// https://github.com/projectcontour/contour
// and is provided here subject to the following:
// Copyright Project Contour Authors
// SPDX-License-Identifier: Apache-2.0

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func TestValidateGRPCFilterRef(t *testing.T) {
	testCases := []struct {
		name     string
		filter   *gwapiv1a2.GRPCRouteFilter
		expected bool
	}{
		{
			name: "request mirror filter",
			filter: &gwapiv1a2.GRPCRouteFilter{
				Type: gwapiv1a2.GRPCRouteFilterRequestMirror,
			},
			expected: true,
		},
		{
			name: "request header modifier filter",
			filter: &gwapiv1a2.GRPCRouteFilter{
				Type: gwapiv1a2.GRPCRouteFilterRequestHeaderModifier,
			},
			expected: true,
		},
		{
			name: "response header modifier filter",
			filter: &gwapiv1a2.GRPCRouteFilter{
				Type: gwapiv1a2.GRPCRouteFilterResponseHeaderModifier,
			},
			expected: true,
		},
		{
			name: "valid extension resource",
			filter: &gwapiv1a2.GRPCRouteFilter{
				Type: gwapiv1a2.GRPCRouteFilterExtensionRef,
				ExtensionRef: &gwapiv1.LocalObjectReference{
					Group: "example.io",
					Kind:  "Foo",
					Name:  "test",
				},
			},
			expected: true,
		},
		{
			name: "unsupported extended filter",
			filter: &gwapiv1a2.GRPCRouteFilter{
				Type: gwapiv1a2.GRPCRouteFilterExtensionRef,
				ExtensionRef: &gwapiv1.LocalObjectReference{
					Group: "UnsupportedGroup",
					Kind:  "UnsupportedKind",
					Name:  "test",
				},
			},
			expected: false,
		},
		{
			name: "empty extended filter",
			filter: &gwapiv1a2.GRPCRouteFilter{
				Type: gwapiv1a2.GRPCRouteFilterExtensionRef,
			},
			expected: false,
		},
		{
			name: "invalid filter type",
			filter: &gwapiv1a2.GRPCRouteFilter{
				Type: "Invalid",
				ExtensionRef: &gwapiv1.LocalObjectReference{
					Group: "example.io",
					Kind:  "Foo",
					Name:  "test",
				},
			},
			expected: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateGRPCRouteFilter(tc.filter, schema.GroupKind{Group: "example.io", Kind: "Foo"})
			if tc.expected {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestValidateHTTPFilterRef(t *testing.T) {
	testCases := []struct {
		name     string
		filter   *gwapiv1.HTTPRouteFilter
		expected bool
	}{
		{
			name: "request mirror filter",
			filter: &gwapiv1.HTTPRouteFilter{
				Type: gwapiv1.HTTPRouteFilterRequestMirror,
			},
			expected: true,
		},
		{
			name: "url rewrite filter",
			filter: &gwapiv1.HTTPRouteFilter{
				Type: gwapiv1.HTTPRouteFilterURLRewrite,
			},
			expected: true,
		},
		{
			name: "request header modifier filter",
			filter: &gwapiv1.HTTPRouteFilter{
				Type: gwapiv1.HTTPRouteFilterRequestHeaderModifier,
			},
			expected: true,
		},
		{
			name: "request redirect filter",
			filter: &gwapiv1.HTTPRouteFilter{
				Type: gwapiv1.HTTPRouteFilterRequestRedirect,
			},
			expected: true,
		},
		{
			name: "unsupported extended filter",
			filter: &gwapiv1.HTTPRouteFilter{
				Type: gwapiv1.HTTPRouteFilterExtensionRef,
				ExtensionRef: &gwapiv1.LocalObjectReference{
					Group: "UnsupportedGroup",
					Kind:  "UnsupportedKind",
					Name:  "test",
				},
			},
			expected: false,
		},
		{
			name: "extended filter with missing reference",
			filter: &gwapiv1.HTTPRouteFilter{
				Type: gwapiv1.HTTPRouteFilterExtensionRef,
			},
			expected: false,
		},
		{
			name: "valid extension resource",
			filter: &gwapiv1.HTTPRouteFilter{
				Type: gwapiv1.HTTPRouteFilterExtensionRef,
				ExtensionRef: &gwapiv1.LocalObjectReference{
					Group: "example.io",
					Kind:  "Foo",
					Name:  "test",
				},
			},
			expected: true,
		},
		{
			name: "invalid filter type",
			filter: &gwapiv1.HTTPRouteFilter{
				Type: "Invalid",
				ExtensionRef: &gwapiv1.LocalObjectReference{
					Group: "example.io",
					Kind:  "Foo",
					Name:  "test",
				},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateHTTPRouteFilter(tc.filter, schema.GroupKind{Group: "example.io", Kind: "Foo"})
			if tc.expected {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
