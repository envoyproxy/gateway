// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func (t *Translator) ProcessBackendTLSPoliciesAncestorRef(backendTLSPolicies []*v1alpha2.BackendTLSPolicy, gateways []*GatewayContext) []*v1alpha2.BackendTLSPolicy {

	var res []*v1alpha2.BackendTLSPolicy

	for _, btlsPolicy := range backendTLSPolicies {

		policy := btlsPolicy.DeepCopy()
		res = append(res, policy)

		if policy.Status.Ancestors != nil {
			for k, status := range policy.Status.Ancestors {
				exist := false
				for _, gwContext := range gateways {
					gw := gwContext.Gateway
					if gw.Name == string(status.AncestorRef.Name) && gw.Namespace == NamespaceDerefOrAlpha(status.AncestorRef.Namespace, "default") {
						for _, lis := range gw.Spec.Listeners {
							if lis.Name == *status.AncestorRef.SectionName {
								exist = true
							}
						}
					}
				}

				if !exist {
					policy.Status.Ancestors = append(policy.Status.Ancestors[:k], policy.Status.Ancestors[k+1:]...)
				}
			}
		} else {
			policy.Status.Ancestors = []v1alpha2.PolicyAncestorStatus{}
		}
	}

	return res
}
