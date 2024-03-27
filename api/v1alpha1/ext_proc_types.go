// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// +kubebuilder:validation:XValidation:rule="has(self.service) ? (!has(self.service.backendRef.group) || self.service.backendRef.group == \"\") : true", message="group is invalid, only the core API group (specified by omitting the group field or setting it to an empty string) is supported"
// +kubebuilder:validation:XValidation:rule="has(self.service) ? (!has(self.service.backendRef.kind) || self.service.backendRef.kind == 'Service') : true", message="kind is invalid, only Service (specified by omitting the kind field or setting it to 'Service') is supported"
//
// ExtProc defines the configuration for External Processing filter.
type ExtProc struct {
	// Service defines the configuration of the external processing service
	BackendRef ExtProcBackendRef `json:"backendRef"`
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
