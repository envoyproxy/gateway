// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package registry

import (
	"testing"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestGetExtensionServerAddress(t *testing.T) {
	tests := []struct {
		Name     string
		Service  *egv1a1.ExtensionService
		Expected string
	}{
		{
			Name: "has an FQDN",
			Service: &egv1a1.ExtensionService{
				BackendEndpoint: egv1a1.BackendEndpoint{
					FQDN: &egv1a1.FQDNEndpoint{
						Hostname: "extserver.svc.cluster.local",
						Port:     5050,
					},
				},
			},
			Expected: "extserver.svc.cluster.local:5050",
		},
		{
			Name: "has an IP",
			Service: &egv1a1.ExtensionService{
				BackendEndpoint: egv1a1.BackendEndpoint{
					IP: &egv1a1.IPEndpoint{
						Address: "10.10.10.10",
						Port:    5050,
					},
				},
			},
			Expected: "10.10.10.10:5050",
		},
		{
			Name: "has a Unix path",
			Service: &egv1a1.ExtensionService{
				BackendEndpoint: egv1a1.BackendEndpoint{
					Unix: &egv1a1.UnixSocket{
						Path: "/some/path",
					},
				},
			},
			Expected: "unix:///some/path",
		},
		{
			Name: "has a Unix path",
			Service: &egv1a1.ExtensionService{
				Host: "foo.bar",
				Port: 5050,
			},
			Expected: "foo.bar:5050",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			out := getExtensionServerAddress(tc.Service)
			require.Equal(t, tc.Expected, out)
		})
	}
}
