// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"

	kube "github.com/envoyproxy/gateway/internal/kubernetes"
	"github.com/envoyproxy/gateway/internal/utils/file"
	netutil "github.com/envoyproxy/gateway/internal/utils/net"
)

const (
	defaultNamespace           = "default"
	defaultEnvoyGatewayPodName = "eg"
)

var _ kube.PortForwarder = &fakePortForwarder{}

type fakePortForwarder struct {
	responseBody []byte
	localPort    int
	l            net.Listener
	mux          *http.ServeMux
}

func newFakePortForwarder(b []byte) (kube.PortForwarder, error) {
	p, err := netutil.LocalAvailablePort()
	if err != nil {
		return nil, err
	}

	fw := &fakePortForwarder{
		responseBody: b,
		localPort:    p,
		mux:          http.NewServeMux(),
	}
	fw.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(fw.responseBody)
	})

	return fw, nil
}

func (fw *fakePortForwarder) Start() error {
	l, err := net.Listen("tcp", fw.Address())
	if err != nil {
		return err
	}
	fw.l = l

	go func() {
		// nolint: gosec
		if err := http.Serve(l, fw.mux); err != nil {
			log.Fatal(err)
		}
	}()

	return nil
}

func (fw *fakePortForwarder) Stop() {}

func (fw *fakePortForwarder) Address() string {
	return fmt.Sprintf("localhost:%d", fw.localPort)
}

func TestExtractAllConfigDump(t *testing.T) {
	input, err := readInputConfig("in.all.json")
	require.NoError(t, err)
	fw, err := newFakePortForwarder(input)
	require.NoError(t, err)
	err = fw.Start()
	require.NoError(t, err)

	cases := []struct {
		output       string
		expected     string
		resourceType string
	}{
		{
			output:   "json",
			expected: "out.all.json",
		},
		{
			output:   "yaml",
			expected: "out.all.yaml",
		},
	}

	for _, tc := range cases {
		t.Run(tc.expected, func(t *testing.T) {
			configDump, err := extractConfigDump(fw, true, AllEnvoyConfigType)
			require.NoError(t, err)
			aggregated := sampleAggregatedConfigDump(configDump)
			got, err := marshalEnvoyProxyConfig(aggregated, tc.output)
			require.NoError(t, err)
			if *overrideTestData {
				require.NoError(t, file.Write(string(got), filepath.Join("testdata", "config", "out", tc.expected)))
			}
			out, err := readOutputConfig(tc.expected)
			require.NoError(t, err)
			if tc.output == "yaml" {
				assert.YAMLEq(t, string(out), string(got))
			} else {
				assert.JSONEq(t, string(out), string(got))
			}
		})
	}

	fw.Stop()
}

func TestExtractSubResourcesConfigDump(t *testing.T) {
	input, err := readInputConfig("in.all.json")
	require.NoError(t, err)
	fw, err := newFakePortForwarder(input)
	require.NoError(t, err)
	err = fw.Start()
	require.NoError(t, err)

	cases := []struct {
		output       string
		expected     string
		resourceType envoyConfigType
	}{
		{
			output:       "json",
			resourceType: BootstrapEnvoyConfigType,
			expected:     "out.bootstrap.json",
		},
		{
			output:       "yaml",
			resourceType: BootstrapEnvoyConfigType,
			expected:     "out.bootstrap.yaml",
		}, {
			output:       "json",
			resourceType: ClusterEnvoyConfigType,
			expected:     "out.cluster.json",
		},
		{
			output:       "yaml",
			resourceType: ClusterEnvoyConfigType,
			expected:     "out.cluster.yaml",
		}, {
			output:       "json",
			resourceType: ListenerEnvoyConfigType,
			expected:     "out.listener.json",
		},
		{
			output:       "yaml",
			resourceType: ListenerEnvoyConfigType,
			expected:     "out.listener.yaml",
		}, {
			output:       "json",
			resourceType: RouteEnvoyConfigType,
			expected:     "out.route.json",
		},
		{
			output:       "yaml",
			resourceType: RouteEnvoyConfigType,
			expected:     "out.route.yaml",
		},
		{
			output:       "json",
			resourceType: EndpointEnvoyConfigType,
			expected:     "out.endpoints.json",
		},
		{
			output:       "yaml",
			resourceType: EndpointEnvoyConfigType,
			expected:     "out.endpoints.yaml",
		},
	}

	for _, tc := range cases {
		t.Run(tc.expected, func(t *testing.T) {
			configDump, err := extractConfigDump(fw, false, tc.resourceType)
			require.NoError(t, err)
			aggregated := sampleAggregatedConfigDump(configDump)
			got, err := marshalEnvoyProxyConfig(aggregated, tc.output)
			require.NoError(t, err)
			if *overrideTestData {
				require.NoError(t, file.Write(string(got), filepath.Join("testdata", "config", "out", tc.expected)))
			}
			out, err := readOutputConfig(tc.expected)
			require.NoError(t, err)
			if tc.output == "yaml" {
				assert.YAMLEq(t, string(out), string(got))
			} else {
				assert.JSONEq(t, string(out), string(got))
			}
		})
	}

	fw.Stop()
}

func TestLabelSelectorBadInput(t *testing.T) {
	podNamespace = "default"

	cases := []struct {
		name   string
		args   []string
		labels []string
	}{
		{
			name:   "no label, no pod name",
			args:   []string{},
			labels: []string{},
		},
		{
			name:   "wrong label, no pod name",
			args:   []string{},
			labels: []string{"foo=bar"},
		},
		{
			name:   "no label, wrong pod name",
			args:   []string{"eg"},
			labels: []string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			labelSelectors = tc.labels
			_, err := retrieveConfigDump(tc.args, false, AllEnvoyConfigType)
			require.Error(t, err, "error not found")
		})
	}
}

func readInputConfig(filename string) ([]byte, error) {
	b, err := os.ReadFile(path.Join("testdata", "config", "in", filename))
	if err != nil {
		return nil, err
	}
	return b, nil
}

func readOutputConfig(filename string) ([]byte, error) {
	b, err := os.ReadFile(path.Join("testdata", "config", "out", filename))
	if err != nil {
		return nil, err
	}
	return b, nil
}

func sampleAggregatedConfigDump(configDump protoreflect.ProtoMessage) aggregatedConfigDump {
	return aggregatedConfigDump{
		defaultNamespace: {
			defaultEnvoyGatewayPodName: configDump,
		},
	}
}
