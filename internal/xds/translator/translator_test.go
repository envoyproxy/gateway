// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"embed"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	ratelimitv3 "github.com/envoyproxy/go-control-plane/ratelimit/config/ratelimit/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/extension/registry"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/field"
	"github.com/envoyproxy/gateway/internal/utils/file"
	"github.com/envoyproxy/gateway/internal/utils/test"
	xtypes "github.com/envoyproxy/gateway/internal/xds/types"
	"github.com/envoyproxy/gateway/internal/xds/utils"
)

var (
	//go:embed testdata/out/*
	outFiles embed.FS
	//go:embed testdata/in/*
	inFiles embed.FS
)

type testFileConfig struct {
	requireEnvoyPatchPolicies bool
	dnsDomain                 string
	errMsg                    string
	runtimeFlags              *egv1a1.RuntimeFlags
}

type rateLimitOutput struct {
	RouteName     string               `json:"routeName" yaml:"routeName"`
	RateLimits    []*routev3.RateLimit `json:"rateLimits" yaml:"rateLimits"`
	CostSpecified bool                 `json:"costSpecified" yaml:"costSpecified"`
}

var testConfigs = map[string]testFileConfig{
	"ratelimit-custom-domain": {
		dnsDomain: "example-cluster.local",
	},
	"jsonpatch": {
		requireEnvoyPatchPolicies: true,
	},
	"jsonpatch-with-jsonpath": {
		requireEnvoyPatchPolicies: true,
	},
	"jsonpatch-with-jsonpath-invalid": {
		requireEnvoyPatchPolicies: true,
	},
	"jsonpatch-add-op-empty-jsonpath": {
		requireEnvoyPatchPolicies: true,
	},
	"jsonpatch-missing-resource": {
		requireEnvoyPatchPolicies: true,
	},
	"jsonpatch-invalid-patch": {
		requireEnvoyPatchPolicies: true,
	},
	"jsonpatch-add-op-without-value": {
		requireEnvoyPatchPolicies: true,
	},
	"jsonpatch-move-op-with-value": {
		requireEnvoyPatchPolicies: true,
	},
	"http-route-invalid": {
		errMsg: "validation failed for xds resource",
	},
	"tcp-route-invalid": {
		errMsg: "validation failed for xds resource",
	},
	"tcp-route-invalid-endpoint": {
		errMsg: "validation failed for xds resource",
	},
	"udp-route-invalid": {
		errMsg: "validation failed for xds resource",
	},
	"jsonpatch-invalid": {
		errMsg: "validation failed for xds resource",
	},
	"jsonpatch-invalid-listener": {
		errMsg: "validation failed for xds resource",
	},
	"accesslog-invalid": {
		errMsg: "validation failed for xds resource",
	},
	"accesslog-without-format": {
		errMsg: "text.Format is nil",
	},
	"tracing-invalid": {
		errMsg: "validation failed for xds resource",
	},
	"tracing-unknown-provider-type": {
		errMsg: "unknown tracing provider type: AwesomeTelemetry",
	},
	"xds-name-scheme-v2": {
		runtimeFlags: &egv1a1.RuntimeFlags{
			Enabled: []egv1a1.RuntimeFlag{egv1a1.XDSNameSchemeV2},
		},
	},
}

