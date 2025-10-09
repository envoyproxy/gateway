// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type CustomBackendMtlsPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CustomBackendMtlsPolicySpec `json:"spec"`
}

type CustomBackendMtlsPolicySpec struct {
	TargetRefs []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName `json:"targetRefs"`

	TrustDomain        string `json:"trustDomain"`
	WorkloadIdentifier string `json:"workloadIdentifier"`

	TargetRoutes []gwapiv1a2.LocalPolicyTargetReference `json:"targetRoutes"`
}

// +kubebuilder:object:root=true
type CustomBackendMtlsPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomBackendMtlsPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomBackendMtlsPolicy{}, &CustomBackendMtlsPolicyList{})
}
