// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"bytes"
	"embed"
	"path/filepath"
	"testing"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	ratelimitserviceconfig "github.com/envoyproxy/ratelimit/src/config"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"sigs.k8s.io/yaml"

	infra "github.com/envoyproxy/gateway/internal/infrastructure/kubernetes"
	"github.com/envoyproxy/gateway/internal/ir"
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
			name: "http-route-rewrite-url-fullpath",
		},
		{
			name: "http-route-rewrite-url-host",
		},
		{
			name: "ratelimit",
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
			tCtx, err := tr.Translate(ir)
			require.NoError(t, err)
			listeners := tCtx.XdsResources[resource.ListenerType]
			routes := tCtx.XdsResources[resource.RouteType]
			clusters := tCtx.XdsResources[resource.ClusterType]
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", tc.name+".listeners.yaml"), requireResourcesToYAMLString(t, listeners))
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", tc.name+".routes.yaml"), requireResourcesToYAMLString(t, routes))
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", tc.name+".clusters.yaml"), requireResourcesToYAMLString(t, clusters))
			if tc.requireSecrets {
				secrets := tCtx.XdsResources[resource.SecretType]
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
	jsonBytes, err := marshalResourcesToJSON(resources)
	require.NoError(t, err)
	data, err := yaml.JSONToYAML(jsonBytes)
	require.NoError(t, err)
	return string(data)
}

func marshalResourcesToJSON(resources []types.Resource) ([]byte, error) {
	msgs := make([]proto.Message, 0)
	for _, resource := range resources {
		msgs = append(msgs, resource.(proto.Message))
	}
	var buffer bytes.Buffer
	buffer.WriteByte('[')
	for idx, msg := range msgs {
		if idx != 0 {
			buffer.WriteByte(',')
		}
		b, err := protojson.Marshal(msg)
		if err != nil {
			return nil, err
		}
		buffer.Write(b)
	}
	buffer.WriteByte(']')
	return buffer.Bytes(), nil
}
