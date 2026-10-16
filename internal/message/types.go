// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import (
	"context"
	"reflect"
	"time"

	"github.com/telepresenceio/watchable"
	"go.opentelemetry.io/otel/trace"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
)

// ProviderResources message
type ProviderResources struct {
	// GatewayAPIResources is a map from a GatewayClass name to
	// a group of gateway API and other related resources with trace context.
	GatewayAPIResources watchable.Map[string, *resource.ControllerResourcesContext]

	// EndpointUpdates is a map from a backend key (EndpointUpdate.Key()) to the
	// backend's current EndpointSlices, published by the provider for the
	// endpoint fast path. Only written when the EndpointFastPath runtime flag
	// is enabled.
	EndpointUpdates watchable.Map[string, *EndpointUpdate]

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
		if v != nil && v.Resources != nil {
			return *v.Resources
		}
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
	// EndpointUpdates is deliberately not closed: it is written from informer
	// event handlers that are not synchronized with this Close, and a Store on
	// a closed watchable.Map panics. Its subscribers terminate through their
	// own subscription contexts instead.
	p.GatewayAPIStatuses.Close()
	p.PolicyStatuses.Close()
	p.ExtensionStatuses.Close()
}

// GatewayAPIStatuses contains gateway API resources statuses
type GatewayAPIStatuses struct {
	GatewayClassStatuses watchable.Map[types.NamespacedName, *gwapiv1.GatewayClassStatus]
	GatewayStatuses      watchable.Map[types.NamespacedName, *gwapiv1.GatewayStatus]
	HTTPRouteStatuses    watchable.Map[types.NamespacedName, *gwapiv1.HTTPRouteStatus]
	GRPCRouteStatuses    watchable.Map[types.NamespacedName, *gwapiv1.GRPCRouteStatus]
	TLSRouteStatuses     watchable.Map[types.NamespacedName, *gwapiv1.TLSRouteStatus]
	TCPRouteStatuses     watchable.Map[types.NamespacedName, *gwapiv1.TCPRouteStatus]
	UDPRouteStatuses     watchable.Map[types.NamespacedName, *gwapiv1.UDPRouteStatus]
	ListenerSetStatuses  watchable.Map[types.NamespacedName, *gwapiv1.ListenerSetStatus]
}

func (s *GatewayAPIStatuses) Close() {
	s.GatewayClassStatuses.Close()
	s.GatewayStatuses.Close()
	s.HTTPRouteStatuses.Close()
	s.GRPCRouteStatuses.Close()
	s.TLSRouteStatuses.Close()
	s.TCPRouteStatuses.Close()
	s.UDPRouteStatuses.Close()
	s.ListenerSetStatuses.Close()
}

type NamespacedNameAndGVK struct {
	types.NamespacedName
	schema.GroupVersionKind
}

// PolicyStatuses contains policy related resources statuses
type PolicyStatuses struct {
	ClientTrafficPolicyStatuses  watchable.Map[types.NamespacedName, *gwapiv1.PolicyStatus]
	BackendTrafficPolicyStatuses watchable.Map[types.NamespacedName, *gwapiv1.PolicyStatus]
	EnvoyPatchPolicyStatuses     watchable.Map[types.NamespacedName, *gwapiv1.PolicyStatus]
	SecurityPolicyStatuses       watchable.Map[types.NamespacedName, *gwapiv1.PolicyStatus]
	BackendTLSPolicyStatuses     watchable.Map[types.NamespacedName, *gwapiv1.PolicyStatus]
	EnvoyExtensionPolicyStatuses watchable.Map[types.NamespacedName, *gwapiv1.PolicyStatus]
	ExtensionPolicyStatuses      watchable.Map[NamespacedNameAndGVK, *gwapiv1.PolicyStatus]

	EnvoyProxyStatuses watchable.Map[types.NamespacedName, *egv1a1.EnvoyProxyStatus]
}

