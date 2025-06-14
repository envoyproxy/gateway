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
	// KindSecurityPolicy is the name of the SecurityPolicy kind.
	KindSecurityPolicy = "SecurityPolicy"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=envoy-gateway,shortName=sp
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// SecurityPolicy allows the user to configure various security settings for a
// Gateway.
type SecurityPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of SecurityPolicy.
	Spec SecurityPolicySpec `json:"spec"`

	// Status defines the current status of SecurityPolicy.
	Status gwapiv1a2.PolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="(has(self.targetRef) && !has(self.targetRefs)) || (!has(self.targetRef) && has(self.targetRefs)) || (has(self.targetSelectors) && self.targetSelectors.size() > 0) ", message="either targetRef or targetRefs must be used"
//
// +kubebuilder:validation:XValidation:rule="has(self.targetRef) ? self.targetRef.group == 'gateway.networking.k8s.io' : true", message="this policy can only have a targetRef.group of gateway.networking.k8s.io"
// +kubebuilder:validation:XValidation:rule="has(self.targetRef) ? self.targetRef.kind in ['Gateway', 'HTTPRoute', 'GRPCRoute'] : true", message="this policy can only have a targetRef.kind of Gateway/HTTPRoute/GRPCRoute"
// +kubebuilder:validation:XValidation:rule="has(self.targetRef) ? self.targetRef.kind == 'Gateway' || !has(self.targetRef.sectionName) : true",message="this policy supports the sectionName field only for kind Gateway"
// +kubebuilder:validation:XValidation:rule="has(self.targetRefs) ? self.targetRefs.all(ref, ref.group == 'gateway.networking.k8s.io') : true ", message="this policy can only have a targetRefs[*].group of gateway.networking.k8s.io"
// +kubebuilder:validation:XValidation:rule="has(self.targetRefs) ? self.targetRefs.all(ref, ref.kind in ['Gateway', 'HTTPRoute', 'GRPCRoute']) : true ", message="this policy can only have a targetRefs[*].kind of Gateway/HTTPRoute/GRPCRoute"
// +kubebuilder:validation:XValidation:rule="has(self.targetRefs) ? self.targetRefs.all(ref, ref.kind == 'Gateway' || !has(ref.sectionName)) : true",message="this policy supports the sectionName field only for kind Gateway"
// +kubebuilder:validation:XValidation:rule="(has(self.authorization) && has(self.authorization.rules) && self.authorization.rules.exists(r, has(r.principal.jwt))) ? has(self.jwt) : true", message="if authorization.rules.principal.jwt is used, jwt must be defined"
//
// SecurityPolicySpec defines the desired state of SecurityPolicy.
type SecurityPolicySpec struct {
	PolicyTargetReferences `json:",inline"`

	// APIKeyAuth defines the configuration for the API Key Authentication.
	//
	// +optional
	APIKeyAuth *APIKeyAuth `json:"apiKeyAuth,omitempty"`

	// CORS defines the configuration for Cross-Origin Resource Sharing (CORS).
	//
	// +optional
	CORS *CORS `json:"cors,omitempty"`

	// BasicAuth defines the configuration for the HTTP Basic Authentication.
	//
	// +optional
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`

	// JWT defines the configuration for JSON Web Token (JWT) authentication.
	//
	// +optional
	JWT *JWT `json:"jwt,omitempty"`

	// OIDC defines the configuration for the OpenID Connect (OIDC) authentication.
	//
	// +optional
	OIDC *OIDC `json:"oidc,omitempty"`

	// ExtAuth defines the configuration for External Authorization.
	//
	// +optional
	ExtAuth *ExtAuth `json:"extAuth,omitempty"`

	// Authorization defines the authorization configuration.
	//
	// +optional
	Authorization *Authorization `json:"authorization,omitempty"`
}

//+kubebuilder:object:root=true

// SecurityPolicyList contains a list of SecurityPolicy resources.
type SecurityPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecurityPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecurityPolicy{}, &SecurityPolicyList{})
}
