// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"github.com/golang/protobuf/ptypes/any"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

const (
	// WasmExtensionPolicy is the name of the WasmExtensionPolicy kind.
	KindWasmExtensionPolicy = "WasmExtensionPolicy"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=envoy-gateway,shortName=sp
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[?(@.type=="Accepted")].reason`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// WasmExtensionPolicy allows the user to configure wasm extensions for a Gateway.
type WasmExtensionPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of WasmExtensionPolicy.
	Spec WasmExtensionPolicySpec `json:"spec"`

	// Status defines the current status of WasmExtensionPolicySpec.
	Status gwapiv1a2.PolicyStatus `json:"status,omitempty"`
}

// WasmExtensionPolicySpec defines the desired state of WasmExtensionPolicy.
type WasmExtensionPolicySpec struct {
	// +kubebuilder:validation:XValidation:rule="self.group == 'gateway.networking.k8s.io'", message="this policy can only have a targetRef.group of gateway.networking.k8s.io"
	// +kubebuilder:validation:XValidation:rule="self.kind in ['Gateway', 'HTTPRoute', 'GRPCRoute']", message="this policy can only have a targetRef.kind of Gateway/HTTPRoute/GRPCRoute"
	// +kubebuilder:validation:XValidation:rule="!has(self.sectionName)",message="this policy does not yet support the sectionName field"
	//
	// TargetRef is the name of the Gateway resource this policy
	// is being attached to.
	// This Policy and the TargetRef MUST be in the same namespace
	// for this Policy to have effect and be applied to the Gateway.
	TargetRef gwapiv1a2.PolicyTargetReferenceWithSectionName `json:"targetRef"`

	// Extensions is a list of Wasm extensions to be loaded by the Gateway.
	// Order matters, as the extensions will be loaded in the order they are
	// defined in this list.
	//
	// +kubebuilder:validation:MinItems=1
	Extensions []WasmExtension `json:"extensions"`
}

// WasmExtension defines an wasm extension.
type WasmExtension struct {
	// Name is a unique name for this Wasm extension. It is used to identify the
	// Wasm extension if multiple extensions are handled by the same vm_id and root_id.
	// It's also used for logging/debugging.
	Name string `json:"name"`

	// VmID is an ID that will be used along with a hash of the wasm code to
	// determine which VM will be used to load the Wasm extension. All extensions
	// that have the same vm_id and code will use the same VM.
	//
	// Note that sharing a VM between plugins can reduce memory utilization and
	// make sharing of data easier, but it may have security implications.
	VmID *string `json:"vmID,omitempty"`

	// RootID is a unique ID for a set of extensions in a VM which will share a
	// RootContext and Contexts if applicable (e.g., an Wasm HttpFilter and an Wasm AccessLog).
	// If left blank, all extensions with a blank root_id with the same vm_id will share Context(s).
	RootID *string `json:"rootID,omitempty"`

	// Code is the wasm code for the extension.
	Code WasmCodeSource `json:"code"`

	// Configuration for the wasm code.
	Config any.Any `json:"config"`
}

// WasmCodeSource defines the source of the wasm code.
type WasmCodeSource struct {
	// ConfigMap is the name of the ConfigMap containing the wasm code.
	//
	// The key in the ConfigMap should be the name of the WasmExtension. For example,
	// if the WasmExtension is named "my-wasm-extension", the ConfigMap should have a key
	// named "my-wasm-extension" and the value should be the wasm code.
	ConfigMap *string `json:"ConfigMap,omitempty"`

	// HTTP is the HTTP URL containing the wasm code.
	//
	// Note that the HTTP server must be accessible from the Envoy proxy.
	HTTP *string `json:"http,omitempty"`

	// Image is the OCI image containing the wasm code.
	// Image *string `json:"image,omitempty"` //TODO: Add support for OCI image in the future.
}

//+kubebuilder:object:root=true

// WasmExtensionPolicyList contains a list of WasmExtensionPolicy resources.
type WasmExtensionPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WasmExtensionPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WasmExtensionPolicy{}, &WasmExtensionPolicyList{})
}
