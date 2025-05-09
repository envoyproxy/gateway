// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"cmp"
	"reflect"

	"golang.org/x/exp/slices"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	gwapixv1a1 "sigs.k8s.io/gateway-api/apisx/v1alpha1"
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

	XBackendTrafficPolicies []*gwapixv1a1.XBackendTrafficPolicy `json:"xBackendTrafficPolicies,omitempty" yaml:"xBackendTrafficPolicies,omitempty"`

	serviceMap map[types.NamespacedName]*corev1.Service
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
// This function creates a HashMap of Services for faster lookup when it's called for the first time.
// Subsequent calls will use the HashMap for lookup.
// Note:
// - This function is not thread-safe.
// - This function should be called after all the Services are added to the Resources.
func (r *Resources) GetService(namespace, name string) *corev1.Service {
	if r.serviceMap == nil {
		r.serviceMap = make(map[types.NamespacedName]*corev1.Service)
		for _, svc := range r.Services {
			r.serviceMap[types.NamespacedName{Namespace: svc.Namespace, Name: svc.Name}] = svc
		}
	}
	return r.serviceMap[types.NamespacedName{Namespace: namespace, Name: name}]
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

// Equal implements the Comparable interface used by watchable.DeepEqual to skip unnecessary updates.
func (c *ControllerResources) Equal(y *ControllerResources) bool {
	// Deep copy to avoid modifying the original ordering.
	c = c.DeepCopy()
	c.sort()
	y = y.DeepCopy()
	y.sort()
	return reflect.DeepEqual(c, y)
}

func (c *ControllerResources) sort() {
	slices.SortFunc(*c, func(c1, c2 *Resources) int {
		return cmp.Compare(c1.GatewayClass.Name, c2.GatewayClass.Name)
	})
}
