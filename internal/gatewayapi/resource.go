// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
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
	GatewayClass          *v1beta1.GatewayClass          `json:"gatewayClass,omitempty"`
	Gateways              []*v1beta1.Gateway             `json:"gateways,omitempty"`
	HTTPRoutes            []*v1beta1.HTTPRoute           `json:"httpRoutes,omitempty"`
	GRPCRoutes            []*v1alpha2.GRPCRoute          `json:"grpcRoutes,omitempty"`
	TLSRoutes             []*v1alpha2.TLSRoute           `json:"tlsRoutes,omitempty"`
	TCPRoutes             []*v1alpha2.TCPRoute           `json:"tcpRoutes,omitempty"`
	UDPRoutes             []*v1alpha2.UDPRoute           `json:"udpRoutes,omitempty"`
	ReferenceGrants       []*v1alpha2.ReferenceGrant     `json:"referenceGrants,omitempty"`
	Namespaces            []*v1.Namespace                `json:"namespaces,omitempty"`
	Services              []*v1.Service                  `json:"services,omitempty"`
	Secrets               []*v1.Secret                   `json:"secrets,omitempty"`
	AuthenticationFilters []*egv1a1.AuthenticationFilter `json:"authenticationFilters,omitempty"`
	RateLimitFilters      []*egv1a1.RateLimitFilter      `json:"rateLimitFilters,omitempty"`
	EnvoyProxy            *egcfgv1a1.EnvoyProxy          `json:"envoyProxy,omitempty"`
	ExtensionRefFilters   []unstructured.Unstructured    `json:"extensionRefFilters,omitempty"`
}

func NewResources() *Resources {
	return &Resources{
		Gateways:              []*v1beta1.Gateway{},
		HTTPRoutes:            []*v1beta1.HTTPRoute{},
		GRPCRoutes:            []*v1alpha2.GRPCRoute{},
		TLSRoutes:             []*v1alpha2.TLSRoute{},
		Services:              []*v1.Service{},
		Secrets:               []*v1.Secret{},
		ReferenceGrants:       []*v1alpha2.ReferenceGrant{},
		Namespaces:            []*v1.Namespace{},
		RateLimitFilters:      []*egv1a1.RateLimitFilter{},
		EnvoyProxy:            new(egcfgv1a1.EnvoyProxy),
		AuthenticationFilters: []*egv1a1.AuthenticationFilter{},
		ExtensionRefFilters:   []unstructured.Unstructured{},
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

func (r *Resources) GetSecret(namespace, name string) *v1.Secret {
	for _, secret := range r.Secrets {
		if secret.Namespace == namespace && secret.Name == name {
			return secret
		}
	}

	return nil
}
