// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"embed"
	"flag"
	"path/filepath"
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

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/extension/testutils"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/field"
	"github.com/envoyproxy/gateway/internal/utils/file"
	xtypes "github.com/envoyproxy/gateway/internal/xds/types"
	"github.com/envoyproxy/gateway/internal/xds/utils"
)

var (
	//go:embed testdata/out/*
	outFiles embed.FS
	//go:embed testdata/in/*
	inFiles embed.FS

	overrideTestData = flag.Bool("override-testdata", false, "if override the test output data.")
)

type testFileConfig struct {
	requireEnvoyPatchPolicies bool
	dnsDomain                 string
	errMsg                    string
}

func TestTranslateXds(t *testing.T) {
	testConfigs := map[string]testFileConfig{
		"ratelimit-custom-domain": {
			dnsDomain: "example-cluster.local",
		},
		"jsonpatch": {
			requireEnvoyPatchPolicies: true,
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
			errMsg:                    "the value field can not be set for the remove operation",
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
		"tracing-invalid": {
			errMsg: "validation failed for xds resource",
		},
	}

	inputFiles, err := filepath.Glob(filepath.Join("testdata", "in", "xds-ir", "*.yaml"))
	require.NoError(t, err)

	for _, inputFile := range inputFiles {
		inputFile := inputFile
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
				GlobalRateLimit: &GlobalRateLimitSettings{
					ServiceURL: ratelimit.GetServiceURL("envoy-gateway-system", dnsDomain),
				},
				FilterOrder: x.FilterOrder,
			}

			tCtx, err := tr.Translate(x)
			if !strings.HasSuffix(inputFileName, "partial-invalid") && len(cfg.errMsg) == 0 {
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
			if *overrideTestData {
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
				if *overrideTestData {
					require.NoError(t, file.Write(requireResourcesToYAMLString(t, secrets), filepath.Join("testdata", "out", "xds-ir", inputFileName+".secrets.yaml")))
				}
				require.Equal(t, requireTestDataOutFile(t, "xds-ir", inputFileName+".secrets.yaml"), requireResourcesToYAMLString(t, secrets))
			}

			if cfg.requireEnvoyPatchPolicies {
				got := tCtx.EnvoyPatchPolicyStatuses
				for _, e := range got {
					require.NoError(t, field.SetValue(e, "LastTransitionTime", metav1.NewTime(time.Time{})))
				}
				if *overrideTestData {
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
		inputFile := inputFile
		inputFileName := testName(inputFile)
		t.Run(inputFileName, func(t *testing.T) {
			in := requireXdsIRListenerFromInputTestData(t, inputFile)
			out := BuildRateLimitServiceConfig(in)
			if *overrideTestData {
				require.NoError(t, file.Write(requireYamlRootToYAMLString(t, out), filepath.Join("testdata", "out", "ratelimit-config", inputFileName+".yaml")))
			}
			require.Equal(t, requireTestDataOutFile(t, "ratelimit-config", inputFileName+".yaml"), requireYamlRootToYAMLString(t, out))
		})
	}
}

func TestTranslateXdsWithExtension(t *testing.T) {
	testConfigs := map[string]testFileConfig{
		"http-route-extension-route-error": {
			errMsg: "route hook resource error",
		},
		"http-route-extension-virtualhost-error": {
			errMsg: "extension post xds virtual host hook error",
		},
		"http-route-extension-listener-error": {
			errMsg: "extension post xds listener hook error",
		},
	}

	inputFiles, err := filepath.Glob(filepath.Join("testdata", "in", "extension-xds-ir", "*.yaml"))
	require.NoError(t, err)

	for _, inputFile := range inputFiles {
		inputFile := inputFile
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
			ext := v1alpha1.ExtensionManager{
				Resources: []v1alpha1.GroupVersionKind{
					{
						Group:   "foo.example.io",
						Version: "v1alpha1",
						Kind:    "examplefilter",
					},
				},
				Hooks: &v1alpha1.ExtensionHooks{
					XDSTranslator: &v1alpha1.XDSTranslatorHooks{
						Post: []v1alpha1.XDSTranslatorHook{
							v1alpha1.XDSRoute,
							v1alpha1.XDSVirtualHost,
							v1alpha1.XDSHTTPListener,
							v1alpha1.XDSTranslation,
						},
					},
				},
			}
			extMgr := testutils.NewManager(ext)
			tr.ExtensionManager = &extMgr

			tCtx, err := tr.Translate(x)
			if len(cfg.errMsg) > 0 {
				require.EqualError(t, err, cfg.errMsg)
			} else {
				require.NoError(t, err)
				listeners := tCtx.XdsResources[resourcev3.ListenerType]
				routes := tCtx.XdsResources[resourcev3.RouteType]
				clusters := tCtx.XdsResources[resourcev3.ClusterType]
				endpoints := tCtx.XdsResources[resourcev3.EndpointType]
				if *overrideTestData {
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
					if *overrideTestData {
						require.NoError(t, file.Write(requireResourcesToYAMLString(t, secrets), filepath.Join("testdata", "out", "extension-xds-ir", inputFileName+".secrets.yaml")))
					}
					require.Equal(t, requireTestDataOutFile(t, "extension-xds-ir", inputFileName+".secrets.yaml"), requireResourcesToYAMLString(t, secrets))
				}
			}
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

func requireXdsIRListenerFromInputTestData(t *testing.T, name string) *ir.HTTPListener {
	t.Helper()
	content, err := inFiles.ReadFile(name)
	require.NoError(t, err)
	listener := &ir.HTTPListener{}
	err = yaml.Unmarshal(content, listener)
	require.NoError(t, err)
	return listener
}

func requireTestDataOutFile(t *testing.T, name ...string) string {
	t.Helper()
	elems := append([]string{"testdata", "out"}, name...)
	content, err := outFiles.ReadFile(filepath.Join(elems...))
	require.NoError(t, err)
	return string(content)
}

func requireYamlRootToYAMLString(t *testing.T, pbRoot *ratelimitv3.RateLimitConfig) string {
	str, err := GetRateLimitServiceConfigStr(pbRoot)
	require.NoError(t, err)
	return str
}

func requireResourcesToYAMLString(t *testing.T, resources []types.Resource) string {
	jsonBytes, err := utils.MarshalResourcesToJSON(resources)
	require.NoError(t, err)
	data, err := yaml.JSONToYAML(jsonBytes)
	require.NoError(t, err)
	return string(data)
}
