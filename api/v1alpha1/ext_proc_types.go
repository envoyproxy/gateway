// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// ExtProc defines the configuration for External Processing filter.
type ExtProc struct {
	// BackendRefs references a Kubernetes object that represents the gRPC service to which
	// the access logs will be sent. Currently only Service is supported.
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=1
	// +kubebuilder:validation:XValidation:message="BackendRefs only supports Service kind.",rule="self.all(f, f.kind == 'Service')"
	// +kubebuilder:validation:XValidation:message="BackendRefs only supports Core group.",rule="self.all(f, f.group == '')"
	BackendRefs []BackendRef `json:"backendRefs"`

	// MessageTimeout is the timeout for a response to be returned from the external processor
	// Default: 200ms
	//
	// +optional
	MessageTimeout *gwapiv1.Duration `json:"messageTimeout,omitempty"`

	// FailOpen defines if requests or responses that cannot be processed due to connectivity to the
	// external processor are terminated or passed-through.
	// Default: false
	//
	// +optional
	FailOpen *bool `json:"failOpen,omitempty"`
}

// ExtProcService defines the gRPC External Processing service using the envoy grpc client
// The processing request and response messages are defined in
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/ext_proc/v3/external_processor.proto
type ExtProcBackendRef struct {
	// BackendObjectReference references a Kubernetes object that represents the
	// backend server to which the processing requests will be sent.
	// Only service Kind is supported for now.
	gwapiv1.BackendObjectReference `json:",inline"`
}
