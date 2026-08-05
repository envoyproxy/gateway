// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// RouteContextExample provides an example extension policy context resource that
// targets an HTTPRoute or one of its rules (via sectionName).
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type RouteContextExample struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RouteContextExampleSpec `json:"spec"`
}

type RouteContextExampleSpec struct {
	TargetRefs []gwapiv1.LocalPolicyTargetReferenceWithSectionName `json:"targetRefs"`

	ResponseHeaderName string `json:"responseHeaderName"`
	ResponseHeaderValue string `json:"responseHeaderValue"`
}

// +kubebuilder:object:root=true
//
// RouteContextExampleList contains a list of RouteContextExample resources.
type RouteContextExampleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RouteContextExample `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RouteContextExample{}, &RouteContextExampleList{})
}
