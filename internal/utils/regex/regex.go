// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package regex

import (
	"fmt"
	"regexp"
	"strings"
)

// Validate validates a regex string.
func Validate(regex string) error {
	if _, err := regexp.Compile(regex); err != nil {
		return fmt.Errorf("regex %q is invalid: %w", regex, err)
	}
	return nil
}

// PathSeparatedPrefixRegex creates a regex pattern that Envoy's PathSeparatedPrefix behavior.
// The pattern matches paths that either exactly match the prefix or have the prefix followed by "/".
// This ensures proper path separation (e.g., "/api" matches "/api" and "/api/v1" but not "/apiv1").
//
// References:
// - Envoy 'path_separated_prefix' : https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#config-route-v3-routematch
func PathSeparatedPrefixRegex(prefix string) string {
	// Remove trailing slash
	trimmedPrefix := strings.TrimSuffix(prefix, "/")

	// Escape special regex characters in the prefix
	escapedPrefix := regexp.QuoteMeta(trimmedPrefix)

	return "^" + escapedPrefix + "(/.*|\\?.*|#.*|;.*|$)"
}
