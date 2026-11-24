// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package deprecation

import (
	"fmt"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// ClientTrafficPolicy returns the deprecated fields used in the policy, keyed by field path.
func ClientTrafficPolicy(policy *egv1a1.ClientTrafficPolicy) map[string]DeprecatedField {
	fields := make(map[string]DeprecatedField)
	if policy.Spec.EnableProxyProtocol != nil {
		fields["spec.enableProxyProtocol"] = scheduledForRemoval("spec.proxyProtocol")
	}
	if policy.Spec.Headers != nil && policy.Spec.Headers.PreserveXRequestID != nil {
		fields["spec.headers.preserveXRequestID"] = scheduledForRemoval("spec.headers.requestID")
	}
	if policy.Spec.TargetRef != nil {
		fields["spec.targetRef"] = scheduledForRemoval("spec.targetRefs")
	}
	// Optional is a non-pointer bool, so deprecated usage is only observable when it changes behavior.
	if policy.Spec.TLS != nil &&
		policy.Spec.TLS.ClientValidation != nil &&
		policy.Spec.TLS.ClientValidation.Optional {
		fields["spec.tls.clientValidation.optional"] = DeprecatedField{Replacement: "spec.tls.clientValidation.mode"}
	}
	return fields
}

// BackendTrafficPolicy returns the deprecated fields used in the policy, keyed by field path.
func BackendTrafficPolicy(policy *egv1a1.BackendTrafficPolicy) map[string]DeprecatedField {
	fields := make(map[string]DeprecatedField)
	if policy.Spec.TargetRef != nil {
		fields["spec.targetRef"] = scheduledForRemoval("spec.targetRefs")
	}
	if len(policy.Spec.Compression) > 0 {
		fields["spec.compression"] = scheduledForRemoval("spec.compressor")
	}
	AddDeprecatedLoadBalancerFields(fields, "spec.loadBalancer", policy.Spec.LoadBalancer)
	return fields
}

// SecurityPolicy returns the deprecated fields used in the policy, keyed by field path.
func SecurityPolicy(policy *egv1a1.SecurityPolicy) map[string]DeprecatedField {
	fields := make(map[string]DeprecatedField)
	if policy.Spec.TargetRef != nil {
		fields["spec.targetRef"] = scheduledForRemoval("spec.targetRefs")
	}
	if policy.Spec.ExtAuth != nil {
		if policy.Spec.ExtAuth.GRPC != nil {
			AddDeprecatedBackendClusterFields(fields, "spec.extAuth.grpc", &policy.Spec.ExtAuth.GRPC.BackendCluster, nil)
		}
		if policy.Spec.ExtAuth.HTTP != nil {
			AddDeprecatedBackendClusterFields(fields, "spec.extAuth.http", &policy.Spec.ExtAuth.HTTP.BackendCluster, nil)
		}
	}
	if policy.Spec.JWT != nil {
		for i := range policy.Spec.JWT.Providers {
			if jwks := policy.Spec.JWT.Providers[i].RemoteJWKS; jwks != nil {
				AddDeprecatedBackendClusterFields(fields, fmt.Sprintf("spec.jwt.providers[%d].remoteJWKS", i), &jwks.BackendCluster, nil)
			}
		}
	}
	if policy.Spec.OIDC != nil {
		AddDeprecatedBackendClusterFields(fields, "spec.oidc.provider", &policy.Spec.OIDC.Provider.BackendCluster, nil)
	}
	return fields
}

// EnvoyExtensionPolicy returns the deprecated fields used in the policy, keyed by field path.
func EnvoyExtensionPolicy(policy *egv1a1.EnvoyExtensionPolicy) map[string]DeprecatedField {
	fields := make(map[string]DeprecatedField)
	if policy.Spec.TargetRef != nil {
		fields["spec.targetRef"] = scheduledForRemoval("spec.targetRefs")
	}
	for i := range policy.Spec.ExtProc {
		AddDeprecatedBackendClusterFields(fields, fmt.Sprintf("spec.extProc[%d]", i),
			&policy.Spec.ExtProc[i].BackendCluster, nil)
	}
	for i, backend := range policy.Spec.Backends {
		if backend.BackendSettings != nil {
			AddDeprecatedLoadBalancerFields(fields,
				fmt.Sprintf("spec.backends[%d].backendSettings.loadBalancer", i), backend.BackendSettings.LoadBalancer)
		}
	}
	return fields
}

// EnvoyProxy returns the deprecated fields used in the EnvoyProxy resource, keyed by field path.
func EnvoyProxy(ep *egv1a1.EnvoyProxy) map[string]DeprecatedField {
	fields := make(map[string]DeprecatedField)
	AddDeprecatedEnvoyProxySpecFields(fields, "spec", &ep.Spec, false)
	return fields
}
