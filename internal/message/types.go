// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import (
	"context"

	"github.com/telepresenceio/watchable"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
	xdstypes "github.com/envoyproxy/gateway/internal/xds/types"
)

// Subscription counts for each resource type.
// Increment if more subscribers are needed for a resource.
const (
	subscriptionCountGatewayAPIResources  = 1
	subscriptionCountGatewayClassStatuses = 1
	subscriptionCountGatewayStatuses      = 1
	subscriptionCountHTTPRouteStatuses    = 1
	subscriptionCountGRPCRouteStatuses    = 1
	subscriptionCountTLSRouteStatuses     = 1
	subscriptionCountTCPRouteStatuses     = 1
	subscriptionCountUDPRouteStatuses     = 1
	subscriptionCountPolicyStatuses       = 1
	subscriptionCountBackendStatuses      = 1
	subscriptionCountXdsIR                = 2
	subscriptionCountInfraIR              = 1
	subscriptionCountXds                  = 1
)

// Subscription for Gateway API Resources
type GatewayAPIResources struct {
	watchable.Map[string, *resource.ControllerResources]
	Subscriptions *SubscriptionList[string, *resource.ControllerResources]
}

// ProviderResources message
type ProviderResources struct {
	// GatewayAPIResources is a map from a GatewayClass name to
	// a group of gateway API and other related resources.
	GatewayAPIResources GatewayAPIResources

	// GatewayAPIStatuses is a group of gateway api
	// resource statuses maps.
	GatewayAPIStatuses

	// PolicyStatuses is a group of policy statuses maps.
	PolicyStatuses

	// ExtensionStatuses is a group of gw-api extension resource statuses map.
	ExtensionStatuses
}

// NewSubscribedProviderResources creates a new ProviderResources instance,
// initializes all subscriptions, and returns the fully subscribed instance.
func NewSubscribedProviderResources(ctx context.Context) *ProviderResources {
	pr := new(ProviderResources)
	pr.Subscribe(ctx)
	return pr
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
	p.ExtensionStatuses.Close()
}

// Subscribe initializes the subscriptions for all the resources under ProviderResources.
// Note that number of subscriptions for each resource is fixed defined in subscriptionCount* constants.
func (p *ProviderResources) Subscribe(ctx context.Context) {
	p.GatewayAPIResources.Subscriptions = NewSubscriptionList(ctx, &p.GatewayAPIResources, subscriptionCountGatewayAPIResources)
	p.GatewayAPIStatuses.Subscribe(ctx)
	p.PolicyStatuses.Subscribe(ctx)
	p.ExtensionStatuses.Subscribe(ctx)
}

// Subscriptions for Gateway API Statuses
type GatewayClassStatuses struct {
	watchable.Map[types.NamespacedName, *gwapiv1.GatewayClassStatus]
	Subscriptions *SubscriptionList[types.NamespacedName, *gwapiv1.GatewayClassStatus]
}

type GatewayStatuses struct {
	watchable.Map[types.NamespacedName, *gwapiv1.GatewayStatus]
	Subscriptions *SubscriptionList[types.NamespacedName, *gwapiv1.GatewayStatus]
}

type HTTPRouteStatuses struct {
	watchable.Map[types.NamespacedName, *gwapiv1.HTTPRouteStatus]
	Subscriptions *SubscriptionList[types.NamespacedName, *gwapiv1.HTTPRouteStatus]
}

type GRPCRouteStatuses struct {
	watchable.Map[types.NamespacedName, *gwapiv1.GRPCRouteStatus]
	Subscriptions *SubscriptionList[types.NamespacedName, *gwapiv1.GRPCRouteStatus]
}

type TLSRouteStatuses struct {
	watchable.Map[types.NamespacedName, *gwapiv1a2.TLSRouteStatus]
	Subscriptions *SubscriptionList[types.NamespacedName, *gwapiv1a2.TLSRouteStatus]
}

type TCPRouteStatuses struct {
	watchable.Map[types.NamespacedName, *gwapiv1a2.TCPRouteStatus]
	Subscriptions *SubscriptionList[types.NamespacedName, *gwapiv1a2.TCPRouteStatus]
}

