// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	// KindBackendTrafficPolicy is the name of the BackendTrafficPolicy kind.
	KindBackendTrafficPolicy = "BackendTrafficPolicy"
)

// BackendTrafficPolicy allows the user to configure the behavior of the connection
// between the Envoy Proxy listener and the backend service.
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=envoy-gateway,shortName=btp
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type BackendTrafficPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of BackendTrafficPolicy.
	Spec BackendTrafficPolicySpec `json:"spec"`

	// status defines the current status of BackendTrafficPolicy.
	Status gwapiv1.PolicyStatus `json:"status,omitempty"`
}

// BackendTrafficPolicySpec defines the desired state of BackendTrafficPolicy.
//
// +kubebuilder:validation:XValidation:rule="(has(self.targetRef) && !has(self.targetRefs)) || (!has(self.targetRef) && has(self.targetRefs)) || (has(self.targetSelectors) && self.targetSelectors.size() > 0) ", message="either targetRef or targetRefs must be used"
// +kubebuilder:validation:XValidation:rule="has(self.targetRef) ? self.targetRef.group == 'gateway.networking.k8s.io' : true ", message="this policy can only have a targetRef.group of gateway.networking.k8s.io"
// +kubebuilder:validation:XValidation:rule="has(self.targetRef) ? self.targetRef.kind in ['Gateway', 'HTTPRoute', 'GRPCRoute', 'UDPRoute', 'TCPRoute', 'TLSRoute'] : true", message="this policy can only have a targetRef.kind of Gateway/HTTPRoute/GRPCRoute/TCPRoute/UDPRoute/TLSRoute"
// +kubebuilder:validation:XValidation:rule="has(self.targetRefs) ? self.targetRefs.all(ref, ref.group == 'gateway.networking.k8s.io') : true ", message="this policy can only have a targetRefs[*].group of gateway.networking.k8s.io"
// +kubebuilder:validation:XValidation:rule="has(self.targetRefs) ? self.targetRefs.all(ref, ref.kind in ['Gateway', 'HTTPRoute', 'GRPCRoute', 'UDPRoute', 'TCPRoute', 'TLSRoute']) : true ", message="this policy can only have a targetRefs[*].kind of Gateway/HTTPRoute/GRPCRoute/TCPRoute/UDPRoute/TLSRoute"
// +kubebuilder:validation:XValidation:rule="!has(self.compression) || !has(self.compressor)", message="either compression or compressor can be set, not both"
// +kubebuilder:validation:XValidation:rule="!has(self.requestBuffer) || self.requestBuffer.mode == 'LimitOnly' || !has(self.httpUpgrade) || self.httpUpgrade.size() == 0", message="requestBuffer with mode FullBuffer cannot be used together with httpUpgrade"
// +kubebuilder:validation:XValidation:rule="!has(self.admissionControl) || ((!has(self.targetRef) || self.targetRef.kind in ['Gateway', 'HTTPRoute', 'GRPCRoute']) && (!has(self.targetRefs) || self.targetRefs.all(ref, ref.kind in ['Gateway', 'HTTPRoute', 'GRPCRoute'])) && (!has(self.targetSelectors) || self.targetSelectors.all(sel, sel.kind in ['Gateway', 'HTTPRoute', 'GRPCRoute'])))", message="admissionControl can only be used with HTTPRoute, GRPCRoute, or Gateway targets"
type BackendTrafficPolicySpec struct {
	PolicyTargetReferences `json:",inline"`
	ClusterSettings        `json:",inline"`

	// MergeType determines how this configuration is merged with existing BackendTrafficPolicy
	// configurations targeting a parent resource. When set, this configuration will be merged
	// into a parent BackendTrafficPolicy (i.e. the one targeting a Gateway or Listener).
	// This field cannot be set when targeting a parent resource (Gateway).
	// If unset, no merging occurs, and only the most specific configuration takes effect.
	//
	// +kubebuilder:validation:XValidation:rule="self != 'Replace'",message="Replace is not a valid MergeType for BackendTrafficPolicySpec"
	// +optional
	MergeType *MergeType `json:"mergeType,omitempty"`

	// RateLimit allows the user to limit the number of incoming requests
	// to a predefined value based on attributes within the traffic flow.
	// +optional
	RateLimit *RateLimitSpec `json:"rateLimit,omitempty"`

	// BandwidthLimit allows the user to limit the bandwidth of traffic
	// sent to and received from the backend.
	// +optional
	BandwidthLimit *BandwidthLimitSpec `json:"bandwidthLimit,omitempty"`

	// FaultInjection defines the fault injection policy to be applied. This configuration can be used to
	// inject delays and abort requests to mimic failure scenarios such as service failures and overloads
	// +optional
	FaultInjection *FaultInjection `json:"faultInjection,omitempty"`

	// AdmissionControl defines the admission control policy to be applied. This configuration
	// probabilistically rejects requests based on the success rate of previous requests in a
	// configurable sliding time window.
	// +optional
	AdmissionControl *AdmissionControl `json:"admissionControl,omitempty"`

	// UseClientProtocol configures Envoy to prefer sending requests to backends using
	// the same HTTP protocol that the incoming request used. Defaults to false, which means
	// that Envoy will use the protocol indicated by the attached BackendRef.
	//
	// +optional
	UseClientProtocol *bool `json:"useClientProtocol,omitempty"`

	// The compression config for the http streams.
	//
	// Deprecated: Use Compressor instead.
	//
	// +patchMergeKey=type
	// +patchStrategy=merge
	//
	// +optional
	Compression []*Compression `json:"compression,omitempty" patchMergeKey:"type" patchStrategy:"merge"`

	// The compressor config for the http streams.
	// This provides more granular control over compression configuration.
	// Order matters: The first compressor in the list is preferred when q-values in Accept-Encoding are equal.
	//
	// +patchMergeKey=type
	// +patchStrategy=merge
	//
	// +optional
	Compressor []*Compression `json:"compressor,omitempty" patchMergeKey:"type" patchStrategy:"merge"`

	// ResponseOverride defines the configuration to override specific responses with a custom one.
	// If multiple configurations are specified, the first one to match wins.
	//
	// +optional
	ResponseOverride []*ResponseOverride `json:"responseOverride,omitempty"`
	// HTTPUpgrade defines the configuration for HTTP protocol upgrades.
	// If not specified, the default upgrade configuration (websocket) will be used.
	// However, if requestBuffer is configured with mode FullBuffer, the default
	// upgrade configuration will be ignored.
	//
	// +patchMergeKey=type
	// +patchStrategy=merge
	//
	// +optional
	HTTPUpgrade []*ProtocolUpgradeConfig `json:"httpUpgrade,omitempty" patchMergeKey:"type" patchStrategy:"merge"`

	// RequestBuffer configures how much of a request body Envoy is allowed to buffer for a route,
	// and whether the gateway fully buffers each request before forwarding it upstream.
	//
	// A request whose buffered body exceeds the configured limit is rejected with HTTP 413 Content Too
	// Large. How much of a request is buffered, and therefore whether the limit acts as a maximum request
	// body size, depends on the mode: see the mode field.
	//
	// Buffering increases memory usage for Envoy that should be accounted for in your deployment settings.
	//
	// +optional
	RequestBuffer *RequestBuffer `json:"requestBuffer,omitempty"`

	// Telemetry configures the telemetry settings for the policy target (Gateway or xRoute).
	// This will override the telemetry settings in the EnvoyProxy resource.
	//
	// +optional
	Telemetry *BackendTelemetry `json:"telemetry,omitempty"`

	// RoutingType can be set to "Service" to use the Service Cluster IP for routing to the backend,
	// or it can be set to "Endpoint" to use Endpoint routing.
	// When specified, this overrides the EnvoyProxy-level setting for the relevant targetRefs.
	// If not specified, the EnvoyProxy-level setting is used.
	//
	// +optional
	RoutingType *RoutingType `json:"routingType,omitempty"`
}

