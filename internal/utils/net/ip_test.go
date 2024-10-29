// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package net

import "testing"

func TestIsIPv6(t *testing.T) {
	cases := []struct {
		ip       string
		expected bool
	}{
		{
			ip:       "",
			expected: false,
		},
		{
			ip:       "127.0.0.1",
			expected: false,
		},
		{
			ip:       "::1",
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.ip, func(t *testing.T) {
			actual := IsIPv6(tc.ip)
			if actual != tc.expected {
				t.Errorf("IsIPv6(%s) = %t; expected %t", tc.ip, actual, tc.expected)
			}
		})
	}
}
