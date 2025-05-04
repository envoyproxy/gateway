// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package bootstrap

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/utils/test"
)

func TestApplyBootstrapConfig(t *testing.T) {
	str, _ := readTestData("enable-prometheus")
	cases := []struct {
		name             string
		boostrapConfig   *egv1a1.ProxyBootstrap
		defaultBootstrap string
	}{
		{
			name: "default",
			boostrapConfig: &egv1a1.ProxyBootstrap{
				Type: ptr.To(egv1a1.BootstrapTypeMerge),
			},
			defaultBootstrap: str,
		},
		{
			name: "merge-user-bootstrap",
			boostrapConfig: &egv1a1.ProxyBootstrap{
				Type: ptr.To(egv1a1.BootstrapTypeMerge),
			},
			defaultBootstrap: str,
		},
		{
			name: "stats_sinks",
			boostrapConfig: &egv1a1.ProxyBootstrap{
				Type: ptr.To(egv1a1.BootstrapTypeMerge),
			},
			defaultBootstrap: str,
		},
		{
			name: "patch-global-config",
			boostrapConfig: &egv1a1.ProxyBootstrap{
				Type: ptr.To(egv1a1.BootstrapTypeJSONPatch),
			},
			defaultBootstrap: str,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in, err := loadData(tc.name, "in")
			require.NoError(t, err)

			switch *tc.boostrapConfig.Type {
			case egv1a1.BootstrapTypeJSONPatch:
				err = yaml.Unmarshal([]byte(in), &tc.boostrapConfig.JSONPatches)
				require.NoError(t, err)
			default:
				tc.boostrapConfig.Value = &in
			}

			data, err := ApplyBootstrapConfig(tc.boostrapConfig, tc.defaultBootstrap)
			require.NoError(t, err)

			if test.OverrideTestData() {
				// nolint:gosec
				err = os.WriteFile(path.Join("testdata", "merge", fmt.Sprintf("%s.out.yaml", tc.name)), []byte(data), 0o644)
				require.NoError(t, err)
				return
			}

			expected, err := loadData(tc.name, "out")
			require.NoError(t, err)
			require.Equal(t, expected, data)
		})
	}
}

func loadData(caseName, inOrOut string) (string, error) {
	filename := path.Join("testdata", "merge", fmt.Sprintf("%s.%s.yaml", caseName, inOrOut))
	b, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
