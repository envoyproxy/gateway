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

func TestExtractConfigDump(t *testing.T) {
	fw, err := newFakePortForwarder(readConfig("in.json"))
	assert.NoError(t, err)
	err = fw.Start()
	assert.NoError(t, err)

	cases := []struct {
		output   string
		expected string
	}{
		{
			output:   "json",
			expected: "out.json",
		},
		{
			output:   "yaml",
			expected: "out.yaml",
		},
	}

	for _, tc := range cases {
		t.Run(tc.output, func(t *testing.T) {
			got, err := extractConfigDump(fw, tc.output)
			assert.NoError(t, err)
			assert.Equal(t, string(readConfig(tc.expected)), string(got))
		})
	}

	fw.Stop()
}

func readConfig(filename string) []byte {
	b, _ := os.ReadFile(path.Join("testdata", "config", filename))
	return b
}
