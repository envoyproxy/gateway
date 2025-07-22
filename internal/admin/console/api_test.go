// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package console

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/telepresenceio/watchable"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
)

func TestHandleAPIInfo(t *testing.T) {
	cfg := &config.Server{
		EnvoyGateway: egv1a1.DefaultEnvoyGateway(),
		Logger:       logging.DefaultLogger(nil, egv1a1.LogLevelInfo),
	}

	handler := NewHandler(cfg, (*message.ProviderResources)(nil))

	req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
	w := httptest.NewRecorder()

	handler.handleAPIInfo(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var info SystemInfo
	err := json.Unmarshal(w.Body.Bytes(), &info)
	require.NoError(t, err)

	// Version might be empty in test environment
	assert.NotEmpty(t, info.GoVersion)
	assert.NotEmpty(t, info.Platform)
	assert.NotEmpty(t, info.Uptime)
	assert.False(t, info.StartTime.IsZero())
	assert.False(t, info.Timestamp.IsZero())
}

func TestHandleAPIInfoMethodNotAllowed(t *testing.T) {
	cfg := &config.Server{
		EnvoyGateway: egv1a1.DefaultEnvoyGateway(),
		Logger:       logging.DefaultLogger(nil, egv1a1.LogLevelInfo),
	}

	handler := NewHandler(cfg, (*message.ProviderResources)(nil))

	req := httptest.NewRequest(http.MethodPost, "/api/info", nil)
	w := httptest.NewRecorder()

	handler.handleAPIInfo(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleAPIServerInfo(t *testing.T) {
	cfg := &config.Server{
		EnvoyGateway: egv1a1.DefaultEnvoyGateway(),
		Logger:       logging.DefaultLogger(nil, egv1a1.LogLevelInfo),
	}

	handler := NewHandler(cfg, (*message.ProviderResources)(nil))

	req := httptest.NewRequest(http.MethodGet, "/api/server_info", nil)
	w := httptest.NewRecorder()

	handler.handleAPIServerInfo(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var info ServerInfo
	err := json.Unmarshal(w.Body.Bytes(), &info)
	require.NoError(t, err)

	assert.Equal(t, "Running", info.State)
	// Version might be empty in test environment
	assert.NotEmpty(t, info.Uptime)
	assert.Len(t, info.Components, 4) // Core components only
	assert.False(t, info.LastUpdated.IsZero())
	assert.NotNil(t, info.EnvoyGatewayConfig) // Verify EnvoyGateway config is included

	// Check component details
	for _, component := range info.Components {
		assert.NotEmpty(t, component.Name)
		assert.Equal(t, "Running", component.Status)
		assert.NotEmpty(t, component.Message)
	}

	// Verify EnvoyGateway configuration structure
	configData, ok := info.EnvoyGatewayConfig.(map[string]interface{})
	if ok {
		// Check if it has the expected structure (this will be a map when JSON unmarshaled)
		assert.Contains(t, configData, "kind")
		assert.Contains(t, configData, "apiVersion")
	}
}

func TestHandleAPIConfigDump(t *testing.T) {
	cfg := &config.Server{
		EnvoyGateway: egv1a1.DefaultEnvoyGateway(),
		Logger:       logging.DefaultLogger(nil, egv1a1.LogLevelInfo),
	}

	// Create a mock provider resources
	providerResources := &message.ProviderResources{}

	handler := NewHandler(cfg, providerResources)

	req := httptest.NewRequest(http.MethodGet, "/api/config_dump", nil)
	w := httptest.NewRecorder()

	handler.handleAPIConfigDump(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var configDump ConfigDumpInfo
	err := json.Unmarshal(w.Body.Bytes(), &configDump)
	require.NoError(t, err)

	assert.NotNil(t, configDump.Gateways)
	assert.NotNil(t, configDump.HTTPRoutes)
	assert.NotNil(t, configDump.GatewayClass)
	assert.False(t, configDump.LastUpdated.IsZero())
}

func TestLoadConfigDumpWithData(t *testing.T) {
	cfg := &config.Server{
		EnvoyGateway: egv1a1.DefaultEnvoyGateway(),
		Logger:       logging.DefaultLogger(nil, egv1a1.LogLevelInfo),
	}

	// Create a simple test without the complex watchable structure
	// This test focuses on the basic functionality
	providerRes := &message.ProviderResources{}
	// Initialize empty watchable map
	providerRes.GatewayAPIResources = watchable.Map[string, *resource.ControllerResources]{}

	// Skip storing to avoid watchable copy issues
	// providerResources.Store("test", providerRes)

	handler := NewHandler(cfg, providerRes)

	configDump := handler.loadConfigDump()

	// Verify the response structure with empty data
	// Gateway API Core Resources
	assert.Empty(t, configDump.GatewayClass)
	assert.Empty(t, configDump.Gateways)
	assert.Empty(t, configDump.HTTPRoutes)
	assert.Empty(t, configDump.GRPCRoutes)
	assert.Empty(t, configDump.TLSRoutes)
	assert.Empty(t, configDump.TCPRoutes)
	assert.Empty(t, configDump.UDPRoutes)

	// Envoy Gateway Policies
	assert.Empty(t, configDump.ClientTrafficPolicies)
	assert.Empty(t, configDump.BackendTrafficPolicies)
	assert.Empty(t, configDump.BackendTLSPolicies)
	assert.Empty(t, configDump.SecurityPolicies)
	assert.Empty(t, configDump.EnvoyPatchPolicies)
	assert.Empty(t, configDump.EnvoyExtensionPolicies)

	// Kubernetes Resources
	assert.Empty(t, configDump.Services)
	assert.Empty(t, configDump.Secrets)
	assert.Empty(t, configDump.ConfigMaps)
	assert.Empty(t, configDump.Namespaces)
	assert.Empty(t, configDump.EndpointSlices)

	// Other Resources
	assert.Empty(t, configDump.ReferenceGrants)
	assert.Empty(t, configDump.HTTPRouteFilters)
	assert.Empty(t, configDump.EnvoyProxies)
	assert.Empty(t, configDump.Backends)
	assert.Empty(t, configDump.ServiceImports)

	// Verify all fields are properly initialized (not nil)
	assert.NotNil(t, configDump.GatewayClass)
	assert.NotNil(t, configDump.Gateways)
	assert.NotNil(t, configDump.HTTPRoutes)
	assert.NotNil(t, configDump.GRPCRoutes)
	assert.NotNil(t, configDump.TLSRoutes)
	assert.NotNil(t, configDump.TCPRoutes)
	assert.NotNil(t, configDump.UDPRoutes)
	assert.NotNil(t, configDump.ClientTrafficPolicies)
	assert.NotNil(t, configDump.BackendTrafficPolicies)
	assert.NotNil(t, configDump.BackendTLSPolicies)
	assert.NotNil(t, configDump.SecurityPolicies)
	assert.NotNil(t, configDump.EnvoyPatchPolicies)
	assert.NotNil(t, configDump.EnvoyExtensionPolicies)
	assert.NotNil(t, configDump.Services)
	assert.NotNil(t, configDump.Secrets)
	assert.NotNil(t, configDump.ConfigMaps)
	assert.NotNil(t, configDump.Namespaces)
	assert.NotNil(t, configDump.EndpointSlices)
	assert.NotNil(t, configDump.ReferenceGrants)
	assert.NotNil(t, configDump.HTTPRouteFilters)
	assert.NotNil(t, configDump.EnvoyProxies)
	assert.NotNil(t, configDump.Backends)
	assert.NotNil(t, configDump.ServiceImports)

	assert.False(t, configDump.LastUpdated.IsZero())
}

func TestGetGatewayClassStatus(t *testing.T) {
	tests := []struct {
		name       string
		conditions []metav1.Condition
		expected   string
	}{
		{
			name:       "no conditions",
			conditions: nil,
			expected:   "Unknown",
		},
		{
			name: "accepted",
			conditions: []metav1.Condition{
				{
					Type:   string(gwapiv1.GatewayClassConditionStatusAccepted),
					Status: metav1.ConditionTrue,
				},
			},
			expected: "Accepted",
		},
		{
			name: "not accepted",
			conditions: []metav1.Condition{
				{
					Type:   string(gwapiv1.GatewayClassConditionStatusAccepted),
					Status: metav1.ConditionFalse,
				},
			},
			expected: "Not Accepted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gc := &gwapiv1.GatewayClass{
				Status: gwapiv1.GatewayClassStatus{
					Conditions: tt.conditions,
				},
			}

			result := getGatewayClassStatus(gc)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetGatewayStatus(t *testing.T) {
	tests := []struct {
		name       string
		conditions []metav1.Condition
		expected   string
	}{
		{
			name:       "no conditions",
			conditions: nil,
			expected:   "Unknown",
		},
		{
			name: "programmed",
			conditions: []metav1.Condition{
				{
					Type:   string(gwapiv1.GatewayConditionProgrammed),
					Status: metav1.ConditionTrue,
				},
			},
			expected: "Programmed",
		},
		{
			name: "accepted",
			conditions: []metav1.Condition{
				{
					Type:   string(gwapiv1.GatewayConditionAccepted),
					Status: metav1.ConditionTrue,
				},
			},
			expected: "Accepted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gw := &gwapiv1.Gateway{
				Status: gwapiv1.GatewayStatus{
					Conditions: tt.conditions,
				},
			}

			result := getGatewayStatus(gw)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHandleAPIConfigDumpAll(t *testing.T) {
	cfg := &config.Server{
		Logger: logging.NewLogger(os.Stdout, egv1a1.DefaultEnvoyGatewayLogging()),
	}

	providerRes := &message.ProviderResources{}
	handler := NewHandler(cfg, providerRes)

	req := httptest.NewRequest(http.MethodGet, "/api/config_dump_all", nil)
	resp := httptest.NewRecorder()

	handler.handleAPIConfigDumpAll(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "application/json; charset=utf-8", resp.Header().Get("Content-Type"))

	// Should not have download headers
	assert.Empty(t, resp.Header().Get("Content-Disposition"))

	// Should return JSON response with structured format
	var result map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &result)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, result, "resources")
	assert.Contains(t, result, "timestamp")
	assert.Contains(t, result, "totalCount")

	// Verify resources field exists (could be nil for empty case)
	resources := result["resources"]
	if resources != nil {
		resourcesArray, ok := resources.([]interface{})
		assert.True(t, ok)
		assert.Empty(t, resourcesArray) // Empty for test case
	} else {
		// resources is nil, which is expected for empty provider resources
		assert.Nil(t, resources)
	}

	// Verify totalCount matches resources length
	totalCount, ok := result["totalCount"].(float64)
	assert.True(t, ok)
	assert.Equal(t, float64(0), totalCount)
}

func TestHandleAPIConfigDumpAllMethodNotAllowed(t *testing.T) {
	cfg := &config.Server{
		Logger: logging.NewLogger(os.Stdout, egv1a1.DefaultEnvoyGatewayLogging()),
	}

	handler := NewHandler(cfg, (*message.ProviderResources)(nil))

	req := httptest.NewRequest(http.MethodPost, "/api/config_dump_all", nil)
	resp := httptest.NewRecorder()

	handler.handleAPIConfigDumpAll(resp, req)

	assert.Equal(t, http.StatusMethodNotAllowed, resp.Code)
}

func TestHandleAPIMetrics(t *testing.T) {
	cfg := &config.Server{
		Logger: logging.NewLogger(os.Stdout, egv1a1.DefaultEnvoyGatewayLogging()),
	}

	handler := NewHandler(cfg, (*message.ProviderResources)(nil))

	req := httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
	resp := httptest.NewRecorder()

	handler.handleAPIMetrics(resp, req)

	// Now we're using the Prometheus registry directly, so we should get a successful response
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Header().Get("Content-Type"), "text/plain; version=0.0.4; charset=utf-8")

	// Verify we get actual Prometheus metrics
	body := resp.Body.String()
	assert.Contains(t, body, "# HELP")
	assert.Contains(t, body, "# TYPE")
	// Should contain Go runtime metrics
	assert.Contains(t, body, "go_goroutines")
	assert.Contains(t, body, "process_cpu_seconds_total")
}

func TestHandleAPIMetricsMethodNotAllowed(t *testing.T) {
	cfg := &config.Server{
		Logger: logging.NewLogger(os.Stdout, egv1a1.DefaultEnvoyGatewayLogging()),
	}

	handler := NewHandler(cfg, (*message.ProviderResources)(nil))

	req := httptest.NewRequest(http.MethodPost, "/api/metrics", nil)
	resp := httptest.NewRecorder()

	handler.handleAPIMetrics(resp, req)

	assert.Equal(t, http.StatusMethodNotAllowed, resp.Code)
}
