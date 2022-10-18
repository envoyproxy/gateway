// Copyright 2022 Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// EnvoyProxy is the Schema for the envoyproxies API
type EnvoyProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EnvoyProxySpec   `json:"spec,omitempty"`
	Status EnvoyProxyStatus `json:"status,omitempty"`
}

// EnvoyProxySpec defines the desired state of EnvoyProxy.
type EnvoyProxySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - define desired state of cluster.
	// Important: Run "make" to regenerate code after modifying this file.
}

// EnvoyProxyStatus defines the observed state of EnvoyProxy
type EnvoyProxyStatus struct {
	// INSERT ADDITIONAL STATUS FIELDS - define observed state of cluster.
	// Important: Run "make" to regenerate code after modifying this file.
}

//+kubebuilder:object:root=true

// EnvoyProxyList contains a list of EnvoyProxy
type EnvoyProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EnvoyProxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EnvoyProxy{}, &EnvoyProxyList{})
}
