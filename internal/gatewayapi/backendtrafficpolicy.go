// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"net"
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/status"
	"github.com/envoyproxy/gateway/internal/utils/ptr"
)

type policyTargetRouteKey struct {
	Kind      string
	Namespace string
	Name      string
}

type policyRouteTargetContext struct {
	RouteContext
	attached bool
}

type policyGatewayTargetContext struct {
	*GatewayContext
	attached bool
}

func (t *Translator) ProcessBackendTrafficPolicies(backendTrafficPolicies []*egv1a1.BackendTrafficPolicy,
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
	for _, policy := range backendTrafficPolicies {
		if policy.Spec.TargetRef.Kind != KindGateway {
			policy := policy.DeepCopy()
			res = append(res, policy)

			// Negative statuses have already been assigned so its safe to skip
			route := resolveBTPolicyRouteTargetRef(policy, routeMap)
			if route == nil {
				continue
			}

			t.translateBackendTrafficPolicyForRoute(policy, route, xdsIR)

			message := "BackendTrafficPolicy has been accepted."
			status.SetBackendTrafficPolicyAcceptedIfUnset(&policy.Status, message)
		}
	}

	// Process the policies targeting Gateways
	for _, policy := range backendTrafficPolicies {
		if policy.Spec.TargetRef.Kind == KindGateway {
			policy := policy.DeepCopy()
			res = append(res, policy)

			// Negative statuses have already been assigned so its safe to skip
			gateway := resolveBTPolicyGatewayTargetRef(policy, gatewayMap)
			if gateway == nil {
				continue
			}

			t.translateBackendTrafficPolicyForGateway(policy, gateway, xdsIR)

			message := "BackendTrafficPolicy has been accepted."
			status.SetBackendTrafficPolicyAcceptedIfUnset(&policy.Status, message)
		}
	}

	return res
}

