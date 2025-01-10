// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// AppProtocolType defines various backend applications protocols supported by Envoy Gateway
//
// +kubebuilder:validation:Enum=gateway.envoyproxy.io/h2c;gateway.envoyproxy.io/ws;gateway.envoyproxy.io/wss

// Backend allows the user to configure the endpoints of a backend and
// the behavior of the connection from Envoy Proxy to the backend.
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=envoy-gateway,shortName=be
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[?(@.type=="Accepted")].reason`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type BandwidthLimitPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of Backend.
	Spec BandwidthLimitSpec `json:"spec"`

	// Status defines the current status of Backend.
	Status BandwidthLimitStatus `json:"status,omitempty"`
}

func (b BandwidthLimitPolicy) DeepCopyObject() runtime.Object {
	//TODO implement me
	panic("implement me")
}

// BackendSpec describes the desired state of BackendSpec.
type BandwidthLimitSpec struct {
	Mode     string `json:"mode"`
	Interval int32  `json:"interval"`
	Limit    int32  `json:"limit"`
	Enable   bool   `json:"enable"`
}

// BackendStatus defines the state of Backend
type BandwidthLimitStatus struct {
	// Conditions describe the current conditions of the Backend.
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// BackendList contains a list of Backend resources.
//
// +kubebuilder:object:root=true
type BandwidthLimitPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BandwidthLimitPolicy `json:"items"`
}

func (b BandwidthLimitPolicyList) DeepCopyObject() runtime.Object {
	//TODO implement me
	panic("implement me")
}

func init() {
	SchemeBuilder.Register(&BandwidthLimitPolicy{}, &BandwidthLimitPolicyList{})
}
