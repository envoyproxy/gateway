// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/internal/utils/field"
	"github.com/envoyproxy/gateway/internal/utils/file"
)

var (
	overrideTestData = flag.Bool("override-testdata", false, "if override the test output data.")
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
			name:   "jwt-single-route-single-match-to-xds",
			from:   "gateway-api",
			to:     "xds",
			output: jsonOutput,
			expect: true,
		},
		{
			name:   "jwt-single-route-single-match-to-xds",
			from:   "gateway-api",
			to:     "xds",
			output: yamlOutput,
			expect: true,
		},
		{
			name:   "jwt-single-route-single-match-to-xds",
			from:   "gateway-api",
			to:     "xds",
			expect: true,
		},
		{
			name:         "jwt-single-route-single-match-to-xds",
			from:         "gateway-api",
			to:           "xds",
			output:       yamlOutput,
			resourceType: "unknown",
			expect:       false,
		},
		{
			name:         "jwt-single-route-single-match-to-xds",
			from:         "gateway-api",
			to:           "xds",
			output:       yamlOutput,
			resourceType: string(AllEnvoyConfigType),
			expect:       true,
		},
		{
			name:         "jwt-single-route-single-match-to-xds",
			from:         "gateway-api",
			to:           "xds",
			output:       yamlOutput,
			resourceType: string(BootstrapEnvoyConfigType),
			expect:       true,
		},
		{
			name:         "jwt-single-route-single-match-to-xds",
			from:         "gateway-api",
			to:           "xds",
			output:       yamlOutput,
			resourceType: string(ClusterEnvoyConfigType),
			expect:       true,
		},
		{
			name:         "jwt-single-route-single-match-to-xds",
			from:         "gateway-api",
			to:           "xds",
			output:       yamlOutput,
			resourceType: string(ListenerEnvoyConfigType),
			expect:       true,
		},
		{
			name:         "jwt-single-route-single-match-to-xds",
			from:         "gateway-api",
			to:           "xds",
			output:       yamlOutput,
			resourceType: string(RouteEnvoyConfigType),
			expect:       true,
		},
		{
			name:         "envoy-patch-policy",
			from:         "gateway-api",
			to:           "xds",
			output:       yamlOutput,
			resourceType: string(AllEnvoyConfigType),
			expect:       false,
		},
		{
			name:      "default-resources",
			from:      "gateway-api",
			to:        "gateway-api,xds",
			expect:    true,
			extraArgs: []string{"--add-missing-resources"},
		},
		{
			name:   "quickstart",
			from:   "gateway-api",
			to:     "ir",
			output: yamlOutput,
			expect: true,
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
		},
		{
			name:         "echo-gateway-api",
			from:         "gateway-api",
			to:           "gateway-api",
			output:       jsonOutput,
			resourceType: string(RouteEnvoyConfigType),
			expect:       true,
		},
		{
			name:         "echo-gateway-api",
			from:         "gateway-api",
			to:           "xds,gateway-api",
			output:       yamlOutput,
			resourceType: string(ClusterEnvoyConfigType),
			expect:       true,
		},
		{
			name: "echo-gateway-api",
			from: "gateway-api",
			// ensure the order doesn't affect the output
			to:           "gateway-api,xds",
			output:       yamlOutput,
			resourceType: string(ClusterEnvoyConfigType),
			expect:       true,
		},
		{
			name:         "multiple-xds",
			from:         "gateway-api",
			to:           "xds",
			output:       jsonOutput,
			resourceType: string(RouteEnvoyConfigType),
			expect:       true,
		},
		{
			name:   "valid-envoyproxy",
			from:   "gateway-api",
			to:     "gateway-api",
			output: yamlOutput,
			expect: true,
		},
		{
			name:   "invalid-envoyproxy",
			from:   "gateway-api",
			to:     "gateway-api",
			output: yamlOutput,
			expect: true,
		},
		{
			name:   "no-gateway-class-resources",
			from:   "gateway-api",
			to:     "xds",
			expect: false,
		},
		{
			name:   "no-gateway-class-resources",
			from:   "gateway-api",
			to:     "gateway-api",
			expect: false,
		},
	}

	flag.Parse()

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name+"|"+tc.resourceType, func(t *testing.T) {
			b := bytes.NewBufferString("")
			root := newTranslateCommand()
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
				require.NoError(t, root.ExecuteContext(context.Background()))
			} else {
				require.Error(t, root.ExecuteContext(context.Background()))
				return
			}

			out, err := io.ReadAll(b)
			require.NoError(t, err)
			got := &TranslationResult{}
			mustUnmarshal(t, out, got)
			var fn string
			require.NoError(t, field.SetValue(got, "LastTransitionTime", metav1.NewTime(time.Time{})))

			if tc.output == jsonOutput {
				fn = tc.name + "." + resourceType + ".json"
				out, err = json.MarshalIndent(got, "", "  ")
				require.NoError(t, err)
			} else {
				fn = tc.name + "." + resourceType + ".yaml"
				out, err = yaml.Marshal(got)
				require.NoError(t, err)
			}
			if *overrideTestData {
				require.NoError(t, file.Write(string(out), filepath.Join("testdata", "translate", "out", fn)))
			}
			want := &TranslationResult{}
			mustUnmarshal(t, requireTestDataOutFile(t, fn), want)
			opts := cmpopts.IgnoreFields(metav1.Condition{}, "LastTransitionTime")
			require.Empty(t, cmp.Diff(want, got, opts))

		})
	}
}

func requireTestDataOutFile(t *testing.T, name ...string) []byte {
	t.Helper()
	elems := append([]string{"testdata", "translate", "out"}, name...)
	content, err := os.ReadFile(filepath.Join(elems...))
	require.NoError(t, err)
	return content
}

func mustUnmarshal(t *testing.T, val []byte, out interface{}) {
	require.NoError(t, yaml.UnmarshalStrict(val, out, yaml.DisallowUnknownFields))
}
