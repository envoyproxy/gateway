// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package envoyformatvalidator

import (
	"fmt"
	"regexp"
	"strings"
)

// commandOperatorRegex matches Envoy command operators: %OPERATOR(args):Z%
var commandOperatorRegex = regexp.MustCompile(`%([A-Z_]+)(\([^)]*\))?(:[0-9]+)?%`)

// operatorsWithAlternatives lists operators that support alternative header lookup with ?
var operatorsWithAlternatives = map[string]bool{
	"REQ":     true,
	"RESP":    true,
	"TRAILER": true,
}

// operatorsRequiringArgs lists operators that must have arguments
var operatorsRequiringArgs = map[string]bool{
	"REQ":               true,
	"RESP":              true,
	"TRAILER":           true,
	"DYNAMIC_METADATA":  true,
	"CLUSTER_METADATA":  true,
	"UPSTREAM_METADATA": true,
	"FILTER_STATE":      true,
	"ENVIRONMENT":       true,
	"QUERY_PARAM":       true,
	"PATH":              true,
}

// validOperators lists all valid Envoy command operators
var validOperators = map[string]bool{
	// Request operators
	"REQ":                   true,
	"REQUEST_HEADERS_BYTES": true,
	"REQUEST_BODY_BYTES":    true,
	"GRPC_STATUS":           true,
	"GRPC_STATUS_NUMBER":    true,
	// Response operators
	"RESP":                    true,
	"RESPONSE_HEADERS_BYTES":  true,
	"RESPONSE_BODY_BYTES":     true,
	"RESPONSE_TRAILERS_BYTES": true,
	"TRAILER":                 true,
	// Connection operators
	"DOWNSTREAM_REMOTE_ADDRESS":                     true,
	"DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT":        true,
	"DOWNSTREAM_LOCAL_ADDRESS":                      true,
	"DOWNSTREAM_LOCAL_ADDRESS_WITHOUT_PORT":         true,
	"DOWNSTREAM_LOCAL_PORT":                         true,
	"DOWNSTREAM_PEER_URI_SAN":                       true,
	"DOWNSTREAM_LOCAL_URI_SAN":                      true,
	"DOWNSTREAM_PEER_SUBJECT":                       true,
	"DOWNSTREAM_LOCAL_SUBJECT":                      true,
	"DOWNSTREAM_PEER_ISSUER":                        true,
	"DOWNSTREAM_TLS_SESSION_ID":                     true,
	"DOWNSTREAM_TLS_CIPHER":                         true,
	"DOWNSTREAM_TLS_VERSION":                        true,
	"DOWNSTREAM_PEER_FINGERPRINT_256":               true,
	"DOWNSTREAM_PEER_FINGERPRINT_1":                 true,
	"DOWNSTREAM_PEER_SERIAL":                        true,
	"DOWNSTREAM_PEER_CERT":                          true,
	"DOWNSTREAM_PEER_CERT_V_START":                  true,
	"DOWNSTREAM_PEER_CERT_V_END":                    true,
	"DOWNSTREAM_DIRECT_REMOTE_ADDRESS":              true,
	"DOWNSTREAM_DIRECT_REMOTE_ADDRESS_WITHOUT_PORT": true,
	"CONNECTION_ID":                                 true,
	"REQUESTED_SERVER_NAME":                         true,
	// Upstream operators
	"UPSTREAM_HOST":                     true,
	"UPSTREAM_CLUSTER":                  true,
	"UPSTREAM_LOCAL_ADDRESS":            true,
	"UPSTREAM_TRANSPORT_FAILURE_REASON": true,
	"UPSTREAM_WIRE_BYTES_SENT":          true,
	"UPSTREAM_WIRE_BYTES_RECEIVED":      true,
	"UPSTREAM_HEADER_BYTES_SENT":        true,
	"UPSTREAM_HEADER_BYTES_RECEIVED":    true,
	"UPSTREAM_METADATA":                 true,
	"UPSTREAM_PEER_URI_SAN":             true,
	"UPSTREAM_PEER_SUBJECT":             true,
	"UPSTREAM_PEER_ISSUER":              true,
	"UPSTREAM_TLS_SESSION_ID":           true,
	"UPSTREAM_TLS_CIPHER":               true,
	"UPSTREAM_TLS_VERSION":              true,
	"UPSTREAM_PEER_CERT_V_START":        true,
	"UPSTREAM_PEER_CERT_V_END":          true,
	"UPSTREAM_LOCAL_URI_SAN":            true,
	"UPSTREAM_LOCAL_SUBJECT":            true,
	// Timing operators
	"START_TIME":                     true,
	"DURATION":                       true,
	"REQUEST_DURATION":               true,
	"REQUEST_TX_DURATION":            true,
	"RESPONSE_DURATION":              true,
	"RESPONSE_TX_DURATION":           true,
	"DOWNSTREAM_HANDSHAKE_DURATION":  true,
	"ROUNDTRIP_DURATION":             true,
	"BYTES_RECEIVED":                 true,
	"PROTOCOL":                       true,
	"RESPONSE_CODE":                  true,
	"RESPONSE_CODE_DETAILS":          true,
	"CONNECTION_TERMINATION_DETAILS": true,
	"BYTES_SENT":                     true,
	"RESPONSE_FLAGS":                 true,
	"RESPONSE_FLAGS_LONG":            true,
	"ROUTE_NAME":                     true,
	"VIRTUAL_CLUSTER_NAME":           true,
	// Metadata operators
	"DYNAMIC_METADATA": true,
	"CLUSTER_METADATA": true,
	"FILTER_STATE":     true,
	// Other operators
	"UNIQUE_ID":    true,
	"ENVIRONMENT":  true,
	"TRACE_ID":     true,
	"QUERY_PARAM":  true,
	"PATH":         true,
	"CUSTOM_FLAGS": true,
}

