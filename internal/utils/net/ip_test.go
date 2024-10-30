// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package net

import (
	"testing"

	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

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

func TestPreferIPFamily(t *testing.T) {
	cases := []struct {
		name       string
		ipv6First  bool
		envoyProxy *egv1a1.EnvoyProxy
		expected   egv1a1.IPFamily
	}{
		{
			name:       "ipv6First=true,envoyProxy=nil",
			ipv6First:  true,
			envoyProxy: nil,
			expected:   egv1a1.IPv6,
		},
		{
			name:       "ipv6First=true,envoyProxy=ipv4",
			ipv6First:  true,
			envoyProxy: &egv1a1.EnvoyProxy{Spec: egv1a1.EnvoyProxySpec{IPFamily: ptr.To(egv1a1.IPv4)}},
			expected:   egv1a1.IPv4,
		},
		{
			name:       "ipv6First=false,envoyProxy=nil",
			ipv6First:  false,
			envoyProxy: nil,
			expected:   egv1a1.IPv4,
		},
		{
			name:       "ipv6First=false,envoyProxy=IPv6",
			ipv6First:  true,
			envoyProxy: &egv1a1.EnvoyProxy{Spec: egv1a1.EnvoyProxySpec{IPFamily: ptr.To(egv1a1.IPv6)}},
			expected:   egv1a1.IPv6,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := PreferIPFamily(tc.ipv6First, tc.envoyProxy)
			if actual != tc.expected {
				t.Errorf("PreferIPFamily(%t, %v) = %v; expected %v", tc.ipv6First, tc.envoyProxy, actual, tc.expected)
			}
		})
	}
}