func TestTranslateXds(t *testing.T) {
	inputFiles, err := filepath.Glob(filepath.Join("testdata", "in", "xds-ir", "*.yaml"))
	require.NoError(t, err)
	keepLinear := make(sets.Set[string])
	keepSublinear := make(sets.Set[string])
	if test.OverrideTestData() {
		require.NoError(t, os.MkdirAll(filepath.Join("testdata", "out", "xds-ir-sublinear"), 0o755))
	}

	for _, inputFile := range inputFiles {
		inputFileName := testName(inputFile)
		t.Run(inputFileName, func(t *testing.T) {
			cfg, ok := testConfigs[inputFileName]
			if !ok {
				cfg = testFileConfig{
					requireEnvoyPatchPolicies: false,
					dnsDomain:                 "",
					errMsg:                    "",
				}
			}

			dnsDomain := cfg.dnsDomain
			if len(dnsDomain) == 0 {
				dnsDomain = "cluster.local"
			}

			x := requireXdsIRFromInputTestData(t, inputFile)

			partialInvalid := strings.HasSuffix(inputFileName, "partial-invalid")

			// Linear mode (default)
			trLinear := &Translator{
				ControllerNamespace: "envoy-gateway-system",
				GlobalRateLimit: &GlobalRateLimitSettings{
					ServiceURL: ratelimit.GetServiceURL("envoy-gateway-system", dnsDomain),
				},
				FilterOrder:  x.FilterOrder,
				RuntimeFlags: cfg.runtimeFlags,
			}
			tCtxLinear, err := trLinear.Translate(x)

			// Handle EnvoyPatchPolicy statuses first, even if there are errors
			if cfg.requireEnvoyPatchPolicies {
				got := tCtxLinear.EnvoyPatchPolicyStatuses
				for _, e := range got {
					require.NoError(t, field.SetValue(e, "LastTransitionTime", metav1.NewTime(time.Time{})))
				}
				if test.OverrideTestData() {
					keepLinear.Insert(inputFileName + ".envoypatchpolicies.yaml")
					out, err := yaml.Marshal(got)
					require.NoError(t, err)
					require.NoError(t, file.Write(string(out), filepath.Join("testdata", "out", "xds-ir", inputFileName+".envoypatchpolicies.yaml")))
				}
				in := requireTestDataOutFile(t, "xds-ir", inputFileName+".envoypatchpolicies.yaml")
				want := xtypes.EnvoyPatchPolicyStatuses{}
				require.NoError(t, yaml.Unmarshal([]byte(in), &want))
				opts := cmpopts.IgnoreFields(metav1.Condition{}, "LastTransitionTime")
				require.Empty(t, cmp.Diff(want, got, opts))
			}

			if !partialInvalid && len(cfg.errMsg) == 0 {
				t.Log(inputFileName)
				require.NoError(t, err)
			} else if len(cfg.errMsg) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), cfg.errMsg)
				return
			}

			listenersLinear := tCtxLinear.XdsResources[resourcev3.ListenerType]
			routesLinear := tCtxLinear.XdsResources[resourcev3.RouteType]
			clustersLinear := tCtxLinear.XdsResources[resourcev3.ClusterType]
			endpointsLinear := tCtxLinear.XdsResources[resourcev3.EndpointType]

			var routesSublinear []types.Resource
			if !partialInvalid {
				sublinearFlags := &egv1a1.RuntimeFlags{Enabled: []egv1a1.RuntimeFlag{egv1a1.SublinearRouteMatching}}
				if cfg.runtimeFlags != nil && len(cfg.runtimeFlags.Enabled) > 0 {
					sublinearFlags.Enabled = append([]egv1a1.RuntimeFlag{egv1a1.SublinearRouteMatching}, cfg.runtimeFlags.Enabled...)
				}
				trSublinear := &Translator{
					ControllerNamespace: "envoy-gateway-system",
					GlobalRateLimit: &GlobalRateLimitSettings{
						ServiceURL: ratelimit.GetServiceURL("envoy-gateway-system", dnsDomain),
					},
					FilterOrder:  x.FilterOrder,
					RuntimeFlags: sublinearFlags,
				}
				tCtxSublinear, errSub := trSublinear.Translate(x)
				require.NoError(t, errSub)

				routesSublinear = tCtxSublinear.XdsResources[resourcev3.RouteType]
				listenersSublinear := tCtxSublinear.XdsResources[resourcev3.ListenerType]
				clustersSublinear := tCtxSublinear.XdsResources[resourcev3.ClusterType]
				endpointsSublinear := tCtxSublinear.XdsResources[resourcev3.EndpointType]

				// Listeners, clusters, endpoints must be identical between linear and sublinear (only routes differ).
				// Skip for edge cases (e.g. no HTTP routes) where the translator may produce different non-route resources.
				skipNonRouteEquality := strings.Contains(inputFileName, "with-no-routes")
				if !skipNonRouteEquality {
					require.Equal(t, requireResourcesToYAMLString(t, listenersLinear), requireResourcesToYAMLString(t, listenersSublinear),
						"listeners must be identical between linear and sublinear mode")
					require.Equal(t, requireResourcesToYAMLString(t, clustersLinear), requireResourcesToYAMLString(t, clustersSublinear),
						"clusters must be identical between linear and sublinear mode")
					require.Equal(t, requireResourcesToYAMLString(t, endpointsLinear), requireResourcesToYAMLString(t, endpointsSublinear),
						"endpoints must be identical between linear and sublinear mode")
				}
			}

			if test.OverrideTestData() {
				keepLinear.Insert(inputFileName + ".listeners.yaml")
				keepLinear.Insert(inputFileName + ".routes.yaml")
				keepLinear.Insert(inputFileName + ".clusters.yaml")
				keepLinear.Insert(inputFileName + ".endpoints.yaml")
				require.NoError(t, file.Write(requireResourcesToYAMLString(t, listenersLinear), filepath.Join("testdata", "out", "xds-ir", inputFileName+".listeners.yaml")))
				require.NoError(t, file.Write(requireResourcesToYAMLString(t, routesLinear), filepath.Join("testdata", "out", "xds-ir", inputFileName+".routes.yaml")))
				require.NoError(t, file.Write(requireResourcesToYAMLString(t, clustersLinear), filepath.Join("testdata", "out", "xds-ir", inputFileName+".clusters.yaml")))
				require.NoError(t, file.Write(requireResourcesToYAMLString(t, endpointsLinear), filepath.Join("testdata", "out", "xds-ir", inputFileName+".endpoints.yaml")))
				if !partialInvalid {
					keepSublinear.Insert(inputFileName + ".routes.yaml")
					require.NoError(t, file.Write(requireResourcesToYAMLString(t, routesSublinear), filepath.Join("testdata", "out", "xds-ir-sublinear", inputFileName+".routes.yaml")))
				}
			}
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", inputFileName+".listeners.yaml"), requireResourcesToYAMLString(t, listenersLinear))
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", inputFileName+".routes.yaml"), requireResourcesToYAMLString(t, routesLinear))
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", inputFileName+".clusters.yaml"), requireResourcesToYAMLString(t, clustersLinear))
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", inputFileName+".endpoints.yaml"), requireResourcesToYAMLString(t, endpointsLinear))
			if !partialInvalid {
				require.Equal(t, requireTestDataOutFile(t, "xds-ir-sublinear", inputFileName+".routes.yaml"), requireResourcesToYAMLString(t, routesSublinear))
			}

			secrets, ok := tCtxLinear.XdsResources[resourcev3.SecretType]
			if ok && len(secrets) > 0 {
				if test.OverrideTestData() {
					keepLinear.Insert(inputFileName + ".secrets.yaml")
					require.NoError(t, file.Write(requireResourcesToYAMLString(t, secrets), filepath.Join("testdata", "out", "xds-ir", inputFileName+".secrets.yaml")))
				}
				require.Equal(t, requireTestDataOutFile(t, "xds-ir", inputFileName+".secrets.yaml"), requireResourcesToYAMLString(t, secrets))
			}
		})
	}

	if test.OverrideTestData() {
		cleanupOutdatedTestData(t, filepath.Join("testdata", "out", "xds-ir"), keepLinear)
		cleanupOutdatedTestData(t, filepath.Join("testdata", "out", "xds-ir-sublinear"), keepSublinear)
	}
}

