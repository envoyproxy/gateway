// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package tempopb

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmarshal(t *testing.T) {
	cases := []struct {
		data     string
		expected *SearchResponse
	}{
		{
			data: `{
    "traces": [
        {
            "traceID": "cfd56d607ee974253874593cd8724c4",
            "rootServiceName": "same-namespace.gateway-conformance-infra",
            "rootTraceName": "ingress",
            "startTimeUnixNano": "1725550510115558000",
            "durationMs": 2
        }
    ],
    "metrics": {
        "inspectedTraces": 1,
        "inspectedBytes": "7692"
    }
}`,
			expected: &SearchResponse{
				Traces: []*TraceSearchMetadata{
					{
						TraceID:           "cfd56d607ee974253874593cd8724c4",
						RootServiceName:   "same-namespace.gateway-conformance-infra",
						RootTraceName:     "ingress",
						StartTimeUnixNano: "1725550510115558000",
						DurationMs:        2,
					},
				},
				Metrics: &SearchMetrics{
					InspectedTraces: 1,
					InspectedBytes:  "7692",
				},
			},
		},
	}

	for _, tc := range cases {
		got := &SearchResponse{}
		err := json.Unmarshal([]byte(tc.data), got)
		require.NoError(t, err)
		require.Equal(t, tc.expected, got)
	}
}