// ExtensionStatuses contains statuses related to gw-api extension resources
type ExtensionStatuses struct {
	BackendStatuses watchable.Map[types.NamespacedName, *egv1a1.BackendStatus]
}

func (p *PolicyStatuses) Close() {
	p.ClientTrafficPolicyStatuses.Close()
	p.BackendTrafficPolicyStatuses.Close()
	p.SecurityPolicyStatuses.Close()
	p.EnvoyPatchPolicyStatuses.Close()
	p.BackendTLSPolicyStatuses.Close()
	p.EnvoyExtensionPolicyStatuses.Close()
	p.ExtensionPolicyStatuses.Close()
	p.EnvoyProxyStatuses.Close()
}

func (e *ExtensionStatuses) Close() {
	e.BackendStatuses.Close()
}

type XdsIRWithContext struct {
	XdsIR   *ir.Xds
	Context context.Context
	// StoredAt is when this value was Store()'d into the watchable map. Subscribers use it
	// to record how long the value sat buffered in the map's internal queue before being
	// dequeued; see RecordQueueWait.
	StoredAt time.Time
}

// ParentContext adds the trace context stashed on x to fallback without replacing
// fallback's cancellation and values.
func (x *XdsIRWithContext) ParentContext(fallback context.Context) context.Context {
	if x == nil || x.Context == nil {
		return fallback
	}
	spanContext := trace.SpanContextFromContext(x.Context)
	if !spanContext.IsValid() {
		return fallback
	}
	return trace.ContextWithSpanContext(fallback, spanContext)
}

// StoredAtTime returns the time x was stored in the watchable map, or the zero Time if x is nil.
func (x *XdsIRWithContext) StoredAtTime() time.Time {
	if x == nil {
		return time.Time{}
	}
	return x.StoredAt
}

// DeepCopy creates a new ControllerResourcesContext.
// The Context field is preserved (not deep copied) since contexts are meant to be passed around.
func (x *XdsIRWithContext) DeepCopy() *XdsIRWithContext {
	if x == nil {
		return nil
	}
	var xdsIRCopy *ir.Xds
	if x.XdsIR != nil {
		xdsIRCopy = x.XdsIR.DeepCopy()
	}
	return &XdsIRWithContext{
		XdsIR:    xdsIRCopy,
		Context:  x.Context,
		StoredAt: x.StoredAt,
	}
}

func (x *XdsIRWithContext) Equal(other *XdsIRWithContext) bool {
	if x == nil && other == nil {
		return true
	}
	if x == nil || other == nil {
		return false
	}
	if x.XdsIR == nil && other.XdsIR == nil {
		return true
	}
	if x.XdsIR == nil || other.XdsIR == nil {
		return false
	}

	return reflect.DeepEqual(x.XdsIR, other.XdsIR)
}

// XdsIR message
type XdsIR struct {
	watchable.Map[string, *XdsIRWithContext]
}

// InfraIRWithContext wraps ir.Infra with trace context for propagating spans across the
// InfraIR watchable-map boundary, the same way XdsIRWithContext does for XdsIR.
type InfraIRWithContext struct {
	Infra   *ir.Infra
	Context context.Context
	// StoredAt is when this value was Store()'d into the watchable map. Subscribers use it to
	// record how long the value sat buffered in the map's internal queue before being
	// dequeued; see RecordQueueWait.
	StoredAt time.Time
}

// ParentContext adds the trace context stashed on x to fallback without replacing
// fallback's cancellation and values.
func (x *InfraIRWithContext) ParentContext(fallback context.Context) context.Context {
	if x == nil || x.Context == nil {
		return fallback
	}
	spanContext := trace.SpanContextFromContext(x.Context)
	if !spanContext.IsValid() {
		return fallback
	}
	return trace.ContextWithSpanContext(fallback, spanContext)
}

// StoredAtTime returns the time x was stored in the watchable map, or the zero Time if x is nil.
func (x *InfraIRWithContext) StoredAtTime() time.Time {
	if x == nil {
		return time.Time{}
	}
	return x.StoredAt
}

