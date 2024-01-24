// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

const (
	// KindClientTrafficPolicy is the name of the ClientTrafficPolicy kind.
	KindClientTrafficPolicy = "ClientTrafficPolicy"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=ctp
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[?(@.type=="Accepted")].reason`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// ClientTrafficPolicy allows the user to configure the behavior of the connection
// between the downstream client and Envoy Proxy listener.
type ClientTrafficPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of ClientTrafficPolicy.
	Spec ClientTrafficPolicySpec `json:"spec"`

	// Status defines the current status of ClientTrafficPolicy.
	Status ClientTrafficPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="has(self.http3) && has(self.tls) && has(self.tls.alpnProtocols) ? self.tls.alpnProtocols.size() == 0 : true",message="alpn protocols can't be set if HTTP/3 is enabled"
// ClientTrafficPolicySpec defines the desired state of ClientTrafficPolicy.
type ClientTrafficPolicySpec struct {
	// +kubebuilder:validation:XValidation:rule="self.group == 'gateway.networking.k8s.io'", message="this policy can only have a targetRef.group of gateway.networking.k8s.io"
	// +kubebuilder:validation:XValidation:rule="self.kind == 'Gateway'", message="this policy can only have a targetRef.kind of Gateway"
	// +kubebuilder:validation:XValidation:rule="!has(self.sectionName)",message="this policy does not yet support the sectionName field"
	//
	// TargetRef is the name of the Gateway resource this policy
	// is being attached to.
	// This Policy and the TargetRef MUST be in the same namespace
	// for this Policy to have effect and be applied to the Gateway.
	// TargetRef
	TargetRef gwapiv1a2.PolicyTargetReferenceWithSectionName `json:"targetRef"`
	// TcpKeepalive settings associated with the downstream client connection.
	// If defined, sets SO_KEEPALIVE on the listener socket to enable TCP Keepalives.
	// Disabled by default.
	//
	// +optional
	TCPKeepalive *TCPKeepalive `json:"tcpKeepalive,omitempty"`
	// SuppressEnvoyHeaders configures the Envoy Router filter to suppress the "x-envoy-'
	// headers from both requests and responses.
	// By default these headers are added to both requests and responses.
	//
	// +optional
	SuppressEnvoyHeaders *bool `json:"suppressEnvoyHeaders,omitempty"`
	// EnableProxyProtocol interprets the ProxyProtocol header and adds the
	// Client Address into the X-Forwarded-For header.
	// Note Proxy Protocol must be present when this field is set, else the connection
	// is closed.
	//
	// +optional
	EnableProxyProtocol *bool `json:"enableProxyProtocol,omitempty"`
	// HTTP3 provides HTTP/3 configuration on the listener.
	//
	// +optional
	HTTP3 *HTTP3Settings `json:"http3,omitempty"`
	// TLS settings configure TLS termination settings with the downstream client.
	//
	// +optional
	TLS *TLSSettings `json:"tls,omitempty"`
	// Path enables managing how the incoming path set by clients can be normalized.
	//
	// +optional
	Path *PathSettings `json:"path,omitempty"`
	// HTTP1 provides HTTP/1 configuration on the listener.
	//
	// +optional
	HTTP1 *HTTP1Settings `json:"http1,omitempty"`
}

// HTTP3Settings provides HTTP/3 configuration on the listener.
type HTTP3Settings struct {
}

// HTTP1Settings provides HTTP/1 configuration on the listener.
type HTTP1Settings struct {
	// EnableTrailers defines if HTTP/1 trailers should be proxied by Envoy.
	// +optional
	EnableTrailers *bool `json:"enableTrailers,omitempty"`
}

// ClientTrafficPolicyStatus defines the state of ClientTrafficPolicy
type ClientTrafficPolicyStatus struct {
	// Conditions describe the current conditions of the ClientTrafficPolicy.
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

const (
	// PolicyConditionOverridden indicates whether the policy has
	// completely attached to all the sections within the target or not.
	//
	// Possible reasons for this condition to be True are:
	//
	// * "Overridden"
	//
	PolicyConditionOverridden gwapiv1a2.PolicyConditionType = "Overridden"

	// PolicyReasonOverridden is used with the "Overridden" condition when the policy
	// has been overridden by another policy targeting a section within the same target.
	PolicyReasonOverridden gwapiv1a2.PolicyConditionReason = "Overridden"
)

//+kubebuilder:object:root=true

// ClientTrafficPolicyList contains a list of ClientTrafficPolicy resources.
type ClientTrafficPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClientTrafficPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClientTrafficPolicy{}, &ClientTrafficPolicyList{})
}
