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

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=envoy-gateway,shortName=btp
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// BackendTrafficPolicy allows the user to configure the behavior of the connection
// between the Envoy Proxy listener and the backend service.
type BackendTrafficPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of BackendTrafficPolicy.
	Spec BackendTrafficPolicySpec `json:"spec"`

	// status defines the current status of BackendTrafficPolicy.
	Status gwapiv1a2.PolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="(has(self.targetRef) && !has(self.targetRefs)) || (!has(self.targetRef) && has(self.targetRefs)) || (has(self.targetSelectors) && self.targetSelectors.size() > 0) ", message="either targetRef or targetRefs must be used"
//
// +kubebuilder:validation:XValidation:rule="has(self.targetRef) ? self.targetRef.group == 'gateway.networking.k8s.io' : true ", message="this policy can only have a targetRef.group of gateway.networking.k8s.io"
// +kubebuilder:validation:XValidation:rule="has(self.targetRef) ? self.targetRef.kind in ['Gateway', 'HTTPRoute', 'GRPCRoute', 'UDPRoute', 'TCPRoute', 'TLSRoute'] : true", message="this policy can only have a targetRef.kind of Gateway/HTTPRoute/GRPCRoute/TCPRoute/UDPRoute/TLSRoute"
// +kubebuilder:validation:XValidation:rule="has(self.targetRef) ? !has(self.targetRef.sectionName) : true",message="this policy does not yet support the sectionName field"
// +kubebuilder:validation:XValidation:rule="has(self.targetRefs) ? self.targetRefs.all(ref, ref.group == 'gateway.networking.k8s.io') : true ", message="this policy can only have a targetRefs[*].group of gateway.networking.k8s.io"
// +kubebuilder:validation:XValidation:rule="has(self.targetRefs) ? self.targetRefs.all(ref, ref.kind in ['Gateway', 'HTTPRoute', 'GRPCRoute', 'UDPRoute', 'TCPRoute', 'TLSRoute']) : true ", message="this policy can only have a targetRefs[*].kind of Gateway/HTTPRoute/GRPCRoute/TCPRoute/UDPRoute/TLSRoute"
// +kubebuilder:validation:XValidation:rule="has(self.targetRefs) ? self.targetRefs.all(ref, !has(ref.sectionName)) : true",message="this policy does not yet support the sectionName field"
//
// BackendTrafficPolicySpec defines the desired state of BackendTrafficPolicy.
type BackendTrafficPolicySpec struct {
	PolicyTargetReferences `json:",inline"`

	// RateLimit allows the user to limit the number of incoming requests
	// to a predefined value based on attributes within the traffic flow.
	// +optional
	RateLimit *RateLimitSpec `json:"rateLimit,omitempty"`

	// LoadBalancer policy to apply when routing traffic from the gateway to
	// the backend endpoints
	// +optional
	LoadBalancer *LoadBalancer `json:"loadBalancer,omitempty"`

	// ProxyProtocol enables the Proxy Protocol when communicating with the backend.
	// +optional
	ProxyProtocol *ProxyProtocol `json:"proxyProtocol,omitempty"`

	// TcpKeepalive settings associated with the upstream client connection.
	// Disabled by default.
	//
	// +optional
	TCPKeepalive *TCPKeepalive `json:"tcpKeepalive,omitempty"`

	// HealthCheck allows gateway to perform active health checking on backends.
	//
	// +optional
	HealthCheck *HealthCheck `json:"healthCheck,omitempty"`

	// FaultInjection defines the fault injection policy to be applied. This configuration can be used to
	// inject delays and abort requests to mimic failure scenarios such as service failures and overloads
	// +optional
	FaultInjection *FaultInjection `json:"faultInjection,omitempty"`

	// Circuit Breaker settings for the upstream connections and requests.
	// If not set, circuit breakers will be enabled with the default thresholds
	//
	// +optional
	CircuitBreaker *CircuitBreaker `json:"circuitBreaker,omitempty"`

	// Retry provides more advanced usage, allowing users to customize the number of retries, retry fallback strategy, and retry triggering conditions.
	// If not set, retry will be disabled.
	// +optional
	Retry *Retry `json:"retry,omitempty"`

	// UseClientProtocol configures Envoy to prefer sending requests to backends using
	// the same HTTP protocol that the incoming request used. Defaults to false, which means
	// that Envoy will use the protocol indicated by the attached BackendRef.
	//
	// +optional
	UseClientProtocol *bool `json:"useClientProtocol,omitempty"`

	// Timeout settings for the backend connections.
	//
	// +optional
	Timeout *Timeout `json:"timeout,omitempty"`

	// The compression config for the http streams.
	//
	// +optional
	// +notImplementedHide
	Compression []*Compression `json:"compression,omitempty"`

	// Connection includes backend connection settings.
	//
	// +optional
	Connection *BackendConnection `json:"connection,omitempty"`
}

// +kubebuilder:object:root=true

// BackendTrafficPolicyList contains a list of BackendTrafficPolicy resources.
type BackendTrafficPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackendTrafficPolicy `json:"items"`
}

// BackendTrafficPolicyConnection allows users to configure connection-level settings of backend
type BackendTrafficPolicyConnection struct {
	// BufferLimit Soft limit on size of the cluster’s connections read and write buffers.
	// If unspecified, an implementation defined default is applied (32768 bytes).
	// For example, 20Mi, 1Gi, 256Ki etc.
	// Note: that when the suffix is not provided, the value is interpreted as bytes.
	//
	// +kubebuilder:validation:XValidation:rule="type(self) == string ? self.matches(r\"^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$\") : type(self) == int",message="BufferLimit must be of the format \"^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$\""
	// +optional
	BufferLimit *resource.Quantity `json:"bufferLimit,omitempty"`
}

func init() {
	SchemeBuilder.Register(&BackendTrafficPolicy{}, &BackendTrafficPolicyList{})
}
