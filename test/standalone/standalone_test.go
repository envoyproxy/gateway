// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build standalone

package standalone

import (
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	testutils "github.com/envoyproxy/gateway/test/utils"
	"github.com/envoyproxy/gateway/test/utils/testotel"
)

//go:embed testdata/otel-grpc-headers/envoy-gateway.yaml
var otelGRPCHeadersEnvoyGateway string

//go:embed testdata/otel-grpc-headers/resources/gateway.yaml
var otelGRPCHeadersGateway string

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

func TestOTELGRPCHeaders(t *testing.T) {
	ports := testutils.RequireRandomPorts(t, 2)
	listenerPort := ports[0]
	adminPort := ports[1]

	// Start a local backend server
	backendListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	backendPort := backendListener.Addr().(*net.TCPAddr).Port
	backendServer := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status": "ok"}`))
		}),
	}
	go func() { _ = backendServer.Serve(backendListener) }()
	defer backendServer.Close()

	collector, err := testotel.StartGRPCCollector()
	require.NoError(t, err)
	defer collector.Close()

	testDir := t.TempDir()
	resourcesDir := filepath.Join(testDir, "resources")
	require.NoError(t, os.Mkdir(resourcesDir, 0o755))

	// Replace default ports with actual ports
	replacements := map[string]string{
		"port: 10080": "port: " + strconv.Itoa(listenerPort),
		"port: 19000": "port: " + strconv.Itoa(adminPort),
		"port: 4317":  "port: " + strconv.Itoa(collector.Port()),
		"port: 8080":  "port: " + strconv.Itoa(backendPort),
	}

	envoyGatewayYAML := replaceTokens(otelGRPCHeadersEnvoyGateway, replacements)
	gatewayYAML := replaceTokens(otelGRPCHeadersGateway, replacements)

	require.NoError(t, os.WriteFile(filepath.Join(testDir, "envoy-gateway.yaml"), []byte(envoyGatewayYAML), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(resourcesDir, "gateway.yaml"), []byte(gatewayYAML), 0o600))

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

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/get", listenerPort))
	require.NoError(t, err)
	resp.Body.Close()

	// Verify collector received an access log with the expected format and Authorization header
	log := collector.TakeLog()
	require.NotNil(t, log)
	require.Contains(t, log.Body.GetStringValue(), `HTTP/1.1" 200`)
	require.Equal(t, "Bearer test-api-key", testotel.GetAttributeString(log.Attributes, "grpc.metadata.authorization"))

	// Verify collector received a trace span with Authorization header
	span := collector.TakeSpan()
	require.NotNil(t, span)
	require.Equal(t, "Bearer test-api-key", testotel.GetAttributeString(span.Attributes, "grpc.metadata.authorization"))

	// Verify collector received a metric with Authorization header
	resourceMetrics := collector.TakeMetric()
	require.NotNil(t, resourceMetrics)
	require.Equal(t, "Bearer test-api-key", testotel.GetAttributeString(resourceMetrics.Resource.Attributes, "grpc.metadata.authorization"))
}

func replaceTokens(content string, replacements map[string]string) string {
	result := content
	for token, value := range replacements {
		result = strings.ReplaceAll(result, token, value)
	}
	return result
}
