// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	runtimev3 "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func jsonValue(raw string) apiextensionsv1.JSON {
	return apiextensionsv1.JSON{Raw: []byte(raw)}
}

func TestProcessRuntime(t *testing.T) {
	testCases := []struct {
		name    string
		runtime map[string]apiextensionsv1.JSON
		// expected holds the wanted layer contents, keyed by runtime key. A nil map means
		// the Runtime resource is expected with no layer at all.
		expected map[string]any
	}{
		{
			// The resource must be emitted even with no values configured, otherwise the
			// proxy's RTDS subscription never resolves and it logs an initial fetch
			// timeout for the Runtime type.
			name:     "no runtime values still emits the resource",
			runtime:  nil,
			expected: nil,
		},
		{
			name: "integer value stays a number",
			runtime: map[string]apiextensionsv1.JSON{
				"circuit_breakers.ai-gateway-extproc-uds.default.max_requests": jsonValue("16384"),
			},
			expected: map[string]any{
				"circuit_breakers.ai-gateway-extproc-uds.default.max_requests": float64(16384),
			},
		},
		{
			// A boolean must survive as a real boolean. Envoy resolves boolean runtime
			// guards from bool_value_ only, so a stringified "true" would be ignored.
			name: "boolean value stays a boolean",
			runtime: map[string]apiextensionsv1.JSON{
				"envoy.reloadable_features.some_feature": jsonValue("true"),
			},
			expected: map[string]any{
				"envoy.reloadable_features.some_feature": true,
			},
		},
		{
			name: "string value stays a string",
			runtime: map[string]apiextensionsv1.JSON{
				"some.string.key": jsonValue(`"hello"`),
			},
			expected: map[string]any{
				"some.string.key": "hello",
			},
		},
		{
			name: "mixed types in one layer",
			runtime: map[string]apiextensionsv1.JSON{
				"circuit_breakers.c1.default.max_requests": jsonValue("2048"),
				"circuit_breakers.c1.high.max_requests":    jsonValue("4096"),
				"envoy.reloadable_features.flag":           jsonValue("false"),
				"overload.global_downstream_max_conns":     jsonValue("50000"),
			},
			expected: map[string]any{
				"circuit_breakers.c1.default.max_requests": float64(2048),
				"circuit_breakers.c1.high.max_requests":    float64(4096),
				"envoy.reloadable_features.flag":           false,
				"overload.global_downstream_max_conns":     float64(50000),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tCtx := new(types.ResourceVersionTable)
			require.NoError(t, processRuntime(tCtx, &ir.Xds{Runtime: tc.runtime}))

			resources := tCtx.XdsResources[resourcev3.RuntimeType]
			require.Len(t, resources, 1, "exactly one Runtime resource must be emitted")

			rt, ok := resources[0].(*runtimev3.Runtime)
			require.True(t, ok)
			// The name must match the rtds_layer name in the bootstrap, otherwise Envoy
			// never matches the resource to its layer.
			require.Equal(t, RuntimeLayerName, rt.GetName())

			if tc.expected == nil {
				require.Nil(t, rt.GetLayer())
				return
			}
			require.Equal(t, tc.expected, rt.GetLayer().AsMap())
		})
	}
}

func TestProcessRuntimeInvalidValue(t *testing.T) {
	tCtx := new(types.ResourceVersionTable)
	err := processRuntime(tCtx, &ir.Xds{Runtime: map[string]apiextensionsv1.JSON{
		"bad.key": jsonValue("{not json"),
	}})
	require.ErrorContains(t, err, `invalid value for runtime key "bad.key"`)
}
