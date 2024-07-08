package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
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
	TargetRefs []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName `json:"targetRefs"`

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
