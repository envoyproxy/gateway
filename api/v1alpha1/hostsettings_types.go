// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// HostSettings provides settings that manage how the incoming Host/Authority header
// set by clients is normalized.
type HostSettings struct {
	// StripTrailingHostDot determines if the trailing dot of the host should be removed
	// from the Host/Authority header before any processing of the request.
	// This affects the upstream host header as well. Without this option, incoming requests
	// with host "example.com." will not match routes with domains set to "example.com".
	// When the host includes a port (for example "example.com.:443"), only the trailing dot
	// from the host section is stripped, leaving the port as-is ("example.com:443").
	// Defaults to false.
	//
	// +optional
	StripTrailingHostDot *bool `json:"stripTrailingHostDot,omitempty"`
}
