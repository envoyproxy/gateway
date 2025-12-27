package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// BackendPolicy allows the user to configure the behavior of the connection
// between the Envoy Proxy and the backend service.
type BackendPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of BackendPolicySpec.
	Spec BackendPolicySpec `json:"spec"`

	// status defines the current status of BackendPolicySpec.
	Status gwapiv1.PolicyStatus `json:"status,omitempty"`
}

// BackendPolicySpec defines the desired state of BackendPolicy.
type BackendPolicySpec struct {
	PolicyTargetReferences `json:",inline"`
	ClusterSettings        `json:",inline"`
}
