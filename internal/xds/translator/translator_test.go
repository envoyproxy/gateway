// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"embed"
	"path/filepath"
	"testing"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	ratelimitserviceconfig "github.com/envoyproxy/ratelimit/src/config"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/extension/testutils"
	infra "github.com/envoyproxy/gateway/internal/infrastructure/kubernetes"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/utils"
)

var (
	//go:embed testdata/out/*
	outFiles embed.FS
	//go:embed testdata/in/*
	inFiles embed.FS
)

func TestTranslateXds(t *testing.T) {
	testCases := []struct {
		name           string
		requireSecrets bool
	}{
		{
			name: "empty",
		},
		{
			name: "http-route",
		},
		{
			name: "http-route-regex",
		},
		{
			name: "http-route-redirect",
		},
		{
			name: "http-route-mirror",
		},
		{
			name: "http-route-direct-response",
		},
		{
			name: "http-route-request-headers",
		},
		{
			name: "http-route-response-add-headers",
		},
		{
			name: "http-route-response-remove-headers",
		},
		{
			name: "http-route-response-add-remove-headers",
		},
		{
			name: "http-route-weighted-invalid-backend",
		},
		{
			name:           "simple-tls",
			requireSecrets: true,
		},
		{
			name: "tls-route-passthrough",
		},
		{
			name: "tcp-route-simple",
		},
		{
			name: "tcp-route-complex",
		},
		{
			name: "multiple-simple-tcp-route-same-port",
		},
		{
			name: "http-route-weighted-backend",
		},
		{
			name: "tcp-route-weighted-backend",
		},
		{
			name:           "multiple-listeners-same-port",
			requireSecrets: true,
		},
		{
			name: "udp-route",
		},
		{
			name: "http2-route",
		},
		{
			name: "http-route-rewrite-url-prefix",
		},
		{
			name: "http-route-rewrite-root-path-url-prefix",
		},
		{
			name: "http-route-rewrite-url-fullpath",
		},
		{
			name: "http-route-rewrite-url-host",
		},
		{
			name: "ratelimit",
		},
		{
			name: "ratelimit-sourceip",
		},
		{
			name: "authn-single-route-single-match",
		},
		{
			name: "authn-multi-route-single-provider",
		},
		{
			name: "authn-multi-route-multi-provider",
		},
		{
			name: "authn-ratelimit",
		},
		{
			name: "http-route-extension-filter",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ir := requireXdsIRFromInputTestData(t, "xds-ir", tc.name+".yaml")
			tr := &Translator{
				GlobalRateLimit: &GlobalRateLimitSettings{
					ServiceURL: infra.GetRateLimitServiceURL("envoy-gateway-system"),
				},
			}
			ext := v1alpha1.Extension{
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

			tCtx, err := tr.Translate(ir, extMgr)
			require.NoError(t, err)
			listeners := tCtx.XdsResources[resourcev3.ListenerType]
			routes := tCtx.XdsResources[resourcev3.RouteType]
			clusters := tCtx.XdsResources[resourcev3.ClusterType]
			endpoints := tCtx.XdsResources[resourcev3.EndpointType]
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", tc.name+".listeners.yaml"), requireResourcesToYAMLString(t, listeners))
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", tc.name+".routes.yaml"), requireResourcesToYAMLString(t, routes))
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", tc.name+".clusters.yaml"), requireResourcesToYAMLString(t, clusters))
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", tc.name+".endpoints.yaml"), requireResourcesToYAMLString(t, endpoints))
			if tc.requireSecrets {
				secrets := tCtx.XdsResources[resourcev3.SecretType]
				require.Equal(t, requireTestDataOutFile(t, "xds-ir", tc.name+".secrets.yaml"), requireResourcesToYAMLString(t, secrets))
			}
		})
	}
}

func TestTranslateRateLimitConfig(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "empty-header-matches",
		},
		{
			name: "distinct-match",
		},
		{
			name: "value-match",
		},
		{
			name: "multiple-matches",
		},
		{
			name: "multiple-rules",
		},
		{
			name: "multiple-routes",
		},
		{
			name: "masked-remote-address-match",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			in := requireXdsIRListenerFromInputTestData(t, "ratelimit-config", tc.name+".yaml")
			out := BuildRateLimitServiceConfig(in)
			require.Equal(t, requireTestDataOutFile(t, "ratelimit-config", tc.name+".yaml"), requireYamlRootToYAMLString(t, out))
		})
	}
}

func requireXdsIRFromInputTestData(t *testing.T, name ...string) *ir.Xds {
	t.Helper()
	elems := append([]string{"testdata", "in"}, name...)
	content, err := inFiles.ReadFile(filepath.Join(elems...))
	require.NoError(t, err)
	ir := &ir.Xds{}
	err = yaml.Unmarshal(content, ir)
	require.NoError(t, err)
	return ir
}

func requireXdsIRListenerFromInputTestData(t *testing.T, name ...string) *ir.HTTPListener {
	t.Helper()
	elems := append([]string{"testdata", "in"}, name...)
	content, err := inFiles.ReadFile(filepath.Join(elems...))
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

func requireYamlRootToYAMLString(t *testing.T, yamlRoot *ratelimitserviceconfig.YamlRoot) string {
	str, err := GetRateLimitServiceConfigStr(yamlRoot)
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