func cleanupOutdatedTestData(t *testing.T, dir string, keep sets.Set[string]) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".yaml" {
			continue
		}
		if _, ok := keep[name]; ok {
			continue
		}
		require.NoError(t, os.Remove(filepath.Join(dir, name)))
	}
}

func TestTranslateRateLimitConfig(t *testing.T) {
	inputFiles, err := filepath.Glob(filepath.Join("testdata", "in", "ratelimit-config", "*.yaml"))
	require.NoError(t, err)

	for _, inputFile := range inputFiles {
		inputFileName := testName(inputFile)
		t.Run(inputFileName, func(t *testing.T) {
			// Get listeners from the test data
			listeners := requireXdsIRListenersFromInputTestData(t, inputFile)

			// Call BuildRateLimitServiceConfig with the list of listeners
			configs := BuildRateLimitServiceConfig(listeners)

			if test.OverrideTestData() {
				require.NoError(t, file.Write(requireRateLimitConfigsToYAMLString(t, configs), filepath.Join("testdata", "out", "ratelimit-config", inputFileName+".yaml")))
			}
			require.Equal(t, requireTestDataOutFile(t, "ratelimit-config", inputFileName+".yaml"), requireRateLimitConfigsToYAMLString(t, configs))
		})
	}
}

func TestBuildRouteRateLimits(t *testing.T) {
	inputFiles, err := filepath.Glob(filepath.Join("testdata", "in", "ratelimit-config", "*.yaml"))
	require.NoError(t, err)

	for _, inputFile := range inputFiles {
		inputFileName := testName(inputFile)
		t.Run(inputFileName, func(t *testing.T) {
			// Get listeners from the test data
			listeners := requireXdsIRListenersFromInputTestData(t, inputFile)

			var outputs []rateLimitOutput

			// Process each route to get rate limit actions
			for _, listener := range listeners {
				for _, route := range listener.Routes {
					rateLimits, costSpecified := buildRouteRateLimits(route)

					output := rateLimitOutput{
						RouteName:     route.Name,
						RateLimits:    rateLimits,
						CostSpecified: costSpecified,
					}
					outputs = append(outputs, output)
				}
			}
			if test.OverrideTestData() {
				require.NoError(t, file.Write(requireRouteRateLimitsToYAMLString(t, outputs), filepath.Join("testdata", "out", "ratelimit-config", inputFileName+".routes.yaml")))
			}
			require.Equal(t, requireTestDataOutFile(t, "ratelimit-config", inputFileName+".routes.yaml"), requireRouteRateLimitsToYAMLString(t, outputs))
		})
	}
}

