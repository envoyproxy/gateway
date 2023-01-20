// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	v1 "k8s.io/api/core/v1"
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
	Gateways              []*v1beta1.Gateway
	HTTPRoutes            []*v1beta1.HTTPRoute
	GRPCRoutes            []*v1alpha2.GRPCRoute
	TLSRoutes             []*v1alpha2.TLSRoute
	TCPRoutes             []*v1alpha2.TCPRoute
	UDPRoutes             []*v1alpha2.UDPRoute
	ReferenceGrants       []*v1alpha2.ReferenceGrant
	Namespaces            []*v1.Namespace
	Services              []*v1.Service
	Secrets               []*v1.Secret
	AuthenticationFilters []*egv1a1.AuthenticationFilter
	RateLimitFilters      []*egv1a1.RateLimitFilter
	EnvoyProxy            *egcfgv1a1.EnvoyProxy
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
