// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
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
type BackendTrafficPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of BackendTrafficPolicy.
	Spec BackendTrafficPolicySpec `json:"spec"`

	// status defines the current status of BackendTrafficPolicy.
	Status gwapiv1a2.PolicyStatus `json:"status,omitempty"`
}

// BackendTrafficPolicySpec defines the desired state of BackendTrafficPolicy.
//
// +kubebuilder:validation:XValidation:rule="(has(self.targetRef) && !has(self.targetRefs)) || (!has(self.targetRef) && has(self.targetRefs)) || (has(self.targetSelectors) && self.targetSelectors.size() > 0) ", message="either targetRef or targetRefs must be used"
// +kubebuilder:validation:XValidation:rule="has(self.targetRef) ? self.targetRef.group == 'gateway.networking.k8s.io' : true ", message="this policy can only have a targetRef.group of gateway.networking.k8s.io"
// +kubebuilder:validation:XValidation:rule="has(self.targetRef) ? self.targetRef.kind in ['Gateway', 'HTTPRoute', 'GRPCRoute', 'UDPRoute', 'TCPRoute', 'TLSRoute'] : true", message="this policy can only have a targetRef.kind of Gateway/HTTPRoute/GRPCRoute/TCPRoute/UDPRoute/TLSRoute"
// +kubebuilder:validation:XValidation:rule="has(self.targetRef) ? !has(self.targetRef.sectionName) : true",message="this policy does not yet support the sectionName field"
// +kubebuilder:validation:XValidation:rule="has(self.targetRefs) ? self.targetRefs.all(ref, ref.group == 'gateway.networking.k8s.io') : true ", message="this policy can only have a targetRefs[*].group of gateway.networking.k8s.io"
// +kubebuilder:validation:XValidation:rule="has(self.targetRefs) ? self.targetRefs.all(ref, ref.kind in ['Gateway', 'HTTPRoute', 'GRPCRoute', 'UDPRoute', 'TCPRoute', 'TLSRoute']) : true ", message="this policy can only have a targetRefs[*].kind of Gateway/HTTPRoute/GRPCRoute/TCPRoute/UDPRoute/TLSRoute"
// +kubebuilder:validation:XValidation:rule="has(self.targetRefs) ? self.targetRefs.all(ref, !has(ref.sectionName)) : true",message="this policy does not yet support the sectionName field"
type BackendTrafficPolicySpec struct {
	PolicyTargetReferences `json:",inline"`
	ClusterSettings        `json:",inline"`

	// MergeType determines how this configuration is merged with existing BackendTrafficPolicy
	// configurations targeting a parent resource. When set, this configuration will be merged
	// into a parent BackendTrafficPolicy (i.e. the one targeting a Gateway or Listener).
	// This field cannot be set when targeting a parent resource (Gateway).
	// If unset, no merging occurs, and only the most specific configuration takes effect.
	// +optional
	MergeType *MergeType `json:"mergeType,omitempty"`

	// RateLimit allows the user to limit the number of incoming requests
	// to a predefined value based on attributes within the traffic flow.
	// +optional
	RateLimit *RateLimitSpec `json:"rateLimit,omitempty"`

	// FaultInjection defines the fault injection policy to be applied. This configuration can be used to
	// inject delays and abort requests to mimic failure scenarios such as service failures and overloads
	// +optional
	FaultInjection *FaultInjection `json:"faultInjection,omitempty"`

	// UseClientProtocol configures Envoy to prefer sending requests to backends using
	// the same HTTP protocol that the incoming request used. Defaults to false, which means
	// that Envoy will use the protocol indicated by the attached BackendRef.
	//
	// +optional
	UseClientProtocol *bool `json:"useClientProtocol,omitempty"`

	// The compression config for the http streams.
	//
	// +optional
	Compression []*Compression `json:"compression,omitempty"`

	// ResponseOverride defines the configuration to override specific responses with a custom one.
	// If multiple configurations are specified, the first one to match wins.
	//
	// +optional
	ResponseOverride []*ResponseOverride `json:"responseOverride,omitempty"`
	// HTTPUpgrade defines the configuration for HTTP protocol upgrades.
	// If not specified, the default upgrade configuration(websocket) will be used.
	//
	// +patchMergeKey=type
	// +patchStrategy=merge
	//
	// +optional
	HTTPUpgrade []*ProtocolUpgradeConfig `json:"httpUpgrade,omitempty" patchMergeKey:"type" patchStrategy:"merge"`

	// RequestBuffer allows the gateway to buffer and fully receive each request from a client before continuing to send the request
	// upstream to the backends. This can be helpful to shield your backend servers from slow clients, and also to enforce a maximum size per request
	// as any requests larger than the buffer size will be rejected.
	//
	// This can have a negative performance impact so should only be enabled when necessary.
	//
	// When enabling this option, you should also configure your connection buffer size to account for these request buffers. There will also be an
	// increase in memory usage for Envoy that should be accounted for in your deployment settings.
	//
	// +notImplementedHide
	// +optional
	RequestBuffer *RequestBuffer `json:"requestBuffer,omitempty"`
	// Telemetry configures the telemetry settings for the policy target (Gateway or xRoute).
	// This will override the telemetry settings in the EnvoyProxy resource.
	//
	// +optional
	Telemetry *BackendTelemetry `json:"telemetry,omitempty"`
}

type BackendTelemetry struct {
	// Tracing configures the tracing settings for the backend or HTTPRoute.
	//
	// +optional
	Tracing *Tracing `json:"tracing,omitempty"`
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
	// Limit specifies the maximum allowed size in bytes for each incoming request buffer.
	// If exceeded, the request will be rejected with HTTP 413 Content Too Large.
	//
	// Accepts values in resource.Quantity format (e.g., "10Mi", "500Ki").
	//
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern="^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$"
	// +notImplementedHide
	Limit resource.Quantity `json:"limit,omitempty"`
}

// BackendTrafficPolicyList contains a list of BackendTrafficPolicy resources.
//
// +kubebuilder:object:root=true
type BackendTrafficPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackendTrafficPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BackendTrafficPolicy{}, &BackendTrafficPolicyList{})
}
