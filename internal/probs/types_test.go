// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package probs

import (
	"testing"
)

func TestSetIndicator(t *testing.T) {
	// Initialize xdsProb before each test
	xdsProb := &xdsHealthProb{
		indicators: make(map[string]bool),
		isReady:    false,
	}

	testCases := []struct {
		name       string
		indicators []string
		expect     bool
	}{
		{
			name:       "No indicators set",
			indicators: []string{},
			expect:     false,
		},
		{
			name:       "GeneratedNewXdsSnapshot only",
			indicators: []string{GeneratedNewXdsSnapshot},
			expect:     true,
		},
		{
			name:       "StartedXdsServer only",
			indicators: []string{StartedXdsServer},
			expect:     false,
		},
		{
			name:       "Both GeneratedNewXdsSnapshot and StartedXdsServer",
			indicators: []string{GeneratedNewXdsSnapshot, StartedXdsServer},
			expect:     true,
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset xdsProb before each test
			xdsProb.indicators = make(map[string]bool)
			xdsProb.isReady = false

			// Set each indicator
			for _, indicator := range tc.indicators {
				xdsProb.SetIndicator(indicator)
			}

			// Check if isReady matches the expected value
			if xdsProb.isReady != tc.expect {
				t.Errorf("Expected isReady to be %v after setting indicators %v, got %v", tc.expect, tc.indicators, xdsProb.isReady)
			}
		})
	}
}
