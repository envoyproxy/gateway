// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import (
	"context"

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
	GatewayAPIResources *preSubscribedWatchableMap[string, *resource.ControllerResources]

	// GatewayAPIStatuses is a group of gateway api
	// resource statuses maps.
	GatewayAPIStatuses

	// PolicyStatuses is a group of policy statuses maps.
	PolicyStatuses

	// ExtensionStatuses is a group of gw-api extension resource statuses map.
	ExtensionStatuses
}

// NewSubscribedProviderResources creates a new ProviderResources with the given context and required subscriptions.
// Update subscription count here if more subscriptions are needed.
func NewSubscribedProviderResources(ctx context.Context) *ProviderResources {
	return &ProviderResources{
		GatewayAPIResources: newPreSubscribedWatchableMap[string, *resource.ControllerResources](ctx, 1),
		GatewayAPIStatuses: GatewayAPIStatuses{
			GatewayClassStatuses: newPreSubscribedWatchableMap[types.NamespacedName, *gwapiv1.GatewayClassStatus](ctx, 1),
			GatewayStatuses:      newPreSubscribedWatchableMap[types.NamespacedName, *gwapiv1.GatewayStatus](ctx, 1),
			HTTPRouteStatuses:    newPreSubscribedWatchableMap[types.NamespacedName, *gwapiv1.HTTPRouteStatus](ctx, 1),
			GRPCRouteStatuses:    newPreSubscribedWatchableMap[types.NamespacedName, *gwapiv1.GRPCRouteStatus](ctx, 1),
			TLSRouteStatuses:     newPreSubscribedWatchableMap[types.NamespacedName, *gwapiv1a2.TLSRouteStatus](ctx, 1),
			TCPRouteStatuses:     newPreSubscribedWatchableMap[types.NamespacedName, *gwapiv1a2.TCPRouteStatus](ctx, 1),
			UDPRouteStatuses:     newPreSubscribedWatchableMap[types.NamespacedName, *gwapiv1a2.UDPRouteStatus](ctx, 1),
		},
		PolicyStatuses: PolicyStatuses{
			ClientTrafficPolicyStatuses:  newPreSubscribedWatchableMap[types.NamespacedName, *gwapiv1a2.PolicyStatus](ctx, 1),
			BackendTrafficPolicyStatuses: newPreSubscribedWatchableMap[types.NamespacedName, *gwapiv1a2.PolicyStatus](ctx, 1),
			EnvoyPatchPolicyStatuses:     newPreSubscribedWatchableMap[types.NamespacedName, *gwapiv1a2.PolicyStatus](ctx, 1),
			SecurityPolicyStatuses:       newPreSubscribedWatchableMap[types.NamespacedName, *gwapiv1a2.PolicyStatus](ctx, 1),
			BackendTLSPolicyStatuses:     newPreSubscribedWatchableMap[types.NamespacedName, *gwapiv1a2.PolicyStatus](ctx, 1),
			EnvoyExtensionPolicyStatuses: newPreSubscribedWatchableMap[types.NamespacedName, *gwapiv1a2.PolicyStatus](ctx, 1),
			ExtensionPolicyStatuses:      newPreSubscribedWatchableMap[NamespacedNameAndGVK, *gwapiv1a2.PolicyStatus](ctx, 1),
		},
		ExtensionStatuses: ExtensionStatuses{
			BackendStatuses: newPreSubscribedWatchableMap[types.NamespacedName, *egv1a1.BackendStatus](ctx, 1),
		},
	}
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
	GatewayClassStatuses *preSubscribedWatchableMap[types.NamespacedName, *gwapiv1.GatewayClassStatus]
	GatewayStatuses      *preSubscribedWatchableMap[types.NamespacedName, *gwapiv1.GatewayStatus]
	HTTPRouteStatuses    *preSubscribedWatchableMap[types.NamespacedName, *gwapiv1.HTTPRouteStatus]
	GRPCRouteStatuses    *preSubscribedWatchableMap[types.NamespacedName, *gwapiv1.GRPCRouteStatus]
	TLSRouteStatuses     *preSubscribedWatchableMap[types.NamespacedName, *gwapiv1a2.TLSRouteStatus]
	TCPRouteStatuses     *preSubscribedWatchableMap[types.NamespacedName, *gwapiv1a2.TCPRouteStatus]
	UDPRouteStatuses     *preSubscribedWatchableMap[types.NamespacedName, *gwapiv1a2.UDPRouteStatus]
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
	ClientTrafficPolicyStatuses  *preSubscribedWatchableMap[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	BackendTrafficPolicyStatuses *preSubscribedWatchableMap[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	EnvoyPatchPolicyStatuses     *preSubscribedWatchableMap[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	SecurityPolicyStatuses       *preSubscribedWatchableMap[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	BackendTLSPolicyStatuses     *preSubscribedWatchableMap[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	EnvoyExtensionPolicyStatuses *preSubscribedWatchableMap[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	ExtensionPolicyStatuses      *preSubscribedWatchableMap[NamespacedNameAndGVK, *gwapiv1a2.PolicyStatus]
}

// ExtensionStatuses contains statuses related to gw-api extension resources
type ExtensionStatuses struct {
	BackendStatuses *preSubscribedWatchableMap[types.NamespacedName, *egv1a1.BackendStatus]
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
type XdsIR = preSubscribedWatchableMap[string, *ir.Xds]

// NewSubscribedXdsIR creates a new XdsIR with the given context and required subscriptions.
// Update subscription count here if more subscriptions are needed.
func NewSubscribedXdsIR(ctx context.Context) *XdsIR {
	return newPreSubscribedWatchableMap[string, *ir.Xds](ctx, 2)
}

// InfraIR message
type InfraIR = preSubscribedWatchableMap[string, *ir.Infra]

// NewSubscribedInfraIR creates a new InfraIR with the given context and required subscriptions.
// Update subscription count here if more subscriptions are needed.
func NewSubscribedInfraIR(ctx context.Context) *InfraIR {
	return newPreSubscribedWatchableMap[string, *ir.Infra](ctx, 1)
}

// Xds message
type Xds = preSubscribedWatchableMap[string, *xdstypes.ResourceVersionTable]

// NewSubscribedXds creates a new Xds with the given context and required subscriptions.
// Update subscription count here if more subscriptions are needed.
func NewSubscribedXds(ctx context.Context) *Xds {
	return newPreSubscribedWatchableMap[string, *xdstypes.ResourceVersionTable](ctx, 1)
}

type MessageName string

const (
	// XDSMessageName is a message containing xds translated from xds-ir
	XDSMessageName MessageName = "xds"
	// XDSIRMessageName is a message containing xds-ir translated from provider-resources
	XDSIRMessageName MessageName = "xds-ir"
	// InfraIRMessageName is a message containing infra-ir translated from provider-resources
	InfraIRMessageName MessageName = "infra-ir"
	// ProviderResourcesMessageName is a message containing gw-api and envoy gateway resources from the provider
	ProviderResourcesMessageName MessageName = "provider-resources"
	// BackendStatusMessageName is a message containing updates to Backend status
	BackendStatusMessageName MessageName = "backend-status"
	// ExtensionServerPoliciesStatusMessageName is a message containing updates to ExtensionServerPolicy status
	ExtensionServerPoliciesStatusMessageName MessageName = "extensionserverpolicies-status"
	// EnvoyExtensionPolicyStatusMessageName is a message containing updates to EnvoyExtensionPolicy status
	EnvoyExtensionPolicyStatusMessageName MessageName = "envoyextensionpolicy-status"
	// EnvoyPatchPolicyStatusMessageName is a message containing updates to EnvoyPatchPolicy status
	EnvoyPatchPolicyStatusMessageName MessageName = "envoypatchpolicy-status"
	// SecurityPolicyStatusMessageName is a message containing updates to SecurityPolicy status
	SecurityPolicyStatusMessageName MessageName = "securitypolicy-status"
	// BackendTrafficPolicyStatusMessageName is a message containing updates to BackendTrafficPolicy status
	BackendTrafficPolicyStatusMessageName MessageName = "backendtrafficpolicy-status"
	// ClientTrafficPolicyStatusMessageName is a message containing updates to ClientTrafficPolicy status
	ClientTrafficPolicyStatusMessageName MessageName = "clienttrafficpolicy-status"
	// BackendTLSPolicyStatusMessageName is a message containing updates to BackendTLSPolicy status
	BackendTLSPolicyStatusMessageName MessageName = "backendtlspolicy-status"
	// UDPRouteStatusMessageName is a message containing updates to UDPRoute status
	UDPRouteStatusMessageName MessageName = "udproute-status"
	// TCPRouteStatusMessageName is a message containing updates to TCPRoute status
	TCPRouteStatusMessageName MessageName = "tcproute-status"
	// TLSRouteStatusMessageName is a message containing updates to TLSRoute status
	TLSRouteStatusMessageName MessageName = "tlsroute-status"
	// GRPCRouteStatusMessageName is a message containing updates to GRPCRoute status
	GRPCRouteStatusMessageName MessageName = "grpcroute-status"
	// HTTPRouteStatusMessageName is a message containing updates to HTTPRoute status
	HTTPRouteStatusMessageName MessageName = "httproute-status"
	// GatewayStatusMessageName is a message containing updates to Gateway status
	GatewayStatusMessageName MessageName = "gateway-status"
	// GatewayClassStatusMessageName is a message containing updates to GatewayClass status
	GatewayClassStatusMessageName MessageName = "gatewayclass-status"
)
