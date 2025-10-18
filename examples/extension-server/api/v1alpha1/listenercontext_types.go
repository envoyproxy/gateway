// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
//
// ListenerContext provides an example extension policy context resource.
type ListenerContextExample struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ListenerContextExampleSpec `json:"spec"`
}

type ListenerContextExampleSpec struct {
	TargetRefs []gwapiv1.LocalPolicyTargetReferenceWithSectionName `json:"targetRefs"`

	Username string `json:"username"`
	Password string `json:"password"`
}

// +kubebuilder:object:root=true
//
// ListenerContextList contains a list of ListenerContext resources.
type ListenerContextExampleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ListenerContextExample `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ListenerContextExample{}, &ListenerContextExampleList{})
}
