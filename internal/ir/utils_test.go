// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

import (
	"testing"

	"github.com/stretchr/testify/assert"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestMapToSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		expected []MapEntry
	}{
		{
			name:     "nil map",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty map",
			input:    map[string]string{},
			expected: nil,
		},
		{
			name:  "single entry",
			input: map[string]string{"key1": "val1"},
			expected: []MapEntry{
				{Key: "key1", Value: "val1"},
			},
		},
		{
			name: "multiple entries sorted by key",
			input: map[string]string{
				"charlie": "3",
				"alpha":   "1",
				"bravo":   "2",
			},
			expected: []MapEntry{
				{Key: "alpha", Value: "1"},
				{Key: "bravo", Value: "2"},
				{Key: "charlie", Value: "3"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := MapToSlice(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSliceToMap(t *testing.T) {
	tests := []struct {
		name     string
		input    []MapEntry
		expected map[string]string
	}{
		{
			name:     "nil slice",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty slice",
			input:    []MapEntry{},
			expected: nil,
		},
		{
			name: "single entry",
			input: []MapEntry{
				{Key: "key1", Value: "val1"},
			},
			expected: map[string]string{"key1": "val1"},
		},
		{
			name: "multiple entries",
			input: []MapEntry{
				{Key: "alpha", Value: "1"},
				{Key: "bravo", Value: "2"},
			},
			expected: map[string]string{
				"alpha": "1",
				"bravo": "2",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := SliceToMap(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCustomTagMapToSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]egv1a1.CustomTag
		expected []CustomTagMapEntry
	}{
		{
			name:     "nil map",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty map",
			input:    map[string]egv1a1.CustomTag{},
			expected: nil,
		},
		{
			name: "multiple entries sorted by key",
			input: map[string]egv1a1.CustomTag{
				"beta": {
					Type: egv1a1.CustomTagTypeLiteral,
					Literal: &egv1a1.LiteralCustomTag{
						Value: "lit-val",
					},
				},
				"alpha": {
					Type: egv1a1.CustomTagTypeEnvironment,
					Environment: &egv1a1.EnvironmentCustomTag{
						Name: "ENV_VAR",
					},
				},
			},
			expected: []CustomTagMapEntry{
				{
					Key: "alpha",
					Value: egv1a1.CustomTag{
						Type: egv1a1.CustomTagTypeEnvironment,
						Environment: &egv1a1.EnvironmentCustomTag{
							Name: "ENV_VAR",
						},
					},
				},
				{
					Key: "beta",
					Value: egv1a1.CustomTag{
						Type: egv1a1.CustomTagTypeLiteral,
						Literal: &egv1a1.LiteralCustomTag{
							Value: "lit-val",
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := CustomTagMapToSlice(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMapToSliceRoundTrip(t *testing.T) {
	input := map[string]string{
		"host":   "example.com",
		"method": "GET",
		"path":   "/api",
	}
	result := SliceToMap(MapToSlice(input))
	assert.Equal(t, input, result)
}
