// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestValidateHostnameRunningOnHost(t *testing.T) {
	cases := []struct {
		name          string
		hostname      string
		runningOnHost bool
		expectedErr   string
	}{
		{
			name:          "domain ok in any case",
			hostname:      "httpbin.org",
			runningOnHost: false,
		},
		{
			name:          "absolute domain ok in any case",
			hostname:      "httpbin.org.",
			runningOnHost: false,
		},
		{
			name:          "single label ok when in k8s",
			hostname:      "ml-service",
			runningOnHost: false,
		},
		{
			name:          "single label ok when running on host",
			hostname:      "otel-tui",
			runningOnHost: true,
		},
		{
			name:          "IP not ok in any case",
			hostname:      "127.0.0.1",
			runningOnHost: true,
			expectedErr:   "hostname 127.0.0.1 is an IP address",
		},
		{
			name:          "localhost not ok when in k8s",
			hostname:      "localhost",
			runningOnHost: false,
			expectedErr:   "hostname localhost is only supported when using the Host infrastructure",
		},
		{
			name:          "absolute localhost not ok when in k8s",
			hostname:      "localhost.",
			runningOnHost: false,
			expectedErr:   "hostname localhost. is only supported when using the Host infrastructure",
		},
		{
			name:          "localhost ok when running on host",
			hostname:      "localhost",
			runningOnHost: true,
		},
		{
			name:          "absolute localhost ok when running on host",
			hostname:      "localhost.",
			runningOnHost: true,
		},
		{
			name:          "underscore not ok when in k8s",
			hostname:      "bad_name",
			runningOnHost: false,
			expectedErr:   "hostname bad_name is not a valid hostname",
		},
		{
			name:          "empty not ok when in k8s",
			hostname:      "",
			runningOnHost: false,
			expectedErr:   "hostname  is not a valid hostname",
		},
		{
			name:          "wildcard not ok when in k8s",
			hostname:      "*.foo.com",
			runningOnHost: false,
			expectedErr:   "hostname *.foo.com is not a valid hostname",
		},
		{
			name:          "IP not ok when in k8s",
			hostname:      "127.0.0.1",
			runningOnHost: false,
			expectedErr:   "hostname 127.0.0.1 is an IP address",
		},
		{
			name:          "63 character single label ok in k8s",
			hostname:      strings.Repeat("a", 63),
			runningOnHost: false,
		},
		{
			name:          "64 character single label not ok in k8s",
			hostname:      strings.Repeat("a", 64),
			runningOnHost: false,
			expectedErr:   "hostname " + strings.Repeat("a", 64) + " is not a valid hostname",
		},
		{
			name:          "253 character single label not ok in k8s",
			hostname:      strings.Repeat("a", 253),
			runningOnHost: false,
			expectedErr:   "hostname " + strings.Repeat("a", 253) + " is not a valid hostname",
		},
		{
			name:          "63 character absolute single label ok in k8s",
			hostname:      strings.Repeat("a", 63) + ".",
			runningOnHost: false,
		},
		{
			name:          "64 character absolute single label not ok in k8s",
			hostname:      strings.Repeat("a", 64) + ".",
			runningOnHost: false,
			expectedErr:   "hostname " + strings.Repeat("a", 64) + ". is not a valid hostname",
		},
		{
			name:          "253 character absolute single label not ok in k8s",
			hostname:      strings.Repeat("a", 253) + ".",
			runningOnHost: false,
			expectedErr:   "hostname " + strings.Repeat("a", 253) + ". is not a valid hostname",
		},
		{
			name:          "63 character single label ok when running on host",
			hostname:      strings.Repeat("a", 63),
			runningOnHost: true,
		},
		{
			name:          "64 character single label not ok when running on host",
			hostname:      strings.Repeat("a", 64),
			runningOnHost: true,
			expectedErr:   "hostname " + strings.Repeat("a", 64) + " is not a valid hostname",
		},
		{
			name:          "multi label domain ok in k8s",
			hostname:      strings.Repeat("a", 63) + ".example.com",
			runningOnHost: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateHostname(tc.hostname, "hostname", tc.runningOnHost)
			if tc.expectedErr == "" {
				require.Nil(t, err)
				return
			}
			require.EqualError(t, err, tc.expectedErr)
		})
	}
}

func TestValidateIPRunningOnHost(t *testing.T) {
	cases := []struct {
		name          string
		address       string
		runningOnHost bool
		expectedErr   string
	}{
		{
			name:          "address ok in any case",
			address:       "10.0.0.1",
			runningOnHost: false,
		},
		{
			name:          "loopback not ok when in k8s",
			address:       "127.0.0.1",
			runningOnHost: false,
			expectedErr:   "IP address 127.0.0.1 in the loopback range is only supported when using the Host infrastructure",
		},
		{
			name:          "loopback ok when running on host",
			address:       "127.0.0.1",
			runningOnHost: true,
		},
		{
			name:          "invalid IP not ok in any case",
			address:       "300.0.0.1",
			runningOnHost: true,
			expectedErr:   "IP address 300.0.0.1 is invalid",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateIP(&egv1a1.IPEndpoint{Address: tc.address}, tc.runningOnHost)
			if tc.expectedErr == "" {
				require.Nil(t, err)
				return
			}
			require.EqualError(t, err, tc.expectedErr)
		})
	}
}
