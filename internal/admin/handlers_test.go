// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleWebUI(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleWebUI)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "text/html", rr.Header().Get("Content-Type"))
	assert.Contains(t, rr.Body.String(), "Envoy Gateway Admin")
}

func TestHandleSVG(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "logo.svg",
			filename: "logo.svg",
		},
		{
			name:     "cncf.svg",
			filename: "cncf.svg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/"+tt.filename, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler := func(w http.ResponseWriter, r *http.Request) {
				handleSVG(w, r, tt.filename)
			}

			http.HandlerFunc(handler).ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, "image/svg+xml", rr.Header().Get("Content-Type"))
			assert.Contains(t, rr.Body.String(), "<svg")
		})
	}
}

func TestHandleStats(t *testing.T) {
	req, err := http.NewRequest("GET", "/admin/api/stats", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleStats)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response StatsResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "runtime")
	assert.Contains(t, response, "system")

	runtime, ok := response["runtime"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, runtime, "goroutines")
	assert.Contains(t, runtime, "memory_alloc")

	system, ok := response["system"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, system, "uptime")
	assert.Contains(t, system, "cpu_count")
}

func TestHandleResourcesWithoutClient(t *testing.T) {
	// Test without k8s client
	SetK8sClient(nil)

	req, err := http.NewRequest("GET", "/admin/api/resources", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleResources)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	assert.Contains(t, rr.Body.String(), "Kubernetes client not available")
}

func TestSetK8sClient(t *testing.T) {
	// Test setting and getting k8s client
	originalClient := k8sClient

	// Set to nil
	SetK8sClient(nil)
	assert.Nil(t, k8sClient)

	// Restore original
	SetK8sClient(originalClient)
	assert.Equal(t, originalClient, k8sClient)
}

func TestStatusResponseStructure(t *testing.T) {
	response := StatusResponse{
		Version:      "test-version",
		Uptime:       "1h30m",
		Status:       "Running",
		ConfigStatus: "Loaded",
	}

	data, err := json.Marshal(response)
	require.NoError(t, err)

	var unmarshaled StatusResponse
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, response.Version, unmarshaled.Version)
	assert.Equal(t, response.Uptime, unmarshaled.Uptime)
	assert.Equal(t, response.Status, unmarshaled.Status)
	assert.Equal(t, response.ConfigStatus, unmarshaled.ConfigStatus)
}

func TestResourceInfoStructure(t *testing.T) {
	info := ResourceInfo{
		Name:      "test-resource",
		Namespace: "test-namespace",
		Status:    "True",
		Reason:    "Accepted",
		Message:   "Resource is valid",
	}

	data, err := json.Marshal(info)
	require.NoError(t, err)

	var unmarshaled ResourceInfo
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, info.Name, unmarshaled.Name)
	assert.Equal(t, info.Namespace, unmarshaled.Namespace)
	assert.Equal(t, info.Status, unmarshaled.Status)
	assert.Equal(t, info.Reason, unmarshaled.Reason)
	assert.Equal(t, info.Message, unmarshaled.Message)
}

func TestResourcesResponseStructure(t *testing.T) {
	response := ResourcesResponse{
		"Gateways": []ResourceInfo{
			{
				Name:      "gateway-1",
				Namespace: "default",
				Status:    "True",
				Reason:    "Programmed",
				Message:   "Gateway is ready",
			},
		},
		"HTTPRoutes": []ResourceInfo{
			{
				Name:      "route-1",
				Namespace: "default",
				Status:    "True",
				Reason:    "ResolvedRefs",
				Message:   "Route is valid",
			},
		},
	}

	data, err := json.Marshal(response)
	require.NoError(t, err)

	var unmarshaled ResourcesResponse
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Len(t, unmarshaled["Gateways"], 1)
	assert.Len(t, unmarshaled["HTTPRoutes"], 1)
	assert.Equal(t, "gateway-1", unmarshaled["Gateways"][0].Name)
	assert.Equal(t, "route-1", unmarshaled["HTTPRoutes"][0].Name)
}

func TestStatsResponseStructure(t *testing.T) {
	response := StatsResponse{
		"runtime": map[string]interface{}{
			"goroutines":   10,
			"memory_alloc": 1024,
		},
		"system": map[string]interface{}{
			"uptime":    "1h30m",
			"cpu_count": 4,
		},
	}

	data, err := json.Marshal(response)
	require.NoError(t, err)

	var unmarshaled StatsResponse
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	runtime, ok := unmarshaled["runtime"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(10), runtime["goroutines"]) // JSON unmarshals numbers as float64

	system, ok := unmarshaled["system"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "1h30m", system["uptime"])
}
