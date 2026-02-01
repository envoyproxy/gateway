// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
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
			name:          "single label not ok when in k8s",
			hostname:      "otel-tui",
			runningOnHost: false,
			expectedErr:   "hostname otel-tui should be a domain with at least two segments separated by dots",
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
