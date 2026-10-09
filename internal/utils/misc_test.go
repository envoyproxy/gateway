// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/validation"
)

func TestLabelValue(t *testing.T) {
	short := "gateway-1"
	require.Equal(t, short, LabelValue(short))

	atLimit := strings.Repeat("a", 63)
	require.Equal(t, atLimit, LabelValue(atLimit))

	long := strings.Repeat("a", 64)
	got := LabelValue(long)
	require.NotEqual(t, long, got)
	require.Equal(t, GetHashedName(long, 54), got)
	require.LessOrEqual(t, len(got), 63)
	require.Empty(t, validation.IsValidLabelValue(got))

	other := strings.Repeat("b", 64)
	require.NotEqual(t, LabelValue(long), LabelValue(other))
	require.Equal(t, LabelValue(long), LabelValue(long))
}

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
		t.Run(tc.name, func(t *testing.T) {
			result := GetHashedName(tc.nsName, tc.length)
			require.Equal(t, tc.expected, result, "Result does not match expected string")
		})
	}
}
