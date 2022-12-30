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

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/ir"
	xdstypes "github.com/envoyproxy/gateway/internal/xds/types"
)

// ProviderResources message
type ProviderResources struct {
	// GatewayAPIResources is a map from a GatewayClass name to
	// a group of gateway API resources.
	GatewayAPIResources watchable.Map[string, *gatewayapi.Resources]

	GatewayStatuses   watchable.Map[types.NamespacedName, *gwapiv1b1.Gateway]
	HTTPRouteStatuses watchable.Map[types.NamespacedName, *gwapiv1b1.HTTPRoute]
	GRPCRouteStatuses watchable.Map[types.NamespacedName, *gwapiv1a2.GRPCRoute]
	TLSRouteStatuses  watchable.Map[types.NamespacedName, *gwapiv1a2.TLSRoute]
	UDPRouteStatuses  watchable.Map[types.NamespacedName, *gwapiv1a2.UDPRoute]
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
	p.TLSRouteStatuses.Close()
	p.UDPRouteStatuses.Close()
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
