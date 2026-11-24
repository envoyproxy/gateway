// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package deprecation

import (
	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// ClientTrafficPolicy returns the deprecated fields used in the policy, keyed by field path.
func ClientTrafficPolicy(policy *egv1a1.ClientTrafficPolicy) map[string]DeprecatedField {
	fields := make(map[string]DeprecatedField)
	// Optional is a non-pointer bool, so deprecated usage is only observable when it changes behavior.
	if policy.Spec.TLS != nil &&
		policy.Spec.TLS.ClientValidation != nil &&
		policy.Spec.TLS.ClientValidation.Optional {
		fields["spec.tls.clientValidation.optional"] = DeprecatedField{Replacement: "spec.tls.clientValidation.mode"}
	}
	return fields
}

// EnvoyProxy returns the deprecated fields used in the EnvoyProxy resource, keyed by field path.
func EnvoyProxy(ep *egv1a1.EnvoyProxy) map[string]DeprecatedField {
	fields := make(map[string]DeprecatedField)
	AddDeprecatedEnvoyProxySpecFields(fields, "spec", &ep.Spec)
	return fields
}
