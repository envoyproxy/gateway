// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"embed"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	ratelimitv3 "github.com/envoyproxy/go-control-plane/ratelimit/config/ratelimit/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
}

func TestTranslateXds(t *testing.T) {
	// this is a hack to make sure EG render same output on macos and linux
	defaultCertificateName = "/etc/ssl/certs/ca-certificates.crt"
	defer func() {
		defaultCertificateName = func() string {
			switch runtime.GOOS {
			case "darwin":
				// TODO: maybe automatically get the keychain cert? That might be macOS version dependent.
				// For now, we'll just use the root cert installed by Homebrew: brew install ca-certificates.
				//
				// See:
				// * https://apple.stackexchange.com/questions/226375/where-are-the-root-cas-stored-on-os-x
				// * https://superuser.com/questions/992167/where-are-digital-certificates-physically-stored-on-a-mac-os-x-machine
				return "/opt/homebrew/etc/ca-certificates/cert.pem"
			default:
				// This is the default location for the system trust store
				// on Debian derivatives like the envoy-proxy image being used by the infrastructure
				// controller.
				// See https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/security/ssl
				return "/etc/ssl/certs/ca-certificates.crt"
			}
		}()
	}()
	testConfigs := map[string]testFileConfig{
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
			errMsg:                    "no jsonPointers were found while evaluating the jsonPath",
		},
		"jsonpatch-add-op-empty-jsonpath": {
			requireEnvoyPatchPolicies: true,
			errMsg:                    "a patch operation must specify a path or jsonPath",
		},
		"jsonpatch-missing-resource": {
			requireEnvoyPatchPolicies: true,
		},
		"jsonpatch-invalid-patch": {
			requireEnvoyPatchPolicies: true,
			errMsg:                    "unable to unmarshal xds resource",
		},
		"jsonpatch-add-op-without-value": {
			requireEnvoyPatchPolicies: true,
			errMsg:                    "the add operation requires a value",
		},
		"jsonpatch-move-op-with-value": {
			requireEnvoyPatchPolicies: true,
			errMsg:                    "value and from can't be specified with the remove operation",
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
	}

	inputFiles, err := filepath.Glob(filepath.Join("testdata", "in", "xds-ir", "*.yaml"))
	require.NoError(t, err)

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
			tr := &Translator{
				ControllerNamespace: "envoy-gateway-system",
				GlobalRateLimit: &GlobalRateLimitSettings{
					ServiceURL: ratelimit.GetServiceURL("envoy-gateway-system", dnsDomain),
				},
				FilterOrder: x.FilterOrder,
			}
			tCtx, err := tr.Translate(x)
			if !strings.HasSuffix(inputFileName, "partial-invalid") && len(cfg.errMsg) == 0 {
				t.Log(inputFileName)
				require.NoError(t, err)
			} else if len(cfg.errMsg) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), cfg.errMsg)
				return
			}

			listeners := tCtx.XdsResources[resourcev3.ListenerType]
			routes := tCtx.XdsResources[resourcev3.RouteType]
			clusters := tCtx.XdsResources[resourcev3.ClusterType]
			endpoints := tCtx.XdsResources[resourcev3.EndpointType]
			if test.OverrideTestData() {
				require.NoError(t, file.Write(requireResourcesToYAMLString(t, listeners), filepath.Join("testdata", "out", "xds-ir", inputFileName+".listeners.yaml")))
				require.NoError(t, file.Write(requireResourcesToYAMLString(t, routes), filepath.Join("testdata", "out", "xds-ir", inputFileName+".routes.yaml")))
				require.NoError(t, file.Write(requireResourcesToYAMLString(t, clusters), filepath.Join("testdata", "out", "xds-ir", inputFileName+".clusters.yaml")))
				require.NoError(t, file.Write(requireResourcesToYAMLString(t, endpoints), filepath.Join("testdata", "out", "xds-ir", inputFileName+".endpoints.yaml")))
			}
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", inputFileName+".listeners.yaml"), requireResourcesToYAMLString(t, listeners))
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", inputFileName+".routes.yaml"), requireResourcesToYAMLString(t, routes))
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", inputFileName+".clusters.yaml"), requireResourcesToYAMLString(t, clusters))
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", inputFileName+".endpoints.yaml"), requireResourcesToYAMLString(t, endpoints))

			secrets, ok := tCtx.XdsResources[resourcev3.SecretType]
			if ok && len(secrets) > 0 {
				if test.OverrideTestData() {
					require.NoError(t, file.Write(requireResourcesToYAMLString(t, secrets), filepath.Join("testdata", "out", "xds-ir", inputFileName+".secrets.yaml")))
				}
				require.Equal(t, requireTestDataOutFile(t, "xds-ir", inputFileName+".secrets.yaml"), requireResourcesToYAMLString(t, secrets))
			}

			if cfg.requireEnvoyPatchPolicies {
				got := tCtx.EnvoyPatchPolicyStatuses
				for _, e := range got {
					require.NoError(t, field.SetValue(e, "LastTransitionTime", metav1.NewTime(time.Time{})))
				}
				if test.OverrideTestData() {
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
		})
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
					},
				},
			}

			extMgr, closeFunc, err := registry.NewInMemoryManager(ext, &testingExtensionServer{})
			require.NoError(t, err)
			defer closeFunc()
			tr.ExtensionManager = &extMgr

			tCtx, err := tr.Translate(x)
			if len(cfg.errMsg) > 0 {
				require.EqualError(t, err, cfg.errMsg)
			} else {
				require.NoError(t, err)
			}
			listeners := tCtx.XdsResources[resourcev3.ListenerType]
			routes := tCtx.XdsResources[resourcev3.RouteType]
			clusters := tCtx.XdsResources[resourcev3.ClusterType]
			endpoints := tCtx.XdsResources[resourcev3.EndpointType]
			if test.OverrideTestData() {
				require.NoError(t, file.Write(requireResourcesToYAMLString(t, listeners), filepath.Join("testdata", "out", "extension-xds-ir", inputFileName+".listeners.yaml")))
				require.NoError(t, file.Write(requireResourcesToYAMLString(t, routes), filepath.Join("testdata", "out", "extension-xds-ir", inputFileName+".routes.yaml")))
				require.NoError(t, file.Write(requireResourcesToYAMLString(t, clusters), filepath.Join("testdata", "out", "extension-xds-ir", inputFileName+".clusters.yaml")))
				require.NoError(t, file.Write(requireResourcesToYAMLString(t, endpoints), filepath.Join("testdata", "out", "extension-xds-ir", inputFileName+".endpoints.yaml")))
			}
			require.Equal(t, requireTestDataOutFile(t, "extension-xds-ir", inputFileName+".listeners.yaml"), requireResourcesToYAMLString(t, listeners))
			require.Equal(t, requireTestDataOutFile(t, "extension-xds-ir", inputFileName+".routes.yaml"), requireResourcesToYAMLString(t, routes))
			require.Equal(t, requireTestDataOutFile(t, "extension-xds-ir", inputFileName+".clusters.yaml"), requireResourcesToYAMLString(t, clusters))
			require.Equal(t, requireTestDataOutFile(t, "extension-xds-ir", inputFileName+".endpoints.yaml"), requireResourcesToYAMLString(t, endpoints))

			secrets, ok := tCtx.XdsResources[resourcev3.SecretType]
			if ok {
				if test.OverrideTestData() {
					require.NoError(t, file.Write(requireResourcesToYAMLString(t, secrets), filepath.Join("testdata", "out", "extension-xds-ir", inputFileName+".secrets.yaml")))
				}
				require.Equal(t, requireTestDataOutFile(t, "extension-xds-ir", inputFileName+".secrets.yaml"), requireResourcesToYAMLString(t, secrets))
			}
		})
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
					},
				},
			}

			extMgr, closeFunc, err := registry.NewInMemoryManager(ext, &testingExtensionServer{})
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
	err = yaml.Unmarshal(content, x)
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
	content, err := outFiles.ReadFile(filepath.Join(elems...))
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

func requireResourcesToYAMLString(t *testing.T, resources []types.Resource) string {
	jsonBytes, err := utils.MarshalResourcesToJSON(resources)
	require.NoError(t, err)
	data, err := yaml.JSONToYAML(jsonBytes)
	require.NoError(t, err)
	return string(data)
}