// Simulate various extension server hooks and ensure that the translator returns the original resources
// when configured to failOpen
func TestTranslateXdsWithExtensionErrorsWhenFailOpen(t *testing.T) {
	testConfigs := map[string]testFileConfig{
		"http-route-extension-route-error":                 {},
		"http-route-extension-virtualhost-error":           {},
		"http-route-extension-listener-error":              {},
		"http-route-extension-translate-error":             {},
		"multiple-listeners-same-port-error":               {},
		"http-route-custom-backend":                        {},
		"http-route-custom-backends-multiple":              {},
		"http-route-custom-backends-partial":               {},
		"http-route-custom-backend-error":                  {},
		"http-route-custom-backend-multiple-backend-error": {},
	}

	inputFiles, err := filepath.Glob(filepath.Join("testdata", "in", "extension-xds-ir", "*.yaml"))
	require.NoError(t, err)

	keepLinear := make(sets.Set[string])
	keepSublinear := make(sets.Set[string])
	if test.OverrideTestData() {
		require.NoError(t, os.MkdirAll(filepath.Join("testdata", "out", "extension-xds-ir-sublinear"), 0o755))
	}

	for _, inputFile := range inputFiles {
		inputFileName := testName(inputFile)
		t.Run(inputFileName, func(t *testing.T) {
			cfg, ok := testConfigs[inputFileName]
			if !ok {
				cfg = testFileConfig{}
			}

			// Testdata for the extension tests is similar to the ir test data
			// New directory is just to keep them separate and easy to understand
			x := requireXdsIRFromInputTestData(t, inputFile)
			ext := egv1a1.ExtensionManager{
				FailOpen: true,
				Resources: []egv1a1.GroupVersionKind{
					{
						Group:   "foo.example.io",
						Version: "v1alpha1",
						Kind:    "examplefilter",
					},
				},
				BackendResources: []egv1a1.GroupVersionKind{
					{
						Group:   "inference.networking.x-k8s.io",
						Version: "v1alpha2",
						Kind:    "InferencePool",
					},
				},
				PolicyResources: []egv1a1.GroupVersionKind{
					{
						Group:   "bar.example.io",
						Version: "v1alpha1",
						Kind:    "ExtensionPolicy",
					},
					{
						Group:   "foo.example.io",
						Version: "v1alpha1",
						Kind:    "Bar",
					},
					{
						Group:   "security.example.io",
						Version: "v1alpha1",
						Kind:    "ExampleExtPolicy",
					},
				},
				Hooks: &egv1a1.ExtensionHooks{
					XDSTranslator: &egv1a1.XDSTranslatorHooks{
						Post: []egv1a1.XDSTranslatorHook{
							egv1a1.XDSCluster,
							egv1a1.XDSRoute,
							egv1a1.XDSVirtualHost,
							egv1a1.XDSHTTPListener,
							egv1a1.XDSCluster,
							egv1a1.XDSTranslation,
						},
						// Enable listeners and routes for PostTranslateModifyHook for these tests
						Translation: &egv1a1.TranslationConfig{
							Listener: &egv1a1.ListenerTranslationConfig{
								IncludeAll: ptr.To(true),
							},
							Route: &egv1a1.RouteTranslationConfig{
								IncludeAll: ptr.To(true),
							},
						},
					},
				},
			}

			extMgr, closeFunc, err := registry.NewInMemoryManager(&ext, &testingExtensionServer{})
			require.NoError(t, err)
			defer closeFunc()

			if len(cfg.errMsg) > 0 {
				tr := &Translator{
					GlobalRateLimit: &GlobalRateLimitSettings{
						ServiceURL: ratelimit.GetServiceURL("envoy-gateway-system", "cluster.local"),
					},
					ExtensionManager: &extMgr,
				}
				_, err := tr.Translate(x)
				require.EqualError(t, err, cfg.errMsg)
				return
			}

			// Success path: run linear then sublinear (sequential — Translate may mutate shared IR).
			trLinear := &Translator{
				GlobalRateLimit: &GlobalRateLimitSettings{
					ServiceURL: ratelimit.GetServiceURL("envoy-gateway-system", "cluster.local"),
				},
				ExtensionManager: &extMgr,
			}
			tCtxLinear, err := trLinear.Translate(x)
			require.NoError(t, err)
			listenersLinear := tCtxLinear.XdsResources[resourcev3.ListenerType]
			routesLinear := tCtxLinear.XdsResources[resourcev3.RouteType]
			clustersLinear := tCtxLinear.XdsResources[resourcev3.ClusterType]
			endpointsLinear := tCtxLinear.XdsResources[resourcev3.EndpointType]

			trSublinear := &Translator{
				GlobalRateLimit: &GlobalRateLimitSettings{
					ServiceURL: ratelimit.GetServiceURL("envoy-gateway-system", "cluster.local"),
				},
				ExtensionManager: &extMgr,
				RuntimeFlags:     &egv1a1.RuntimeFlags{Enabled: []egv1a1.RuntimeFlag{egv1a1.SublinearRouteMatching}},
			}
			tCtxSublinear, errSub := trSublinear.Translate(x)
			require.NoError(t, errSub)
			routesSublinear := tCtxSublinear.XdsResources[resourcev3.RouteType]
			listenersSublinear := tCtxSublinear.XdsResources[resourcev3.ListenerType]
			clustersSublinear := tCtxSublinear.XdsResources[resourcev3.ClusterType]
			endpointsSublinear := tCtxSublinear.XdsResources[resourcev3.EndpointType]

			require.Equal(t, requireResourcesToYAMLString(t, listenersLinear), requireResourcesToYAMLString(t, listenersSublinear),
				"listeners must be identical between linear and sublinear mode")
			require.Equal(t, requireResourcesToYAMLString(t, clustersLinear), requireResourcesToYAMLString(t, clustersSublinear),
				"clusters must be identical between linear and sublinear mode")
			require.Equal(t, requireResourcesToYAMLString(t, endpointsLinear), requireResourcesToYAMLString(t, endpointsSublinear),
				"endpoints must be identical between linear and sublinear mode")

			if test.OverrideTestData() {
				keepLinear.Insert(inputFileName + ".listeners.yaml")
				keepLinear.Insert(inputFileName + ".routes.yaml")
				keepLinear.Insert(inputFileName + ".clusters.yaml")
				keepLinear.Insert(inputFileName + ".endpoints.yaml")
				require.NoError(t, file.Write(requireResourcesToYAMLString(t, listenersLinear), filepath.Join("testdata", "out", "extension-xds-ir", inputFileName+".listeners.yaml")))
				require.NoError(t, file.Write(requireResourcesToYAMLString(t, routesLinear), filepath.Join("testdata", "out", "extension-xds-ir", inputFileName+".routes.yaml")))
				require.NoError(t, file.Write(requireResourcesToYAMLString(t, clustersLinear), filepath.Join("testdata", "out", "extension-xds-ir", inputFileName+".clusters.yaml")))
				require.NoError(t, file.Write(requireResourcesToYAMLString(t, endpointsLinear), filepath.Join("testdata", "out", "extension-xds-ir", inputFileName+".endpoints.yaml")))
				keepSublinear.Insert(inputFileName + ".routes.yaml")
				require.NoError(t, file.Write(requireResourcesToYAMLString(t, routesSublinear), filepath.Join("testdata", "out", "extension-xds-ir-sublinear", inputFileName+".routes.yaml")))
			}
			require.Equal(t, requireTestDataOutFile(t, "extension-xds-ir", inputFileName+".listeners.yaml"), requireResourcesToYAMLString(t, listenersLinear))
			require.Equal(t, requireTestDataOutFile(t, "extension-xds-ir", inputFileName+".routes.yaml"), requireResourcesToYAMLString(t, routesLinear))
			require.Equal(t, requireTestDataOutFile(t, "extension-xds-ir", inputFileName+".clusters.yaml"), requireResourcesToYAMLString(t, clustersLinear))
			require.Equal(t, requireTestDataOutFile(t, "extension-xds-ir", inputFileName+".endpoints.yaml"), requireResourcesToYAMLString(t, endpointsLinear))
			require.Equal(t, requireTestDataOutFile(t, "extension-xds-ir-sublinear", inputFileName+".routes.yaml"), requireResourcesToYAMLString(t, routesSublinear))

			secrets, ok := tCtxLinear.XdsResources[resourcev3.SecretType]
			if ok {
				if test.OverrideTestData() {
					keepLinear.Insert(inputFileName + ".secrets.yaml")
					require.NoError(t, file.Write(requireResourcesToYAMLString(t, secrets), filepath.Join("testdata", "out", "extension-xds-ir", inputFileName+".secrets.yaml")))
				}
				require.Equal(t, requireTestDataOutFile(t, "extension-xds-ir", inputFileName+".secrets.yaml"), requireResourcesToYAMLString(t, secrets))
			}
		})
	}
	if test.OverrideTestData() {
		cleanupOutdatedTestData(t, filepath.Join("testdata", "out", "extension-xds-ir"), keepLinear)
		cleanupOutdatedTestData(t, filepath.Join("testdata", "out", "extension-xds-ir-sublinear"), keepSublinear)
	}
}

