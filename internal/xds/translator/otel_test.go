// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	cfgcore "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/stretchr/testify/require"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestBuildGrpcInitialMetadata(t *testing.T) {
	tests := []struct {
		name     string
		headers  []gwapiv1.HTTPHeader
		expected []*cfgcore.HeaderValue
	}{
		{
			name:     "nil headers",
			headers:  nil,
			expected: nil,
		},
		{
			name:     "empty headers",
			headers:  []gwapiv1.HTTPHeader{},
			expected: nil,
		},
		{
			name: "single header",
			headers: []gwapiv1.HTTPHeader{
				{Name: "X-Custom-Header", Value: "custom-value"},
			},
			expected: []*cfgcore.HeaderValue{
				{Key: "X-Custom-Header", Value: "custom-value"},
			},
		},
		{
			name: "multiple headers",
			headers: []gwapiv1.HTTPHeader{
				{Name: "X-Foo", Value: "bar"},
				{Name: "X-Custom-Header", Value: "custom-value"},
			},
			expected: []*cfgcore.HeaderValue{
				{Key: "X-Foo", Value: "bar"},
				{Key: "X-Custom-Header", Value: "custom-value"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := buildGrpcInitialMetadata(tc.headers)
			require.Equal(t, tc.expected, actual)
		})
	}
}
