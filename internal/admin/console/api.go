// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package console

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/cmd/version"
)

// SystemInfo represents basic system information
type SystemInfo struct {
	Version   string    `json:"version"`
	StartTime time.Time `json:"startTime"`
	Uptime    string    `json:"uptime"`
	GoVersion string    `json:"goVersion"`
	Platform  string    `json:"platform"`
	Timestamp time.Time `json:"timestamp"`
}

// ServerInfo represents server status information
type ServerInfo struct {
	State              string            `json:"state"`
	Version            string            `json:"version"`
	Uptime             string            `json:"uptime"`
	Components         []ComponentStatus `json:"components"`
	EnvoyGatewayConfig interface{}       `json:"envoyGatewayConfig"`
	LastUpdated        time.Time         `json:"lastUpdated"`
}

// ComponentStatus represents the status of a system component
type ComponentStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ConfigDumpInfo represents configuration dump information
type ConfigDumpInfo struct {
	// Gateway API Core Resources
	Gateways     []ResourceSummary `json:"gateways"`
	HTTPRoutes   []ResourceSummary `json:"httpRoutes"`
	GRPCRoutes   []ResourceSummary `json:"grpcRoutes"`
	TLSRoutes    []ResourceSummary `json:"tlsRoutes"`
	TCPRoutes    []ResourceSummary `json:"tcpRoutes"`
	UDPRoutes    []ResourceSummary `json:"udpRoutes"`
	GatewayClass []ResourceSummary `json:"gatewayClass"`

	// Envoy Gateway Policies
	ClientTrafficPolicies  []ResourceSummary `json:"clientTrafficPolicies"`
	BackendTrafficPolicies []ResourceSummary `json:"backendTrafficPolicies"`
	BackendTLSPolicies     []ResourceSummary `json:"backendTLSPolicies"`
	SecurityPolicies       []ResourceSummary `json:"securityPolicies"`
	EnvoyPatchPolicies     []ResourceSummary `json:"envoyPatchPolicies"`
	EnvoyExtensionPolicies []ResourceSummary `json:"envoyExtensionPolicies"`

	// Kubernetes Resources
	Services       []ResourceSummary `json:"services"`
	Secrets        []ResourceSummary `json:"secrets"`
	ConfigMaps     []ResourceSummary `json:"configMaps"`
	Namespaces     []ResourceSummary `json:"namespaces"`
	EndpointSlices []ResourceSummary `json:"endpointSlices"`

	// Other Resources
	ReferenceGrants  []ResourceSummary `json:"referenceGrants"`
	HTTPRouteFilters []ResourceSummary `json:"httpRouteFilters"`
	EnvoyProxies     []ResourceSummary `json:"envoyProxies"`
	Backends         []ResourceSummary `json:"backends"`
	ServiceImports   []ResourceSummary `json:"serviceImports"`

	LastUpdated time.Time `json:"lastUpdated"`
}

// ResourceSummary represents a simplified summary of a Kubernetes resource
type ResourceSummary struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

var startTime = time.Now()