// Simulate various extension server hooks and ensure that the translator returns an error
// when configured with to not fail open.
func TestTranslateXdsWithExtensionErrorsWhenFailClosed(t *testing.T) {
	testConfigs := map[string]testFileConfig{
		"http-route-extension-route-error": {
			errMsg: "rpc error: code = Unknown desc = route hook resource error",
		},
		"http-route-extension-virtualhost-error": {
			errMsg: "rpc error: code = Unknown desc = extension post xds virtual host hook error",
		},
		"http-route-extension-listener-error": {
			errMsg: "rpc error: code = Unknown desc = extension post xds listener hook error",
		},
		"http-route-extension-translate-error": {
			errMsg: "rpc error: code = Unknown desc = cluster hook resource error: fail-close-error",
		},
		"multiple-listeners-same-port-error": {
			errMsg: "rpc error: code = Unknown desc = simulate error when there is no default filter chain in the original resources",
		},
		"extensionpolicy-extension-server-error": {
			errMsg: "rpc error: code = Unknown desc = invalid extension policy : ext-server-policy-invalid-test",
		},
		"http-route-custom-backend-error": {
			errMsg: "rpc error: code = Unknown desc = inference pool target port number is 0",
		},
		"http-route-custom-backend-multiple-backend-error": {
			errMsg: "rpc error: code = Unknown desc = inference pool only support one per rule",
		},
	}

	inputFiles, err := filepath.Glob(filepath.Join("testdata", "in", "extension-xds-ir", "*-error.yaml"))
	require.NoError(t, err)

	for _, inputFile := range inputFiles {
		inputFileName := testName(inputFile)
		t.Run(inputFileName, func(t *testing.T) {
			cfg, ok := testConfigs[inputFileName]
			if !ok {
				cfg = testFileConfig{}
			}

			// Testdata for the extension tests is similar to the ir test data
			// New directory is just to keep them separate and easy to understand
			x := requireXdsIRFromInputTestData(t, inputFile)
			tr := &Translator{
				GlobalRateLimit: &GlobalRateLimitSettings{
					ServiceURL: ratelimit.GetServiceURL("envoy-gateway-system", "cluster.local"),
				},
			}
			ext := egv1a1.ExtensionManager{
				FailOpen: false,
				Resources: []egv1a1.GroupVersionKind{
					{
						Group:   "foo.example.io",
						Version: "v1alpha1",
						Kind:    "examplefilter",
					},
				},
				BackendResources: []egv1a1.GroupVersionKind{
					{
						Group:   "inference.networking.x-k8s.io",
						Version: "v1alpha2",
						Kind:    "InferencePool",
					},
				},
				PolicyResources: []egv1a1.GroupVersionKind{
					{
						Group:   "bar.example.io",
						Version: "v1alpha1",
						Kind:    "ExtensionPolicy",
					},
					{
						Group:   "foo.example.io",
						Version: "v1alpha1",
						Kind:    "Bar",
					},
					{
						Group:   "security.example.io",
						Version: "v1alpha1",
						Kind:    "ExampleExtPolicy",
					},
				},
				Hooks: &egv1a1.ExtensionHooks{
					XDSTranslator: &egv1a1.XDSTranslatorHooks{
						Post: []egv1a1.XDSTranslatorHook{
							egv1a1.XDSCluster,
							egv1a1.XDSRoute,
							egv1a1.XDSVirtualHost,
							egv1a1.XDSHTTPListener,
							egv1a1.XDSTranslation,
						},
						// Enable listeners and routes for PostTranslateModifyHook for these tests
						Translation: &egv1a1.TranslationConfig{
							Listener: &egv1a1.ListenerTranslationConfig{
								IncludeAll: ptr.To(true),
							},
							Route: &egv1a1.RouteTranslationConfig{
								IncludeAll: ptr.To(true),
							},
						},
					},
				},
			}

			extMgr, closeFunc, err := registry.NewInMemoryManager(&ext, &testingExtensionServer{})
			require.NoError(t, err)
			defer closeFunc()
			tr.ExtensionManager = &extMgr

			_, err = tr.Translate(x)
			require.EqualError(t, err, cfg.errMsg)
		})
	}
}

