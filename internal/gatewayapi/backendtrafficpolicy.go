// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"
	"math"
	"net"
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/status"
	"github.com/envoyproxy/gateway/internal/utils/regex"
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
	var (
		rl *ir.RateLimit
		lb *ir.LoadBalancer
		pp *ir.ProxyProtocol
		cb *ir.CircuitBreaker
		fi *ir.FaultInjection
	)

	// Build IR
	if policy.Spec.RateLimit != nil {
		rl = t.buildRateLimit(policy)
	}
	if policy.Spec.LoadBalancer != nil {
		lb = t.buildLoadBalancer(policy)
	}
	if policy.Spec.ProxyProtocol != nil {
		pp = t.buildProxyProtocol(policy)
	}
	if policy.Spec.CircuitBreaker != nil {
		cb = t.buildCircuitBreaker(policy)
	}

	if policy.Spec.FaultInjection != nil {
		fi = t.buildFaultInjection(policy)
	}
	// Apply IR to all relevant routes
	prefix := irRoutePrefix(route)
	for _, ir := range xdsIR {
		for _, http := range ir.HTTP {
			for _, r := range http.Routes {
				// Apply if there is a match
				if strings.HasPrefix(r.Name, prefix) {
					r.RateLimit = rl
					r.LoadBalancer = lb
					r.ProxyProtocol = pp
					r.CircuitBreaker = cb
					r.FaultInjection = fi
				}
			}
		}

	}
}

func (t *Translator) translateBackendTrafficPolicyForGateway(policy *egv1a1.BackendTrafficPolicy, gateway *GatewayContext, xdsIR XdsIRMap) {
	var (
		rl *ir.RateLimit
		lb *ir.LoadBalancer
		pp *ir.ProxyProtocol
		cb *ir.CircuitBreaker
		fi *ir.FaultInjection
	)

	// Build IR
	if policy.Spec.RateLimit != nil {
		rl = t.buildRateLimit(policy)
	}
	if policy.Spec.LoadBalancer != nil {
		lb = t.buildLoadBalancer(policy)
	}
	if policy.Spec.ProxyProtocol != nil {
		pp = t.buildProxyProtocol(policy)
	}
	if policy.Spec.CircuitBreaker != nil {
		cb = t.buildCircuitBreaker(policy)
	}
	if policy.Spec.FaultInjection != nil {
		fi = t.buildFaultInjection(policy)
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
			if r.LoadBalancer == nil {
				r.LoadBalancer = lb
			}
			if r.ProxyProtocol == nil {
				r.ProxyProtocol = pp
			}
			if r.CircuitBreaker == nil {
				r.CircuitBreaker = cb
			}
			if r.FaultInjection == nil {
				r.FaultInjection = fi
			}
		}

	}
}

func (t *Translator) buildRateLimit(policy *egv1a1.BackendTrafficPolicy) *ir.RateLimit {
	switch policy.Spec.RateLimit.Type {
	case egv1a1.GlobalRateLimitType:
		return t.buildGlobalRateLimit(policy)
	case egv1a1.LocalRateLimitType:
		return t.buildLocalRateLimit(policy)
	}

	status.SetBackendTrafficPolicyCondition(policy,
		gwv1a2.PolicyConditionAccepted,
		metav1.ConditionFalse,
		gwv1a2.PolicyReasonInvalid,
		"Invalid rateLimit type",
	)
	return nil
}

