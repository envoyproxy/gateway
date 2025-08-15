// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package labels

import "testing"

func TestMatches(t *testing.T) {
	cases := []struct {
		l, r     map[string]string
		expected bool
	}{
		{
			l:        map[string]string{"app": "test"},
			r:        map[string]string{"app": "test"},
			expected: true,
		},
		{
			l:        map[string]string{"app": "test"},
			r:        map[string]string{"app": "test", "env": "prod"},
			expected: true,
		},
		{
			l:        map[string]string{"app": "test", "env": "prod"},
			r:        map[string]string{"app": "test"},
			expected: false,
		},
		{
			l:        nil,
			r:        map[string]string{"app": "test"},
			expected: false,
		},
		{
			l:        map[string]string{},
			r:        map[string]string{"app": "test"},
			expected: true,
		},
		{
			l: map[string]string{"app": "test"},
			r: map[string]string{},

			expected: false,
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			match, err := Matches(c.l, c.r)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if match != c.expected {
				t.Errorf("expected %v, got %v", c.expected, match)
			}
		})
	}
}
