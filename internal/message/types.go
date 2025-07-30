// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import (
	"github.com/telepresenceio/watchable"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
)

// ProviderResources message
type ProviderResources struct {
	// GatewayAPIResources is a map from a GatewayClass name to
	// a group of gateway API and other related resources.
	GatewayAPIResources watchable.Map[string, *resource.ControllerResources]

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
	GatewayClassStatuses watchable.Map[types.NamespacedName, *gwapiv1.GatewayClassStatus]
	GatewayStatuses      watchable.Map[types.NamespacedName, *gwapiv1.GatewayStatus]
	HTTPRouteStatuses    watchable.Map[types.NamespacedName, *gwapiv1.HTTPRouteStatus]
	GRPCRouteStatuses    watchable.Map[types.NamespacedName, *gwapiv1.GRPCRouteStatus]
	TLSRouteStatuses     watchable.Map[types.NamespacedName, *gwapiv1a2.TLSRouteStatus]
	TCPRouteStatuses     watchable.Map[types.NamespacedName, *gwapiv1a2.TCPRouteStatus]
	UDPRouteStatuses     watchable.Map[types.NamespacedName, *gwapiv1a2.UDPRouteStatus]
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
	ClientTrafficPolicyStatuses  watchable.Map[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	BackendTrafficPolicyStatuses watchable.Map[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	EnvoyPatchPolicyStatuses     watchable.Map[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	SecurityPolicyStatuses       watchable.Map[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	BackendTLSPolicyStatuses     watchable.Map[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	EnvoyExtensionPolicyStatuses watchable.Map[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	ExtensionPolicyStatuses      watchable.Map[NamespacedNameAndGVK, *gwapiv1a2.PolicyStatus]
}

// ExtensionStatuses contains statuses related to gw-api extension resources
type ExtensionStatuses struct {
	BackendStatuses watchable.Map[types.NamespacedName, *egv1a1.BackendStatus]
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
	watchable.Map[string, *ir.Xds]
}

// InfraIR message
type InfraIR struct {
	watchable.Map[string, *ir.Infra]
}

type MessageName string

const (
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