type BackendTelemetry struct {
	// Tracing configures the tracing settings for the backend or HTTPRoute.
	//
	// This takes precedence over EnvoyProxy tracing when set.
	//
	// +optional
	Tracing *Tracing `json:"tracing,omitempty"`
	// Metrics defines metrics configuration for the backend or Route.
	//
	// +optional
	Metrics *BackendMetrics `json:"metrics,omitempty"`
}

type BackendMetrics struct {
	// RouteStatName defines the value of the Route stat_prefix, determining how the route stats are named.
	// For more details, see envoy docs: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#config-route-v3-route
	// The supported operators for this pattern are:
	// %ROUTE_NAME%: name of Gateway API xRoute resource
	// %ROUTE_NAMESPACE%: namespace of Gateway API xRoute resource
	// %ROUTE_KIND%: kind of Gateway API xRoute resource
	// Example: %ROUTE_KIND%/%ROUTE_NAMESPACE%/%ROUTE_NAME% => httproute/my-ns/my-route
	// Disabled by default.
	//
	// +optional
	RouteStatName *string `json:"routeStatName,omitempty"`
}

// ProtocolUpgradeConfig specifies the configuration for protocol upgrades.
//
// +kubebuilder:validation:XValidation:rule="!has(self.connect) || self.type == 'CONNECT'",message="The connect configuration is only allowed when the type is CONNECT."
type ProtocolUpgradeConfig struct {
	// Type is the case-insensitive type of protocol upgrade.
	// e.g. `websocket`, `CONNECT`, `spdy/3.1` etc.
	//
	// +kubebuilder:validation:Required
	Type string `json:"type"`
	// Connect specifies the configuration for the CONNECT config.
	// This is allowed only when type is CONNECT.
	//
	// +optional
	Connect *ConnectConfig `json:"connect,omitempty"`
}

