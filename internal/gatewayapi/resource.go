// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"cmp"
	"reflect"

	"golang.org/x/exp/slices"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	mcsapi "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

type XdsIRMap map[string]*ir.Xds
type InfraIRMap map[string]*ir.Infra

// Resources holds the Gateway API and related
// resources that the translators needs as inputs.
// +k8s:deepcopy-gen=true
type Resources struct {
	// This field is only used for marshalling/unmarshalling purposes and is not used by
	// the translator
	GatewayClass           *gwapiv1.GatewayClass          `json:"gatewayClass,omitempty" yaml:"gatewayClass,omitempty"`
	Gateways               []*gwapiv1.Gateway             `json:"gateways,omitempty" yaml:"gateways,omitempty"`
	HTTPRoutes             []*gwapiv1.HTTPRoute           `json:"httpRoutes,omitempty" yaml:"httpRoutes,omitempty"`
	GRPCRoutes             []*gwapiv1a2.GRPCRoute         `json:"grpcRoutes,omitempty" yaml:"grpcRoutes,omitempty"`
	TLSRoutes              []*gwapiv1a2.TLSRoute          `json:"tlsRoutes,omitempty" yaml:"tlsRoutes,omitempty"`
	TCPRoutes              []*gwapiv1a2.TCPRoute          `json:"tcpRoutes,omitempty" yaml:"tcpRoutes,omitempty"`
	UDPRoutes              []*gwapiv1a2.UDPRoute          `json:"udpRoutes,omitempty" yaml:"udpRoutes,omitempty"`
	ReferenceGrants        []*gwapiv1b1.ReferenceGrant    `json:"referenceGrants,omitempty" yaml:"referenceGrants,omitempty"`
	Namespaces             []*v1.Namespace                `json:"namespaces,omitempty" yaml:"namespaces,omitempty"`
	Services               []*v1.Service                  `json:"services,omitempty" yaml:"services,omitempty"`
	ServiceImports         []*mcsapi.ServiceImport        `json:"serviceImports,omitempty" yaml:"serviceImports,omitempty"`
	EndpointSlices         []*discoveryv1.EndpointSlice   `json:"endpointSlices,omitempty" yaml:"endpointSlices,omitempty"`
	Secrets                []*v1.Secret                   `json:"secrets,omitempty" yaml:"secrets,omitempty"`
	ConfigMaps             []*v1.ConfigMap                `json:"configMaps,omitempty" yaml:"configMaps,omitempty"`
	EnvoyProxy             *egv1a1.EnvoyProxy             `json:"envoyProxy,omitempty" yaml:"envoyProxy,omitempty"`
	ExtensionRefFilters    []unstructured.Unstructured    `json:"extensionRefFilters,omitempty" yaml:"extensionRefFilters,omitempty"`
	EnvoyPatchPolicies     []*egv1a1.EnvoyPatchPolicy     `json:"envoyPatchPolicies,omitempty" yaml:"envoyPatchPolicies,omitempty"`
	ClientTrafficPolicies  []*egv1a1.ClientTrafficPolicy  `json:"clientTrafficPolicies,omitempty" yaml:"clientTrafficPolicies,omitempty"`
	BackendTrafficPolicies []*egv1a1.BackendTrafficPolicy `json:"backendTrafficPolicies,omitempty" yaml:"backendTrafficPolicies,omitempty"`
	SecurityPolicies       []*egv1a1.SecurityPolicy       `json:"securityPolicies,omitempty" yaml:"securityPolicies,omitempty"`
	BackendTLSPolicies     []*gwapiv1a2.BackendTLSPolicy  `json:"backendTLSPolicies,omitempty" yaml:"backendTLSPolicies,omitempty"`
}

func NewResources() *Resources {
	return &Resources{
		Gateways:               []*gwapiv1.Gateway{},
		HTTPRoutes:             []*gwapiv1.HTTPRoute{},
		GRPCRoutes:             []*gwapiv1a2.GRPCRoute{},
		TLSRoutes:              []*gwapiv1a2.TLSRoute{},
		Services:               []*v1.Service{},
		EndpointSlices:         []*discoveryv1.EndpointSlice{},
		Secrets:                []*v1.Secret{},
		ConfigMaps:             []*v1.ConfigMap{},
		ReferenceGrants:        []*gwapiv1b1.ReferenceGrant{},
		Namespaces:             []*v1.Namespace{},
		ExtensionRefFilters:    []unstructured.Unstructured{},
		EnvoyPatchPolicies:     []*egv1a1.EnvoyPatchPolicy{},
		ClientTrafficPolicies:  []*egv1a1.ClientTrafficPolicy{},
		BackendTrafficPolicies: []*egv1a1.BackendTrafficPolicy{},
		SecurityPolicies:       []*egv1a1.SecurityPolicy{},
		BackendTLSPolicies:     []*gwapiv1a2.BackendTLSPolicy{},
	}
}

func (r *Resources) GetNamespace(name string) *v1.Namespace {
	for _, ns := range r.Namespaces {
		if ns.Name == name {
			return ns
		}
	}

	return nil
}

func (r *Resources) GetService(namespace, name string) *v1.Service {
	for _, svc := range r.Services {
		if svc.Namespace == namespace && svc.Name == name {
			return svc
		}
	}

	return nil
}

func (r *Resources) GetServiceImport(namespace, name string) *mcsapi.ServiceImport {
	for _, svcImp := range r.ServiceImports {
		if svcImp.Namespace == namespace && svcImp.Name == name {
			return svcImp
		}
	}

	return nil
}

func (r *Resources) GetSecret(namespace, name string) *v1.Secret {
	for _, secret := range r.Secrets {
		if secret.Namespace == namespace && secret.Name == name {
			return secret
		}
	}

	return nil
}

func (r *Resources) GetConfigMap(namespace, name string) *v1.ConfigMap {
	for _, configMap := range r.ConfigMaps {
		if configMap.Namespace == namespace && configMap.Name == name {
			return configMap
		}
	}

	return nil
}

func (r *Resources) GetEndpointSlicesForBackend(svcNamespace, svcName string, backendKind string) []*discoveryv1.EndpointSlice {
	var endpointSlices []*discoveryv1.EndpointSlice
	for _, endpointSlice := range r.EndpointSlices {
		var backendSelectorLabel string
		switch backendKind {
		case KindService:
			backendSelectorLabel = discoveryv1.LabelServiceName
		case KindServiceImport:
			backendSelectorLabel = mcsapi.LabelServiceName
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