// handleAPIInfo returns basic system information
func (h *Handler) handleAPIInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uptime := time.Since(startTime)
	versionInfo := version.Get()
	info := SystemInfo{
		Version:   versionInfo.EnvoyGatewayVersion,
		StartTime: startTime,
		Uptime:    uptime.String(),
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleAPIServerInfo returns server status information
func (h *Handler) handleAPIServerInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uptime := time.Since(startTime)
	versionInfo := version.Get()

	// Since Admin Server starts last, if we can respond, all components are healthy
	// This reflects the actual startup order in server.go
	components := []ComponentStatus{
		{
			Name:    "Provider Service",
			Status:  "Running",
			Message: "Fetches resources from configured provider and publishes them",
		},
		{
			Name:    "GatewayAPI Translator",
			Status:  "Running",
			Message: "Translates provider resources to xDS IR and infra IR",
		},
		{
			Name:    "XDS Translator",
			Status:  "Running",
			Message: "Translates xDS IR into xDS resources and computes policy statuses",
		},
		{
			Name:    "Infrastructure Manager",
			Status:  "Running",
			Message: "Translates infra IR into Envoy Proxy infrastructure resources",
		},
	}

	info := ServerInfo{
		State:              "Running",
		Version:            versionInfo.EnvoyGatewayVersion,
		Uptime:             uptime.String(),
		Components:         components,
		EnvoyGatewayConfig: h.cfg.EnvoyGateway,
		LastUpdated:        time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleAPIConfigDump returns configuration dump information
func (h *Handler) handleAPIConfigDump(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if resource=all parameter is provided
	resourceParam := r.URL.Query().Get("resource")
	if resourceParam == "all" {
		// Return complete provider resources dump
		h.handleCompleteConfigDump(w, r)
		return
	}

	// Load provider resources from watchable (summary view)
	configDump := h.loadConfigDump()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(configDump); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleCompleteConfigDump handles requests for complete provider resources dump
func (h *Handler) handleCompleteConfigDump(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if h.providerResources == nil {
		// Return empty structure if no provider resources
		emptyResponse := map[string]interface{}{
			"message":   "No provider resources available",
			"timestamp": time.Now(),
		}
		if err := json.NewEncoder(w).Encode(emptyResponse); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
		return
	}

	// Get the actual resources using GetResources() method
	resources := h.providerResources.GetResources()

	// Create a structured response with the actual resource data
	response := map[string]interface{}{
		"resources":  resources,
		"timestamp":  time.Now(),
		"totalCount": len(resources),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode provider resources", http.StatusInternalServerError)
		return
	}
}

// loadConfigDump loads configuration data from provider resources
func (h *Handler) loadConfigDump() ConfigDumpInfo {
	configDump := ConfigDumpInfo{
		// Gateway API Core Resources
		Gateways:     []ResourceSummary{},
		HTTPRoutes:   []ResourceSummary{},
		GRPCRoutes:   []ResourceSummary{},
		TLSRoutes:    []ResourceSummary{},
		TCPRoutes:    []ResourceSummary{},
		UDPRoutes:    []ResourceSummary{},
		GatewayClass: []ResourceSummary{},

		// Envoy Gateway Policies
		ClientTrafficPolicies:  []ResourceSummary{},
		BackendTrafficPolicies: []ResourceSummary{},
		BackendTLSPolicies:     []ResourceSummary{},
		SecurityPolicies:       []ResourceSummary{},
		EnvoyPatchPolicies:     []ResourceSummary{},
		EnvoyExtensionPolicies: []ResourceSummary{},

		// Kubernetes Resources
		Services:       []ResourceSummary{},
		Secrets:        []ResourceSummary{},
		ConfigMaps:     []ResourceSummary{},
		Namespaces:     []ResourceSummary{},
		EndpointSlices: []ResourceSummary{},

		// Other Resources
		ReferenceGrants:  []ResourceSummary{},
		HTTPRouteFilters: []ResourceSummary{},
		EnvoyProxies:     []ResourceSummary{},
		Backends:         []ResourceSummary{},
		ServiceImports:   []ResourceSummary{},

		LastUpdated: time.Now(),
	}

	if h.providerResources != nil {
		// Load controller resources directly from the provider resources
		controllerResources := h.providerResources.GatewayAPIResources.LoadAll()

		for _, resources := range controllerResources {
			if resources == nil {
				continue
			}

			for _, res := range *resources {
				if res == nil {
					continue
				}

				// Process GatewayClass
				if res.GatewayClass != nil {
					configDump.GatewayClass = append(configDump.GatewayClass, ResourceSummary{
						Name:      res.GatewayClass.Name,
						Namespace: "", // GatewayClass is cluster-scoped
					})
				}

				// Process Gateways
				for _, gateway := range res.Gateways {
					if gateway != nil {
						configDump.Gateways = append(configDump.Gateways, ResourceSummary{
							Name:      gateway.Name,
							Namespace: gateway.Namespace,
						})
					}
				}

				// Process HTTPRoutes
				for _, httpRoute := range res.HTTPRoutes {
					if httpRoute != nil {
						configDump.HTTPRoutes = append(configDump.HTTPRoutes, ResourceSummary{
							Name:      httpRoute.Name,
							Namespace: httpRoute.Namespace,
						})
					}
				}

				// Process GRPCRoutes
				for _, grpcRoute := range res.GRPCRoutes {
					if grpcRoute != nil {
						configDump.GRPCRoutes = append(configDump.GRPCRoutes, ResourceSummary{
							Name:      grpcRoute.Name,
							Namespace: grpcRoute.Namespace,
						})
					}
				}

				// Process TLSRoutes
				for _, tlsRoute := range res.TLSRoutes {
					if tlsRoute != nil {
						configDump.TLSRoutes = append(configDump.TLSRoutes, ResourceSummary{
							Name:      tlsRoute.Name,
							Namespace: tlsRoute.Namespace,
						})
					}
				}

				// Process TCPRoutes
				for _, tcpRoute := range res.TCPRoutes {
					if tcpRoute != nil {
						configDump.TCPRoutes = append(configDump.TCPRoutes, ResourceSummary{
							Name:      tcpRoute.Name,
							Namespace: tcpRoute.Namespace,
						})
					}
				}

				// Process UDPRoutes
				for _, udpRoute := range res.UDPRoutes {
					if udpRoute != nil {
						configDump.UDPRoutes = append(configDump.UDPRoutes, ResourceSummary{
							Name:      udpRoute.Name,
							Namespace: udpRoute.Namespace,
						})
					}
				}

				// Process Envoy Gateway Policies
				for _, policy := range res.ClientTrafficPolicies {
					if policy != nil {
						configDump.ClientTrafficPolicies = append(configDump.ClientTrafficPolicies, ResourceSummary{
							Name:      policy.Name,
							Namespace: policy.Namespace,
						})
					}
				}

				for _, policy := range res.BackendTrafficPolicies {
					if policy != nil {
						configDump.BackendTrafficPolicies = append(configDump.BackendTrafficPolicies, ResourceSummary{
							Name:      policy.Name,
							Namespace: policy.Namespace,
						})
					}
				}

				for _, policy := range res.BackendTLSPolicies {
					if policy != nil {
						configDump.BackendTLSPolicies = append(configDump.BackendTLSPolicies, ResourceSummary{
							Name:      policy.Name,
							Namespace: policy.Namespace,
						})
					}
				}

				for _, policy := range res.SecurityPolicies {
					if policy != nil {
						configDump.SecurityPolicies = append(configDump.SecurityPolicies, ResourceSummary{
							Name:      policy.Name,
							Namespace: policy.Namespace,
						})
					}
				}

				for _, policy := range res.EnvoyPatchPolicies {
					if policy != nil {
						configDump.EnvoyPatchPolicies = append(configDump.EnvoyPatchPolicies, ResourceSummary{
							Name:      policy.Name,
							Namespace: policy.Namespace,
						})
					}
				}

				for _, policy := range res.EnvoyExtensionPolicies {
					if policy != nil {
						configDump.EnvoyExtensionPolicies = append(configDump.EnvoyExtensionPolicies, ResourceSummary{
							Name:      policy.Name,
							Namespace: policy.Namespace,
						})
					}
				}

				// Process Kubernetes Resources
				for _, service := range res.Services {
					if service != nil {
						configDump.Services = append(configDump.Services, ResourceSummary{
							Name:      service.Name,
							Namespace: service.Namespace,
						})
					}
				}

				for _, secret := range res.Secrets {
					if secret != nil {
						configDump.Secrets = append(configDump.Secrets, ResourceSummary{
							Name:      secret.Name,
							Namespace: secret.Namespace,
						})
					}
				}

				for _, configMap := range res.ConfigMaps {
					if configMap != nil {
						configDump.ConfigMaps = append(configDump.ConfigMaps, ResourceSummary{
							Name:      configMap.Name,
							Namespace: configMap.Namespace,
						})
					}
				}

				for _, namespace := range res.Namespaces {
					if namespace != nil {
						configDump.Namespaces = append(configDump.Namespaces, ResourceSummary{
							Name:      namespace.Name,
							Namespace: "", // Namespaces are cluster-scoped
						})
					}
				}

				for _, endpointSlice := range res.EndpointSlices {
					if endpointSlice != nil {
						configDump.EndpointSlices = append(configDump.EndpointSlices, ResourceSummary{
							Name:      endpointSlice.Name,
							Namespace: endpointSlice.Namespace,
						})
					}
				}

				// Process Other Resources
				for _, refGrant := range res.ReferenceGrants {
					if refGrant != nil {
						configDump.ReferenceGrants = append(configDump.ReferenceGrants, ResourceSummary{
							Name:      refGrant.Name,
							Namespace: refGrant.Namespace,
						})
					}
				}

				for _, filter := range res.HTTPRouteFilters {
					if filter != nil {
						configDump.HTTPRouteFilters = append(configDump.HTTPRouteFilters, ResourceSummary{
							Name:      filter.Name,
							Namespace: filter.Namespace,
						})
					}
				}

				// Process EnvoyProxy for GatewayClass
				if res.EnvoyProxyForGatewayClass != nil {
					configDump.EnvoyProxies = append(configDump.EnvoyProxies, ResourceSummary{
						Name:      res.EnvoyProxyForGatewayClass.Name,
						Namespace: res.EnvoyProxyForGatewayClass.Namespace,
					})
				}

				// Process EnvoyProxies for Gateways
				for _, proxy := range res.EnvoyProxiesForGateways {
					if proxy != nil {
						configDump.EnvoyProxies = append(configDump.EnvoyProxies, ResourceSummary{
							Name:      proxy.Name,
							Namespace: proxy.Namespace,
						})
					}
				}

				for _, backend := range res.Backends {
					if backend != nil {
						configDump.Backends = append(configDump.Backends, ResourceSummary{
							Name:      backend.Name,
							Namespace: backend.Namespace,
						})
					}
				}

				for _, serviceImport := range res.ServiceImports {
					if serviceImport != nil {
						configDump.ServiceImports = append(configDump.ServiceImports, ResourceSummary{
							Name:      serviceImport.Name,
							Namespace: serviceImport.Namespace,
						})
					}
				}
			}
		}
	}

	return configDump
}

// handleAPIMetrics handles requests for metrics using the Prometheus registry
func (h *Handler) handleAPIMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create a Prometheus handler using the controller-runtime registry
	// This includes all Envoy Gateway control plane metrics
	handler := createCombinedMetricsHandler()

	// Serve the metrics using the Prometheus handler
	handler.ServeHTTP(w, r)
}

// Helper functions to extract status information from Gateway API resources

func getGatewayClassStatus(gc *gwapiv1.GatewayClass) string {
	if len(gc.Status.Conditions) == 0 {
		return "Unknown"
	}

	for _, condition := range gc.Status.Conditions {
		if condition.Type == string(gwapiv1.GatewayClassConditionStatusAccepted) {
			if condition.Status == metav1.ConditionTrue {
				return "Accepted"
			}
			return "Not Accepted"
		}
	}
	return "Unknown"
}

func getGatewayStatus(gw *gwapiv1.Gateway) string {
	if len(gw.Status.Conditions) == 0 {
		return "Unknown"
	}

	// Check for Programmed condition first
	for _, condition := range gw.Status.Conditions {
		if condition.Type == string(gwapiv1.GatewayConditionProgrammed) {
			if condition.Status == metav1.ConditionTrue {
				return "Programmed"
			}
			return "Not Programmed"
		}
	}

	// Check for Accepted condition
	for _, condition := range gw.Status.Conditions {
		if condition.Type == string(gwapiv1.GatewayConditionAccepted) {
			if condition.Status == metav1.ConditionTrue {
				return "Accepted"
			}
			return "Not Accepted"
		}
	}

	return "Unknown"
}
