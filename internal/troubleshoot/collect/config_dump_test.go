// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package collect

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfigDump_getTimeout(t *testing.T) {
	tests := []struct {
		name     string
		timeout  time.Duration
		expected time.Duration
	}{
		{
			name:     "returns default when timeout is zero",
			timeout:  0,
			expected: DefaultConfigDumpTimeout,
		},
		{
			name:     "returns default when timeout is negative",
			timeout:  -5 * time.Second,
			expected: DefaultConfigDumpTimeout,
		},
		{
			name:     "returns configured timeout when set",
			timeout:  60 * time.Second,
			expected: 60 * time.Second,
		},
		{
			name:     "returns configured timeout when less than default",
			timeout:  10 * time.Second,
			expected: 10 * time.Second,
		},
		{
			name:     "returns configured timeout when greater than default",
			timeout:  120 * time.Second,
			expected: 120 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cd := ConfigDump{
				Timeout: tt.timeout,
			}
			got := cd.getTimeout()
			assert.Equal(t, tt.expected, got, "getTimeout() should return expected duration")
		})
	}
}

func TestDefaultConfigDumpTimeout(t *testing.T) {
	// Verify the default timeout constant is set to the expected value
	expected := 30 * time.Second
	assert.Equal(t, expected, DefaultConfigDumpTimeout, "DefaultConfigDumpTimeout should be 30 seconds")
}

func TestConfigDump_TimeoutField(t *testing.T) {
	// Test that the Timeout field can be set and retrieved
	cd := ConfigDump{
		Timeout: 45 * time.Second,
	}
	assert.Equal(t, 45*time.Second, cd.Timeout, "Timeout field should be settable")
}
