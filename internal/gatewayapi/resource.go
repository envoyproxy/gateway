// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
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
	GatewayClass          *gwapiv1.GatewayClass          `json:"gatewayClass,omitempty" yaml:"gatewayClass,omitempty"`
	Gateways              []*gwapiv1.Gateway             `json:"gateways,omitempty" yaml:"gateways,omitempty"`
	HTTPRoutes            []*gwapiv1.HTTPRoute           `json:"httpRoutes,omitempty" yaml:"httpRoutes,omitempty"`
	GRPCRoutes            []*gwapiv1a2.GRPCRoute         `json:"grpcRoutes,omitempty" yaml:"grpcRoutes,omitempty"`
	TLSRoutes             []*gwapiv1a2.TLSRoute          `json:"tlsRoutes,omitempty" yaml:"tlsRoutes,omitempty"`
	TCPRoutes             []*gwapiv1a2.TCPRoute          `json:"tcpRoutes,omitempty" yaml:"tcpRoutes,omitempty"`
	UDPRoutes             []*gwapiv1a2.UDPRoute          `json:"udpRoutes,omitempty" yaml:"udpRoutes,omitempty"`
	ReferenceGrants       []*gwapiv1a2.ReferenceGrant    `json:"referenceGrants,omitempty" yaml:"referenceGrants,omitempty"`
	Namespaces            []*v1.Namespace                `json:"namespaces,omitempty" yaml:"namespaces,omitempty"`
	Services              []*v1.Service                  `json:"services,omitempty" yaml:"services,omitempty"`
	ServiceImports        []*mcsapi.ServiceImport        `json:"serviceImports,omitempty" yaml:"serviceImports,omitempty"`
	EndpointSlices        []*discoveryv1.EndpointSlice   `json:"endpointSlices,omitempty" yaml:"endpointSlices,omitempty"`
	Secrets               []*v1.Secret                   `json:"secrets,omitempty" yaml:"secrets,omitempty"`
	AuthenticationFilters []*egv1a1.AuthenticationFilter `json:"authenticationFilters,omitempty" yaml:"authenticationFilters,omitempty"`
	RateLimitFilters      []*egv1a1.RateLimitFilter      `json:"rateLimitFilters,omitempty" yaml:"rateLimitFilters,omitempty"`
	EnvoyProxy            *egv1a1.EnvoyProxy             `json:"envoyProxy,omitempty" yaml:"envoyProxy,omitempty"`
	ExtensionRefFilters   []unstructured.Unstructured    `json:"extensionRefFilters,omitempty" yaml:"extensionRefFilters,omitempty"`
	EnvoyPatchPolicies    []*egv1a1.EnvoyPatchPolicy     `json:"envoyPatchPolicies,omitempty" yaml:"envoyPatchPolicies,omitempty"`
	ClientTrafficPolicies []*egv1a1.ClientTrafficPolicy  `json:"clientTrafficPolicies,omitempty" yaml:"clientTrafficPolicies,omitempty"`
}

func NewResources() *Resources {
	return &Resources{
		Gateways:              []*gwapiv1.Gateway{},
		HTTPRoutes:            []*gwapiv1.HTTPRoute{},
		GRPCRoutes:            []*gwapiv1a2.GRPCRoute{},
		TLSRoutes:             []*gwapiv1a2.TLSRoute{},
		Services:              []*v1.Service{},
		EndpointSlices:        []*discoveryv1.EndpointSlice{},
		Secrets:               []*v1.Secret{},
		ReferenceGrants:       []*gwapiv1a2.ReferenceGrant{},
		Namespaces:            []*v1.Namespace{},
		RateLimitFilters:      []*egv1a1.RateLimitFilter{},
		AuthenticationFilters: []*egv1a1.AuthenticationFilter{},
		ExtensionRefFilters:   []unstructured.Unstructured{},
		EnvoyPatchPolicies:    []*egv1a1.EnvoyPatchPolicy{},
		ClientTrafficPolicies: []*egv1a1.ClientTrafficPolicy{},
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

func (r *Resources) GetEndpointSlicesForBackend(svcNamespace, svcName string, backendKind string) []*discoveryv1.EndpointSlice {
	endpointSlices := []*discoveryv1.EndpointSlice{}
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
