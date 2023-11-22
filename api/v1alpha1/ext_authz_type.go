// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// ExtAuthz defines the configuration for External Authorization.
type ExtAuthz struct {
	// GRPCURI defines the gRPC cluster name to use in that route.
	GRPCURI string `json:"grpcURI,omitempty" yaml:"grpcURI"`
}