type ConnectConfig struct {
	// Terminate the CONNECT request, and forwards the payload as raw TCP data.
	//
	// +optional
	Terminate *bool `json:"terminate,omitempty"`
}

type RequestBuffer struct {
	// Limit specifies the maximum size in bytes that Envoy may buffer for an incoming request body.
	// If a request's buffered body exceeds this limit, the request is rejected with HTTP 413 Content
	// Too Large.
	//
	// In FullBuffer mode the entire body is always buffered, so this acts as a maximum request body size.
	// In LimitOnly mode only what a filter later in the chain actually buffers counts against the limit,
	// so a streamed request that nothing buffers can exceed it and still be forwarded upstream.
	//
	// Accepts values in resource.Quantity format (e.g., "10Mi", "500Ki").
	//
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern="^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$"
	Limit resource.Quantity `json:"limit,omitempty"`

	// Mode determines how Limit is enforced. Defaults to FullBuffer.
	//
	// FullBuffer makes the gateway buffer and fully receive each request from a client before continuing
	// to send the request upstream to the backends. This can be helpful to shield your backend servers
	// from slow clients. Envoy sets the request buffer limit for the stream to Limit, so there is no need
	// to also raise the connection buffer size for it to take effect.
	// This can have a negative performance impact so should only be used when necessary. It is also
	// incompatible with streaming APIs and protocol upgrades such as gRPC streaming and WebSocket: do not
	// use this mode on routes that need those protocols, because requests can hang instead of being
	// forwarded upstream.
	//
	// LimitOnly only raises Envoy's request body buffer limit for the route, without enabling full request
	// buffering. Use this mode when a filter later in the chain (ext_proc, Lua, WASM, ...) buffers the
	// request body and the default buffer limit is too small. Unlike FullBuffer, this mode is compatible
	// with streaming APIs and protocol upgrades, because Envoy does not wait for the whole request body
	// before forwarding it upstream.
	//
	// In both modes, the request body buffer limit is independent of the connection buffer limits
	// configured by ClientTrafficPolicy and BackendTrafficPolicy. The connection buffer limits control
	// downstream and upstream connection read/write buffering and back pressure, while Limit sets the
	// maximum size of an individual request body that Envoy may buffer for HTTP processing.
	//
	// For HTTP/1, connection buffer limits can still affect how request body data is read and buffered
	// before or while the request body limit is enforced. Configure connection buffer limits large
	// enough for the request body buffering you expect to allow on those connections.
	//
	// For HTTP/2, initial stream and connection window sizes control HTTP/2 flow control: how much data
	// can be in flight before back pressure applies. They affect upload throughput and buffering
	// behavior, but they are not a replacement for this request body size limit.
	//
	// +kubebuilder:default=FullBuffer
	// +optional
	Mode *RequestBufferMode `json:"mode,omitempty"`
}

// RequestBufferMode determines how RequestBuffer.Limit is applied.
//
// +kubebuilder:validation:Enum=FullBuffer;LimitOnly
type RequestBufferMode string

const (
	// RequestBufferModeFullBuffer buffers the entire request body in the gateway before forwarding the
	// request upstream. It is incompatible with streaming APIs and protocol upgrades.
	RequestBufferModeFullBuffer RequestBufferMode = "FullBuffer"

	// RequestBufferModeLimitOnly only raises the request body buffer limit for the route, without
	// enabling full request buffering.
	RequestBufferModeLimitOnly RequestBufferMode = "LimitOnly"
)

// BackendTrafficPolicyList contains a list of BackendTrafficPolicy resources.
//
// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type BackendTrafficPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackendTrafficPolicy `json:"items"`
}

func init() {
	localSchemeBuilder.Register(&BackendTrafficPolicy{}, &BackendTrafficPolicyList{})
}
