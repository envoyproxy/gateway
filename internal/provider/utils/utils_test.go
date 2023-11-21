// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetHashedName(t *testing.T) {
	testCases := []struct {
		name     string
		nsName   string
		length   int
		expected string
	}{
		{"test default name", "http", 6, "http-e0603c49"},
		{"test removing trailing slash", "namespace/name", 10, "namespace-18a6500f"},
		{"test removing trailing hyphen", "envoy-gateway-system/eg/http", 6, "envoy-2ecf157b"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := GetHashedName(tc.nsName, tc.length)
			require.Equal(t, tc.expected, result, "Result does not match expected string")
		})
	}
}
