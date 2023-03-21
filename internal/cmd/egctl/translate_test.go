// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

func noopFilter(in string) string {
	return in
}

func fieldFilter(in string, output string, exp interface{}) string {
	var out []byte
	if output == jsonOutput {
		err := json.Unmarshal([]byte(in), &[]interface{}{exp})
		if err != nil {
			panic(err)
		}

		out, err = json.Marshal(exp)
		if err != nil {
			panic(err)
		}

	} else {
		err := yaml.Unmarshal([]byte(in), exp)
		if err != nil {
			panic(err)
		}

		out, err = yaml.Marshal(exp)
		if err != nil {
			panic(err)
		}
	}

	return string(out)
}

func gatewayAPIWithXdsYamlFilter(in string, exp interface{}) string {
	yamls := strings.SplitN(in, "---\n", 2)
	err := yaml.Unmarshal([]byte(yamls[0]), exp)
	if err != nil {
		panic(err)
	}

	out, err := yaml.Marshal(exp)
	if err != nil {
		panic(err)
	}

	return string(out) + "---\n" + yamls[1]
}

func TestTranslate(t *testing.T) {
	type ExpectHTTPRoutes struct {
		HTTPRoutes []struct {
			Metadata struct {
				Namespace string
				Name      string
			}
			Spec   interface{}
			Status struct {
				Parents []struct {
					Conditions []struct {
						Status string
						Reason string
					}
					ControllerName string
					ParentRef      struct {
						Name string
					}
				}
			}
		}
	}

	testCases := []struct {
		name         string
		from         string
		to           string
		output       string
		resourceType string
		extraArgs    []string
		expect       bool
		filterFunc   func(string) string
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
		{
			name:         "rejected-http-route",
			from:         "gateway-api",
			to:           "gateway-api",
			output:       yamlOutput,
			resourceType: string(RouteEnvoyConfigType),
			expect:       true,
			filterFunc: func(in string) string {
				// Need to filter out the fields we care about, otherwise the fields
				// will changed when we update the gatewayapi library
				return fieldFilter(in, yamlOutput, &ExpectHTTPRoutes{})
			},
		},
		{
			name:         "echo-gateway-api",
			from:         "gateway-api",
			to:           "gateway-api",
			output:       jsonOutput,
			resourceType: string(RouteEnvoyConfigType),
			expect:       true,
			filterFunc: func(in string) string {
				return fieldFilter(in, jsonOutput, &ExpectHTTPRoutes{})
			},
		},
		{
			name:         "echo-gateway-api",
			from:         "gateway-api",
			to:           "xds,gateway-api",
			output:       yamlOutput,
			resourceType: string(ClusterEnvoyConfigType),
			expect:       true,
			filterFunc: func(in string) string {
				return gatewayAPIWithXdsYamlFilter(in, &ExpectHTTPRoutes{})
			},
		},
		{
			name: "echo-gateway-api",
			from: "gateway-api",
			// ensure the order doesn't affect the output
			to:           "gateway-api,xds",
			output:       yamlOutput,
			resourceType: string(ClusterEnvoyConfigType),
			expect:       true,
			filterFunc: func(in string) string {
				return gatewayAPIWithXdsYamlFilter(in, &ExpectHTTPRoutes{})
			},
		},
		{
			name:         "multiple-xds",
			from:         "gateway-api",
			to:           "xds",
			output:       jsonOutput,
			resourceType: string(RouteEnvoyConfigType),
			expect:       true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		if tc.filterFunc == nil {
			tc.filterFunc = noopFilter
		}

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
				require.JSONEq(t, tc.filterFunc(requireTestDataOutFile(t, fn)),
					tc.filterFunc(string(out)), "failure in "+fn)
			} else {
				fn := tc.name + "." + resourceType + ".yaml"
				require.YAMLEq(t, tc.filterFunc(requireTestDataOutFile(t, fn)),
					tc.filterFunc(string(out)), "failure in "+fn)
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