func (t *Translator) buildLocalRateLimit(policy *egv1a1.BackendTrafficPolicy) *ir.RateLimit {
	if policy.Spec.RateLimit.Local == nil {
		message := "Local configuration empty for rateLimit."
		status.SetBackendTrafficPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwv1a2.PolicyReasonInvalid,
			message,
		)
		return nil
	}

	local := policy.Spec.RateLimit.Local

	// Envoy local rateLimit requires a default limit to be set for a route.
	// EG uses the first rule without clientSelectors as the default route-level
	// limit. If no such rule is found, EG uses a default limit of uint32 max.
	var defaultLimit *ir.RateLimitValue
	for _, rule := range local.Rules {
		if rule.ClientSelectors == nil || len(rule.ClientSelectors) == 0 {
			if defaultLimit != nil {
				message := "Local rateLimit can not have more than one rule without clientSelectors."
				status.SetBackendTrafficPolicyCondition(policy,
					gwv1a2.PolicyConditionAccepted,
					metav1.ConditionFalse,
					gwv1a2.PolicyReasonInvalid,
					message,
				)
				return nil
			}
			defaultLimit = &ir.RateLimitValue{
				Requests: rule.Limit.Requests,
				Unit:     ir.RateLimitUnit(rule.Limit.Unit),
			}
		}
	}
	// If no rule without clientSelectors is found, use uint32 max as the default
	// limit, which effectively make the default limit unlimited.
	if defaultLimit == nil {
		defaultLimit = &ir.RateLimitValue{
			Requests: math.MaxUint32,
			Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
		}
	}

	// Validate that the rule limit unit is a multiple of the default limit unit.
	// This is required by Envoy local rateLimit implementation.
	// see https://github.com/envoyproxy/envoy/blob/6d9a6e995f472526de2b75233abca69aa00021ed/source/extensions/filters/common/local_ratelimit/local_ratelimit_impl.cc#L49
	defaultLimitUnit := ratelimitUnitToDuration(egv1a1.RateLimitUnit(defaultLimit.Unit))
	for _, rule := range local.Rules {
		ruleLimitUint := ratelimitUnitToDuration(rule.Limit.Unit)
		if defaultLimitUnit == 0 || ruleLimitUint%defaultLimitUnit != 0 {
			message := "Local rateLimit rule limit unit must be a multiple of the default limit unit."
			status.SetBackendTrafficPolicyCondition(policy,
				gwv1a2.PolicyConditionAccepted,
				metav1.ConditionFalse,
				gwv1a2.PolicyReasonInvalid,
				message,
			)
			return nil
		}
	}

	var err error
	var irRule *ir.RateLimitRule
	var irRules = make([]*ir.RateLimitRule, 0)
	for _, rule := range local.Rules {
		// We don't process the rule without clientSelectors here because it's
		// previously used as the default route-level limit.
		if len(rule.ClientSelectors) == 0 {
			continue
		}

		irRule, err = buildRateLimitRule(rule)
		if err != nil {
			status.SetBackendTrafficPolicyCondition(policy,
				gwv1a2.PolicyConditionAccepted,
				metav1.ConditionFalse,
				gwv1a2.PolicyReasonInvalid,
				status.Error2ConditionMsg(err),
			)
			return nil
		}
		if irRule.CIDRMatch != nil && irRule.CIDRMatch.Distinct {
			message := "Local rateLimit does not support distinct CIDRMatch."
			status.SetBackendTrafficPolicyCondition(policy,
				gwv1a2.PolicyConditionAccepted,
				metav1.ConditionFalse,
				gwv1a2.PolicyReasonInvalid,
				message,
			)
			return nil
		}
		for _, match := range irRule.HeaderMatches {
			if match.Distinct {
				message := "Local rateLimit does not support distinct HeaderMatch."
				status.SetBackendTrafficPolicyCondition(policy,
					gwv1a2.PolicyConditionAccepted,
					metav1.ConditionFalse,
					gwv1a2.PolicyReasonInvalid,
					message,
				)
				return nil
			}
		}
		irRules = append(irRules, irRule)
	}

	rateLimit := &ir.RateLimit{
		Local: &ir.LocalRateLimit{
			Default: *defaultLimit,
			Rules:   irRules,
		},
	}
	return rateLimit
}

func (t *Translator) buildGlobalRateLimit(policy *egv1a1.BackendTrafficPolicy) *ir.RateLimit {
	if policy.Spec.RateLimit.Global == nil {
		message := "Global configuration empty for rateLimit."
		status.SetBackendTrafficPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwv1a2.PolicyReasonInvalid,
			message,
		)
		return nil
	}

	if !t.GlobalRateLimitEnabled {
		message := "Enable Ratelimit in the EnvoyGateway config to configure global rateLimit."
		status.SetBackendTrafficPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwv1a2.PolicyReasonInvalid,
			message,
		)
		return nil
	}

	global := policy.Spec.RateLimit.Global
	rateLimit := &ir.RateLimit{
		Global: &ir.GlobalRateLimit{
			Rules: make([]*ir.RateLimitRule, len(global.Rules)),
		},
	}

	irRules := rateLimit.Global.Rules
	var err error
	for i, rule := range global.Rules {
		irRules[i], err = buildRateLimitRule(rule)
		if err != nil {
			status.SetBackendTrafficPolicyCondition(policy,
				gwv1a2.PolicyConditionAccepted,
				metav1.ConditionFalse,
				gwv1a2.PolicyReasonInvalid,
				status.Error2ConditionMsg(err),
			)
			return nil
		}
	}

	return rateLimit
}

