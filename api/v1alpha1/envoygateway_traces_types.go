// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// EnvoyGatewayTraces defines control plane tracing configurations.
type EnvoyGatewayTraces struct {
	// Disable disables the traces.
	//
	// +optional
	Disable bool `json:"disable,omitempty"`
}
