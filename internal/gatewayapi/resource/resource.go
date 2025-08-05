// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"fmt"
	"sort"

	certificatesv1b1 "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	mcsapiv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

type (
	XdsIRMap   map[string]*ir.Xds
	InfraIRMap map[string]*ir.Infra
)

// Resources holds the Gateway API and related
// resources that the translators needs as inputs.
// +k8s:deepcopy-gen=true
type Resources struct {
	// This field is only used for marshalling/unmarshalling purposes and is not used by
	// the translator

	// EnvoyProxyForGatewayClass holds EnvoyProxy attached to GatewayClass
	EnvoyProxyForGatewayClass *egv1a1.EnvoyProxy `json:"envoyProxyForGatewayClass,omitempty" yaml:"envoyProxyForGatewayClass,omitempty"`
	// EnvoyProxiesForGateways holds EnvoyProxiesForGateways attached to Gateways
	EnvoyProxiesForGateways []*egv1a1.EnvoyProxy `json:"envoyProxiesForGateways,omitempty" yaml:"envoyProxiesForGateways,omitempty"`

	GatewayClass            *gwapiv1.GatewayClass          `json:"gatewayClass,omitempty" yaml:"gatewayClass,omitempty"`
	Gateways                []*gwapiv1.Gateway             `json:"gateways,omitempty" yaml:"gateways,omitempty"`
	HTTPRoutes              []*gwapiv1.HTTPRoute           `json:"httpRoutes,omitempty" yaml:"httpRoutes,omitempty"`
	GRPCRoutes              []*gwapiv1.GRPCRoute           `json:"grpcRoutes,omitempty" yaml:"grpcRoutes,omitempty"`
	TLSRoutes               []*gwapiv1a2.TLSRoute          `json:"tlsRoutes,omitempty" yaml:"tlsRoutes,omitempty"`
	TCPRoutes               []*gwapiv1a2.TCPRoute          `json:"tcpRoutes,omitempty" yaml:"tcpRoutes,omitempty"`
	UDPRoutes               []*gwapiv1a2.UDPRoute          `json:"udpRoutes,omitempty" yaml:"udpRoutes,omitempty"`
	ReferenceGrants         []*gwapiv1b1.ReferenceGrant    `json:"referenceGrants,omitempty" yaml:"referenceGrants,omitempty"`
	Namespaces              []*corev1.Namespace            `json:"namespaces,omitempty" yaml:"namespaces,omitempty"`
	Services                []*corev1.Service              `json:"services,omitempty" yaml:"services,omitempty"`
	ServiceImports          []*mcsapiv1a1.ServiceImport    `json:"serviceImports,omitempty" yaml:"serviceImports,omitempty"`
	EndpointSlices          []*discoveryv1.EndpointSlice   `json:"endpointSlices,omitempty" yaml:"endpointSlices,omitempty"`
	Secrets                 []*corev1.Secret               `json:"secrets,omitempty" yaml:"secrets,omitempty"`
	ConfigMaps              []*corev1.ConfigMap            `json:"configMaps,omitempty" yaml:"configMaps,omitempty"`
	ExtensionRefFilters     []unstructured.Unstructured    `json:"extensionRefFilters,omitempty" yaml:"extensionRefFilters,omitempty"`
	EnvoyPatchPolicies      []*egv1a1.EnvoyPatchPolicy     `json:"envoyPatchPolicies,omitempty" yaml:"envoyPatchPolicies,omitempty"`
	ClientTrafficPolicies   []*egv1a1.ClientTrafficPolicy  `json:"clientTrafficPolicies,omitempty" yaml:"clientTrafficPolicies,omitempty"`
	BackendTrafficPolicies  []*egv1a1.BackendTrafficPolicy `json:"backendTrafficPolicies,omitempty" yaml:"backendTrafficPolicies,omitempty"`
	SecurityPolicies        []*egv1a1.SecurityPolicy       `json:"securityPolicies,omitempty" yaml:"securityPolicies,omitempty"`
	BackendTLSPolicies      []*gwapiv1a3.BackendTLSPolicy  `json:"backendTLSPolicies,omitempty" yaml:"backendTLSPolicies,omitempty"`
	EnvoyExtensionPolicies  []*egv1a1.EnvoyExtensionPolicy `json:"envoyExtensionPolicies,omitempty" yaml:"envoyExtensionPolicies,omitempty"`
	ExtensionServerPolicies []unstructured.Unstructured    `json:"extensionServerPolicies,omitempty" yaml:"extensionServerPolicies,omitempty"`
	Backends                []*egv1a1.Backend              `json:"backends,omitempty" yaml:"backends,omitempty"`
	HTTPRouteFilters        []*egv1a1.HTTPRouteFilter      `json:"httpFilters,omitempty" yaml:"httpFilters,omitempty"`

	ClusterTrustBundles []*certificatesv1b1.ClusterTrustBundle `json:"clusterTrustBundles,omitempty" yaml:"clusterTrustBundles,omitempty"`
}