// ValidateEnvoyFormatString validates an Envoy format string containing command operators
func ValidateEnvoyFormatString(value string) error {
	// Empty strings are valid
	if value == "" {
		return nil
	}

	// Find all command operators in the string
	matches := commandOperatorRegex.FindAllStringSubmatch(value, -1)

	// No operators found - this is fine, could be plain text
	if len(matches) == 0 {
		return nil
	}

	// Check for malformed % usage (single % without matching operator pattern)
	// We need to check if there are any % that aren't part of valid operators
	tempValue := value
	for _, match := range matches {
		tempValue = strings.Replace(tempValue, match[0], "", 1)
	}
	if strings.Contains(tempValue, "%") {
		return fmt.Errorf("malformed command operator: unpaired or invalid '%%' character found")
	}

	// Validate each operator
	for _, match := range matches {
		fullMatch := match[0]  // e.g., %REQ(X-Real-IP?x-client-ip)%
		operator := match[1]   // e.g., REQ
		args := match[2]       // e.g., (X-Real-IP?x-client-ip) or empty
		truncation := match[3] // e.g., :10 or empty

		// Validate operator is known
		if !validOperators[operator] {
			return fmt.Errorf("unknown command operator: %s in %q", operator, fullMatch)
		}

		// Validate operators that require arguments
		if operatorsRequiringArgs[operator] && args == "" {
			return fmt.Errorf("command operator %s requires arguments in %q", operator, fullMatch)
		}

		// Validate operators with alternatives (only one ? allowed)
		if operatorsWithAlternatives[operator] && args != "" {
			// Remove parentheses
			argsContent := strings.TrimPrefix(args, "(")
			argsContent = strings.TrimSuffix(argsContent, ")")

			// Count question marks
			questionMarkCount := strings.Count(argsContent, "?")
			if questionMarkCount > 1 {
				return fmt.Errorf("more than 1 alternative header specified in token: %s", argsContent)
			}
		}

		// Validate truncation syntax if present
		if truncation != "" {
			if !regexp.MustCompile(`^:[0-9]+$`).MatchString(truncation) {
				return fmt.Errorf("invalid truncation syntax in %q", fullMatch)
			}
		}
	}

	return nil
}
