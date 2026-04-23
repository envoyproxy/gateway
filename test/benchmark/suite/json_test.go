// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark

package suite

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetLatencyPercentiles(t *testing.T) {
	result := fakeCaseResult()
	require.NotNil(t, result)
	require.NotEmpty(t, result.Statistics)

	percentiles := getLatencyPercentiles(result.Statistics[1].Percentiles)

	require.InDelta(t, 0.438127, percentiles.P50, 1e-9)
	require.InDelta(t, 0.546047, percentiles.P75, 1e-9)
	require.InDelta(t, 0.584511, percentiles.P80, 1e-9)
	require.InDelta(t, 0.751775, percentiles.P90, 1e-9)
	require.InDelta(t, 1.100415, percentiles.P95, 1e-9)
	require.InDelta(t, 55.601151, percentiles.P99, 1e-9)
	require.InDelta(t, 60.461055, percentiles.P999, 1e-9)
}
