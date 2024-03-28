// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import (
	"github.com/telepresenceio/watchable"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/ir"
	xdstypes "github.com/envoyproxy/gateway/internal/xds/types"
)

// ProviderResources message
type ProviderResources struct {
	// GatewayAPIResources is a map from a GatewayClass name to
	// a group of gateway API and other related resources.
	GatewayAPIResources watchable.Map[string, *gatewayapi.ControllerResources]

	// GatewayAPIStatuses is a group of gateway api
	// resource statuses maps.
	GatewayAPIStatuses

	// PolicyStatuses is a group of policy statuses maps.
	PolicyStatuses
}

func (p *ProviderResources) GetResources() []*gatewayapi.Resources {
	if p.GatewayAPIResources.Len() == 0 {
		return nil
	}

	for _, v := range p.GatewayAPIResources.LoadAll() {
		return *v
	}

	return nil
}

func (p *ProviderResources) GetResourcesByGatewayClass(name string) *gatewayapi.Resources {
	for _, r := range p.GetResources() {
		if r != nil && r.GatewayClass != nil && r.GatewayClass.Name == name {
			return r
		}
	}

	return nil
}

func (p *ProviderResources) GetResourcesKey() string {
	if p.GatewayAPIResources.Len() == 0 {
		return ""
	}
	for k := range p.GatewayAPIResources.LoadAll() {
		return k
	}
	return ""
}

func (p *ProviderResources) Close() {
	p.GatewayAPIResources.Close()
	p.GatewayAPIStatuses.Close()
	p.PolicyStatuses.Close()
}

// GatewayAPIStatuses contains gateway API resources statuses
type GatewayAPIStatuses struct {
	GatewayStatuses   watchable.Map[types.NamespacedName, *gwapiv1.GatewayStatus]
	HTTPRouteStatuses watchable.Map[types.NamespacedName, *gwapiv1.HTTPRouteStatus]
	GRPCRouteStatuses watchable.Map[types.NamespacedName, *gwapiv1a2.GRPCRouteStatus]
	TLSRouteStatuses  watchable.Map[types.NamespacedName, *gwapiv1a2.TLSRouteStatus]
	TCPRouteStatuses  watchable.Map[types.NamespacedName, *gwapiv1a2.TCPRouteStatus]
	UDPRouteStatuses  watchable.Map[types.NamespacedName, *gwapiv1a2.UDPRouteStatus]
}

func (s *GatewayAPIStatuses) Close() {
	s.GatewayStatuses.Close()
	s.HTTPRouteStatuses.Close()
	s.GRPCRouteStatuses.Close()
	s.TLSRouteStatuses.Close()
	s.TCPRouteStatuses.Close()
	s.UDPRouteStatuses.Close()
}

// PolicyStatuses contains policy related resources statuses
type PolicyStatuses struct {
	ClientTrafficPolicyStatuses  watchable.Map[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	BackendTrafficPolicyStatuses watchable.Map[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	EnvoyPatchPolicyStatuses     watchable.Map[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	SecurityPolicyStatuses       watchable.Map[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	BackendTLSPolicyStatuses     watchable.Map[types.NamespacedName, *gwapiv1a2.PolicyStatus]
}

func (p *PolicyStatuses) Close() {
	p.ClientTrafficPolicyStatuses.Close()
	p.SecurityPolicyStatuses.Close()
	p.EnvoyPatchPolicyStatuses.Close()
	p.BackendTLSPolicyStatuses.Close()
}

// XdsIR message
type XdsIR struct {
	watchable.Map[string, *ir.Xds]
}

// InfraIR message
type InfraIR struct {
	watchable.Map[string, *ir.Infra]
}

// Xds message
type Xds struct {
	watchable.Map[string, *xdstypes.ResourceVersionTable]
}
