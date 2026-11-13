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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

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

const xdsClusterPath = "$.static_resources.clusters[?(@.name=='xds_cluster')]"

func newJSONPatchBootstrap(patches ...egv1a1.JSONPatchOperation) *egv1a1.ProxyBootstrap {
	return &egv1a1.ProxyBootstrap{
		Type:        new(egv1a1.BootstrapTypeJSONPatch),
		JSONPatches: patches,
	}
}

func patchOp(op egv1a1.JSONPatchOperationType, path, value string) egv1a1.JSONPatchOperation {
	p := egv1a1.JSONPatchOperation{Op: op, Path: &path}
	if value != "" {
		p.Value = &apiextensionsv1.JSON{Raw: []byte(value)}
	}
	return p
}

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
		{
			name: "valid json patch replaces overload manager refresh interval",
			bootstrap: newJSONPatchBootstrap(
				patchOp("replace", "/overload_manager/refresh_interval", `"0.5s"`),
			),
			expected: true,
		},
		{
			name: "valid json patch adds stats config field",
			bootstrap: newJSONPatchBootstrap(
				patchOp("add", "/stats_config/use_all_default_tags", `false`),
			),
			expected: true,
		},
		{
			name: "json patch removes dynamic resources",
			bootstrap: newJSONPatchBootstrap(
				patchOp("remove", "/dynamic_resources", ""),
			),
			expected: false,
		},
		{
			name: "json patch changes dynamic resources api type",
			bootstrap: newJSONPatchBootstrap(
				patchOp("replace", "/dynamic_resources/ads_config/api_type", `"GRPC"`),
			),
			expected: false,
		},
		{
			name: "json patch changes xds_cluster address",
			bootstrap: newJSONPatchBootstrap(egv1a1.JSONPatchOperation{
				Op:       "replace",
				JSONPath: new(xdsClusterPath),
				Path:     new("/load_assignment/endpoints/0/lb_endpoints/0/endpoint/address/socket_address/address"),
				Value:    &apiextensionsv1.JSON{Raw: []byte(`"example.com"`)},
			}),
			expected: false,
		},
		{
			name: "json patch with invalid path",
			bootstrap: newJSONPatchBootstrap(
				patchOp("remove", "/no_such_field", ""),
			),
			expected: false,
		},
		{
			name: "valid merge type",
			bootstrap: &egv1a1.ProxyBootstrap{
				Type:  new(egv1a1.BootstrapTypeMerge),
				Value: new("layered_runtime:\n  layers:\n  - name: extra\n    static_layer:\n      foo: bar\n"),
			},
			expected: true,
		},
		{
			name: "merge type overriding dynamic resources",
			bootstrap: &egv1a1.ProxyBootstrap{
				Type:  new(egv1a1.BootstrapTypeMerge),
				Value: new("dynamic_resources:\n  ads_config:\n    api_type: GRPC\n"),
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

			viaValidator := Validator{}.Validate(tc.bootstrap)
			if err == nil {
				require.NoError(t, viaValidator)
			} else {
				require.EqualError(t, viaValidator, err.Error())
			}
		})
	}
}
