// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

const (
	// KindEnvoyPatchPolicy is the name of the EnvoyPatchPolicy kind.
	KindEnvoyPatchPolicy = "EnvoyPatchPolicy"
)

// +kubebuilder:object:root=true

// EnvoyPatchPolicy allows the user to modify the generated Envoy xDS
// resources by Envoy Gateway using this patch API
type EnvoyPatchPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of EnvoyPatchPolicy.
	Spec EnvoyPatchPolicySpec `json:"spec"`

	// Status defines the current status of EnvoyPatchPolicy.
	Status EnvoyPatchPolicyStatus `json:"status,omitempty"`
}

// EnvoyPatchPolicySpec defines the desired state of EnvoyPatchPolicy.
// +union
type EnvoyPatchPolicySpec struct {
	// Type decides the type of patch.
	// Valid EnvoyPatchType values are "JSONPatch".
	//
	// +unionDiscriminator
	Type EnvoyPatchType `json:"type"`
	// JSONPatch defines the JSONPatch configuration.
	//
	// +optional
	JSONPatches []EnvoyJSONPatchConfig `json:"jsonPatches,omitempty"`
	// TargetRef is the name of the Gateway API resource this policy
	// is being attached to.
	// Currently only attaching to Gateway is supported
	// This Policy and the TargetRef MUST be in the same namespace
	// for this Policy to have effect and be applied to the Gateway
	// TargetRef
	TargetRef gwapiv1a2.PolicyTargetReference `json:"targetRef"`
	// Priority of the EnvoyPatchPolicy.
	// If multiple EnvoyPatchPolicies are applied to the same
	// TargetRef, they will be applied in the ascending order of
	// the priority i.e. int32.min has the highest priority and
	// int32.max has the lowest priority.
	// Defaults to 0.
	Priority int32 `json:"priority"`
}

// EnvoyPatchType specifies the types of Envoy patching mechanisms.
// +kubebuilder:validation:Enum=JSONPatch
type EnvoyPatchType string

const (
	// JSONPatchEnvoyPatchType allows the user to patch the generated xDS resources using JSONPatch semantics.
	// For more details on the semantics, please refer to https://datatracker.ietf.org/doc/html/rfc6902
	JSONPatchEnvoyPatchType EnvoyPatchType = "JSONPatch"
)

// EnvoyJSONPatchConfig defines the configuration for patching a Envoy xDS Resource
// using JSONPatch semantic
type EnvoyJSONPatchConfig struct {
	// Type is the typed URL of the Envoy xDS Resource
	Type EnvoyResourceType `json:"type"`
	// Name is the name of the resource
	Name string `json:"name"`
	// Patch defines the JSON Patch Operation
	Operation JSONPatchOperation `json:"operation"`
}

// EnvoyResourceType specifies the type URL of the Envoy resource.
// +kubebuilder:validation:Enum=type.googleapis.com/envoy.config.listener.v3.Listener;type.googleapis.com/envoy.config.route.v3.RouteConfiguration;type.googleapis.com/envoy.config.cluster.v3.Cluster;type.googleapis.com/envoy.config.endpoint.v3.ClusterLoadAssignment
type EnvoyResourceType string

const (
	// ListenerEnvoyResourceType defines the Type URL of the Listener resource
	ListenerEnvoyResourceType EnvoyResourceType = "type.googleapis.com/envoy.config.listener.v3.Listener"
	// RouteConfigurationEnvoyResourceType defines the Type URL of the RouteConfiguration resource
	RouteConfigurationEnvoyResourceType EnvoyResourceType = "type.googleapis.com/envoy.config.route.v3.RouteConfiguration"
	// ClusterEnvoyResourceType defines the Type URL of the Cluster resource
	ClusterEnvoyResourceType EnvoyResourceType = "type.googleapis.com/envoy.config.cluster.v3.Cluster"
	// ClusterLoadAssignmentEnvoyResourceType defines the Type URL of the ClusterLoadAssignment resource
	ClusterLoadAssignmentEnvoyResourceType EnvoyResourceType = "type.googleapis.com/envoy.config.endpoint.v3.ClusterLoadAssignment"
)

// JSONPatchOperationType specifies the JSON Patch operations that can be performed.
// +kubebuilder:validation:Enum=add;remove;replace;move;copy;test
type JSONPatchOperationType string

// JSONPatchOperation defines the JSON Patch Operation as defined in
// https://datatracker.ietf.org/doc/html/rfc6902
type JSONPatchOperation struct {
	// Op is the type of operation to perform
	Op JSONPatchOperationType `json:"op"`
	// Path is the location of the target document/field where the operation will be performed
	// Refer to https://datatracker.ietf.org/doc/html/rfc6901 for more details.
	Path string `json:"path"`
	// Value is the new value of the path location.
	Value apiextensionsv1.JSON `json:"value"`
}

// EnvoyPatchPolicyStatus defines the state of EnvoyPatchPolicy
type EnvoyPatchPolicyStatus struct {
	// Conditions describe the current conditions of the EnvoyPatchPolicy.
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true

// EnvoyPatchPolicyList contains a list of EnvoyPatchPolicy resources.
type EnvoyPatchPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EnvoyPatchPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EnvoyPatchPolicy{}, &EnvoyPatchPolicyList{})
}
