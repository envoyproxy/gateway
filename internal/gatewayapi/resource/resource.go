// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
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
	labelsutil "github.com/envoyproxy/gateway/internal/utils/labels"
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
	TLSRoutes               []*gwapiv1a3.TLSRoute          `json:"tlsRoutes,omitempty" yaml:"tlsRoutes,omitempty"`
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
		TLSRoutes:               []*gwapiv1a3.TLSRoute{},
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
		match, _ := labelsutil.Matches(labels, svc.Labels)
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
	for i, res := range *c {
		if res != nil {
			out[i] = res.DeepCopy()
		}
	}
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
		c[idx].Sort()
	}
}

func (r *Resources) Sort() {
	// Sort gateways based on timestamp.
	// Initially, gateways sort by creation timestamp
	// or sort alphabetically by “{namespace}/{name}” if multiple gateways share same timestamp.
	sort.Slice(r.Gateways, func(i, j int) bool {
		if r.Gateways[i].CreationTimestamp.Equal(&(r.Gateways[j].CreationTimestamp)) {
			if r.Gateways[i].Namespace != r.Gateways[j].Namespace {
				return r.Gateways[i].Namespace < r.Gateways[j].Namespace
			}
			return r.Gateways[i].Name < r.Gateways[j].Name
		}
		// Not identical CreationTimestamps

		return r.Gateways[i].CreationTimestamp.Before(&(r.Gateways[j].CreationTimestamp))
	})

	// Sort HTTPRoutes by creation timestamp, then namespace/name
	sort.Slice(r.HTTPRoutes, func(i, j int) bool {
		if r.HTTPRoutes[i].CreationTimestamp.Equal(&(r.HTTPRoutes[j].CreationTimestamp)) {
			if r.HTTPRoutes[i].Namespace != r.HTTPRoutes[j].Namespace {
				return r.HTTPRoutes[i].Namespace < r.HTTPRoutes[j].Namespace
			}
			return r.HTTPRoutes[i].Name < r.HTTPRoutes[j].Name
		}
		return r.HTTPRoutes[i].CreationTimestamp.Before(&(r.HTTPRoutes[j].CreationTimestamp))
	})

	// Sort GRPCRoutes by creation timestamp, then namespace/name
	sort.Slice(r.GRPCRoutes, func(i, j int) bool {
		if r.GRPCRoutes[i].CreationTimestamp.Equal(&(r.GRPCRoutes[j].CreationTimestamp)) {
			if r.GRPCRoutes[i].Namespace != r.GRPCRoutes[j].Namespace {
				return r.GRPCRoutes[i].Namespace < r.GRPCRoutes[j].Namespace
			}
			return r.GRPCRoutes[i].Name < r.GRPCRoutes[j].Name
		}
		return r.GRPCRoutes[i].CreationTimestamp.Before(&(r.GRPCRoutes[j].CreationTimestamp))
	})

	// Sort TLSRoutes by creation timestamp, then namespace/name
	sort.Slice(r.TLSRoutes, func(i, j int) bool {
		if r.TLSRoutes[i].CreationTimestamp.Equal(&(r.TLSRoutes[j].CreationTimestamp)) {
			if r.TLSRoutes[i].Namespace != r.TLSRoutes[j].Namespace {
				return r.TLSRoutes[i].Namespace < r.TLSRoutes[j].Namespace
			}
			return r.TLSRoutes[i].Name < r.TLSRoutes[j].Name
		}
		return r.TLSRoutes[i].CreationTimestamp.Before(&(r.TLSRoutes[j].CreationTimestamp))
	})

	// Sort TCPRoutes by creation timestamp, then namespace/name
	sort.Slice(r.TCPRoutes, func(i, j int) bool {
		if r.TCPRoutes[i].CreationTimestamp.Equal(&(r.TCPRoutes[j].CreationTimestamp)) {
			if r.TCPRoutes[i].Namespace != r.TCPRoutes[j].Namespace {
				return r.TCPRoutes[i].Namespace < r.TCPRoutes[j].Namespace
			}
			return r.TCPRoutes[i].Name < r.TCPRoutes[j].Name
		}
		return r.TCPRoutes[i].CreationTimestamp.Before(&(r.TCPRoutes[j].CreationTimestamp))
	})

	// Sort UDPRoutes by creation timestamp, then namespace/name
	sort.Slice(r.UDPRoutes, func(i, j int) bool {
		if r.UDPRoutes[i].CreationTimestamp.Equal(&(r.UDPRoutes[j].CreationTimestamp)) {
			if r.UDPRoutes[i].Namespace != r.UDPRoutes[j].Namespace {
				return r.UDPRoutes[i].Namespace < r.UDPRoutes[j].Namespace
			}
			return r.UDPRoutes[i].Name < r.UDPRoutes[j].Name
		}
		return r.UDPRoutes[i].CreationTimestamp.Before(&(r.UDPRoutes[j].CreationTimestamp))
	})

	// Sort ReferenceGrants by creation timestamp, then namespace/name
	sort.Slice(r.ReferenceGrants, func(i, j int) bool {
		if r.ReferenceGrants[i].CreationTimestamp.Equal(&(r.ReferenceGrants[j].CreationTimestamp)) {
			if r.ReferenceGrants[i].Namespace != r.ReferenceGrants[j].Namespace {
				return r.ReferenceGrants[i].Namespace < r.ReferenceGrants[j].Namespace
			}
			return r.ReferenceGrants[i].Name < r.ReferenceGrants[j].Name
		}
		return r.ReferenceGrants[i].CreationTimestamp.Before(&(r.ReferenceGrants[j].CreationTimestamp))
	})

	// Sort Namespaces by creation timestamp, then name
	sort.Slice(r.Namespaces, func(i, j int) bool {
		if r.Namespaces[i].CreationTimestamp.Equal(&(r.Namespaces[j].CreationTimestamp)) {
			return r.Namespaces[i].Name < r.Namespaces[j].Name
		}
		return r.Namespaces[i].CreationTimestamp.Before(&(r.Namespaces[j].CreationTimestamp))
	})

	// Sort Services by creation timestamp, then namespace/name
	sort.Slice(r.Services, func(i, j int) bool {
		if r.Services[i].CreationTimestamp.Equal(&(r.Services[j].CreationTimestamp)) {
			if r.Services[i].Namespace != r.Services[j].Namespace {
				return r.Services[i].Namespace < r.Services[j].Namespace
			}
			return r.Services[i].Name < r.Services[j].Name
		}
		return r.Services[i].CreationTimestamp.Before(&(r.Services[j].CreationTimestamp))
	})

	// Sort ServiceImports by creation timestamp, then namespace/name
	sort.Slice(r.ServiceImports, func(i, j int) bool {
		if r.ServiceImports[i].CreationTimestamp.Equal(&(r.ServiceImports[j].CreationTimestamp)) {
			if r.ServiceImports[i].Namespace != r.ServiceImports[j].Namespace {
				return r.ServiceImports[i].Namespace < r.ServiceImports[j].Namespace
			}
			return r.ServiceImports[i].Name < r.ServiceImports[j].Name
		}
		return r.ServiceImports[i].CreationTimestamp.Before(&(r.ServiceImports[j].CreationTimestamp))
	})

	// Sort EndpointSlices by creation timestamp, then namespace/name
	sort.Slice(r.EndpointSlices, func(i, j int) bool {
		if r.EndpointSlices[i].CreationTimestamp.Equal(&(r.EndpointSlices[j].CreationTimestamp)) {
			if r.EndpointSlices[i].Namespace != r.EndpointSlices[j].Namespace {
				return r.EndpointSlices[i].Namespace < r.EndpointSlices[j].Namespace
			}
			return r.EndpointSlices[i].Name < r.EndpointSlices[j].Name
		}
		return r.EndpointSlices[i].CreationTimestamp.Before(&(r.EndpointSlices[j].CreationTimestamp))
	})

	// Sort Secrets by creation timestamp, then namespace/name
	sort.Slice(r.Secrets, func(i, j int) bool {
		if r.Secrets[i].CreationTimestamp.Equal(&(r.Secrets[j].CreationTimestamp)) {
			if r.Secrets[i].Namespace != r.Secrets[j].Namespace {
				return r.Secrets[i].Namespace < r.Secrets[j].Namespace
			}
			return r.Secrets[i].Name < r.Secrets[j].Name
		}
		return r.Secrets[i].CreationTimestamp.Before(&(r.Secrets[j].CreationTimestamp))
	})

	// Sort ConfigMaps by creation timestamp, then namespace/name
	sort.Slice(r.ConfigMaps, func(i, j int) bool {
		if r.ConfigMaps[i].CreationTimestamp.Equal(&(r.ConfigMaps[j].CreationTimestamp)) {
			if r.ConfigMaps[i].Namespace != r.ConfigMaps[j].Namespace {
				return r.ConfigMaps[i].Namespace < r.ConfigMaps[j].Namespace
			}
			return r.ConfigMaps[i].Name < r.ConfigMaps[j].Name
		}
		return r.ConfigMaps[i].CreationTimestamp.Before(&(r.ConfigMaps[j].CreationTimestamp))
	})

	// Sort EnvoyPatchPolicies by priority first, then creation timestamp, then namespace/name
	sort.Slice(r.EnvoyPatchPolicies, func(i, j int) bool {
		if r.EnvoyPatchPolicies[i].Spec.Priority == r.EnvoyPatchPolicies[j].Spec.Priority {
			if r.EnvoyPatchPolicies[i].CreationTimestamp.Equal(&(r.EnvoyPatchPolicies[j].CreationTimestamp)) {
				if r.EnvoyPatchPolicies[i].Namespace != r.EnvoyPatchPolicies[j].Namespace {
					return r.EnvoyPatchPolicies[i].Namespace < r.EnvoyPatchPolicies[j].Namespace
				}
				return r.EnvoyPatchPolicies[i].Name < r.EnvoyPatchPolicies[j].Name
			}
			return r.EnvoyPatchPolicies[i].CreationTimestamp.Before(&(r.EnvoyPatchPolicies[j].CreationTimestamp))
		}
		return r.EnvoyPatchPolicies[i].Spec.Priority < r.EnvoyPatchPolicies[j].Spec.Priority
	})

	// Sort ClientTrafficPolicies by creation timestamp, then namespace/name
	sort.Slice(r.ClientTrafficPolicies, func(i, j int) bool {
		if r.ClientTrafficPolicies[i].CreationTimestamp.Equal(&(r.ClientTrafficPolicies[j].CreationTimestamp)) {
			if r.ClientTrafficPolicies[i].Namespace != r.ClientTrafficPolicies[j].Namespace {
				return r.ClientTrafficPolicies[i].Namespace < r.ClientTrafficPolicies[j].Namespace
			}
			return r.ClientTrafficPolicies[i].Name < r.ClientTrafficPolicies[j].Name
		}
		return r.ClientTrafficPolicies[i].CreationTimestamp.Before(&(r.ClientTrafficPolicies[j].CreationTimestamp))
	})

	// Sort BackendTrafficPolicies by creation timestamp, then namespace/name
	sort.Slice(r.BackendTrafficPolicies, func(i, j int) bool {
		if r.BackendTrafficPolicies[i].CreationTimestamp.Equal(&(r.BackendTrafficPolicies[j].CreationTimestamp)) {
			if r.BackendTrafficPolicies[i].Namespace != r.BackendTrafficPolicies[j].Namespace {
				return r.BackendTrafficPolicies[i].Namespace < r.BackendTrafficPolicies[j].Namespace
			}
			return r.BackendTrafficPolicies[i].Name < r.BackendTrafficPolicies[j].Name
		}
		return r.BackendTrafficPolicies[i].CreationTimestamp.Before(&(r.BackendTrafficPolicies[j].CreationTimestamp))
	})

	// Sort SecurityPolicies by creation timestamp, then namespace/name
	sort.Slice(r.SecurityPolicies, func(i, j int) bool {
		if r.SecurityPolicies[i].CreationTimestamp.Equal(&(r.SecurityPolicies[j].CreationTimestamp)) {
			if r.SecurityPolicies[i].Namespace != r.SecurityPolicies[j].Namespace {
				return r.SecurityPolicies[i].Namespace < r.SecurityPolicies[j].Namespace
			}
			return r.SecurityPolicies[i].Name < r.SecurityPolicies[j].Name
		}
		return r.SecurityPolicies[i].CreationTimestamp.Before(&(r.SecurityPolicies[j].CreationTimestamp))
	})

	// Sort BackendTLSPolicies by creation timestamp, then namespace/name
	sort.Slice(r.BackendTLSPolicies, func(i, j int) bool {
		if r.BackendTLSPolicies[i].CreationTimestamp.Equal(&(r.BackendTLSPolicies[j].CreationTimestamp)) {
			if r.BackendTLSPolicies[i].Namespace != r.BackendTLSPolicies[j].Namespace {
				return r.BackendTLSPolicies[i].Namespace < r.BackendTLSPolicies[j].Namespace
			}
			return r.BackendTLSPolicies[i].Name < r.BackendTLSPolicies[j].Name
		}
		return r.BackendTLSPolicies[i].CreationTimestamp.Before(&(r.BackendTLSPolicies[j].CreationTimestamp))
	})

	// Sort EnvoyExtensionPolicies by creation timestamp, then namespace/name
	sort.Slice(r.EnvoyExtensionPolicies, func(i, j int) bool {
		if r.EnvoyExtensionPolicies[i].CreationTimestamp.Equal(&(r.EnvoyExtensionPolicies[j].CreationTimestamp)) {
			if r.EnvoyExtensionPolicies[i].Namespace != r.EnvoyExtensionPolicies[j].Namespace {
				return r.EnvoyExtensionPolicies[i].Namespace < r.EnvoyExtensionPolicies[j].Namespace
			}
			return r.EnvoyExtensionPolicies[i].Name < r.EnvoyExtensionPolicies[j].Name
		}
		return r.EnvoyExtensionPolicies[i].CreationTimestamp.Before(&(r.EnvoyExtensionPolicies[j].CreationTimestamp))
	})

	// Sort Backends by creation timestamp, then namespace/name
	sort.Slice(r.Backends, func(i, j int) bool {
		if r.Backends[i].CreationTimestamp.Equal(&(r.Backends[j].CreationTimestamp)) {
			if r.Backends[i].Namespace != r.Backends[j].Namespace {
				return r.Backends[i].Namespace < r.Backends[j].Namespace
			}
			return r.Backends[i].Name < r.Backends[j].Name
		}
		return r.Backends[i].CreationTimestamp.Before(&(r.Backends[j].CreationTimestamp))
	})

	// Sort HTTPRouteFilters by creation timestamp, then namespace/name
	sort.Slice(r.HTTPRouteFilters, func(i, j int) bool {
		if r.HTTPRouteFilters[i].CreationTimestamp.Equal(&(r.HTTPRouteFilters[j].CreationTimestamp)) {
			if r.HTTPRouteFilters[i].Namespace != r.HTTPRouteFilters[j].Namespace {
				return r.HTTPRouteFilters[i].Namespace < r.HTTPRouteFilters[j].Namespace
			}
			return r.HTTPRouteFilters[i].Name < r.HTTPRouteFilters[j].Name
		}
		return r.HTTPRouteFilters[i].CreationTimestamp.Before(&(r.HTTPRouteFilters[j].CreationTimestamp))
	})

	// Sort ClusterTrustBundles by creation timestamp, then name (cluster-scoped)
	sort.Slice(r.ClusterTrustBundles, func(i, j int) bool {
		if r.ClusterTrustBundles[i].CreationTimestamp.Equal(&(r.ClusterTrustBundles[j].CreationTimestamp)) {
			return r.ClusterTrustBundles[i].Name < r.ClusterTrustBundles[j].Name
		}
		return r.ClusterTrustBundles[i].CreationTimestamp.Before(&(r.ClusterTrustBundles[j].CreationTimestamp))
	})

	// Sort ExtensionRefFilters by creation timestamp, then namespace/name (unstructured resources)
	sort.Slice(r.ExtensionRefFilters, func(i, j int) bool {
		tsI := r.ExtensionRefFilters[i].GetCreationTimestamp()
		tsJ := r.ExtensionRefFilters[j].GetCreationTimestamp()
		if tsI.Equal(&tsJ) {
			if r.ExtensionRefFilters[i].GetNamespace() != r.ExtensionRefFilters[j].GetNamespace() {
				return r.ExtensionRefFilters[i].GetNamespace() < r.ExtensionRefFilters[j].GetNamespace()
			}
			return r.ExtensionRefFilters[i].GetName() < r.ExtensionRefFilters[j].GetName()
		}
		return tsI.Before(&tsJ)
	})

	// Sort ExtensionServerPolicies by creation timestamp, then namespace/name (unstructured resources)
	sort.Slice(r.ExtensionServerPolicies, func(i, j int) bool {
		tsI := r.ExtensionServerPolicies[i].GetCreationTimestamp()
		tsJ := r.ExtensionServerPolicies[j].GetCreationTimestamp()
		if tsI.Equal(&tsJ) {
			if r.ExtensionServerPolicies[i].GetNamespace() != r.ExtensionServerPolicies[j].GetNamespace() {
				return r.ExtensionServerPolicies[i].GetNamespace() < r.ExtensionServerPolicies[j].GetNamespace()
			}
			return r.ExtensionServerPolicies[i].GetName() < r.ExtensionServerPolicies[j].GetName()
		}
		return tsI.Before(&tsJ)
	})
}
