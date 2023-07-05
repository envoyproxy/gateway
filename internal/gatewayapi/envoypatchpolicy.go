// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"sort"

	gwv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func (t *Translator) ProcessEnvoyPatchPolicies(envoyPatchPolicies []*egv1a1.EnvoyPatchPolicy, xdsIR XdsIRMap) {
	// Sort based on priority
	sort.Slice(envoyPatchPolicies, func(i, j int) bool {
		return envoyPatchPolicies[i].Spec.Priority < envoyPatchPolicies[j].Spec.Priority
	})

	for _, policy := range envoyPatchPolicies {
		// Ensure policy can only target a Gateway
		if policy.Spec.TargetRef.Group != gwv1b1.GroupName || policy.Spec.TargetRef.Kind != KindGateway {
			// TODO: Update Status
			continue
		}

		// Ensure Policy and target Gateway are in the same namespace
		targetNs := policy.Spec.TargetRef.Namespace
		if targetNs == nil || policy.Namespace != string(*targetNs) {
			// TODO: Update Status
			continue
		}

		// Get the IR
		// It must exist since the gateways have already been processed
		irKey := irStringKey(string(*targetNs), string(policy.Spec.TargetRef.Name))
		gwXdsIR, ok := xdsIR[irKey]
		if !ok {
			// TODO: Update Status
			continue
		}

		// Save the patch
		for _, patch := range policy.Spec.JSONPatches {
			irPatch := ir.JSONPatchConfig{}
			irPatch.Type = string(patch.Type)
			irPatch.Name = patch.Name
			irPatch.Operation.Op = string(patch.Operation.Op)
			irPatch.Operation.Path = patch.Operation.Path
			irPatch.Operation.Value = patch.Operation.Value

			gwXdsIR.JSONPatches = append(gwXdsIR.JSONPatches, &irPatch)
		}
	}
}
