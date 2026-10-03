// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"testing"
	"weak"

	"github.com/stretchr/testify/require"
	"github.com/telepresenceio/watchable"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
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

func TestCoalesceUpdatesReleasesDiscardedValues(t *testing.T) {
	t.Parallel()
	logger := logging.DefaultLogger(io.Discard, egv1a1.LogLevelError)
	updates, discarded := func() ([]watchable.Update[string, *[1024]byte], weak.Pointer[[1024]byte]) {
		old, latest := new([1024]byte), new([1024]byte)
		latest[0] = 1
		ref := weak.Make(old)
		return coalesceUpdates(logger, []watchable.Update[string, *[1024]byte]{
			{Key: "gateway", Value: old},
			{Key: "gateway", Value: latest},
		}), ref
	}()

	runtime.GC()
	require.Nil(t, discarded.Value())
	require.Len(t, updates, 1)
	require.Equal(t, byte(1), updates[0].Value[0])
	runtime.KeepAlive(updates)
}

func BenchmarkCoalesceUpdates(b *testing.B) {
	logger := logging.DefaultLogger(io.Discard, egv1a1.LogLevelError)
	for _, count := range []int{1, 8, 32, 128} {
		keyCounts := []int{1}
		if count > 1 {
			keyCounts = append(keyCounts, count)
		}
		for _, keys := range keyCounts {
			b.Run(fmt.Sprintf("updates=%d/keys=%d", count, keys), func(b *testing.B) {
				input := make([]watchable.Update[int, *int], count)
				for i := range input {
					input[i] = watchable.Update[int, *int]{Key: i % keys, Value: new(i)}
				}
				b.ReportAllocs()
				for b.Loop() {
					updates := append([]watchable.Update[int, *int](nil), input...)
					runtime.KeepAlive(coalesceUpdates(logger, updates))
				}
			})
		}
	}
}
