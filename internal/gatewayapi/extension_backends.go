// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
)

func (t *Translator) buildExtensionBackends(policy *egv1a1.EnvoyExtensionPolicy,
	owners *envoyExtensionPolicyOwners, resources *resource.Resources, gateway *GatewayContext,
) (map[string]*ir.ExtensionBackend, error) {
	if len(policy.Spec.Backends) == 0 {
		return nil, nil
	}

	// Inherited references use the namespace of the policy that declared them.
	owner := policyOwnerOr(owners.backends, policy)
	backends := make(map[string]*ir.ExtensionBackend, len(policy.Spec.Backends))
	// Standalone input bypasses CRD validation.
	for i, backend := range policy.Spec.Backends {
		if problems := validation.IsDNS1123Label(backend.Name); len(problems) > 0 {
			return nil, fmt.Errorf("backends[%d].name %q is invalid: %s", i, backend.Name, strings.Join(problems, ", "))
		}
		if _, exists := backends[backend.Name]; exists {
			return nil, fmt.Errorf("backends[%d].name %q is duplicated", i, backend.Name)
		}
		// Gateway context can change client TLS settings even for the same policy.
		// Include its identity so merged Gateways do not reuse different clusters.
		destination, err := t.translateExtServiceBackendRefs(owner,
			[]egv1a1.BackendRef{{BackendObjectReference: backend.BackendRef}}, ir.HTTP,
			resources, gateway, "backend/"+backend.Name+"/gateway/"+gateway.Namespace+"/"+gateway.Name, 0)
		if err != nil {
			return nil, fmt.Errorf("backends[%d] %q: %w", i, backend.Name, err)
		}
		var traffic *ir.TrafficFeatures
		if settings := backend.BackendSettings; settings != nil {
			if settings.LoadBalancer != nil && settings.LoadBalancer.Type == egv1a1.DynamicModuleLoadBalancerType {
				return nil, fmt.Errorf("backends[%d] %q: DynamicModule load balancing is not supported for extension backends", i, backend.Name)
			}
			if settings.Timeout != nil && settings.Timeout.HTTP != nil &&
				(settings.Timeout.HTTP.RequestTimeout != nil || settings.Timeout.HTTP.StreamIdleTimeout != nil) {
				return nil, fmt.Errorf("backends[%d] %q: requestTimeout and streamIdleTimeout are controlled by the extension", i, backend.Name)
			}
			traffic, err = translateTrafficFeatures(&egv1a1.BackendSettings{ClusterSettings: *settings})
			if err != nil {
				return nil, fmt.Errorf("backends[%d] %q backendSettings: %w", i, backend.Name, err)
			}
		}
		backends[backend.Name] = &ir.ExtensionBackend{RouteDestination: *destination, Traffic: traffic}
	}
	return backends, nil
}
