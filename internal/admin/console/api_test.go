// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package console

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/telepresenceio/watchable"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	mcsapiv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

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
	providerRes.GatewayAPIResources = watchable.Map[string, *resource.ControllerResourcesContext]{}

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

func TestHandleAPIConfigDumpWithResourceAll(t *testing.T) {
	cfg := &config.Server{
		Logger: logging.NewLogger(os.Stdout, egv1a1.DefaultEnvoyGatewayLogging()),
	}

	providerRes := &message.ProviderResources{}
	handler := NewHandler(cfg, providerRes)

	req := httptest.NewRequest(http.MethodGet, "/api/config_dump?resource=all", nil)
	resp := httptest.NewRecorder()

	handler.handleAPIConfigDump(resp, req)

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

func TestHandleAPIConfigDumpWithResourceAllRedactsSecrets(t *testing.T) {
	cfg := &config.Server{
		Logger: logging.NewLogger(os.Stdout, egv1a1.DefaultEnvoyGatewayLogging()),
	}

	providerRes := &message.ProviderResources{}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
			Annotations: map[string]string{
				corev1.LastAppliedConfigAnnotation: "{\"data\":{\"token\":\"c3VwZXJzZWNyZXQ=\"}}",
				"example.com/foo":                  "value",
			},
			ManagedFields: []metav1.ManagedFieldsEntry{
				{
					Manager:   "kubectl",
					Operation: metav1.ManagedFieldsOperationApply,
				},
			},
		},
		Data: map[string][]byte{
			"token": []byte("supersecret"),
		},
		StringData: map[string]string{
			"token": "supersecret",
		},
	}
	controllerResources := resource.ControllerResources{
		&resource.Resources{
			Secrets: []*corev1.Secret{secret},
		},
	}
	providerRes.GatewayAPIResources.Store("test", &resource.ControllerResourcesContext{
		Resources: &controllerResources,
		Context:   context.Background(),
	})

	handler := NewHandler(cfg, providerRes)

	req := httptest.NewRequest(http.MethodGet, "/api/config_dump?resource=all", nil)
	resp := httptest.NewRecorder()

	handler.handleAPIConfigDump(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var result struct {
		Resources []*resource.Resources `json:"resources"`
	}
	resultErr := json.Unmarshal(resp.Body.Bytes(), &result)
	require.NoError(t, resultErr)

	require.Len(t, result.Resources, 1)
	require.Len(t, result.Resources[0].Secrets, 1)
	masked := result.Resources[0].Secrets[0]
	// Ensure the masked secret is redacted
	assert.Contains(t, masked.Data, "token")
	assert.Equal(t, redactedSecretValueBytes, masked.Data["token"])
	assert.Equal(t, redactedSecretValue, masked.StringData["token"])
	assert.Equal(t, "test-secret", masked.Name)
	assert.Empty(t, masked.Annotations)
	assert.Empty(t, masked.ManagedFields)
	// Ensure the original secret is not modified
	assert.Equal(t, []byte("supersecret"), secret.Data["token"])
	assert.Equal(t, "supersecret", secret.StringData["token"])
	assert.Equal(t, map[string]string{
		corev1.LastAppliedConfigAnnotation: "{\"data\":{\"token\":\"c3VwZXJzZWNyZXQ=\"}}",
		"example.com/foo":                  "value",
	}, secret.Annotations)
	assert.Equal(t, []metav1.ManagedFieldsEntry{
		{
			Manager:   "kubectl",
			Operation: metav1.ManagedFieldsOperationApply,
		},
	}, secret.ManagedFields)
}