func testName(inputFile string) string {
	_, fileName := filepath.Split(inputFile)
	return strings.TrimSuffix(fileName, ".yaml")
}

func requireXdsIRFromInputTestData(t *testing.T, name string) *ir.Xds {
	t.Helper()
	content, err := inFiles.ReadFile(name)
	require.NoError(t, err)
	x := &ir.Xds{}
	err = yaml.Unmarshal(content, x, yaml.DisallowUnknownFields)
	require.NoError(t, err)
	return x
}

func requireXdsIRListenersFromInputTestData(t *testing.T, name string) []*ir.HTTPListener {
	t.Helper()
	content, err := inFiles.ReadFile(name)
	require.NoError(t, err)

	// Try to unmarshal as a list of listeners first (new format)
	xdsIR := struct {
		HTTP []*ir.HTTPListener `yaml:"http"`
	}{}

	err = yaml.Unmarshal(content, &xdsIR)
	if err == nil && len(xdsIR.HTTP) > 0 {
		// Return the list of listeners
		return xdsIR.HTTP
	}

	// Fall back to the old format (single listener)
	listener := &ir.HTTPListener{}
	err = yaml.Unmarshal(content, listener)
	require.NoError(t, err)
	return []*ir.HTTPListener{listener}
}

func requireTestDataOutFile(t *testing.T, name ...string) string {
	t.Helper()
	elems := append([]string{"testdata", "out"}, name...)
	path := filepath.Join(elems...)

	// When overriding, read from disk so we compare against what we just wrote
	if test.OverrideTestData() {
		if diskContent, diskErr := os.ReadFile(path); diskErr == nil {
			return string(diskContent)
		}
	}
	content, err := outFiles.ReadFile(path)
	require.NoError(t, err)
	return string(content)
}