func buildRateLimitRule(rule egv1a1.RateLimitRule) (*ir.RateLimitRule, error) {
	irRule := &ir.RateLimitRule{
		Limit: ir.RateLimitValue{
			Requests: rule.Limit.Requests,
			Unit:     ir.RateLimitUnit(rule.Limit.Unit),
		},
		HeaderMatches: make([]*ir.StringMatch, 0),
	}

	for _, match := range rule.ClientSelectors {
		if len(match.Headers) == 0 && match.SourceCIDR == nil {
			return nil, errors.New(
				"unable to translate rateLimit. At least one of the" +
					" header or sourceCIDR must be specified")
		}
		for _, header := range match.Headers {
			switch {
			case header.Type == nil && header.Value != nil:
				fallthrough
			case *header.Type == egv1a1.HeaderMatchExact && header.Value != nil:
				m := &ir.StringMatch{
					Name:  header.Name,
					Exact: header.Value,
				}
				irRule.HeaderMatches = append(irRule.HeaderMatches, m)
			case *header.Type == egv1a1.HeaderMatchRegularExpression && header.Value != nil:
				if err := regex.Validate(*header.Value); err != nil {
					return nil, err
				}
				m := &ir.StringMatch{
					Name:      header.Name,
					SafeRegex: header.Value,
				}
				irRule.HeaderMatches = append(irRule.HeaderMatches, m)
			case *header.Type == egv1a1.HeaderMatchDistinct && header.Value == nil:
				m := &ir.StringMatch{
					Name:     header.Name,
					Distinct: true,
				}
				irRule.HeaderMatches = append(irRule.HeaderMatches, m)
			default:
				return nil, errors.New(
					"unable to translate rateLimit. Either the header." +
						"Type is not valid or the header is missing a value")
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
				return nil, errors.New("unable to translate rateLimit")
			}

			mask, _ := ipn.Mask.Size()
			irRule.CIDRMatch = &ir.CIDRMatch{
				CIDR:     ipn.String(),
				IPv6:     ip.To4() == nil,
				MaskLen:  mask,
				Distinct: distinct,
			}
		}
	}
	return irRule, nil
}

func (t *Translator) buildLoadBalancer(policy *egv1a1.BackendTrafficPolicy) *ir.LoadBalancer {
	var lb *ir.LoadBalancer
	switch policy.Spec.LoadBalancer.Type {
	case egv1a1.ConsistentHashLoadBalancerType:
		lb = &ir.LoadBalancer{
			ConsistentHash: &ir.ConsistentHash{},
		}
		if policy.Spec.LoadBalancer.ConsistentHash != nil &&
			policy.Spec.LoadBalancer.ConsistentHash.Type == egv1a1.SourceIPConsistentHashType {
			lb.ConsistentHash.SourceIP = ptr.To(true)
		}
	case egv1a1.LeastRequestLoadBalancerType:
		lb = &ir.LoadBalancer{}
		if policy.Spec.LoadBalancer.SlowStart != nil {
			if policy.Spec.LoadBalancer.SlowStart.Window != nil {
				lb.LeastRequest = &ir.LeastRequest{
					SlowStart: &ir.SlowStart{
						Window: policy.Spec.LoadBalancer.SlowStart.Window,
					},
				}
			}
		}
	case egv1a1.RandomLoadBalancerType:
		lb = &ir.LoadBalancer{
			Random: &ir.Random{},
		}
	case egv1a1.RoundRobinLoadBalancerType:
		lb = &ir.LoadBalancer{
			RoundRobin: &ir.RoundRobin{
				SlowStart: &ir.SlowStart{},
			},
		}
		if policy.Spec.LoadBalancer.SlowStart != nil {
			if policy.Spec.LoadBalancer.SlowStart.Window != nil {
				lb.RoundRobin = &ir.RoundRobin{
					SlowStart: &ir.SlowStart{
						Window: policy.Spec.LoadBalancer.SlowStart.Window,
					},
				}
			}
		}
	}

	return lb
}