// DeepCopy creates a new InfraIRWithContext.
// The Context field is preserved (not deep copied) since contexts are meant to be passed around.
func (x *InfraIRWithContext) DeepCopy() *InfraIRWithContext {
	if x == nil {
		return nil
	}
	var infraCopy *ir.Infra
	if x.Infra != nil {
		infraCopy = x.Infra.DeepCopy()
	}
	return &InfraIRWithContext{
		Infra:    infraCopy,
		Context:  x.Context,
		StoredAt: x.StoredAt,
	}
}

func (x *InfraIRWithContext) Equal(other *InfraIRWithContext) bool {
	if x == nil && other == nil {
		return true
	}
	if x == nil || other == nil {
		return false
	}
	if x.Infra == nil && other.Infra == nil {
		return true
	}
	if x.Infra == nil || other.Infra == nil {
		return false
	}

	return reflect.DeepEqual(x.Infra, other.Infra)
}

// EndpointUpdate carries the current set of EndpointSlices for one backend
// (Service or ServiceImport). It is published by the provider whenever the
// backend's endpoints change and consumed by the xDS runner's endpoint fast
// path.
type EndpointUpdate struct {
	// Kind is the backend kind: Service or ServiceImport.
	Kind string
	// Namespace of the backend.
	Namespace string
	// Name of the backend.
	Name string
	// EndpointSlices is the full current set of EndpointSlices for the
	// backend, so each update is self-contained under coalescing.
	EndpointSlices []*discoveryv1.EndpointSlice
}

// BackendKey returns the EndpointUpdates map key identifying a backend. It is
// the single definition of that format: the provider keys published updates by
// it, the xDS runner indexes endpoint contexts by it, and the reconcile prunes
// retained updates by it — they must agree.
func BackendKey(kind, namespace, name string) string {
	return kind + "/" + namespace + "/" + name
}

// Key returns the watchable map key for this update's backend.
func (e *EndpointUpdate) Key() string {
	return BackendKey(e.Kind, e.Namespace, e.Name)
}

// DeepCopy creates a new EndpointUpdate.
func (e *EndpointUpdate) DeepCopy() *EndpointUpdate {
	if e == nil {
		return nil
	}
	out := &EndpointUpdate{
		Kind:      e.Kind,
		Namespace: e.Namespace,
		Name:      e.Name,
	}
	if e.EndpointSlices != nil {
		out.EndpointSlices = make([]*discoveryv1.EndpointSlice, len(e.EndpointSlices))
		for i, s := range e.EndpointSlices {
			out.EndpointSlices[i] = s.DeepCopy()
		}
	}
	return out
}

// Equal compares two EndpointUpdates.
func (e *EndpointUpdate) Equal(other *EndpointUpdate) bool {
	if e == nil || other == nil {
		return e == other
	}
	return e.Kind == other.Kind &&
		e.Namespace == other.Namespace &&
		e.Name == other.Name &&
		reflect.DeepEqual(e.EndpointSlices, other.EndpointSlices)
}

// InfraIR message
type InfraIR struct {
	watchable.Map[string, *InfraIRWithContext]
}

type MessageName string

const (
	// XDSIRMessageName is a message containing xds-ir translated from provider-resources
	XDSIRMessageName MessageName = "xds-ir"
	// EndpointSlicesMessageName is a message containing a backend's EndpointSlices,
	// published by the provider for the endpoint fast path
	EndpointSlicesMessageName MessageName = "endpointslices"
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
	// ListenerSetStatusMessageName is a message containing updates to ListenerSet status
	ListenerSetStatusMessageName MessageName = "listenerset-status"
	// GatewayClassStatusMessageName is a message containing updates to GatewayClass status
	GatewayClassStatusMessageName MessageName = "gatewayclass-status"
	// EnvoyProxyStatusMessageName is a message containing updates to EnvoyProxy status
	EnvoyProxyStatusMessageName MessageName = "envoyproxy-status"
)
