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
	"testing"

	"github.com/stretchr/testify/assert"

	kube "github.com/envoyproxy/gateway/internal/kubernetes"
	netutil "github.com/envoyproxy/gateway/internal/utils/net"
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
	assert.NoError(t, err)
	fw, err := newFakePortForwarder(input)
	assert.NoError(t, err)
	err = fw.Start()
	assert.NoError(t, err)

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
		t.Run(tc.output, func(t *testing.T) {
			configDump, err := extractConfigDump(fw)
			assert.NoError(t, err)
			got, err := marshalEnvoyProxyConfig(configDump, tc.output)
			assert.NoError(t, err)
			out, err := readOutputConfig(tc.expected)
			assert.NoError(t, err)
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
	assert.NoError(t, err)
	fw, err := newFakePortForwarder(input)
	assert.NoError(t, err)
	err = fw.Start()
	assert.NoError(t, err)

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
	}

	for _, tc := range cases {
		t.Run(tc.output, func(t *testing.T) {
			configDump, err := extractConfigDump(fw)
			assert.NoError(t, err)
			resource, err := findXDSResourceFromConfigDump(tc.resourceType, configDump)
			assert.NoError(t, err)
			got, err := marshalEnvoyProxyConfig(resource, tc.output)
			assert.NoError(t, err)
			out, err := readOutputConfig(tc.expected)
			assert.NoError(t, err)
			if tc.output == "yaml" {
				assert.YAMLEq(t, string(out), string(got))
			} else {
				assert.JSONEq(t, string(out), string(got))
			}
		})
	}

	fw.Stop()
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
