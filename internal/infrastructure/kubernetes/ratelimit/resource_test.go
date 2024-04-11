// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ratelimit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckTraceEndpointScheme(t *testing.T) {

	cases := []struct {
		caseName    string
		actualURL   string
		expectedURL string
	}{
		{
			caseName:    "normal url with http prefix",
			actualURL:   "http://collector.observability.svc.cluster.local:4318",
			expectedURL: "http://collector.observability.svc.cluster.local:4318",
		},
		{
			caseName:    "abnormal url without http prefix",
			actualURL:   "collector.observability.svc.cluster.local:4318",
			expectedURL: "http://collector.observability.svc.cluster.local:4318",
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			actual := checkTraceEndpointScheme(tc.actualURL)
			require.Equal(t, tc.expectedURL, actual)
		})
	}

}
