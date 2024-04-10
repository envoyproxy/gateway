// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package bootstrap

import (
	"flag"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

var (
	overrideTestData = flag.Bool("override-testdata", false, "if override the test output data.")
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
			name: "stats_sinks",
			boostrapConfig: &egv1a1.ProxyBootstrap{
				Type: ptr.To(egv1a1.BootstrapTypeMerge),
			},
			defaultBootstrap: str,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in, err := loadData(tc.name, "in")
			require.NoError(t, err)

			tc.boostrapConfig.Value = in
			data, err := ApplyBootstrapConfig(tc.boostrapConfig, tc.defaultBootstrap)
			require.NoError(t, err)

			if *overrideTestData {
				// nolint:gosec
				err = os.WriteFile(path.Join("testdata", "merge", fmt.Sprintf("%s.out.yaml", tc.name)), []byte(data), 0644)
				require.NoError(t, err)
				return
			}

			expected, err := loadData(tc.name, "out")
			require.NoError(t, err)
			require.Equal(t, expected, data)
		})
	}
}

func loadData(caseName string, inOrOut string) (string, error) {
	filename := path.Join("testdata", "merge", fmt.Sprintf("%s.%s.yaml", caseName, inOrOut))
	b, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
