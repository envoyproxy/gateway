// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// The Original Src filter binds upstream connections to the original source address determined
// for the connection. This address could come from something like the Proxy Protocol filter, or it
// could come from trusted http headers.
type OriginalSrc struct {
	// Whether to bind the port to the one used in the original downstream connection.
	BindPort bool `json:"bindPort"`
	// Sets the SO_MARK option on the upstream connection's socket to the provided value. Used to
	// ensure that non-local addresses may be routed back through envoy when binding to the original
	// source address. The option will not be applied if the mark is 0.
	Mark int32 `json:"mark"`
}