func NewResources() *Resources {
	return &Resources{
		Gateways:                []*gwapiv1.Gateway{},
		HTTPRoutes:              []*gwapiv1.HTTPRoute{},
		GRPCRoutes:              []*gwapiv1.GRPCRoute{},
		TLSRoutes:               []*gwapiv1a2.TLSRoute{},
		Services:                []*corev1.Service{},
		EndpointSlices:          []*discoveryv1.EndpointSlice{},
		Secrets:                 []*corev1.Secret{},
		ConfigMaps:              []*corev1.ConfigMap{},
		ReferenceGrants:         []*gwapiv1b1.ReferenceGrant{},
		Namespaces:              []*corev1.Namespace{},
		ExtensionRefFilters:     []unstructured.Unstructured{},
		EnvoyPatchPolicies:      []*egv1a1.EnvoyPatchPolicy{},
		ClientTrafficPolicies:   []*egv1a1.ClientTrafficPolicy{},
		BackendTrafficPolicies:  []*egv1a1.BackendTrafficPolicy{},
		SecurityPolicies:        []*egv1a1.SecurityPolicy{},
		BackendTLSPolicies:      []*gwapiv1a3.BackendTLSPolicy{},
		EnvoyExtensionPolicies:  []*egv1a1.EnvoyExtensionPolicy{},
		ExtensionServerPolicies: []unstructured.Unstructured{},
		Backends:                []*egv1a1.Backend{},
		HTTPRouteFilters:        []*egv1a1.HTTPRouteFilter{},
	}
}

func (r *Resources) GetNamespace(name string) *corev1.Namespace {
	for _, ns := range r.Namespaces {
		if ns.Name == name {
			return ns
		}
	}

	return nil
}

func (r *Resources) GetEnvoyProxy(namespace, name string) *egv1a1.EnvoyProxy {
	for _, ep := range r.EnvoyProxiesForGateways {
		if ep.Namespace == namespace && ep.Name == name {
			return ep
		}
	}

	return nil
}

// GetService returns the Service with the given namespace and name.
func (r *Resources) GetService(namespace, name string) *corev1.Service {
	for _, svc := range r.Services {
		if svc.Namespace == namespace && svc.Name == name {
			return svc
		}
	}
	return nil
}

// GetServiceByLabels returns the Service matching the given labels and namespace target.
func (r *Resources) GetServiceByLabels(labels map[string]string, namespace string) *corev1.Service {
	for _, svc := range r.Services {
		if (namespace != "" && svc.Namespace != namespace) || svc.Labels == nil {
			continue
		}
		match := true
		for k, v := range labels {
			if svc.Labels[k] != v {
				match = false
				break
			}
		}
		if match {
			return svc
		}
	}
	return nil
}

func (r *Resources) GetServiceImport(namespace, name string) *mcsapiv1a1.ServiceImport {
	for _, svcImp := range r.ServiceImports {
		if svcImp.Namespace == namespace && svcImp.Name == name {
			return svcImp
		}
	}

	return nil
}

func (r *Resources) GetBackend(namespace, name string) *egv1a1.Backend {
	for _, be := range r.Backends {
		if be.Namespace == namespace && be.Name == name {
			return be
		}
	}

	return nil
}

func (r *Resources) GetSecret(namespace, name string) *corev1.Secret {
	for _, secret := range r.Secrets {
		if secret.Namespace == namespace && secret.Name == name {
			return secret
		}
	}

	return nil
}

func (r *Resources) GetClusterTrustBundle(name string) *certificatesv1b1.ClusterTrustBundle {
	for _, ctb := range r.ClusterTrustBundles {
		if ctb.Name == name {
			return ctb
		}
	}

	return nil
}

