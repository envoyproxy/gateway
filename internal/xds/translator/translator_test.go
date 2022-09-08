package translator

import (
	"bytes"
	"embed"
	"path/filepath"
	"testing"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/internal/ir"
)

var (
	//go:embed testdata/out/*
	outFiles embed.FS
	//go:embed testdata/in/*
	inFiles embed.FS
)

func TestTranslate(t *testing.T) {
	testCases := []struct {
		name string
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
			name: "http-route-direct-response",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ir := requireXdsIRFromInputTestData(t, "xds-ir", tc.name+".yaml")
			tCtx, err := Translate(ir)
			require.NoError(t, err)
			listeners := tCtx.XdsResources[resource.ListenerType]
			routes := tCtx.XdsResources[resource.RouteType]
			clusters := tCtx.XdsResources[resource.ClusterType]
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", tc.name+".listeners.yaml"), requireResourcesToYAMLString(t, listeners))
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", tc.name+".routes.yaml"), requireResourcesToYAMLString(t, routes))
			require.Equal(t, requireTestDataOutFile(t, "xds-ir", tc.name+".clusters.yaml"), requireResourcesToYAMLString(t, clusters))
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

func requireTestDataOutFile(t *testing.T, name ...string) string {
	t.Helper()
	elems := append([]string{"testdata", "out"}, name...)
	content, err := outFiles.ReadFile(filepath.Join(elems...))
	require.NoError(t, err)
	return string(content)
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