func resolveBTPolicyGatewayTargetRef(policy *egv1a1.BackendTrafficPolicy, gateways map[types.NamespacedName]*policyGatewayTargetContext) *GatewayContext {
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

	// Find the Gateway
	key := types.NamespacedName{
		Name:      string(policy.Spec.TargetRef.Name),
		Namespace: string(*targetNs),
	}
	gateway, ok := gateways[key]

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

	// Check if another policy targeting the same Gateway exists
	if gateway.attached {
		message := "Unable to target Gateway, another BackendTrafficPolicy has already attached to it"

		status.SetBackendTrafficPolicyCondition(policy,
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

func resolveBTPolicyRouteTargetRef(policy *egv1a1.BackendTrafficPolicy, routes map[policyTargetRouteKey]*policyRouteTargetContext) RouteContext {
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

		status.SetBackendTrafficPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwv1a2.PolicyReasonTargetNotFound,
			message,
		)
		return nil
	}

	// Check if another policy targeting the same xRoute exists
	if route.attached {
		message := fmt.Sprintf("Unable to target %s, another BackendTrafficPolicy has already attached to it", string(policy.Spec.TargetRef.Kind))

		status.SetBackendTrafficPolicyCondition(policy,
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

func (t *Translator) translateBackendTrafficPolicyForRoute(policy *egv1a1.BackendTrafficPolicy, route RouteContext, xdsIR XdsIRMap) {
	// Build IR
	var rl *ir.RateLimit
	if policy.Spec.RateLimit != nil {
		rl = t.buildRateLimit(policy)
	}

	// Apply IR to all relevant routes
	prefix := irRoutePrefix(route)
	for _, ir := range xdsIR {
		for _, http := range ir.HTTP {
			for _, r := range http.Routes {
				// Apply if there is a match
				if strings.HasPrefix(r.Name, prefix) {
					r.RateLimit = rl
				}
			}
		}

	}
}

func (t *Translator) translateBackendTrafficPolicyForGateway(policy *egv1a1.BackendTrafficPolicy, gateway *GatewayContext, xdsIR XdsIRMap) {
	// Build IR
	var rl *ir.RateLimit
	if policy.Spec.RateLimit != nil {
		rl = t.buildRateLimit(policy)
	}

	// Apply IR to all the routes within the specific Gateway
	// If the feature is already set, then skip it, since it must be have
	// set by a policy attaching to the route
	irKey := t.getIRKey(gateway.Gateway)
	// Should exist since we've validated this
	ir := xdsIR[irKey]

	for _, http := range ir.HTTP {
		for _, r := range http.Routes {
			// Apply if not already set
			if r.RateLimit == nil {
				r.RateLimit = rl
			}
		}
	}

}

func (t *Translator) buildRateLimit(policy *egv1a1.BackendTrafficPolicy) *ir.RateLimit {
	if policy.Spec.RateLimit.Global == nil {
		message := "Global configuration empty for rateLimit"
		status.SetBackendTrafficPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwv1a2.PolicyReasonInvalid,
			message,
		)
		return nil
	}
	if !t.GlobalRateLimitEnabled {
		message := "Enable Ratelimit in the EnvoyGateway config to configure global rateLimit"
		status.SetBackendTrafficPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwv1a2.PolicyReasonInvalid,
			message,
		)
		return nil
	}
	rateLimit := &ir.RateLimit{
		Global: &ir.GlobalRateLimit{
			Rules: make([]*ir.RateLimitRule, len(policy.Spec.RateLimit.Global.Rules)),
		},
	}

	rules := rateLimit.Global.Rules
	for i, rule := range policy.Spec.RateLimit.Global.Rules {
		rules[i] = &ir.RateLimitRule{
			Limit: &ir.RateLimitValue{
				Requests: rule.Limit.Requests,
				Unit:     ir.RateLimitUnit(rule.Limit.Unit),
			},
			HeaderMatches: make([]*ir.StringMatch, 0),
		}
		for _, match := range rule.ClientSelectors {
			for _, header := range match.Headers {
				switch {
				case header.Type == nil && header.Value != nil:
					fallthrough
				case *header.Type == egv1a1.HeaderMatchExact && header.Value != nil:
					m := &ir.StringMatch{
						Name:  header.Name,
						Exact: header.Value,
					}
					rules[i].HeaderMatches = append(rules[i].HeaderMatches, m)
				case *header.Type == egv1a1.HeaderMatchRegularExpression && header.Value != nil:
					m := &ir.StringMatch{
						Name:      header.Name,
						SafeRegex: header.Value,
					}
					rules[i].HeaderMatches = append(rules[i].HeaderMatches, m)
				case *header.Type == egv1a1.HeaderMatchDistinct && header.Value == nil:
					m := &ir.StringMatch{
						Name:     header.Name,
						Distinct: true,
					}
					rules[i].HeaderMatches = append(rules[i].HeaderMatches, m)
				default:
					// set negative status condition.
					message := "Unable to translate rateLimit. Either the header.Type is not valid or the header is missing a value"
					status.SetBackendTrafficPolicyCondition(policy,
						gwv1a2.PolicyConditionAccepted,
						metav1.ConditionFalse,
						gwv1a2.PolicyReasonInvalid,
						message,
					)

					return nil
				}
			}

			if match.SourceCIDR != nil {
				// distinct means that each IP Address within the specified Source IP CIDR is treated as a
				// distinct client selector and uses a separate rate limit bucket/counter.
				distinct := false
				sourceCIDR := match.SourceCIDR.Value
				if match.SourceCIDR.Type != nil && *match.SourceCIDR.Type == egv1a1.SourceMatchDistinct {
					distinct = true
				}

				ip, ipn, err := net.ParseCIDR(sourceCIDR)
				if err != nil {
					message := "Unable to translate rateLimit"
					status.SetBackendTrafficPolicyCondition(policy,
						gwv1a2.PolicyConditionAccepted,
						metav1.ConditionFalse,
						gwv1a2.PolicyReasonInvalid,
						message,
					)

					return nil
				}

				mask, _ := ipn.Mask.Size()
				rules[i].CIDRMatch = &ir.CIDRMatch{
					CIDR:     ipn.String(),
					IPv6:     ip.To4() == nil,
					MaskLen:  mask,
					Distinct: distinct,
				}
			}
		}
	}

	return rateLimit
}