func (r *Resources) GetConfigMap(namespace, name string) *corev1.ConfigMap {
	for _, configMap := range r.ConfigMaps {
		if configMap.Namespace == namespace && configMap.Name == name {
			return configMap
		}
	}

	return nil
}

func (r *Resources) GetEndpointSlicesForBackend(svcNamespace, svcName, backendKind string) []*discoveryv1.EndpointSlice {
	var endpointSlices []*discoveryv1.EndpointSlice
	for _, endpointSlice := range r.EndpointSlices {
		var backendSelectorLabel string
		switch backendKind {
		case KindService:
			backendSelectorLabel = discoveryv1.LabelServiceName
		case KindServiceImport:
			backendSelectorLabel = mcsapiv1a1.LabelServiceName
		}
		if svcNamespace == endpointSlice.Namespace &&
			endpointSlice.GetLabels()[backendSelectorLabel] == svcName {
			endpointSlices = append(endpointSlices, endpointSlice)
		}
	}
	return endpointSlices
}

// ControllerResources holds all the GatewayAPI resources per GatewayClass
type ControllerResources []*Resources

// DeepCopy creates a new ControllerResources.
// It is handwritten since the tooling was unable to copy into a new slice
func (c *ControllerResources) DeepCopy() *ControllerResources {
	if c == nil {
		return nil
	}
	out := make(ControllerResources, len(*c))
	copy(out, *c)
	return &out
}

