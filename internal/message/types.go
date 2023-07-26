// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import (
	"github.com/telepresenceio/watchable"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/ir"
	xdstypes "github.com/envoyproxy/gateway/internal/xds/types"
)

// ProviderResources message
type ProviderResources struct {
	// GatewayAPIResources is a map from a GatewayClass name to
	// a group of gateway API resources.
	GatewayAPIResources watchable.Map[string, *gatewayapi.Resources]

	GatewayStatuses   watchable.Map[types.NamespacedName, *gwapiv1b1.GatewayStatus]
	HTTPRouteStatuses watchable.Map[types.NamespacedName, *gwapiv1b1.HTTPRouteStatus]
	GRPCRouteStatuses watchable.Map[types.NamespacedName, *gwapiv1a2.GRPCRouteStatus]
	TLSRouteStatuses  watchable.Map[types.NamespacedName, *gwapiv1a2.TLSRouteStatus]
	TCPRouteStatuses  watchable.Map[types.NamespacedName, *gwapiv1a2.TCPRouteStatus]
	UDPRouteStatuses  watchable.Map[types.NamespacedName, *gwapiv1a2.UDPRouteStatus]
}

func (p *ProviderResources) GetResources() *gatewayapi.Resources {
	if p.GatewayAPIResources.Len() == 0 {
		return nil
	}
	for _, v := range p.GatewayAPIResources.LoadAll() {
		return v
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
	p.GatewayStatuses.Close()
	p.HTTPRouteStatuses.Close()
	p.GRPCRouteStatuses.Close()
	p.TLSRouteStatuses.Close()
	p.TCPRouteStatuses.Close()
	p.UDPRouteStatuses.Close()
}

// EnvoyPatchPolicyStatuses message
type EnvoyPatchPolicyStatuses struct {
	watchable.Map[types.NamespacedName, *egv1a1.EnvoyPatchPolicyStatus]
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
