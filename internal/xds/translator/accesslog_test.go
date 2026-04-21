// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestResolveEGExtFilterStateOperators(t *testing.T) {
	extProcs := []ir.ExtProc{
		{
			Name:       "ns/extproc/my-eep/0",
			CustomName: "auth",
		},
		{
			Name:       "ns/extproc/my-eep/1",
			CustomName: "enrichment",
		},
		{
			Name: "ns/extproc/my-eep/2",
			// no CustomName — not referenceable
		},
	}

	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{
			name:     "resolves labeled ext-proc in text format",
			format:   "[%START_TIME%] %EG_EXT_FILTER_STATE(auth:latency_ns)%",
			expected: "[%START_TIME%] %FILTER_STATE(envoy.filters.http.ext_proc/ns/extproc/my-eep/0:latency_ns)%",
		},
		{
			name:     "resolves second labeled ext-proc",
			format:   "%EG_EXT_FILTER_STATE(enrichment:grpc_status_code)%",
			expected: "%FILTER_STATE(envoy.filters.http.ext_proc/ns/extproc/my-eep/1:grpc_status_code)%",
		},
		{
			name:     "multiple operators in same string",
			format:   "%EG_EXT_FILTER_STATE(auth:latency_ns)% %EG_EXT_FILTER_STATE(enrichment:grpc_status_code)%",
			expected: "%FILTER_STATE(envoy.filters.http.ext_proc/ns/extproc/my-eep/0:latency_ns)% %FILTER_STATE(envoy.filters.http.ext_proc/ns/extproc/my-eep/1:grpc_status_code)%",
		},
		{
			name:     "passes through serialization type",
			format:   "%EG_EXT_FILTER_STATE(auth:latency_ns:TYPED)%",
			expected: "%FILTER_STATE(envoy.filters.http.ext_proc/ns/extproc/my-eep/0:latency_ns:TYPED)%",
		},
		{
			name:     "passes through serialization type and max length",
			format:   "%EG_EXT_FILTER_STATE(auth:latency_ns:PLAIN:64)%",
			expected: "%FILTER_STATE(envoy.filters.http.ext_proc/ns/extproc/my-eep/0:latency_ns:PLAIN:64)%",
		},
		{
			name:     "unknown label replaced with dash",
			format:   "%EG_EXT_FILTER_STATE(unknown:latency_ns)%",
			expected: "-",
		},
		{
			name:     "mixed: known and unknown labels in same string",
			format:   "%EG_EXT_FILTER_STATE(auth:latency_ns)% %EG_EXT_FILTER_STATE(unknown:grpc_status_code)%",
			expected: "%FILTER_STATE(envoy.filters.http.ext_proc/ns/extproc/my-eep/0:latency_ns)% -",
		},
		{
			name:     "no EG_EXT_FILTER_STATE operator — unchanged",
			format:   "[%START_TIME%] %REQ(:METHOD)%",
			expected: "[%START_TIME%] %REQ(:METHOD)%",
		},
		{
			name:     "empty format string",
			format:   "",
			expected: "",
		},
		{
			name:     "resolves with nil extProcs list — replaced with dash",
			format:   "%EG_EXT_FILTER_STATE(auth:latency_ns)%",
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eps := extProcs
			if tt.name == "resolves with nil extProcs list — replaced with dash" {
				eps = nil
			}
			got := resolveEGExtFilterStateOperators(tt.format, eps)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestCollectHCMExtensionsForLogExpansion(t *testing.T) {
	ep0 := ir.ExtProc{Name: "ep0", CustomName: "auth"}
	ep1 := ir.ExtProc{Name: "ep1", CustomName: "enrichment"}

	listener := &ir.HTTPListener{
		Routes: []*ir.HTTPRoute{
			{
				EnvoyExtensions: &ir.EnvoyExtensionFeatures{
					ExtProcs: []ir.ExtProc{ep0},
				},
			},
			{
				EnvoyExtensions: &ir.EnvoyExtensionFeatures{
					ExtProcs: []ir.ExtProc{ep0, ep1}, // ep0 duplicated across routes
				},
			},
			{
				// no ext-procs
			},
		},
	}

	got := collectHCMExtensionsForLogExpansion(listener)
	assert.Len(t, got, 2)
	names := []string{got[0].Name, got[1].Name}
	assert.Contains(t, names, "ep0")
	assert.Contains(t, names, "ep1")
}
