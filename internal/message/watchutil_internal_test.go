// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/telepresenceio/watchable"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/metrics"
)

func TestCoalesceUpdates(t *testing.T) {
	t.Parallel()
	logger := logging.NewLogger(os.Stdout, egv1a1.DefaultEnvoyGatewayLogging())
	tests := []struct {
		name     string
		input    []watchable.Update[string, int]
		expected []watchable.Update[string, int]
	}{
		{
			name:     "empty input returns nil",
			input:    []watchable.Update[string, int]{},
			expected: []watchable.Update[string, int]{},
		},
		{
			name: "simple updates without repeats",
			input: []watchable.Update[string, int]{
				{Key: "foo", Value: 1},
				{Key: "bar", Value: 2},
				{Key: "baz", Value: 3},
			},
			expected: []watchable.Update[string, int]{
				{Key: "foo", Value: 1},
				{Key: "bar", Value: 2},
				{Key: "baz", Value: 3},
			},
		},
		{
			name: "latest update per key wins",
			input: []watchable.Update[string, int]{
				{Key: "foo", Value: 1},
				{Key: "bar", Delete: true, Value: 2},
				{Key: "baz", Value: 3},
				{Key: "bar", Value: 4},
				{Key: "foo", Value: 5},
				{Key: "baz", Delete: true, Value: 6},
				{Key: "bar", Value: 7},
			},
			expected: []watchable.Update[string, int]{
				{Key: "foo", Value: 5},
				{Key: "baz", Delete: true, Value: 6},
				{Key: "bar", Value: 7},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := coalesceUpdates(logger, tc.input)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func collectWatchableDepth(t *testing.T, reader *sdkmetric.ManualReader) float64 {
	t.Helper()

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &rm))

	for _, scopeMetric := range rm.ScopeMetrics {
		for _, metric := range scopeMetric.Metrics {
			if metric.Name != "watchable_depth" {
				continue
			}

			gauge, ok := metric.Data.(metricdata.Gauge[float64])
			require.True(t, ok, "watchable_depth should export a float64 gauge")
			require.Len(t, gauge.DataPoints, 1)

			return gauge.DataPoints[0].Value
		}
	}

	t.Fatal("watchable_depth was never recorded")
	return 0
}

// handleSnapshot runs HandleSubscription over a single snapshot carrying the
// given updates, and returns the depth recorded for it. The channel is
// unbuffered, matching what watchable.Map.Subscribe returns, so reading the
// depth off the channel rather than the snapshot yields 0 here too.
func handleSnapshot(t *testing.T, reader *sdkmetric.ManualReader, updates []watchable.Update[string, int]) float64 {
	t.Helper()

	ch := make(chan watchable.Snapshot[string, int])
	go func() {
		defer close(ch)
		// Consumed as the initial state, before the update loop.
		ch <- watchable.Snapshot[string, int]{State: map[string]int{}}
		ch <- watchable.Snapshot[string, int]{State: map[string]int{}, Updates: updates}
	}()

	HandleSubscription(
		logging.NewLogger(t.Output(), egv1a1.DefaultEnvoyGatewayLogging()),
		Metadata{Runner: "demo", Message: "demo"},
		ch,
		func(_ Update[string, int], _ chan error) {},
	)

	return collectWatchableDepth(t, reader)
}

// TestHandleSubscriptionRecordsCoalescedUpdateCount pins watchable_depth to the
// number of updates in the snapshot being handled. Recording len(subscription)
// instead reports 0 forever, because watchable's Subscribe returns an
// unbuffered channel.
//
// watchable_depth is an async observable gauge, so both scenarios share one
// meter provider and run in sequence; installing a new provider per case would
// drop the callback registration.
func TestHandleSubscriptionRecordsCoalescedUpdateCount(t *testing.T) {
	// The first SetMeterProvider in a process does not reach an async instrument
	// created immediately after it, so resolve the global delegate first.
	otel.SetMeterProvider(noop.NewMeterProvider())

	reader := sdkmetric.NewManualReader()
	otel.SetMeterProvider(sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)))

	previousDepth := watchableDepth
	watchableDepth = metrics.NewGauge(
		"watchable_depth",
		"Number of updates coalesced into the snapshot being handled.",
	)
	t.Cleanup(func() {
		watchableDepth = previousDepth
		otel.SetMeterProvider(noop.NewMeterProvider())
	})

	backlog := []watchable.Update[string, int]{
		{Key: "foo", Value: 1},
		{Key: "bar", Value: 2},
		{Key: "baz", Value: 3},
	}
	require.Equal(t, float64(3), handleSnapshot(t, reader, backlog), "a backlog of three updates")

	single := []watchable.Update[string, int]{{Key: "foo", Value: 1}}
	require.Equal(t, float64(1), handleSnapshot(t, reader, single), "a single update")
}
