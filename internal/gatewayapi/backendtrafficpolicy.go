// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/status"
	"github.com/envoyproxy/gateway/internal/utils/ptr"
)

type routePolicyTarget struct {
	kind gwv1b1.Kind
	key  types.NamespacedName
}

func ProcessBackendTrafficPolicies(backendTrafficPolicies []*egv1a1.BackendTrafficPolicy,
	gateways []*GatewayContext,
	routes []RouteContext,
	xdsIR XdsIRMap) []*egv1a1.BackendTrafficPolicy {
	var res []*egv1a1.BackendTrafficPolicy

	// Sort based on timestamp
	sort.Slice(backendTrafficPolicies, func(i, j int) bool {
		return backendTrafficPolicies[i].CreationTimestamp.Before(&(backendTrafficPolicies[j].CreationTimestamp))
	})

	// First build a map out of the routes and gateways for faster lookup since users might have thousands of routes or more.
	// For gateways this probably isn't quite as necessary.
	routeMap := map[types.NamespacedName]RouteContext{}
	for _, route := range routes {
		routeMap[types.NamespacedName{
			Name:      route.GetName(),
			Namespace: route.GetNamespace(),
		}] = route
	}
	gatewayMap := map[types.NamespacedName]*GatewayContext{}
	for _, gw := range gateways {
		gatewayMap[types.NamespacedName{
			Name:      gw.GetName(),
			Namespace: gw.GetNamespace(),
		}] = gw
	}

	targetedGateways := make(map[types.NamespacedName]sets.Set[string])
	targetedRoutes := map[routePolicyTarget]bool{}

	// Keep track of all the parent gateways of routes that are being targeted by these policies so that we can set the
	// overridden status on them if needed
	overriddenPolicyTargets := map[types.NamespacedName]bool{}

	// Translate
	// 1. First translate Policies targeting xRoutes
	// 2. Next translate Policies targeting Gateways with a sectionName
	// 3. Finally, the policies targeting Gateways but without a sectionName

	// TODO: Import sort order to ensure policy with same section always appear
	// before policy with no section so below loops can be flattened.
	// This function is messy due to all the context passing/parsing to set the appropriate conditions for all the policies.
	// There is definitely room for some cleanup and optimization here.

	// Process the policies targeting xRoutes
	for _, policy := range backendTrafficPolicies {
		if policy.Spec.TargetRef.Kind != KindGateway {
			policy := policy.DeepCopy()
			res = append(res, policy)

			// Negative statuses have already been assigned so its safe to skip
			target := resolveBTPolicyTargetRef(policy, gatewayMap, routeMap)
			if target == nil {
				continue
			}
			routeKey := routePolicyTarget{key: *target, kind: policy.Spec.TargetRef.Kind}

			// Check if another policy targeting the same xRoute exists
			if _, ok := targetedRoutes[routeKey]; ok {
				message := fmt.Sprintf("Unable to target %s, another BackendTrafficPolicy has already attached to it", routeKey.kind)

				status.SetBackendTrafficPolicyCondition(policy,
					gwv1a2.PolicyConditionAccepted,
					metav1.ConditionFalse,
					gwv1a2.PolicyReasonConflicted,
					message,
				)

				continue
			}

			targetedRoutes[routeKey] = true
			route := routeMap[*target]
			for _, parentRef := range GetParentReferences(route) {
				namespace := ""
				if parentRef.Namespace != nil {
					namespace = string(*parentRef.Namespace)
				}
				overriddenPolicyTargets[types.NamespacedName{
					Name:      string(parentRef.Name),
					Namespace: namespace,
				}] = true
			}
			translateBackendTrafficPolicy(policy, xdsIR)

			// Set Accepted=True
			status.SetBackendTrafficPolicyCondition(policy,
				gwv1a2.PolicyConditionAccepted,
				metav1.ConditionTrue,
				gwv1a2.PolicyReasonAccepted,
				"BackendTrafficPolicy has been accepted.",
			)
		}
	}

	// Process the policies targeting Gateways with a section name
	for _, policy := range backendTrafficPolicies {
		if policy.Spec.TargetRef.SectionName != nil {
			policy := policy.DeepCopy()
			res = append(res, policy)

			// Negative statuses have already been assigned so its safe to skip
			gatewayKey := resolveBTPolicyTargetRef(policy, gatewayMap, routeMap)
			if gatewayKey == nil {
				continue
			}

			// Check if another policy targeting the same section exists
			section := string(*(policy.Spec.TargetRef.SectionName))
			s, ok := targetedGateways[*gatewayKey]
			if ok && s.Has(section) {
				message := "Unable to target section, another BackendTrafficPolicy has already attached to it"

				status.SetBackendTrafficPolicyCondition(policy,
					gwv1a2.PolicyConditionAccepted,
					metav1.ConditionFalse,
					gwv1a2.PolicyReasonConflicted,
					message,
				)
				continue
			}

			// Check if any policies that target xRoute resources are overriding this policy so we can set the status
			if _, ok := overriddenPolicyTargets[*gatewayKey]; ok {
				message := "There are existing BackendTrafficPolicies that are overriding this config at the route level"

				status.SetBackendTrafficPolicyCondition(policy,
					egv1a1.PolicyConditionOverridden,
					metav1.ConditionTrue,
					egv1a1.PolicyReasonOverridden,
					message,
				)
			}

			// Add section to targeted gateways map
			if s == nil {
				targetedGateways[*gatewayKey] = sets.New[string]()
			}
			targetedGateways[*gatewayKey].Insert(section)

			translateBackendTrafficPolicy(policy, xdsIR)

			// Set Accepted=True
			status.SetBackendTrafficPolicyCondition(policy,
				gwv1a2.PolicyConditionAccepted,
				metav1.ConditionTrue,
				gwv1a2.PolicyReasonAccepted,
				"BackendTrafficPolicy has been accepted.",
			)
		}
	}

	// Process the policies targeting Gateways with no section name set
	for _, policy := range backendTrafficPolicies {
		if policy.Spec.TargetRef.SectionName == nil && policy.Spec.TargetRef.Kind == KindGateway {

			policy := policy.DeepCopy()
			res = append(res, policy)

			// Negative statuses have already been assigned so its safe to skip
			gatewayKey := resolveBTPolicyTargetRef(policy, gatewayMap, routeMap)
			if gatewayKey == nil {
				continue
			}

			// Check if another policy targeting the same Gateway exists
			s, ok := targetedGateways[*gatewayKey]
			if ok && s.Has(AllSections) {
				message := "Unable to target Gateway, another BackendTrafficPolicy has already attached to it"

				status.SetBackendTrafficPolicyCondition(policy,
					gwv1a2.PolicyConditionAccepted,
					metav1.ConditionFalse,
					gwv1a2.PolicyReasonConflicted,
					message,
				)

				continue

			}

			if ok && (s.Len() > 0) {
				// Maintain order here to ensure status/string does not change with same data
				sections := s.UnsortedList()
				sort.Strings(sections)
				message := fmt.Sprintf("There are existing BackendTrafficPolicies that are overriding these sections %v", sections)

				status.SetBackendTrafficPolicyCondition(policy,
					egv1a1.PolicyConditionOverridden,
					metav1.ConditionTrue,
					egv1a1.PolicyReasonOverridden,
					message,
				)
			}

			// Add section to targeted gateways map
			if s == nil {
				targetedGateways[*gatewayKey] = sets.New[string]()
			}
			targetedGateways[*gatewayKey].Insert(AllSections)

			translateBackendTrafficPolicy(policy, xdsIR)

			// Set Accepted=True
			status.SetBackendTrafficPolicyCondition(policy,
				gwv1a2.PolicyConditionAccepted,
				metav1.ConditionTrue,
				gwv1a2.PolicyReasonAccepted,
				"BackendTrafficPolicy has been accepted.",
			)
		}
	}

	return res
}

