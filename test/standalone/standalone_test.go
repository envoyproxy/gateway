// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build standalone

package standalone

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/require"

	testutils "github.com/envoyproxy/gateway/test/utils"
	"github.com/envoyproxy/gateway/test/utils/testotel"
)

//go:embed testdata/accesslog-otel-headers/envoy-gateway.yaml.tmpl
//go:embed testdata/accesslog-otel-headers/resources/envoy-proxy.yaml.tmpl
//go:embed testdata/accesslog-otel-headers/resources/gateway.yaml.tmpl
var testdataFS embed.FS

// gatewayBin is the path to the compiled envoy-gateway binary.
var gatewayBin string

// TestMain builds the binary once for all tests.
func TestMain(m *testing.M) {
	var err error
	if gatewayBin, err = testutils.BuildGoBinaryOnDemand("ENVOY_GATEWAY_BIN", "envoy-gateway", "./cmd/envoy-gateway"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build envoy-gateway: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// TestConfig contains template parameters for generating test resources.
type TestConfig struct {
	ListenerPort  int
	CollectorPort int
	XDSPort       int
	WASMPort      int
	AdminPort     int
}

func TestAccessLogOTELHeaders(t *testing.T) {
	ports := testutils.RequireRandomPorts(t, 4)
	listenerPort := ports[0]
	xdsPort := ports[1]
	wasmPort := ports[2]
	adminPort := ports[3]

	collector, err := testotel.StartGRPCCollector()
	require.NoError(t, err)
	defer collector.Close()

	testDir := t.TempDir()
	cfg := TestConfig{
		ListenerPort:  listenerPort,
		CollectorPort: collector.Port(),
		XDSPort:       xdsPort,
		WASMPort:      wasmPort,
		AdminPort:     adminPort,
	}
	generateFromTemplates(t, testDir, cfg)

	cmd := exec.Command(gatewayBin, "server", "--config-path", filepath.Join(testDir, "envoy-gateway.yaml"))
	cmd.Dir = testDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Start())
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	// Wait for Envoy to be ready
	require.Eventually(t, func() bool {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", listenerPort))
		if err != nil {
			return false
		}
		resp.Body.Close()
		return true
	}, 30*time.Second, 500*time.Millisecond, "envoy not ready on port %d", listenerPort)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/get", listenerPort), nil)
	require.NoError(t, err)
	req.Host = "127.0.0.1.nip.io"
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	// Verify collector received the configured Authorization header
	require.True(t, collector.WaitForHeader("authorization", "Bearer test-api-key", 10*time.Second),
		"expected Authorization header not received")
}

func generateFromTemplates(t *testing.T, dir string, cfg TestConfig) {
	t.Helper()

	resourcesDir := filepath.Join(dir, "resources")
	require.NoError(t, os.MkdirAll(resourcesDir, 0o755))

	// Process each .yaml.tmpl file and write .yaml output
	templates := []struct {
		src string
		dst string
	}{
		{"testdata/accesslog-otel-headers/envoy-gateway.yaml.tmpl", "envoy-gateway.yaml"},
		{"testdata/accesslog-otel-headers/resources/envoy-proxy.yaml.tmpl", "resources/envoy-proxy.yaml"},
		{"testdata/accesslog-otel-headers/resources/gateway.yaml.tmpl", "resources/gateway.yaml"},
	}

	for _, tmplDef := range templates {
		content, err := testdataFS.ReadFile(tmplDef.src)
		require.NoError(t, err)

		tmpl, err := template.New(tmplDef.src).Parse(string(content))
		require.NoError(t, err)

		f, err := os.Create(filepath.Join(dir, tmplDef.dst))
		require.NoError(t, err)

		require.NoError(t, tmpl.Execute(f, cfg))
		require.NoError(t, f.Close())
	}
}
