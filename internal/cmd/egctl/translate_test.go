// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslate(t *testing.T) {
	testCases := []struct {
		name         string
		from         string
		to           string
		output       string
		resourceType string
		extraArgs    []string
		expect       bool
	}{
		{
			name:   "from-gateway-api-to-xds",
			from:   "gateway-api",
			to:     "xds",
			output: jsonOutput,
			expect: true,
		},
		{
			name:   "from-gateway-api-to-xds",
			from:   "gateway-api",
			to:     "xds",
			output: yamlOutput,
			expect: true,
		},
		{
			name:   "from-gateway-api-to-xds",
			from:   "gateway-api",
			to:     "xds",
			expect: true,
		},
		{
			name:         "from-gateway-api-to-xds",
			from:         "gateway-api",
			to:           "xds",
			output:       yamlOutput,
			resourceType: "unknown",
			expect:       false,
		},
		{
			name:         "from-gateway-api-to-xds",
			from:         "gateway-api",
			to:           "xds",
			output:       yamlOutput,
			resourceType: string(AllEnvoyConfigType),
			expect:       true,
		},
		{
			name:         "from-gateway-api-to-xds",
			from:         "gateway-api",
			to:           "xds",
			output:       yamlOutput,
			resourceType: string(BootstrapEnvoyConfigType),
			expect:       true,
		},
		{
			name:         "from-gateway-api-to-xds",
			from:         "gateway-api",
			to:           "xds",
			output:       yamlOutput,
			resourceType: string(ClusterEnvoyConfigType),
			expect:       true,
		},
		{
			name:         "from-gateway-api-to-xds",
			from:         "gateway-api",
			to:           "xds",
			output:       yamlOutput,
			resourceType: string(ListenerEnvoyConfigType),
			expect:       true,
		},
		{
			name:         "from-gateway-api-to-xds",
			from:         "gateway-api",
			to:           "xds",
			output:       yamlOutput,
			resourceType: string(RouteEnvoyConfigType),
			expect:       true,
		},
		{
			name:      "default-resources",
			from:      "gateway-api",
			to:        "xds",
			expect:    true,
			extraArgs: []string{"--add-missing-resources"},
		},
		{
			name:         "quickstart",
			from:         "gateway-api",
			to:           "xds",
			output:       yamlOutput,
			resourceType: string(RouteEnvoyConfigType),
			expect:       true,
		},
		{
			name:         "from-gateway-api-to-xds",
			from:         "gateway-api",
			to:           "xds",
			output:       yamlOutput,
			resourceType: string(EndpointEnvoyConfigType),
			expect:       true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			b := bytes.NewBufferString("")
			root := NewTranslateCommand()
			root.SetOut(b)
			root.SetErr(b)
			args := []string{
				"translate",
				"--from",
				tc.from,
				"--to",
				tc.to,
				"--file",
				"testdata/translate/in/" + tc.name + ".yaml",
			}

			if tc.output == yamlOutput {
				args = append(args, "--output", yamlOutput)
			} else if tc.output == jsonOutput {
				args = append(args, "--output", jsonOutput)
			}

			var resourceType string
			if tc.resourceType == "" {
				resourceType = string(AllEnvoyConfigType)
			} else {
				resourceType = tc.resourceType
				args = append(args, "--type", tc.resourceType)
			}

			if len(tc.extraArgs) > 0 {
				args = append(args, tc.extraArgs...)
			}

			root.SetArgs(args)
			if tc.expect {
				assert.NoError(t, root.ExecuteContext(context.Background()))
			} else {
				assert.Error(t, root.ExecuteContext(context.Background()))
				return
			}

			out, err := io.ReadAll(b)
			assert.NoError(t, err)

			if tc.output == jsonOutput {
				fn := tc.name + "." + resourceType + ".json"
				require.JSONEq(t, requireTestDataOutFile(t, fn), string(out), "failure in "+fn)
			} else {
				fn := tc.name + "." + resourceType + ".yaml"
				require.YAMLEq(t, requireTestDataOutFile(t, fn), string(out), "failure in "+fn)
			}
		})
	}
}

func requireTestDataOutFile(t *testing.T, name ...string) string {
	t.Helper()
	elems := append([]string{"testdata", "translate", "out"}, name...)
	content, err := os.ReadFile(filepath.Join(elems...))
	require.NoError(t, err)
	return string(content)
}