func resolveBTPolicyTargetRef(policy *egv1a1.BackendTrafficPolicy, gateways map[types.NamespacedName]*GatewayContext, routes map[types.NamespacedName]RouteContext) *types.NamespacedName {
	targetNs := policy.Spec.TargetRef.Namespace
	// If empty, default to namespace of policy
	if targetNs == nil {
		targetNs = ptr.To(gwv1b1.Namespace(policy.Namespace))
	}

	// Ensure Policy and target are in the same namespace
	if policy.Namespace != string(*targetNs) {

		message := fmt.Sprintf("Namespace:%s TargetRef.Namespace:%s, BackendTrafficPolicy can only target a resource in the same namespace.",
			policy.Namespace, *targetNs)
		status.SetBackendTrafficPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwv1a2.PolicyReasonInvalid,
			message,
		)
		return nil
	}

	// Check if it targets a Gateway, if not then it must target an xRoute.
	// CRD Validation covers all the possible route types, so no need to check them explicitly here.
	if policy.Spec.TargetRef.Kind == KindGateway {
		// Find the Gateway
		var gateway *GatewayContext
		gateway, ok := gateways[types.NamespacedName{Name: string(policy.Spec.TargetRef.Name), Namespace: string(*targetNs)}]

		// Gateway not found
		if !ok {
			message := fmt.Sprintf("Gateway:%s not found.", policy.Spec.TargetRef.Name)

			status.SetBackendTrafficPolicyCondition(policy,
				gwv1a2.PolicyConditionAccepted,
				metav1.ConditionFalse,
				gwv1a2.PolicyReasonTargetNotFound,
				message,
			)
			return nil
		}

		// If sectionName is set, make sure its valid
		if policy.Spec.TargetRef.SectionName != nil {
			found := false
			for _, l := range gateway.Spec.Listeners {
				if l.Name == *(policy.Spec.TargetRef.SectionName) {
					found = true
					break
				}
			}
			if !found {
				message := fmt.Sprintf("SectionName(Listener):%s not found.", *(policy.Spec.TargetRef.SectionName))
				status.SetBackendTrafficPolicyCondition(policy,
					gwv1a2.PolicyConditionAccepted,
					metav1.ConditionFalse,
					gwv1a2.PolicyReasonTargetNotFound,
					message,
				)
				return nil
			}
		}

		return &types.NamespacedName{
			Name:      gateway.Name,
			Namespace: string(*targetNs),
		}
	}

	// Policy is targeting an xRoute
	// Check if the route exists
	_, ok := routes[types.NamespacedName{Name: string(policy.Spec.TargetRef.Name), Namespace: string(*targetNs)}]

	// Route not found
	if !ok {
		message := fmt.Sprintf("%s:%s not found.", policy.Spec.TargetRef.Kind, policy.Spec.TargetRef.Name)

		status.SetBackendTrafficPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwv1a2.PolicyReasonTargetNotFound,
			message,
		)
		return nil
	}

	return &types.NamespacedName{
		Name:      string(policy.Spec.TargetRef.Name),
		Namespace: string(*targetNs),
	}
}

func translateBackendTrafficPolicy(policy *egv1a1.BackendTrafficPolicy, xdsIR XdsIRMap) {
	// TODO
}
