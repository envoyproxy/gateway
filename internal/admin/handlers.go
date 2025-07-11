// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/cmd/version"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

// StatusResponse represents the overall status of Envoy Gateway
type StatusResponse struct {
	Version      string `json:"version"`
	Uptime       string `json:"uptime"`
	Status       string `json:"status"`
	ConfigStatus string `json:"configStatus"`
	Config       any    `json:"config"`
}

// ResourceInfo represents information about a Gateway resource
type ResourceInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Status    string `json:"status"`
	Reason    string `json:"reason,omitempty"`
	Message   string `json:"message,omitempty"`
}

// ResourcesResponse represents all Gateway resources
type ResourcesResponse map[string][]ResourceInfo

// StatsResponse represents statistics information
type StatsResponse map[string]interface{}

var (
	startTime = time.Now()
	k8sClient client.Client
)

// SetK8sClient sets the Kubernetes client for resource queries
func SetK8sClient(client client.Client) {
	k8sClient = client
}

// handleWebUI serves the web UI HTML page
func handleWebUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Read the embedded HTML file
	content, err := webuiFS.ReadFile("webui.html")
	if err != nil {
		http.Error(w, "Failed to load web UI", http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(content); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

// handleSVG serves SVG files
func handleSVG(w http.ResponseWriter, _ *http.Request, filename string) {
	w.Header().Set("Content-Type", "image/svg+xml")

	// Read the embedded SVG file
	content, err := webuiFS.ReadFile(filename)
	if err != nil {
		http.Error(w, "Failed to load SVG file", http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(content); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

// handleStatus returns the overall status of Envoy Gateway
func handleStatus(cfg *config.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		uptime := time.Since(startTime).String()
		response := StatusResponse{
			Version:      version.Get().EnvoyGatewayVersion,
			Uptime:       uptime,
			Status:       "Running",
			ConfigStatus: "Loaded",
			Config:       cfg.ServerConfiguration,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		}
	}
}

// handleResources returns information about Gateway resources
func handleResources(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if k8sClient == nil {
		http.Error(w, "Kubernetes client not available", http.StatusServiceUnavailable)
		return
	}

	ctx := context.Background()
	response := make(ResourcesResponse)

	// Get GatewayClasses
	if gatewayClasses, err := getGatewayClasses(ctx); err == nil {
		response["GatewayClasses"] = gatewayClasses
	}

	// Get Gateways
	if gateways, err := getGateways(ctx); err == nil {
		response["Gateways"] = gateways
	}

	// Get HTTPRoutes
	if httpRoutes, err := getHTTPRoutes(ctx); err == nil {
		response["HTTPRoutes"] = httpRoutes
	}

	// Get GRPCRoutes
	if grpcRoutes, err := getGRPCRoutes(ctx); err == nil {
		response["GRPCRoutes"] = grpcRoutes
	}

	// Get TCPRoutes
	if tcpRoutes, err := getTCPRoutes(ctx); err == nil {
		response["TCPRoutes"] = tcpRoutes
	}

	// Get UDPRoutes
	if udpRoutes, err := getUDPRoutes(ctx); err == nil {
		response["UDPRoutes"] = udpRoutes
	}

	// Get TLSRoutes
	if tlsRoutes, err := getTLSRoutes(ctx); err == nil {
		response["TLSRoutes"] = tlsRoutes
	}

	// Get BackendTLSPolicies
	if backendTLSPolicies, err := getBackendTLSPolicies(ctx); err == nil {
		response["BackendTLSPolicies"] = backendTLSPolicies
	}

	// Get BackendTrafficPolicies
	if backendTrafficPolicies, err := getBackendTrafficPolicies(ctx); err == nil {
		response["BackendTrafficPolicies"] = backendTrafficPolicies
	}

	// Get ClientTrafficPolicies
	if clientTrafficPolicies, err := getClientTrafficPolicies(ctx); err == nil {
		response["ClientTrafficPolicies"] = clientTrafficPolicies
	}

	// Get SecurityPolicies
	if securityPolicies, err := getSecurityPolicies(ctx); err == nil {
		response["SecurityPolicies"] = securityPolicies
	}

	// Get EnvoyPatchPolicies
	if envoyPatchPolicies, err := getEnvoyPatchPolicies(ctx); err == nil {
		response["EnvoyPatchPolicies"] = envoyPatchPolicies
	}

	// Get EnvoyExtensionPolicies
	if envoyExtensionPolicies, err := getEnvoyExtensionPolicies(ctx); err == nil {
		response["EnvoyExtensionPolicies"] = envoyExtensionPolicies
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
	}
}

// handleStats returns statistics information
func handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	response := StatsResponse{
		"runtime": map[string]interface{}{
			"goroutines":   runtime.NumGoroutine(),
			"memory_alloc": memStats.Alloc,
			"memory_total": memStats.TotalAlloc,
			"memory_sys":   memStats.Sys,
			"gc_runs":      memStats.NumGC,
			"last_gc":      time.Unix(0, int64(memStats.LastGC)).Format(time.RFC3339),
		},
		"system": map[string]interface{}{
			"uptime":     time.Since(startTime).String(),
			"cpu_count":  runtime.NumCPU(),
			"go_version": runtime.Version(),
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
	}
}

// Helper functions to get resources
func getGatewayClasses(ctx context.Context) ([]ResourceInfo, error) {
	var list gwapiv1.GatewayClassList
	if err := k8sClient.List(ctx, &list); err != nil {
		return nil, err
	}

	var resources []ResourceInfo
	for _, item := range list.Items {
		info := ResourceInfo{
			Name: item.Name,
		}

		// Get the latest condition
		if len(item.Status.Conditions) > 0 {
			condition := item.Status.Conditions[len(item.Status.Conditions)-1]
			info.Status = string(condition.Status)
			info.Reason = condition.Reason
			info.Message = condition.Message
		}

		resources = append(resources, info)
	}

	return resources, nil
}

func getGateways(ctx context.Context) ([]ResourceInfo, error) {
	var list gwapiv1.GatewayList
	if err := k8sClient.List(ctx, &list); err != nil {
		return nil, err
	}

	var resources []ResourceInfo
	for _, item := range list.Items {
		info := ResourceInfo{
			Name:      item.Name,
			Namespace: item.Namespace,
		}

		// Get the latest condition
		if len(item.Status.Conditions) > 0 {
			condition := item.Status.Conditions[len(item.Status.Conditions)-1]
			info.Status = string(condition.Status)
			info.Reason = condition.Reason
			info.Message = condition.Message
		}

		resources = append(resources, info)
	}

	return resources, nil
}

func getHTTPRoutes(ctx context.Context) ([]ResourceInfo, error) {
	var list gwapiv1.HTTPRouteList
	if err := k8sClient.List(ctx, &list); err != nil {
		return nil, err
	}

	var resources []ResourceInfo
	for _, item := range list.Items {
		info := ResourceInfo{
			Name:      item.Name,
			Namespace: item.Namespace,
		}

		// Get the latest condition from parents
		if len(item.Status.Parents) > 0 && len(item.Status.Parents[0].Conditions) > 0 {
			condition := item.Status.Parents[0].Conditions[0]
			info.Status = string(condition.Status)
			info.Reason = condition.Reason
			info.Message = condition.Message
		}

		resources = append(resources, info)
	}

	return resources, nil
}

func getGRPCRoutes(ctx context.Context) ([]ResourceInfo, error) {
	var list gwapiv1.GRPCRouteList
	if err := k8sClient.List(ctx, &list); err != nil {
		return nil, err
	}

	var resources []ResourceInfo
	for _, item := range list.Items {
		info := ResourceInfo{
			Name:      item.Name,
			Namespace: item.Namespace,
		}

		// Get the latest condition from parents
		if len(item.Status.Parents) > 0 && len(item.Status.Parents[0].Conditions) > 0 {
			condition := item.Status.Parents[0].Conditions[0]
			info.Status = string(condition.Status)
			info.Reason = condition.Reason
			info.Message = condition.Message
		}

		resources = append(resources, info)
	}

	return resources, nil
}

func getTCPRoutes(ctx context.Context) ([]ResourceInfo, error) {
	var list gwapiv1a2.TCPRouteList
	if err := k8sClient.List(ctx, &list); err != nil {
		return nil, err
	}

	var resources []ResourceInfo
	for _, item := range list.Items {
		info := ResourceInfo{
			Name:      item.Name,
			Namespace: item.Namespace,
		}

		// Get the latest condition from parents
		if len(item.Status.Parents) > 0 && len(item.Status.Parents[0].Conditions) > 0 {
			condition := item.Status.Parents[0].Conditions[0]
			info.Status = string(condition.Status)
			info.Reason = condition.Reason
			info.Message = condition.Message
		}

		resources = append(resources, info)
	}

	return resources, nil
}

func getUDPRoutes(ctx context.Context) ([]ResourceInfo, error) {
	var list gwapiv1a2.UDPRouteList
	if err := k8sClient.List(ctx, &list); err != nil {
		return nil, err
	}

	var resources []ResourceInfo
	for _, item := range list.Items {
		info := ResourceInfo{
			Name:      item.Name,
			Namespace: item.Namespace,
		}

		// Get the latest condition from parents
		if len(item.Status.Parents) > 0 && len(item.Status.Parents[0].Conditions) > 0 {
			condition := item.Status.Parents[0].Conditions[0]
			info.Status = string(condition.Status)
			info.Reason = condition.Reason
			info.Message = condition.Message
		}

		resources = append(resources, info)
	}

	return resources, nil
}

func getTLSRoutes(ctx context.Context) ([]ResourceInfo, error) {
	var list gwapiv1a2.TLSRouteList
	if err := k8sClient.List(ctx, &list); err != nil {
		return nil, err
	}

	var resources []ResourceInfo
	for _, item := range list.Items {
		info := ResourceInfo{
			Name:      item.Name,
			Namespace: item.Namespace,
		}

		// Get the latest condition from parents
		if len(item.Status.Parents) > 0 && len(item.Status.Parents[0].Conditions) > 0 {
			condition := item.Status.Parents[0].Conditions[0]
			info.Status = string(condition.Status)
			info.Reason = condition.Reason
			info.Message = condition.Message
		}

		resources = append(resources, info)
	}

	return resources, nil
}

func getBackendTLSPolicies(ctx context.Context) ([]ResourceInfo, error) {
	var list gwapiv1a3.BackendTLSPolicyList
	if err := k8sClient.List(ctx, &list); err != nil {
		return nil, err
	}

	var resources []ResourceInfo
	for _, item := range list.Items {
		info := ResourceInfo{
			Name:      item.Name,
			Namespace: item.Namespace,
		}

		// Get the latest condition from ancestors
		if len(item.Status.Ancestors) > 0 && len(item.Status.Ancestors[0].Conditions) > 0 {
			condition := item.Status.Ancestors[0].Conditions[len(item.Status.Ancestors[0].Conditions)-1]
			info.Status = string(condition.Status)
			info.Reason = condition.Reason
			info.Message = condition.Message
		}

		resources = append(resources, info)
	}

	return resources, nil
}

func getBackendTrafficPolicies(ctx context.Context) ([]ResourceInfo, error) {
	var list egv1a1.BackendTrafficPolicyList
	if err := k8sClient.List(ctx, &list); err != nil {
		return nil, err
	}

	var resources []ResourceInfo
	for _, item := range list.Items {
		info := ResourceInfo{
			Name:      item.Name,
			Namespace: item.Namespace,
		}

		// Get the latest condition from ancestors
		if len(item.Status.Ancestors) > 0 && len(item.Status.Ancestors[0].Conditions) > 0 {
			condition := item.Status.Ancestors[0].Conditions[len(item.Status.Ancestors[0].Conditions)-1]
			info.Status = string(condition.Status)
			info.Reason = condition.Reason
			info.Message = condition.Message
		}

		resources = append(resources, info)
	}

	return resources, nil
}

func getClientTrafficPolicies(ctx context.Context) ([]ResourceInfo, error) {
	var list egv1a1.ClientTrafficPolicyList
	if err := k8sClient.List(ctx, &list); err != nil {
		return nil, err
	}

	var resources []ResourceInfo
	for _, item := range list.Items {
		info := ResourceInfo{
			Name:      item.Name,
			Namespace: item.Namespace,
		}

		// Get the latest condition from ancestors
		if len(item.Status.Ancestors) > 0 && len(item.Status.Ancestors[0].Conditions) > 0 {
			condition := item.Status.Ancestors[0].Conditions[len(item.Status.Ancestors[0].Conditions)-1]
			info.Status = string(condition.Status)
			info.Reason = condition.Reason
			info.Message = condition.Message
		}

		resources = append(resources, info)
	}

	return resources, nil
}

func getSecurityPolicies(ctx context.Context) ([]ResourceInfo, error) {
	var list egv1a1.SecurityPolicyList
	if err := k8sClient.List(ctx, &list); err != nil {
		return nil, err
	}

	var resources []ResourceInfo
	for _, item := range list.Items {
		info := ResourceInfo{
			Name:      item.Name,
			Namespace: item.Namespace,
		}

		// Get the latest condition from ancestors
		if len(item.Status.Ancestors) > 0 && len(item.Status.Ancestors[0].Conditions) > 0 {
			condition := item.Status.Ancestors[0].Conditions[len(item.Status.Ancestors[0].Conditions)-1]
			info.Status = string(condition.Status)
			info.Reason = condition.Reason
			info.Message = condition.Message
		}

		resources = append(resources, info)
	}

	return resources, nil
}

func getEnvoyPatchPolicies(ctx context.Context) ([]ResourceInfo, error) {
	var list egv1a1.EnvoyPatchPolicyList
	if err := k8sClient.List(ctx, &list); err != nil {
		return nil, err
	}

	var resources []ResourceInfo
	for _, item := range list.Items {
		info := ResourceInfo{
			Name:      item.Name,
			Namespace: item.Namespace,
		}

		// Get the latest condition from ancestors
		if len(item.Status.Ancestors) > 0 && len(item.Status.Ancestors[0].Conditions) > 0 {
			condition := item.Status.Ancestors[0].Conditions[len(item.Status.Ancestors[0].Conditions)-1]
			info.Status = string(condition.Status)
			info.Reason = condition.Reason
			info.Message = condition.Message
		}

		resources = append(resources, info)
	}

	return resources, nil
}

func getEnvoyExtensionPolicies(ctx context.Context) ([]ResourceInfo, error) {
	var list egv1a1.EnvoyExtensionPolicyList
	if err := k8sClient.List(ctx, &list); err != nil {
		return nil, err
	}

	var resources []ResourceInfo
	for _, item := range list.Items {
		info := ResourceInfo{
			Name:      item.Name,
			Namespace: item.Namespace,
		}

		// Get the latest condition from ancestors
		if len(item.Status.Ancestors) > 0 && len(item.Status.Ancestors[0].Conditions) > 0 {
			condition := item.Status.Ancestors[0].Conditions[len(item.Status.Ancestors[0].Conditions)-1]
			info.Status = string(condition.Status)
			info.Reason = condition.Reason
			info.Message = condition.Message
		}

		resources = append(resources, info)
	}

	return resources, nil
}
