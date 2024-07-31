// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	// KindSecurityPolicy is the name of the SecurityPolicy kind.
	KindVirtualBackend = "VirtualBackend"
)

type Body string

type StatusCode uint32

type ResponseHeader string

// VirtualBackend defines the configuration for direct response.
type VirtualBackend struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VirtualBackendSpec `json:"spec"`
}

type VirtualBackendSpec struct {
	Body                 Body             `json:"body,omitempty" yaml:"body"`
	StatusCode           StatusCode       `json:"statusCode,omitempty" yaml:"statusCode"`
	ResponseHeadersToAdd []ResponseHeader `json:"responseHeadersToAdd,omitempty" yaml:"responseHeadersToAdd,omitempty"`
}
