// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	// KindEnvoyExtensionPolicy is the name of the EnvoyExtensionPolicy kind.
	KindEnvoyExtensionPolicy = "EnvoyExtensionPolicy"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=eep
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// EnvoyExtensionPolicy allows the user to configure various envoy extensibility options for the Gateway.
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type EnvoyExtensionPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of EnvoyExtensionPolicy.
	Spec EnvoyExtensionPolicySpec `json:"spec"`

	// Status defines the current status of EnvoyExtensionPolicy.
	Status gwapiv1.PolicyStatus `json:"status,omitempty"`
}

// EnvoyExtensionPolicySpec defines the desired state of EnvoyExtensionPolicy.
//
// +kubebuilder:validation:XValidation:rule="(has(self.targetRef) && !has(self.targetRefs)) || (!has(self.targetRef) && has(self.targetRefs)) || (has(self.targetSelectors) && self.targetSelectors.size() > 0) ", message="either targetRef or targetRefs must be used"
// +kubebuilder:validation:XValidation:rule="has(self.targetRef) ? self.targetRef.group == 'gateway.networking.k8s.io' : true", message="this policy can only have a targetRef.group of gateway.networking.k8s.io"
// +kubebuilder:validation:XValidation:rule="has(self.targetRef) ? self.targetRef.kind in ['Gateway', 'ListenerSet', 'HTTPRoute', 'GRPCRoute', 'UDPRoute', 'TCPRoute', 'TLSRoute'] : true", message="this policy can only have a targetRef.kind of Gateway/ListenerSet/HTTPRoute/GRPCRoute/TCPRoute/UDPRoute/TLSRoute"
// +kubebuilder:validation:XValidation:rule="has(self.targetRefs) ? self.targetRefs.all(ref, ref.group == 'gateway.networking.k8s.io') : true ", message="this policy can only have a targetRefs[*].group of gateway.networking.k8s.io"
// +kubebuilder:validation:XValidation:rule="has(self.targetRefs) ? self.targetRefs.all(ref, ref.kind in ['Gateway', 'ListenerSet', 'HTTPRoute', 'GRPCRoute', 'UDPRoute', 'TCPRoute', 'TLSRoute']) : true ", message="this policy can only have a targetRefs[*].kind of Gateway/ListenerSet/HTTPRoute/GRPCRoute/TCPRoute/UDPRoute/TLSRoute"
// +kubebuilder:validation:XValidation:rule="!has(self.mergeType) || ((!has(self.targetRef) || self.targetRef.kind in ['HTTPRoute', 'GRPCRoute', 'UDPRoute', 'TCPRoute', 'TLSRoute']) && (!has(self.targetRefs) || self.targetRefs.all(ref, ref.kind in ['HTTPRoute', 'GRPCRoute', 'UDPRoute', 'TCPRoute', 'TLSRoute'])) && (!has(self.targetSelectors) || self.targetSelectors.all(sel, sel.kind in ['HTTPRoute', 'GRPCRoute', 'UDPRoute', 'TCPRoute', 'TLSRoute'])))", message="mergeType can only be used with xRoute targets"
type EnvoyExtensionPolicySpec struct {
	PolicyTargetReferences `json:",inline"`

	// MergeType determines how this configuration is merged with existing EnvoyExtensionPolicy
	// configurations targeting a parent resource. When set, this configuration will be merged
	// into the closest parent EnvoyExtensionPolicy in the route's attachment hierarchy (for
	// example, one targeting a Gateway, Gateway listener, ListenerSet, or ListenerSet
	// listener).
	// Currently, this field can only be set when targeting xRoute resources.
	// If unset, no merging occurs, and only the most specific configuration takes effect.
	//
	// +kubebuilder:validation:XValidation:rule="self != 'Replace'",message="Replace is not a valid MergeType for EnvoyExtensionPolicy"
	// +optional
	MergeType *MergeType `json:"mergeType,omitempty"`

	// Backends declares HTTP callout dependencies for Lua, Wasm and dynamic modules.
	// Each name is an alias in this policy. Extensions read the generated cluster
	// name from route metadata during a request. Initialization and independent
	// background callbacks have no request route. Asynchronous callbacks can use
	// bindings resolved for their request.
	// When merging policies, a nonempty list replaces the parent's list.
	// Inherited references keep the namespace of the policy that declared them.
	//
	// +kubebuilder:validation:MaxItems=16
	// +listType=map
	// +listMapKey=name
	// +optional
	Backends []ExtensionBackend `json:"backends,omitempty"`

	// Wasm is a list of Wasm extensions to be loaded by the Gateway.
	// Order matters, as the extensions will be loaded in the order they are
	// defined in this list.
	//
	// +kubebuilder:validation:MaxItems=16
	// +optional
	Wasm []Wasm `json:"wasm,omitempty"`

	// ExtProc is an ordered list of external processing filters
	// that should be added to the envoy filter chain
	//
	// +kubebuilder:validation:MaxItems=16
	// +optional
	ExtProc []ExtProc `json:"extProc,omitempty"`

	// Lua is an ordered list of Lua filters
	// that should be added to the envoy filter chain
	//
	// +kubebuilder:validation:MaxItems=16
	// +optional
	Lua []Lua `json:"lua,omitempty"`

	// DynamicModule is an ordered list of dynamic module HTTP filters
	// that should be added to the envoy filter chain.
	// Each module must be registered in the EnvoyProxy resource's dynamicModules
	// allowlist.
	// Order matters, as the filters will be loaded in the order they are
	// defined in this list.
	//
	// +kubebuilder:validation:MaxItems=16
	// +optional
	DynamicModule []DynamicModule `json:"dynamicModule,omitempty"`
}

// ExtensionBackend binds an extension's backend alias to a backend resource.
type ExtensionBackend struct {
	// Name is the alias used by extensions. It must be a lowercase DNS label
	// and unique within this policy.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	Name string `json:"name"`

	// BackendRef references the backend for HTTP callouts. References to another
	// namespace require a ReferenceGrant allowing the policy that declares this binding.
	BackendRef gwapiv1.BackendObjectReference `json:"backendRef"`

	// BackendSettings configures connections and traffic for this binding's cluster.
	// Request timeouts and retries are controlled by the extension's callout API.
	// RequestTimeout and StreamIdleTimeout are not supported here.
	// DynamicModule load balancing is not supported here.
	//
	// +kubebuilder:validation:XValidation:rule="!has(self.timeout) || !has(self.timeout.http) || (!has(self.timeout.http.requestTimeout) && !has(self.timeout.http.streamIdleTimeout))",message="requestTimeout and streamIdleTimeout are controlled by the extension"
	// +kubebuilder:validation:XValidation:rule="!has(self.loadBalancer) || self.loadBalancer.type != 'DynamicModule'",message="DynamicModule load balancing is not supported for extension backends"
	// +optional
	BackendSettings *ClusterSettings `json:"backendSettings,omitempty"`
}

//+kubebuilder:object:root=true

// EnvoyExtensionPolicyList contains a list of EnvoyExtensionPolicy resources.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type EnvoyExtensionPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EnvoyExtensionPolicy `json:"items"`
}

func init() {
	localSchemeBuilder.Register(&EnvoyExtensionPolicy{}, &EnvoyExtensionPolicyList{})
}