func (c ControllerResources) Sort() {
	// Top level sort based on gatewayClass contents
	// Sort gatewayClass based on timestamp.
	// Initially, sort by creation timestamp
	// or sort alphabetically by “{namespace}/{name}” if multiple gatewayclasses share same timestamp.
	sort.Slice(c, func(i, j int) bool {
		if c[i].GatewayClass.CreationTimestamp.Equal(&(c[j].GatewayClass.CreationTimestamp)) {
			return c[i].GatewayClass.Name < c[j].GatewayClass.Name
		}
		// Not identical CreationTimestamps
		return c[i].GatewayClass.CreationTimestamp.Before(&(c[j].GatewayClass.CreationTimestamp))
	})

	// Then, run Sort for each item
	for idx := range c {
		// Sort gateways based on timestamp.
		// Initially, gateways sort by creation timestamp
		// or sort alphabetically by “{namespace}/{name}” if multiple gateways share same timestamp.
		sort.Slice(c[idx].Gateways, func(i, j int) bool {
			if c[idx].Gateways[i].CreationTimestamp.Equal(&(c[idx].Gateways[j].CreationTimestamp)) {
				gatewayKeyI := fmt.Sprintf("%s/%s", c[idx].Gateways[i].Namespace, c[idx].Gateways[i].Name)
				gatewayKeyJ := fmt.Sprintf("%s/%s", c[idx].Gateways[j].Namespace, c[idx].Gateways[j].Name)
				return gatewayKeyI < gatewayKeyJ
			}
			// Not identical CreationTimestamps

			return c[idx].Gateways[i].CreationTimestamp.Before(&(c[idx].Gateways[j].CreationTimestamp))
		})

		// Sort HTTPRoutes by creation timestamp, then namespace/name
		sort.Slice(c[idx].HTTPRoutes, func(i, j int) bool {
			if c[idx].HTTPRoutes[i].CreationTimestamp.Equal(&(c[idx].HTTPRoutes[j].CreationTimestamp)) {
				keyI := fmt.Sprintf("%s/%s", c[idx].HTTPRoutes[i].Namespace, c[idx].HTTPRoutes[i].Name)
				keyJ := fmt.Sprintf("%s/%s", c[idx].HTTPRoutes[j].Namespace, c[idx].HTTPRoutes[j].Name)
				return keyI < keyJ
			}
			return c[idx].HTTPRoutes[i].CreationTimestamp.Before(&(c[idx].HTTPRoutes[j].CreationTimestamp))
		})

		// Sort GRPCRoutes by creation timestamp, then namespace/name
		sort.Slice(c[idx].GRPCRoutes, func(i, j int) bool {
			if c[idx].GRPCRoutes[i].CreationTimestamp.Equal(&(c[idx].GRPCRoutes[j].CreationTimestamp)) {
				keyI := fmt.Sprintf("%s/%s", c[idx].GRPCRoutes[i].Namespace, c[idx].GRPCRoutes[i].Name)
				keyJ := fmt.Sprintf("%s/%s", c[idx].GRPCRoutes[j].Namespace, c[idx].GRPCRoutes[j].Name)
				return keyI < keyJ
			}
			return c[idx].GRPCRoutes[i].CreationTimestamp.Before(&(c[idx].GRPCRoutes[j].CreationTimestamp))
		})

		// Sort TLSRoutes by creation timestamp, then namespace/name
		sort.Slice(c[idx].TLSRoutes, func(i, j int) bool {
			if c[idx].TLSRoutes[i].CreationTimestamp.Equal(&(c[idx].TLSRoutes[j].CreationTimestamp)) {
				keyI := fmt.Sprintf("%s/%s", c[idx].TLSRoutes[i].Namespace, c[idx].TLSRoutes[i].Name)
				keyJ := fmt.Sprintf("%s/%s", c[idx].TLSRoutes[j].Namespace, c[idx].TLSRoutes[j].Name)
				return keyI < keyJ
			}
			return c[idx].TLSRoutes[i].CreationTimestamp.Before(&(c[idx].TLSRoutes[j].CreationTimestamp))
		})

		// Sort TCPRoutes by creation timestamp, then namespace/name
		sort.Slice(c[idx].TCPRoutes, func(i, j int) bool {
			if c[idx].TCPRoutes[i].CreationTimestamp.Equal(&(c[idx].TCPRoutes[j].CreationTimestamp)) {
				keyI := fmt.Sprintf("%s/%s", c[idx].TCPRoutes[i].Namespace, c[idx].TCPRoutes[i].Name)
				keyJ := fmt.Sprintf("%s/%s", c[idx].TCPRoutes[j].Namespace, c[idx].TCPRoutes[j].Name)
				return keyI < keyJ
			}
			return c[idx].TCPRoutes[i].CreationTimestamp.Before(&(c[idx].TCPRoutes[j].CreationTimestamp))
		})

		// Sort UDPRoutes by creation timestamp, then namespace/name
		sort.Slice(c[idx].UDPRoutes, func(i, j int) bool {
			if c[idx].UDPRoutes[i].CreationTimestamp.Equal(&(c[idx].UDPRoutes[j].CreationTimestamp)) {
				keyI := fmt.Sprintf("%s/%s", c[idx].UDPRoutes[i].Namespace, c[idx].UDPRoutes[i].Name)
				keyJ := fmt.Sprintf("%s/%s", c[idx].UDPRoutes[j].Namespace, c[idx].UDPRoutes[j].Name)
				return keyI < keyJ
			}
			return c[idx].UDPRoutes[i].CreationTimestamp.Before(&(c[idx].UDPRoutes[j].CreationTimestamp))
		})

		// Sort ReferenceGrants by creation timestamp, then namespace/name
		sort.Slice(c[idx].ReferenceGrants, func(i, j int) bool {
			if c[idx].ReferenceGrants[i].CreationTimestamp.Equal(&(c[idx].ReferenceGrants[j].CreationTimestamp)) {
				keyI := fmt.Sprintf("%s/%s", c[idx].ReferenceGrants[i].Namespace, c[idx].ReferenceGrants[i].Name)
				keyJ := fmt.Sprintf("%s/%s", c[idx].ReferenceGrants[j].Namespace, c[idx].ReferenceGrants[j].Name)
				return keyI < keyJ
			}
			return c[idx].ReferenceGrants[i].CreationTimestamp.Before(&(c[idx].ReferenceGrants[j].CreationTimestamp))
		})

		// Sort Namespaces by creation timestamp, then name
		sort.Slice(c[idx].Namespaces, func(i, j int) bool {
			if c[idx].Namespaces[i].CreationTimestamp.Equal(&(c[idx].Namespaces[j].CreationTimestamp)) {
				return c[idx].Namespaces[i].Name < c[idx].Namespaces[j].Name
			}
			return c[idx].Namespaces[i].CreationTimestamp.Before(&(c[idx].Namespaces[j].CreationTimestamp))
		})

		// Sort Services by creation timestamp, then namespace/name
		sort.Slice(c[idx].Services, func(i, j int) bool {
			if c[idx].Services[i].CreationTimestamp.Equal(&(c[idx].Services[j].CreationTimestamp)) {
				keyI := fmt.Sprintf("%s/%s", c[idx].Services[i].Namespace, c[idx].Services[i].Name)
				keyJ := fmt.Sprintf("%s/%s", c[idx].Services[j].Namespace, c[idx].Services[j].Name)
				return keyI < keyJ
			}
			return c[idx].Services[i].CreationTimestamp.Before(&(c[idx].Services[j].CreationTimestamp))
		})

		// Sort ServiceImports by creation timestamp, then namespace/name
		sort.Slice(c[idx].ServiceImports, func(i, j int) bool {
			if c[idx].ServiceImports[i].CreationTimestamp.Equal(&(c[idx].ServiceImports[j].CreationTimestamp)) {
				keyI := fmt.Sprintf("%s/%s", c[idx].ServiceImports[i].Namespace, c[idx].ServiceImports[i].Name)
				keyJ := fmt.Sprintf("%s/%s", c[idx].ServiceImports[j].Namespace, c[idx].ServiceImports[j].Name)
				return keyI < keyJ
			}
			return c[idx].ServiceImports[i].CreationTimestamp.Before(&(c[idx].ServiceImports[j].CreationTimestamp))
		})

		// Sort EndpointSlices by creation timestamp, then namespace/name
		sort.Slice(c[idx].EndpointSlices, func(i, j int) bool {
			if c[idx].EndpointSlices[i].CreationTimestamp.Equal(&(c[idx].EndpointSlices[j].CreationTimestamp)) {
				keyI := fmt.Sprintf("%s/%s", c[idx].EndpointSlices[i].Namespace, c[idx].EndpointSlices[i].Name)
				keyJ := fmt.Sprintf("%s/%s", c[idx].EndpointSlices[j].Namespace, c[idx].EndpointSlices[j].Name)
				return keyI < keyJ
			}
			return c[idx].EndpointSlices[i].CreationTimestamp.Before(&(c[idx].EndpointSlices[j].CreationTimestamp))
		})

		// Sort Secrets by creation timestamp, then namespace/name
		sort.Slice(c[idx].Secrets, func(i, j int) bool {
			if c[idx].Secrets[i].CreationTimestamp.Equal(&(c[idx].Secrets[j].CreationTimestamp)) {
				keyI := fmt.Sprintf("%s/%s", c[idx].Secrets[i].Namespace, c[idx].Secrets[i].Name)
				keyJ := fmt.Sprintf("%s/%s", c[idx].Secrets[j].Namespace, c[idx].Secrets[j].Name)
				return keyI < keyJ
			}
			return c[idx].Secrets[i].CreationTimestamp.Before(&(c[idx].Secrets[j].CreationTimestamp))
		})

		// Sort ConfigMaps by creation timestamp, then namespace/name
		sort.Slice(c[idx].ConfigMaps, func(i, j int) bool {
			if c[idx].ConfigMaps[i].CreationTimestamp.Equal(&(c[idx].ConfigMaps[j].CreationTimestamp)) {
				keyI := fmt.Sprintf("%s/%s", c[idx].ConfigMaps[i].Namespace, c[idx].ConfigMaps[i].Name)
				keyJ := fmt.Sprintf("%s/%s", c[idx].ConfigMaps[j].Namespace, c[idx].ConfigMaps[j].Name)
				return keyI < keyJ
			}
			return c[idx].ConfigMaps[i].CreationTimestamp.Before(&(c[idx].ConfigMaps[j].CreationTimestamp))
		})

		// Sort EnvoyPatchPolicies by priority first, then creation timestamp, then namespace/name
		sort.Slice(c[idx].EnvoyPatchPolicies, func(i, j int) bool {
			if c[idx].EnvoyPatchPolicies[i].Spec.Priority == c[idx].EnvoyPatchPolicies[j].Spec.Priority {
				if c[idx].EnvoyPatchPolicies[i].CreationTimestamp.Equal(&(c[idx].EnvoyPatchPolicies[j].CreationTimestamp)) {
					keyI := fmt.Sprintf("%s/%s", c[idx].EnvoyPatchPolicies[i].Namespace, c[idx].EnvoyPatchPolicies[i].Name)
					keyJ := fmt.Sprintf("%s/%s", c[idx].EnvoyPatchPolicies[j].Namespace, c[idx].EnvoyPatchPolicies[j].Name)
					return keyI < keyJ
				}
				return c[idx].EnvoyPatchPolicies[i].CreationTimestamp.Before(&(c[idx].EnvoyPatchPolicies[j].CreationTimestamp))
			}
			return c[idx].EnvoyPatchPolicies[i].Spec.Priority < c[idx].EnvoyPatchPolicies[j].Spec.Priority
		})

		// Sort ClientTrafficPolicies by creation timestamp, then namespace/name
		sort.Slice(c[idx].ClientTrafficPolicies, func(i, j int) bool {
			if c[idx].ClientTrafficPolicies[i].CreationTimestamp.Equal(&(c[idx].ClientTrafficPolicies[j].CreationTimestamp)) {
				keyI := fmt.Sprintf("%s/%s", c[idx].ClientTrafficPolicies[i].Namespace, c[idx].ClientTrafficPolicies[i].Name)
				keyJ := fmt.Sprintf("%s/%s", c[idx].ClientTrafficPolicies[j].Namespace, c[idx].ClientTrafficPolicies[j].Name)
				return keyI < keyJ
			}
			return c[idx].ClientTrafficPolicies[i].CreationTimestamp.Before(&(c[idx].ClientTrafficPolicies[j].CreationTimestamp))
		})

		// Sort BackendTrafficPolicies by creation timestamp, then namespace/name
		sort.Slice(c[idx].BackendTrafficPolicies, func(i, j int) bool {
			if c[idx].BackendTrafficPolicies[i].CreationTimestamp.Equal(&(c[idx].BackendTrafficPolicies[j].CreationTimestamp)) {
				keyI := fmt.Sprintf("%s/%s", c[idx].BackendTrafficPolicies[i].Namespace, c[idx].BackendTrafficPolicies[i].Name)
				keyJ := fmt.Sprintf("%s/%s", c[idx].BackendTrafficPolicies[j].Namespace, c[idx].BackendTrafficPolicies[j].Name)
				return keyI < keyJ
			}
			return c[idx].BackendTrafficPolicies[i].CreationTimestamp.Before(&(c[idx].BackendTrafficPolicies[j].CreationTimestamp))
		})

		// Sort SecurityPolicies by creation timestamp, then namespace/name
		sort.Slice(c[idx].SecurityPolicies, func(i, j int) bool {
			if c[idx].SecurityPolicies[i].CreationTimestamp.Equal(&(c[idx].SecurityPolicies[j].CreationTimestamp)) {
				keyI := fmt.Sprintf("%s/%s", c[idx].SecurityPolicies[i].Namespace, c[idx].SecurityPolicies[i].Name)
				keyJ := fmt.Sprintf("%s/%s", c[idx].SecurityPolicies[j].Namespace, c[idx].SecurityPolicies[j].Name)
				return keyI < keyJ
			}
			return c[idx].SecurityPolicies[i].CreationTimestamp.Before(&(c[idx].SecurityPolicies[j].CreationTimestamp))
		})

		// Sort BackendTLSPolicies by creation timestamp, then namespace/name
		sort.Slice(c[idx].BackendTLSPolicies, func(i, j int) bool {
			if c[idx].BackendTLSPolicies[i].CreationTimestamp.Equal(&(c[idx].BackendTLSPolicies[j].CreationTimestamp)) {
				keyI := fmt.Sprintf("%s/%s", c[idx].BackendTLSPolicies[i].Namespace, c[idx].BackendTLSPolicies[i].Name)
				keyJ := fmt.Sprintf("%s/%s", c[idx].BackendTLSPolicies[j].Namespace, c[idx].BackendTLSPolicies[j].Name)
				return keyI < keyJ
			}
			return c[idx].BackendTLSPolicies[i].CreationTimestamp.Before(&(c[idx].BackendTLSPolicies[j].CreationTimestamp))
		})

		// Sort EnvoyExtensionPolicies by creation timestamp, then namespace/name
		sort.Slice(c[idx].EnvoyExtensionPolicies, func(i, j int) bool {
			if c[idx].EnvoyExtensionPolicies[i].CreationTimestamp.Equal(&(c[idx].EnvoyExtensionPolicies[j].CreationTimestamp)) {
				keyI := fmt.Sprintf("%s/%s", c[idx].EnvoyExtensionPolicies[i].Namespace, c[idx].EnvoyExtensionPolicies[i].Name)
				keyJ := fmt.Sprintf("%s/%s", c[idx].EnvoyExtensionPolicies[j].Namespace, c[idx].EnvoyExtensionPolicies[j].Name)
				return keyI < keyJ
			}
			return c[idx].EnvoyExtensionPolicies[i].CreationTimestamp.Before(&(c[idx].EnvoyExtensionPolicies[j].CreationTimestamp))
		})

		// Sort Backends by creation timestamp, then namespace/name
		sort.Slice(c[idx].Backends, func(i, j int) bool {
			if c[idx].Backends[i].CreationTimestamp.Equal(&(c[idx].Backends[j].CreationTimestamp)) {
				keyI := fmt.Sprintf("%s/%s", c[idx].Backends[i].Namespace, c[idx].Backends[i].Name)
				keyJ := fmt.Sprintf("%s/%s", c[idx].Backends[j].Namespace, c[idx].Backends[j].Name)
				return keyI < keyJ
			}
			return c[idx].Backends[i].CreationTimestamp.Before(&(c[idx].Backends[j].CreationTimestamp))
		})

		// Sort HTTPRouteFilters by creation timestamp, then namespace/name
		sort.Slice(c[idx].HTTPRouteFilters, func(i, j int) bool {
			if c[idx].HTTPRouteFilters[i].CreationTimestamp.Equal(&(c[idx].HTTPRouteFilters[j].CreationTimestamp)) {
				keyI := fmt.Sprintf("%s/%s", c[idx].HTTPRouteFilters[i].Namespace, c[idx].HTTPRouteFilters[i].Name)
				keyJ := fmt.Sprintf("%s/%s", c[idx].HTTPRouteFilters[j].Namespace, c[idx].HTTPRouteFilters[j].Name)
				return keyI < keyJ
			}
			return c[idx].HTTPRouteFilters[i].CreationTimestamp.Before(&(c[idx].HTTPRouteFilters[j].CreationTimestamp))
		})

		// Sort ClusterTrustBundles by creation timestamp, then name (cluster-scoped)
		sort.Slice(c[idx].ClusterTrustBundles, func(i, j int) bool {
			if c[idx].ClusterTrustBundles[i].CreationTimestamp.Equal(&(c[idx].ClusterTrustBundles[j].CreationTimestamp)) {
				return c[idx].ClusterTrustBundles[i].Name < c[idx].ClusterTrustBundles[j].Name
			}
			return c[idx].ClusterTrustBundles[i].CreationTimestamp.Before(&(c[idx].ClusterTrustBundles[j].CreationTimestamp))
		})

		// Sort ExtensionRefFilters by creation timestamp, then namespace/name (unstructured resources)
		sort.Slice(c[idx].ExtensionRefFilters, func(i, j int) bool {
			tsI := c[idx].ExtensionRefFilters[i].GetCreationTimestamp()
			tsJ := c[idx].ExtensionRefFilters[j].GetCreationTimestamp()
			if tsI.Equal(&tsJ) {
				keyI := fmt.Sprintf("%s/%s", c[idx].ExtensionRefFilters[i].GetNamespace(), c[idx].ExtensionRefFilters[i].GetName())
				keyJ := fmt.Sprintf("%s/%s", c[idx].ExtensionRefFilters[j].GetNamespace(), c[idx].ExtensionRefFilters[j].GetName())
				return keyI < keyJ
			}
			return tsI.Before(&tsJ)
		})

		// Sort ExtensionServerPolicies by creation timestamp, then namespace/name (unstructured resources)
		sort.Slice(c[idx].ExtensionServerPolicies, func(i, j int) bool {
			tsI := c[idx].ExtensionServerPolicies[i].GetCreationTimestamp()
			tsJ := c[idx].ExtensionServerPolicies[j].GetCreationTimestamp()
			if tsI.Equal(&tsJ) {
				keyI := fmt.Sprintf("%s/%s", c[idx].ExtensionServerPolicies[i].GetNamespace(), c[idx].ExtensionServerPolicies[i].GetName())
				keyJ := fmt.Sprintf("%s/%s", c[idx].ExtensionServerPolicies[j].GetNamespace(), c[idx].ExtensionServerPolicies[j].GetName())
				return keyI < keyJ
			}
			return tsI.Before(&tsJ)
		})
	}
}