type UDPRouteStatuses struct {
	watchable.Map[types.NamespacedName, *gwapiv1a2.UDPRouteStatus]
	Subscriptions *SubscriptionList[types.NamespacedName, *gwapiv1a2.UDPRouteStatus]
}

// GatewayAPIStatuses contains gateway API resources statuses
type GatewayAPIStatuses struct {
	GatewayClassStatuses GatewayClassStatuses
	GatewayStatuses      GatewayStatuses
	HTTPRouteStatuses    HTTPRouteStatuses
	GRPCRouteStatuses    GRPCRouteStatuses
	TLSRouteStatuses     TLSRouteStatuses
	TCPRouteStatuses     TCPRouteStatuses
	UDPRouteStatuses     UDPRouteStatuses
}

func (s *GatewayAPIStatuses) Close() {
	s.GatewayClassStatuses.Close()
	s.GatewayStatuses.Close()
	s.HTTPRouteStatuses.Close()
	s.GRPCRouteStatuses.Close()
	s.TLSRouteStatuses.Close()
	s.TCPRouteStatuses.Close()
	s.UDPRouteStatuses.Close()
}

func (s *GatewayAPIStatuses) Subscribe(ctx context.Context) {
	s.GatewayClassStatuses.Subscriptions = NewSubscriptionList(ctx, &s.GatewayClassStatuses, subscriptionCountGatewayClassStatuses)
	s.GatewayStatuses.Subscriptions = NewSubscriptionList(ctx, &s.GatewayStatuses, subscriptionCountGatewayStatuses)
	s.HTTPRouteStatuses.Subscriptions = NewSubscriptionList(ctx, &s.HTTPRouteStatuses, subscriptionCountHTTPRouteStatuses)
	s.GRPCRouteStatuses.Subscriptions = NewSubscriptionList(ctx, &s.GRPCRouteStatuses, subscriptionCountGRPCRouteStatuses)
	s.TLSRouteStatuses.Subscriptions = NewSubscriptionList(ctx, &s.TLSRouteStatuses, subscriptionCountTLSRouteStatuses)
	s.TCPRouteStatuses.Subscriptions = NewSubscriptionList(ctx, &s.TCPRouteStatuses, subscriptionCountTCPRouteStatuses)
	s.UDPRouteStatuses.Subscriptions = NewSubscriptionList(ctx, &s.UDPRouteStatuses, subscriptionCountUDPRouteStatuses)
}

type NamespacedNameAndGVK struct {
	types.NamespacedName
	schema.GroupVersionKind
}

