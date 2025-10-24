// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/telepresenceio/watchable"
)

func TestMergeUpdates(t *testing.T) {
	t.Parallel()

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
			name: "latest update per key delete state wins",
			input: []watchable.Update[string, int]{
				{Key: "foo", Value: 1},
				{Key: "bar", Delete: true, Value: 10},
				{Key: "baz", Value: 5},
				{Key: "bar", Delete: true, Value: 11},
				{Key: "foo", Value: 2},
			},
			expected: []watchable.Update[string, int]{
				{Key: "baz", Value: 5},
				{Key: "bar", Delete: true, Value: 11},
				{Key: "foo", Value: 2},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := coalesceUpdates("test-runner", tc.input)
			require.Equal(t, tc.expected, actual)
		})
	}
}
