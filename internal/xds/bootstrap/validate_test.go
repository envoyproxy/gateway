// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package bootstrap

import (
	// Register embed
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

var (
	//go:embed testdata/validate/valid-user-bootstrap.yaml
	validUserBootstrap string
	//go:embed testdata/validate/missing-admin-address-user-bootstrap.yaml
	missingAdminAddressUserBootstrap string
	//go:embed testdata/validate/different-dynamic-resources-user-bootstrap.yaml
	differentDynamicResourcesUserBootstrap string
	//go:embed testdata/validate/different-xds-cluster-address-bootstrap.yaml
	differentXdsClusterAddressBootstrap string
)

func TestValidateBootstrap(t *testing.T) {
	testCases := []struct {
		name      string
		bootstrap *egv1a1.ProxyBootstrap
		expected  bool
	}{
		{
			name: "valid user bootstrap replace type",
			bootstrap: &egv1a1.ProxyBootstrap{
				Value: &validUserBootstrap,
			},
			expected: true,
		},
		{
			name: "user bootstrap with missing admin address",
			bootstrap: &egv1a1.ProxyBootstrap{
				Value: &missingAdminAddressUserBootstrap,
			},
			expected: false,
		},
		{
			name: "user bootstrap with different dynamic resources",
			bootstrap: &egv1a1.ProxyBootstrap{
				Value: &differentDynamicResourcesUserBootstrap,
			},
			expected: false,
		},
		{
			name: "user bootstrap with different xds_cluster endpoint",
			bootstrap: &egv1a1.ProxyBootstrap{
				Value: &differentXdsClusterAddressBootstrap,
			},
			expected: false,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.bootstrap)
			if tc.expected {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