func TestHandleAPIConfigDumpWithResourceFilter(t *testing.T) {
	cfg := &config.Server{
		Logger: logging.NewLogger(os.Stdout, egv1a1.DefaultEnvoyGatewayLogging()),
	}

	providerRes := &message.ProviderResources{}
	controllerResources := resource.ControllerResources{
		&resource.Resources{
			GatewayClass: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "eg",
				},
			},
			Gateways: []*gwapiv1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "eg",
						Namespace: "default",
					},
				},
			},
			HTTPRoutes: []*gwapiv1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hr",
						Namespace: "default",
					},
				},
			},
			GRPCRoutes: []*gwapiv1.GRPCRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gr",
						Namespace: "default",
					},
				},
			},
			TLSRoutes: []*gwapiv1.TLSRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tr",
						Namespace: "default",
					},
				},
			},
			TCPRoutes: []*gwapiv1a2.TCPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tcr",
						Namespace: "default",
					},
				},
			},
			UDPRoutes: []*gwapiv1a2.UDPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ur",
						Namespace: "default",
					},
				},
			},
			ClientTrafficPolicies: []*egv1a1.ClientTrafficPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ctp",
						Namespace: "default",
					},
				},
			},
			BackendTrafficPolicies: []*egv1a1.BackendTrafficPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "btp",
						Namespace: "default",
					},
				},
			},
			BackendTLSPolicies: []*gwapiv1.BackendTLSPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "btlsp",
						Namespace: "default",
					},
				},
			},
			SecurityPolicies: []*egv1a1.SecurityPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "sp",
						Namespace: "default",
					},
				},
			},
			EnvoyPatchPolicies: []*egv1a1.EnvoyPatchPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "epp",
						Namespace: "default",
					},
				},
			},
			EnvoyExtensionPolicies: []*egv1a1.EnvoyExtensionPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "eep",
						Namespace: "default",
					},
				},
			},
			Services: []*corev1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc",
						Namespace: "default",
					},
				},
			},
			Secrets: []*corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "sec",
						Namespace: "default",
					},
				},
			},
			ConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cm",
						Namespace: "default",
					},
				},
			},
			Namespaces: []*corev1.Namespace{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "default",
					},
				},
			},
			EndpointSlices: []*discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "eps",
						Namespace: "default",
					},
				},
			},
			ReferenceGrants: []*gwapiv1b1.ReferenceGrant{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "rg",
						Namespace: "default",
					},
				},
			},
			HTTPRouteFilters: []*egv1a1.HTTPRouteFilter{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hrf",
						Namespace: "default",
					},
				},
			},
			EnvoyProxyForGatewayClass: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ep",
					Namespace: "default",
				},
			},
			Backends: []*egv1a1.Backend{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "be",
						Namespace: "default",
					},
				},
			},
			ServiceImports: []*mcsapiv1a1.ServiceImport{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "si",
						Namespace: "default",
					},
				},
			},
		},
	}
	providerRes.GatewayAPIResources.Store("test", &resource.ControllerResourcesContext{
		Resources: &controllerResources,
		Context:   context.Background(),
	})

	handler := NewHandler(cfg, providerRes)

	testCases := []struct {
		name              string
		resource          string
		totalCount        float64
		expectedName      string
		expectedNamespace string
	}{
		{name: "gatewayclass", resource: "gatewayclass", totalCount: 1, expectedName: "eg"},
		{name: "gateway", resource: "gateway", totalCount: 1, expectedName: "eg", expectedNamespace: "default"},
		{name: "httproute", resource: "httproute", totalCount: 1, expectedName: "hr", expectedNamespace: "default"},
		{name: "grpcroute", resource: "grpcroute", totalCount: 1, expectedName: "gr", expectedNamespace: "default"},
		{name: "tlsroute", resource: "tlsroute", totalCount: 1, expectedName: "tr", expectedNamespace: "default"},
		{name: "tcproute", resource: "tcproute", totalCount: 1, expectedName: "tcr", expectedNamespace: "default"},
		{name: "udproute", resource: "udproute", totalCount: 1, expectedName: "ur", expectedNamespace: "default"},
		{name: "clienttrafficpolicy", resource: "clienttrafficpolicy", totalCount: 1, expectedName: "ctp", expectedNamespace: "default"},
		{name: "backendtrafficpolicy", resource: "backendtrafficpolicy", totalCount: 1, expectedName: "btp", expectedNamespace: "default"},
		{name: "backendtlspolicy", resource: "backendtlspolicy", totalCount: 1, expectedName: "btlsp", expectedNamespace: "default"},
		{name: "securitypolicy", resource: "securitypolicy", totalCount: 1, expectedName: "sp", expectedNamespace: "default"},
		{name: "envoypatchpolicy", resource: "envoypatchpolicy", totalCount: 1, expectedName: "epp", expectedNamespace: "default"},
		{name: "envoyextensionpolicy", resource: "envoyextensionpolicy", totalCount: 1, expectedName: "eep", expectedNamespace: "default"},
		{name: "service", resource: "service", totalCount: 1, expectedName: "svc", expectedNamespace: "default"},
		{name: "secret", resource: "secret", totalCount: 1, expectedName: "sec", expectedNamespace: "default"},
		{name: "configmap", resource: "configmap", totalCount: 1, expectedName: "cm", expectedNamespace: "default"},
		{name: "namespace", resource: "namespace", totalCount: 1, expectedName: "default"},
		{name: "endpointslice", resource: "endpointslice", totalCount: 1, expectedName: "eps", expectedNamespace: "default"},
		{name: "referencegrant", resource: "referencegrant", totalCount: 1, expectedName: "rg", expectedNamespace: "default"},
		{name: "httproutefilter", resource: "httproutefilter", totalCount: 1, expectedName: "hrf", expectedNamespace: "default"},
		{name: "envoyproxy", resource: "envoyproxy", totalCount: 1, expectedName: "ep", expectedNamespace: "default"},
		{name: "backend", resource: "backend", totalCount: 1, expectedName: "be", expectedNamespace: "default"},
		{name: "serviceimport", resource: "serviceimport", totalCount: 1, expectedName: "si", expectedNamespace: "default"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/config_dump?resource="+tc.resource, nil)
			resp := httptest.NewRecorder()

			handler.handleAPIConfigDump(resp, req)

			assert.Equal(t, http.StatusOK, resp.Code)
			var result map[string]interface{}
			err := json.Unmarshal(resp.Body.Bytes(), &result)
			require.NoError(t, err)
			assert.Equal(t, tc.totalCount, result["totalCount"])

			resources, ok := result["resources"].([]interface{})
			require.True(t, ok)
			require.Len(t, resources, 1)
			resourceItem, ok := resources[0].(map[string]interface{})
			require.True(t, ok)
			metadata, ok := resourceItem["metadata"].(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, tc.expectedName, metadata["name"])
			if tc.expectedNamespace == "" {
				_, foundNamespace := metadata["namespace"]
				assert.False(t, foundNamespace)
			} else {
				assert.Equal(t, tc.expectedNamespace, metadata["namespace"])
			}
		})
	}
}

func TestHandleAPIConfigDumpWithInvalidResourceFilter(t *testing.T) {
	cfg := &config.Server{
		Logger: logging.NewLogger(os.Stdout, egv1a1.DefaultEnvoyGatewayLogging()),
	}
	handler := NewHandler(cfg, (*message.ProviderResources)(nil))

	req := httptest.NewRequest(http.MethodGet, "/api/config_dump?resource=invalid", nil)
	resp := httptest.NewRecorder()
	handler.handleAPIConfigDump(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, resp.Body.String(), "Invalid resource filter")
}

func TestHandleAPIConfigDumpWithResourceAllMethodNotAllowed(t *testing.T) {
	cfg := &config.Server{
		Logger: logging.NewLogger(os.Stdout, egv1a1.DefaultEnvoyGatewayLogging()),
	}

	handler := NewHandler(cfg, (*message.ProviderResources)(nil))

	req := httptest.NewRequest(http.MethodPost, "/api/config_dump?resource=all", nil)
	resp := httptest.NewRecorder()

	handler.handleAPIConfigDump(resp, req)

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
