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

func TestContainsAllLabels(t *testing.T) {
	type args struct {
		labels        map[string]string
		labelsToCheck []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"test all labels present", args{map[string]string{"label1": "foo", "label2": "bar"}, []string{"label1", "label2"}}, true},
		{"test some labels missing", args{map[string]string{"label1": "foo", "label2": "bar"}, []string{"label1", "label3"}}, false},
		{"test empty map", args{map[string]string{}, []string{"label1", "label2"}}, false},
		{"test empty labelsToCheck", args{map[string]string{"label1": "foo", "label2": "bar"}, []string{}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsAllLabels(tt.args.labels, tt.args.labelsToCheck); got != tt.want {
				t.Errorf("ContainsAllLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
