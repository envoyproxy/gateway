// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/gateway/internal/ir"
)

func buildHeaderMatches(irRoute *ir.HTTPRoute) []*ir.StringMatch {
	// Create header matches:
	// copy original headers (excluding :method) + add CORS headers (:method=OPTIONS, origin, access-control-request-method)
	headerMatches := make([]*ir.StringMatch, 0, len(irRoute.HeaderMatches)+2)
	for _, headerMatch := range irRoute.HeaderMatches {
		// Skip the original method match for CORS preflight route to avoid conflicting method requirements.
		if headerMatch.Name == ":method" {
			continue
		}
		headerMatches = append(headerMatches, headerMatch)
	}

	corsHeaders := []*ir.StringMatch{
		{
			Name:  ":method",
			Exact: ptr.To("OPTIONS"),
		},
		{
			Name:      "origin",
			SafeRegex: ptr.To(".*"),
		},
		{
			Name:      "access-control-request-method",
			SafeRegex: ptr.To(".*"),
		},
	}
	headerMatches = append(headerMatches, corsHeaders...)
	return headerMatches
}
