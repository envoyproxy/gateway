// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// HostSettings provides settings that manage how the incoming Host/Authority header
// set by clients is normalized.
type HostSettings struct {
	// StripPort determines whether the port is removed from the Host/Authority header.
	// It maps to Envoy's strip_any_host_port, which strips the port unconditionally.
	// The port is removed before route matching, and this affects the upstream host header
	// as well: backends and access logs see the normalized Host/:authority value without
	// the port.
	// If not set, no port stripping is performed (Envoy default).
	//
	// Stripping only the port that matches the listener port (Envoy's
	// strip_matching_host_port) is intentionally not offered: Envoy compares against the
	// Envoy listener port, which differs from the user-facing Gateway listener port (for
	// example, a Gateway listener on port 80 is translated to an Envoy listener on port
	// 10080). Matching-port stripping would therefore be silently ineffective for the
	// port clients actually use, so only unconditional stripping is exposed.
	//
	// +optional
	StripPort *bool `json:"stripPort,omitempty"`
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