func (t *Translator) buildProxyProtocol(policy *egv1a1.BackendTrafficPolicy) *ir.ProxyProtocol {
	var pp *ir.ProxyProtocol
	switch policy.Spec.ProxyProtocol.Version {
	case egv1a1.ProxyProtocolVersionV1:
		pp = &ir.ProxyProtocol{
			Version: ir.ProxyProtocolVersionV1,
		}
	case egv1a1.ProxyProtocolVersionV2:
		pp = &ir.ProxyProtocol{
			Version: ir.ProxyProtocolVersionV2,
		}
	}

	return pp
}

func ratelimitUnitToDuration(unit egv1a1.RateLimitUnit) int64 {
	var seconds int64

	switch unit {
	case egv1a1.RateLimitUnitSecond:
		seconds = 1
	case egv1a1.RateLimitUnitMinute:
		seconds = 60
	case egv1a1.RateLimitUnitHour:
		seconds = 60 * 60
	case egv1a1.RateLimitUnitDay:
		seconds = 60 * 60 * 24
	}
	return seconds
}

func (t *Translator) buildCircuitBreaker(policy *egv1a1.BackendTrafficPolicy) *ir.CircuitBreaker {
	var cb *ir.CircuitBreaker
	pcb := policy.Spec.CircuitBreaker

	if pcb != nil {
		cb = &ir.CircuitBreaker{}

		if pcb.MaxConnections != nil {
			if ui32, ok := int64ToUint32(*pcb.MaxConnections); ok {
				cb.MaxConnections = &ui32
			} else {
				setCircuitBreakerPolicyErrorCondition(policy, fmt.Sprintf("invalid MaxConnections value %d", *pcb.MaxConnections))
				return nil
			}
		}

		if pcb.MaxParallelRequests != nil {
			if ui32, ok := int64ToUint32(*pcb.MaxParallelRequests); ok {
				cb.MaxParallelRequests = &ui32
			} else {
				setCircuitBreakerPolicyErrorCondition(policy, fmt.Sprintf("invalid MaxParallelRequests value %d", *pcb.MaxParallelRequests))
				return nil
			}
		}

		if pcb.MaxPendingRequests != nil {
			if ui32, ok := int64ToUint32(*pcb.MaxPendingRequests); ok {
				cb.MaxPendingRequests = &ui32
			} else {
				setCircuitBreakerPolicyErrorCondition(policy, fmt.Sprintf("invalid MaxPendingRequests value %d", *pcb.MaxPendingRequests))
				return nil
			}
		}
	}

	return cb
}

func setCircuitBreakerPolicyErrorCondition(policy *egv1a1.BackendTrafficPolicy, errMsg string) {
	message := fmt.Sprintf("Unable to translate Circuit Breaker: %s", errMsg)
	status.SetBackendTrafficPolicyCondition(policy,
		gwv1a2.PolicyConditionAccepted,
		metav1.ConditionFalse,
		gwv1a2.PolicyReasonInvalid,
		message,
	)
}

func int64ToUint32(in int64) (uint32, bool) {
	if in >= 0 && in <= math.MaxUint32 {
		return uint32(in), true
	}
	return 0, false
}

func (t *Translator) buildFaultInjection(policy *egv1a1.BackendTrafficPolicy) *ir.FaultInjection {
	var fi *ir.FaultInjection
	if policy.Spec.FaultInjection != nil {
		fi = &ir.FaultInjection{}

		if policy.Spec.FaultInjection.Delay != nil {
			fi.Delay = &ir.FaultInjectionDelay{
				Percentage: policy.Spec.FaultInjection.Delay.Percentage,
				FixedDelay: policy.Spec.FaultInjection.Delay.FixedDelay,
			}
		}
		if policy.Spec.FaultInjection.Abort != nil {
			fi.Abort = &ir.FaultInjectionAbort{
				Percentage: policy.Spec.FaultInjection.Abort.Percentage,
			}

			if policy.Spec.FaultInjection.Abort.GrpcStatus != nil {
				fi.Abort.GrpcStatus = policy.Spec.FaultInjection.Abort.GrpcStatus
			}
			if policy.Spec.FaultInjection.Abort.HTTPStatus != nil {
				fi.Abort.HTTPStatus = policy.Spec.FaultInjection.Abort.HTTPStatus
			}
		}
	}
	return fi
}
