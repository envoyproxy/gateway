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
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/status"
	"github.com/envoyproxy/gateway/internal/utils/ptr"
)

func ProcessSecurityPolicies(securityPolicies []*egv1a1.SecurityPolicy,
	gateways []*GatewayContext,
	routes []RouteContext,
	xdsIR XdsIRMap) []*egv1a1.SecurityPolicy {
	var res []*egv1a1.SecurityPolicy

	// Sort based on timestamp
	sort.Slice(securityPolicies, func(i, j int) bool {
		return securityPolicies[i].CreationTimestamp.Before(&(securityPolicies[j].CreationTimestamp))
	})

	// First build a map out of the routes and gateways for faster lookup since users might have thousands of routes or more.
	// For gateways this probably isn't quite as necessary.
	routeMap := map[policyTargetRouteKey]*policyRouteTargetContext{}
	for _, route := range routes {
		key := policyTargetRouteKey{
			Kind:      string(GetRouteType(route)),
			Name:      route.GetName(),
			Namespace: route.GetNamespace(),
		}
		routeMap[key] = &policyRouteTargetContext{RouteContext: route}
	}
	gatewayMap := map[types.NamespacedName]*policyGatewayTargetContext{}
	for _, gw := range gateways {
		key := types.NamespacedName{
			Name:      gw.GetName(),
			Namespace: gw.GetNamespace(),
		}
		gatewayMap[key] = &policyGatewayTargetContext{GatewayContext: gw}
	}

	// Translate
	// 1. First translate Policies targeting xRoutes
	// 2.. Finally, the policies targeting Gateways

	// Process the policies targeting xRoutes
	for _, policy := range securityPolicies {
		if policy.Spec.TargetRef.Kind != KindGateway {
			policy := policy.DeepCopy()
			res = append(res, policy)

			// Negative statuses have already been assigned so its safe to skip
			route := resolveSecurityPolicyRouteTargetRef(policy, routeMap)
			if route == nil {
				continue
			}

			translateSecurityPolicy(policy, xdsIR)

			// Set Accepted=True
			status.SetSecurityPolicyCondition(policy,
				gwv1a2.PolicyConditionAccepted,
				metav1.ConditionTrue,
				gwv1a2.PolicyReasonAccepted,
				"SecurityPolicy has been accepted.",
			)
		}
	}

	// Process the policies targeting Gateways with a section name
	for _, policy := range securityPolicies {
		if policy.Spec.TargetRef.Kind == KindGateway {
			policy := policy.DeepCopy()
			res = append(res, policy)

			// Negative statuses have already been assigned so its safe to skip
			gatewayKey := resolveSecurityPolicyGatewayTargetRef(policy, gatewayMap)
			if gatewayKey == nil {
				continue
			}

			translateSecurityPolicy(policy, xdsIR)

			// Set Accepted=True
			status.SetSecurityPolicyCondition(policy,
				gwv1a2.PolicyConditionAccepted,
				metav1.ConditionTrue,
				gwv1a2.PolicyReasonAccepted,
				"SecurityPolicy has been accepted.",
			)
		}
	}

	return res
}

func resolveSecurityPolicyGatewayTargetRef(policy *egv1a1.SecurityPolicy, gateways map[types.NamespacedName]*policyGatewayTargetContext) *GatewayContext {
	targetNs := policy.Spec.TargetRef.Namespace
	// If empty, default to namespace of policy
	if targetNs == nil {
		targetNs = ptr.To(gwv1b1.Namespace(policy.Namespace))
	}

	// Ensure Policy and target are in the same namespace
	if policy.Namespace != string(*targetNs) {

		message := fmt.Sprintf("Namespace:%s TargetRef.Namespace:%s, SecurityPolicy can only target a resource in the same namespace.",
			policy.Namespace, *targetNs)
		status.SetSecurityPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwv1a2.PolicyReasonInvalid,
			message,
		)
		return nil
	}

	// Find the Gateway
	key := types.NamespacedName{
		Name:      string(policy.Spec.TargetRef.Name),
		Namespace: string(*targetNs),
	}
	gateway, ok := gateways[key]

	// Gateway not found
	if !ok {
		message := fmt.Sprintf("Gateway:%s not found.", policy.Spec.TargetRef.Name)

		status.SetSecurityPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwv1a2.PolicyReasonTargetNotFound,
			message,
		)
		return nil
	}

	// Check if another policy targeting the same Gateway exists
	if gateway.attached {
		message := "Unable to target Gateway, another SecurityPolicy has already attached to it"

		status.SetSecurityPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwv1a2.PolicyReasonConflicted,
			message,
		)
		return nil
	}

	// Set context and save
	gateway.attached = true
	gateways[key] = gateway

	return gateway.GatewayContext
}

func resolveSecurityPolicyRouteTargetRef(policy *egv1a1.SecurityPolicy, routes map[policyTargetRouteKey]*policyRouteTargetContext) RouteContext {
	targetNs := policy.Spec.TargetRef.Namespace
	// If empty, default to namespace of policy
	if targetNs == nil {
		targetNs = ptr.To(gwv1b1.Namespace(policy.Namespace))
	}

	// Ensure Policy and target are in the same namespace
	if policy.Namespace != string(*targetNs) {

		message := fmt.Sprintf("Namespace:%s TargetRef.Namespace:%s, SecurityPolicy can only target a resource in the same namespace.",
			policy.Namespace, *targetNs)
		status.SetSecurityPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwv1a2.PolicyReasonInvalid,
			message,
		)
		return nil
	}

	// Check if the route exists
	key := policyTargetRouteKey{
		Kind:      string(policy.Spec.TargetRef.Kind),
		Name:      string(policy.Spec.TargetRef.Name),
		Namespace: string(*targetNs),
	}
	route, ok := routes[key]

	// Route not found
	if !ok {
		message := fmt.Sprintf("%s/%s/%s not found.", policy.Spec.TargetRef.Kind, string(*targetNs), policy.Spec.TargetRef.Name)

		status.SetSecurityPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwv1a2.PolicyReasonTargetNotFound,
			message,
		)
		return nil
	}

	// Check if another policy targeting the same xRoute exists
	if route.attached {
		message := fmt.Sprintf("Unable to target %s, another SecurityPolicy has already attached to it", string(policy.Spec.TargetRef.Kind))

		status.SetSecurityPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwv1a2.PolicyReasonConflicted,
			message,
		)
		return nil
	}

	// Set context and save
	route.attached = true
	routes[key] = route

	return route.RouteContext
}
func translateSecurityPolicy(policy *egv1a1.SecurityPolicy, xdsIR XdsIRMap) {
}
