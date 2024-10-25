// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
	xdstypes "github.com/envoyproxy/gateway/internal/xds/types"
)

// ProviderResources message
type ProviderResources struct {
	// GatewayAPIResources is a map from a GatewayClass name to
	// a group of gateway API and other related resources.
	GatewayAPIResources WatchableMap[string, *resource.ControllerResources]

	// GatewayAPIStatuses is a group of gateway api
	// resource statuses maps.
	GatewayAPIStatuses

	// PolicyStatuses is a group of policy statuses maps.
	PolicyStatuses

	// ExtensionStatuses is a group of gw-api extension resource statuses map.
	ExtensionStatuses
}

func (p *ProviderResources) GetResources() []*resource.Resources {
	if p.GatewayAPIResources.Len() == 0 {
		return nil
	}

	for _, v := range p.GatewayAPIResources.LoadAll() {
		return *v
	}

	return nil
}

func (p *ProviderResources) GetResourcesByGatewayClass(name string) *resource.Resources {
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
	GatewayStatuses   WatchableMap[types.NamespacedName, *gwapiv1.GatewayStatus]
	HTTPRouteStatuses WatchableMap[types.NamespacedName, *gwapiv1.HTTPRouteStatus]
	GRPCRouteStatuses WatchableMap[types.NamespacedName, *gwapiv1.GRPCRouteStatus]
	TLSRouteStatuses  WatchableMap[types.NamespacedName, *gwapiv1a2.TLSRouteStatus]
	TCPRouteStatuses  WatchableMap[types.NamespacedName, *gwapiv1a2.TCPRouteStatus]
	UDPRouteStatuses  WatchableMap[types.NamespacedName, *gwapiv1a2.UDPRouteStatus]
}

func (s *GatewayAPIStatuses) Close() {
	s.GatewayStatuses.Close()
	s.HTTPRouteStatuses.Close()
	s.GRPCRouteStatuses.Close()
	s.TLSRouteStatuses.Close()
	s.TCPRouteStatuses.Close()
	s.UDPRouteStatuses.Close()
}

type NamespacedNameAndGVK struct {
	types.NamespacedName
	schema.GroupVersionKind
}

// PolicyStatuses contains policy related resources statuses
type PolicyStatuses struct {
	ClientTrafficPolicyStatuses  WatchableMap[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	BackendTrafficPolicyStatuses WatchableMap[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	EnvoyPatchPolicyStatuses     WatchableMap[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	SecurityPolicyStatuses       WatchableMap[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	BackendTLSPolicyStatuses     WatchableMap[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	EnvoyExtensionPolicyStatuses WatchableMap[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	ExtensionPolicyStatuses      WatchableMap[NamespacedNameAndGVK, *gwapiv1a2.PolicyStatus]
}

// ExtensionStatuses contains statuses related to gw-api extension resources
type ExtensionStatuses struct {
	BackendStatuses WatchableMap[types.NamespacedName, *egv1a1.BackendStatus]
}

func (p *PolicyStatuses) Close() {
	p.ClientTrafficPolicyStatuses.Close()
	p.SecurityPolicyStatuses.Close()
	p.EnvoyPatchPolicyStatuses.Close()
	p.BackendTLSPolicyStatuses.Close()
	p.EnvoyExtensionPolicyStatuses.Close()
	p.ExtensionPolicyStatuses.Close()
}

// XdsIR message
type XdsIR struct {
	WatchableMap[string, *ir.Xds]
}

// InfraIR message
type InfraIR struct {
	WatchableMap[string, *ir.Infra]
}

// Xds message
type Xds struct {
	WatchableMap[string, *xdstypes.ResourceVersionTable]
}