// Subscriptions for Policy Statuses
type ClientTrafficPolicyStatuses struct {
	watchable.Map[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	Subscriptions *SubscriptionList[types.NamespacedName, *gwapiv1a2.PolicyStatus]
}

type BackendTrafficPolicyStatuses struct {
	watchable.Map[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	Subscriptions *SubscriptionList[types.NamespacedName, *gwapiv1a2.PolicyStatus]
}

type EnvoyPatchPolicyStatuses struct {
	watchable.Map[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	Subscriptions *SubscriptionList[types.NamespacedName, *gwapiv1a2.PolicyStatus]
}

type SecurityPolicyStatuses struct {
	watchable.Map[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	Subscriptions *SubscriptionList[types.NamespacedName, *gwapiv1a2.PolicyStatus]
}

type BackendTLSPolicyStatuses struct {
	watchable.Map[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	Subscriptions *SubscriptionList[types.NamespacedName, *gwapiv1a2.PolicyStatus]
}

type EnvoyExtensionPolicyStatuses struct {
	watchable.Map[types.NamespacedName, *gwapiv1a2.PolicyStatus]
	Subscriptions *SubscriptionList[types.NamespacedName, *gwapiv1a2.PolicyStatus]
}

type ExtensionPolicyStatuses struct {
	watchable.Map[NamespacedNameAndGVK, *gwapiv1a2.PolicyStatus]
	Subscriptions *SubscriptionList[NamespacedNameAndGVK, *gwapiv1a2.PolicyStatus]
}

// PolicyStatuses contains policy related resources statuses
type PolicyStatuses struct {
	ClientTrafficPolicyStatuses  ClientTrafficPolicyStatuses
	BackendTrafficPolicyStatuses BackendTrafficPolicyStatuses
	EnvoyPatchPolicyStatuses     EnvoyPatchPolicyStatuses
	SecurityPolicyStatuses       SecurityPolicyStatuses
	BackendTLSPolicyStatuses     BackendTLSPolicyStatuses
	EnvoyExtensionPolicyStatuses EnvoyExtensionPolicyStatuses
	ExtensionPolicyStatuses      ExtensionPolicyStatuses
}

// Subscriptions for Extension Statuses
type BackendStatuses struct {
	watchable.Map[types.NamespacedName, *egv1a1.BackendStatus]
	Subscriptions *SubscriptionList[types.NamespacedName, *egv1a1.BackendStatus]
}

// ExtensionStatuses contains statuses related to gw-api extension resources
type ExtensionStatuses struct {
	BackendStatuses BackendStatuses
}

func (e *ExtensionStatuses) Close() {
	e.BackendStatuses.Close()
}

func (e *ExtensionStatuses) Subscribe(ctx context.Context) {
	e.BackendStatuses.Subscriptions = NewSubscriptionList(ctx, &e.BackendStatuses, subscriptionCountBackendStatuses)
}

func (p *PolicyStatuses) Close() {
	p.ClientTrafficPolicyStatuses.Close()
	p.BackendTrafficPolicyStatuses.Close()
	p.EnvoyPatchPolicyStatuses.Close()
	p.SecurityPolicyStatuses.Close()
	p.BackendTLSPolicyStatuses.Close()
	p.EnvoyExtensionPolicyStatuses.Close()
	p.ExtensionPolicyStatuses.Close()
}

func (p *PolicyStatuses) Subscribe(ctx context.Context) {
	p.ClientTrafficPolicyStatuses.Subscriptions = NewSubscriptionList(ctx, &p.ClientTrafficPolicyStatuses, subscriptionCountPolicyStatuses)
	p.BackendTrafficPolicyStatuses.Subscriptions = NewSubscriptionList(ctx, &p.BackendTrafficPolicyStatuses, subscriptionCountPolicyStatuses)
	p.EnvoyPatchPolicyStatuses.Subscriptions = NewSubscriptionList(ctx, &p.EnvoyPatchPolicyStatuses, subscriptionCountPolicyStatuses)
	p.SecurityPolicyStatuses.Subscriptions = NewSubscriptionList(ctx, &p.SecurityPolicyStatuses, subscriptionCountPolicyStatuses)
	p.BackendTLSPolicyStatuses.Subscriptions = NewSubscriptionList(ctx, &p.BackendTLSPolicyStatuses, subscriptionCountPolicyStatuses)
	p.EnvoyExtensionPolicyStatuses.Subscriptions = NewSubscriptionList(ctx, &p.EnvoyExtensionPolicyStatuses, subscriptionCountPolicyStatuses)
	p.ExtensionPolicyStatuses.Subscriptions = NewSubscriptionList(ctx, &p.ExtensionPolicyStatuses, subscriptionCountPolicyStatuses)
}

// XdsIR message
type XdsIR struct {
	watchable.Map[string, *ir.Xds]
	Subscriptions *SubscriptionList[string, *ir.Xds]
}

func NewSubscribedXdsIR(ctx context.Context) *XdsIR {
	x := new(XdsIR)
	x.Subscriptions = NewSubscriptionList(ctx, x, subscriptionCountXdsIR)
	return x
}

// InfraIR message
type InfraIR struct {
	watchable.Map[string, *ir.Infra]
	Subscriptions *SubscriptionList[string, *ir.Infra]
}

func NewSubscribedInfraIR(ctx context.Context) *InfraIR {
	x := new(InfraIR)
	x.Subscriptions = NewSubscriptionList(ctx, x, subscriptionCountInfraIR)
	return x
}

// Xds message
type Xds struct {
	watchable.Map[string, *xdstypes.ResourceVersionTable]
	Subscriptions *SubscriptionList[string, *xdstypes.ResourceVersionTable]
}

func NewSubscribedXds(ctx context.Context) *Xds {
	x := new(Xds)
	x.Subscriptions = NewSubscriptionList(ctx, x, subscriptionCountXds)
	return x
}