func requireRateLimitConfigsToYAMLString(t *testing.T, configs []*ratelimitv3.RateLimitConfig) string {
	if len(configs) == 0 {
		return ""
	}

	// Sort configs by domain to ensure consistent output regardless of map iteration order
	sort.Slice(configs, func(i, j int) bool {
		return configs[i].Domain < configs[j].Domain
	})

	var result string
	for i, config := range configs {
		str, err := GetRateLimitServiceConfigStr(config)
		require.NoError(t, err)

		if i > 0 {
			result += "---\n"
		}
		result += str
	}
	return result
}

func requireRouteRateLimitsToYAMLString(t *testing.T, rlOutputs []rateLimitOutput) string {
	if len(rlOutputs) == 0 {
		return ""
	}

	results := make([]string, 0, len(rlOutputs))
	for _, output := range rlOutputs {
		jsonBytes, err := json.Marshal(output)
		require.NoError(t, err)
		yamlBytes, err := yaml.JSONToYAML(jsonBytes)
		require.NoError(t, err)

		results = append(results, string(yamlBytes))
	}

	return strings.Join(results, "---\n")
}

func requireResourcesToYAMLString(t *testing.T, resources []types.Resource) string {
	jsonBytes, err := utils.MarshalResourcesToJSON(resources)
	require.NoError(t, err)
	data, err := yaml.JSONToYAML(jsonBytes)
	require.NoError(t, err)
	return test.NormalizeCertPath(string(data))
}
