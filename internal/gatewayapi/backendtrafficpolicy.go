// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"
	"time"

	perr "github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/utils/regex"
)

const (
	MaxConsistentHashTableSize = 5000011 // https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/cluster.proto#config-cluster-v3-cluster-maglevlbconfig
)

func (t *Translator) ProcessBackendTrafficPolicies(backendTrafficPolicies []*egv1a1.BackendTrafficPolicy,
	gateways []*GatewayContext,
	routes []RouteContext,
	xdsIR XdsIRMap,
) []*egv1a1.BackendTrafficPolicy {
	res := []*egv1a1.BackendTrafficPolicy{}

	// Sort based on timestamp
	sort.Slice(backendTrafficPolicies, func(i, j int) bool {
		return backendTrafficPolicies[i].CreationTimestamp.Before(&(backendTrafficPolicies[j].CreationTimestamp))
	})

	// First build a map out of the routes and gateways for faster lookup since users might have thousands of routes or more.
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
		key := utils.NamespacedName(gw)
		gatewayMap[key] = &policyGatewayTargetContext{GatewayContext: gw}
	}

	// Map of Gateway to the routes attached to it
	gatewayRouteMap := make(map[string]sets.Set[string])

	handledPolicies := make(map[types.NamespacedName]*egv1a1.BackendTrafficPolicy)

	// Translate
	// 1. First translate Policies targeting xRoutes
	// 2. Finally, the policies targeting Gateways

	// Process the policies targeting xRoutes
	for _, currPolicy := range backendTrafficPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, routes)
		for _, currTarget := range targetRefs {
			if currTarget.Kind != KindGateway {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy.DeepCopy()
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}

				// Negative statuses have already been assigned so its safe to skip
				route, resolveErr := resolveBTPolicyRouteTargetRef(policy, currTarget, routeMap)
				if route == nil {
					continue
				}

				// Find the Gateway that the route belongs to and add it to the
				// gatewayRouteMap and ancestor list, which will be used to check
				// policy overrides and populate its ancestor status.
				parentRefs := GetParentReferences(route)
				ancestorRefs := make([]gwapiv1a2.ParentReference, 0, len(parentRefs))
				for _, p := range parentRefs {
					if p.Kind == nil || *p.Kind == KindGateway {
						namespace := route.GetNamespace()
						if p.Namespace != nil {
							namespace = string(*p.Namespace)
						}
						gwNN := types.NamespacedName{
							Namespace: namespace,
							Name:      string(p.Name),
						}

						key := gwNN.String()
						if _, ok := gatewayRouteMap[key]; !ok {
							gatewayRouteMap[key] = make(sets.Set[string])
						}
						gatewayRouteMap[key].Insert(utils.NamespacedName(route).String())

						// Do need a section name since the policy is targeting to a route
						ancestorRefs = append(ancestorRefs, getAncestorRefForPolicy(gwNN, p.SectionName))
					}
				}

				// Set conditions for resolve error, then skip current xroute
				if resolveErr != nil {
					status.SetResolveErrorForPolicyAncestors(&policy.Status,
						ancestorRefs,
						t.GatewayControllerName,
						policy.Generation,
						resolveErr,
					)

					continue
				}

				// Set conditions for translation error if it got any
				if err := t.translateBackendTrafficPolicyForRoute(policy, route, xdsIR); err != nil {
					status.SetTranslationErrorForPolicyAncestors(&policy.Status,
						ancestorRefs,
						t.GatewayControllerName,
						policy.Generation,
						status.Error2ConditionMsg(err),
					)
				}

				// Set Accepted condition if it is unset
				status.SetAcceptedForPolicyAncestors(&policy.Status, ancestorRefs, t.GatewayControllerName)
			}
		}
	}

	// Process the policies targeting Gateways
	for _, currPolicy := range backendTrafficPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, gateways)
		for _, currTarget := range targetRefs {
			if currTarget.Kind == KindGateway {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy.DeepCopy()
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}

				// Negative statuses have already been assigned so its safe to skip
				gateway, resolveErr := resolveBTPolicyGatewayTargetRef(policy, currTarget, gatewayMap)
				if gateway == nil {
					continue
				}

				// Find its ancestor reference by resolved gateway, even with resolve error
				gatewayNN := utils.NamespacedName(gateway)
				ancestorRefs := []gwapiv1a2.ParentReference{
					// Don't need a section name since the policy is targeting to a gateway
					getAncestorRefForPolicy(gatewayNN, nil),
				}

				// Set conditions for resolve error, then skip current gateway
				if resolveErr != nil {
					status.SetResolveErrorForPolicyAncestors(&policy.Status,
						ancestorRefs,
						t.GatewayControllerName,
						policy.Generation,
						resolveErr,
					)

					continue
				}

				// Set conditions for translation error if it got any
				if err := t.translateBackendTrafficPolicyForGateway(policy, currTarget, gateway, xdsIR); err != nil {
					status.SetTranslationErrorForPolicyAncestors(&policy.Status,
						ancestorRefs,
						t.GatewayControllerName,
						policy.Generation,
						status.Error2ConditionMsg(err),
					)
				}

				// Set Accepted condition if it is unset
				status.SetAcceptedForPolicyAncestors(&policy.Status, ancestorRefs, t.GatewayControllerName)

				// Check if this policy is overridden by other policies targeting at
				// route level
				if r, ok := gatewayRouteMap[gatewayNN.String()]; ok {
					// Maintain order here to ensure status/string does not change with the same data
					routes := r.UnsortedList()
					sort.Strings(routes)
					message := fmt.Sprintf("This policy is being overridden by other backendTrafficPolicies for these routes: %v", routes)

					status.SetConditionForPolicyAncestors(&policy.Status,
						ancestorRefs,
						t.GatewayControllerName,
						egv1a1.PolicyConditionOverridden,
						metav1.ConditionTrue,
						egv1a1.PolicyReasonOverridden,
						message,
						policy.Generation,
					)
				}
			}
		}
	}

	return res
}

func resolveBTPolicyGatewayTargetRef(policy *egv1a1.BackendTrafficPolicy, target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName, gateways map[types.NamespacedName]*policyGatewayTargetContext) (*GatewayContext, *status.PolicyResolveError) {
	// Check if the gateway exists
	key := types.NamespacedName{
		Name:      string(target.Name),
		Namespace: policy.Namespace,
	}
	gateway, ok := gateways[key]

	// Gateway not found
	if !ok {
		return nil, nil
	}

	// Check if another policy targeting the same Gateway exists
	if gateway.attached {
		message := fmt.Sprintf("Unable to target Gateway %s, another BackendTrafficPolicy has already attached to it",
			string(target.Name))

		return gateway.GatewayContext, &status.PolicyResolveError{
			Reason:  gwapiv1a2.PolicyReasonConflicted,
			Message: message,
		}
	}

	// Set context and save
	gateway.attached = true
	gateways[key] = gateway

	return gateway.GatewayContext, nil
}

func resolveBTPolicyRouteTargetRef(policy *egv1a1.BackendTrafficPolicy, target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName, routes map[policyTargetRouteKey]*policyRouteTargetContext) (RouteContext, *status.PolicyResolveError) {
	// Check if the route exists
	key := policyTargetRouteKey{
		Kind:      string(target.Kind),
		Name:      string(target.Name),
		Namespace: policy.Namespace,
	}

	route, ok := routes[key]
	// Route not found
	if !ok {
		return nil, nil
	}

	// Check if another policy targeting the same xRoute exists
	if route.attached {
		message := fmt.Sprintf("Unable to target %s %s, another BackendTrafficPolicy has already attached to it",
			string(target.Kind), string(target.Name))

		return route.RouteContext, &status.PolicyResolveError{
			Reason:  gwapiv1a2.PolicyReasonConflicted,
			Message: message,
		}
	}

	// Set context and save
	route.attached = true
	routes[key] = route

	return route.RouteContext, nil
}

func (t *Translator) translateBackendTrafficPolicyForRoute(policy *egv1a1.BackendTrafficPolicy, route RouteContext, xdsIR XdsIRMap) error {
	var (
		rl        *ir.RateLimit
		lb        *ir.LoadBalancer
		pp        *ir.ProxyProtocol
		hc        *ir.HealthCheck
		cb        *ir.CircuitBreaker
		fi        *ir.FaultInjection
		to        *ir.Timeout
		ka        *ir.TCPKeepalive
		rt        *ir.Retry
		bc        *ir.BackendConnection
		ds        *ir.DNS
		err, errs error
	)

	// Build IR
	if policy.Spec.RateLimit != nil {
		if rl, err = t.buildRateLimit(policy); err != nil {
			err = perr.WithMessage(err, "RateLimit")
			errs = errors.Join(errs, err)
		}
	}
	if policy.Spec.LoadBalancer != nil {
		if lb, err = t.buildLoadBalancer(policy); err != nil {
			err = perr.WithMessage(err, "LoadBalancer")
			errs = errors.Join(errs, err)
		}
	}
	if policy.Spec.ProxyProtocol != nil {
		pp = t.buildProxyProtocol(policy)
	}
	if policy.Spec.HealthCheck != nil {
		hc = t.buildHealthCheck(policy)
	}
	if policy.Spec.CircuitBreaker != nil {
		if cb, err = t.buildCircuitBreaker(policy); err != nil {
			err = perr.WithMessage(err, "CircuitBreaker")
			errs = errors.Join(errs, err)
		}
	}
	if policy.Spec.FaultInjection != nil {
		fi = t.buildFaultInjection(policy)
	}
	if policy.Spec.TCPKeepalive != nil {
		if ka, err = t.buildTCPKeepAlive(policy); err != nil {
			err = perr.WithMessage(err, "TCPKeepalive")
			errs = errors.Join(errs, err)
		}
	}
	if policy.Spec.Retry != nil {
		rt = t.buildRetry(policy)
	}
	if policy.Spec.Timeout != nil {
		if to, err = t.buildTimeout(policy, nil); err != nil {
			err = perr.WithMessage(err, "Timeout")
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.Connection != nil {
		if bc, err = t.buildBackendConnection(policy); err != nil {
			err = perr.WithMessage(err, "BackendConnection")
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.DNS != nil {
		ds = t.translateDNS(policy)
	}

	// Early return if got any errors
	if errs != nil {
		return errs
	}

	// Apply IR to all relevant routes
	prefix := irRoutePrefix(route)

	for _, x := range xdsIR {
		for _, tcp := range x.TCP {
			for _, r := range tcp.Routes {
				if strings.HasPrefix(r.Destination.Name, prefix) {
					r.LoadBalancer = lb
					r.ProxyProtocol = pp
					r.HealthCheck = hc
					r.CircuitBreaker = cb
					r.TCPKeepalive = ka
					r.Timeout = to
					r.BackendConnection = bc
					r.DNS = ds
				}
			}
		}

		for _, udp := range x.UDP {
			if udp.Route != nil {
				r := udp.Route

				if strings.HasPrefix(r.Destination.Name, prefix) {
					r.LoadBalancer = lb
					r.Timeout = to
					r.BackendConnection = bc
					r.DNS = ds
				}
			}
		}

		for _, http := range x.HTTP {
			for _, r := range http.Routes {
				// Apply if there is a match
				if strings.HasPrefix(r.Name, prefix) {
					r.Traffic = &ir.TrafficFeatures{
						RateLimit:         rl,
						LoadBalancer:      lb,
						ProxyProtocol:     pp,
						HealthCheck:       hc,
						CircuitBreaker:    cb,
						FaultInjection:    fi,
						TCPKeepalive:      ka,
						Retry:             rt,
						BackendConnection: bc,
					}

					r.DNS = ds
					// Update the Host field in HealthCheck, now that we have access to the Route Hostname.
					r.Traffic.HealthCheck.SetHTTPHostIfAbsent(r.Hostname)

					// Some timeout setting originate from the route.
					if policy.Spec.Timeout != nil {
						if to, err = t.buildTimeout(policy, r); err == nil {
							r.Traffic.Timeout = to
						}
					}

					if policy.Spec.UseClientProtocol != nil {
						r.UseClientProtocol = policy.Spec.UseClientProtocol
					}
				}
			}
		}
	}

	return nil
}

func (t *Translator) translateBackendTrafficPolicyForGateway(policy *egv1a1.BackendTrafficPolicy, target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName, gateway *GatewayContext, xdsIR XdsIRMap) error {
	var (
		rl        *ir.RateLimit
		lb        *ir.LoadBalancer
		pp        *ir.ProxyProtocol
		hc        *ir.HealthCheck
		cb        *ir.CircuitBreaker
		fi        *ir.FaultInjection
		ct        *ir.Timeout
		ka        *ir.TCPKeepalive
		rt        *ir.Retry
		ds        *ir.DNS
		err, errs error
	)

	// Build IR
	if policy.Spec.RateLimit != nil {
		if rl, err = t.buildRateLimit(policy); err != nil {
			err = perr.WithMessage(err, "RateLimit")
			errs = errors.Join(errs, err)
		}
	}
	if policy.Spec.LoadBalancer != nil {
		if lb, err = t.buildLoadBalancer(policy); err != nil {
			err = perr.WithMessage(err, "LoadBalancer")
			errs = errors.Join(errs, err)
		}
	}
	if policy.Spec.ProxyProtocol != nil {
		pp = t.buildProxyProtocol(policy)
	}
	if policy.Spec.HealthCheck != nil {
		hc = t.buildHealthCheck(policy)
	}
	if policy.Spec.CircuitBreaker != nil {
		if cb, err = t.buildCircuitBreaker(policy); err != nil {
			err = perr.WithMessage(err, "CircuitBreaker")
			errs = errors.Join(errs, err)
		}
	}
	if policy.Spec.FaultInjection != nil {
		fi = t.buildFaultInjection(policy)
	}
	if policy.Spec.TCPKeepalive != nil {
		if ka, err = t.buildTCPKeepAlive(policy); err != nil {
			err = perr.WithMessage(err, "TCPKeepalive")
			errs = errors.Join(errs, err)
		}
	}
	if policy.Spec.Retry != nil {
		rt = t.buildRetry(policy)
	}
	if policy.Spec.Timeout != nil {
		if ct, err = t.buildTimeout(policy, nil); err != nil {
			err = perr.WithMessage(err, "Timeout")
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.DNS != nil {
		ds = t.translateDNS(policy)
	}

	// Early return if got any errors
	if errs != nil {
		return errs
	}

	// Apply IR to all the routes within the specific Gateway
	// If the feature is already set, then skip it, since it must be have
	// set by a policy attaching to the route
	irKey := t.getIRKey(gateway.Gateway)
	// Should exist since we've validated this
	x := xdsIR[irKey]

	policyTarget := irStringKey(policy.Namespace, string(target.Name))

	for _, tcp := range x.TCP {
		gatewayName := tcp.Name[0:strings.LastIndex(tcp.Name, "/")]
		if t.MergeGateways && gatewayName != policyTarget {
			continue
		}

		for _, r := range tcp.Routes {
			// policy(targeting xRoute) has already set it, so we skip it.
			if r.LoadBalancer != nil || r.ProxyProtocol != nil ||
				r.HealthCheck != nil || r.CircuitBreaker != nil ||
				r.TCPKeepalive != nil || r.Timeout != nil {
				continue
			}

			r.LoadBalancer = lb
			r.ProxyProtocol = pp
			r.HealthCheck = hc
			r.CircuitBreaker = cb
			r.TCPKeepalive = ka

			if r.Timeout == nil {
				r.Timeout = ct
			}

			if r.DNS == nil {
				r.DNS = ds
			}
		}
	}

	for _, udp := range x.UDP {
		gatewayName := udp.Name[0:strings.LastIndex(udp.Name, "/")]
		if t.MergeGateways && gatewayName != policyTarget {
			continue
		}

		if udp.Route == nil {
			continue
		}

		route := udp.Route

		// policy(targeting xRoute) has already set it, so we skip it.
		if route.LoadBalancer != nil || route.Timeout != nil {
			continue
		}

		route.LoadBalancer = lb
		if route.Timeout == nil {
			route.Timeout = ct
		}

		if route.DNS == nil {
			route.DNS = ds
		}
	}

	for _, http := range x.HTTP {
		gatewayName := http.Name[0:strings.LastIndex(http.Name, "/")]
		if t.MergeGateways && gatewayName != policyTarget {
			continue
		}

		// A Policy targeting the most specific scope(xRoute) wins over a policy
		// targeting a lesser specific scope(Gateway).
		for _, r := range http.Routes {
			// If any of the features are already set, it means that a more specific
			// policy(targeting xRoute) has already set it, so we skip it.
			if r.Traffic != nil {
				continue
			}

			r.Traffic = &ir.TrafficFeatures{
				RateLimit:      rl,
				LoadBalancer:   lb,
				ProxyProtocol:  pp,
				HealthCheck:    hc,
				CircuitBreaker: cb,
				FaultInjection: fi,
				TCPKeepalive:   ka,
				Retry:          rt,
			}

			if r.DNS == nil {
				r.DNS = ds
			}

			// Update the Host field in HealthCheck, now that we have access to the Route Hostname.
			r.Traffic.HealthCheck.SetHTTPHostIfAbsent(r.Hostname)

			if policy.Spec.Timeout != nil {
				if ct, err = t.buildTimeout(policy, r); err == nil {
					r.Traffic.Timeout = ct
				}
			}

			if policy.Spec.UseClientProtocol != nil {
				if r.UseClientProtocol == nil {
					r.UseClientProtocol = policy.Spec.UseClientProtocol
				}
			}
		}
	}

	return nil
}

func (t *Translator) buildRateLimit(policy *egv1a1.BackendTrafficPolicy) (*ir.RateLimit, error) {
	switch policy.Spec.RateLimit.Type {
	case egv1a1.GlobalRateLimitType:
		return t.buildGlobalRateLimit(policy)
	case egv1a1.LocalRateLimitType:
		return t.buildLocalRateLimit(policy)
	}

	return nil, fmt.Errorf("invalid rateLimit type: %s", policy.Spec.RateLimit.Type)
}

func (t *Translator) buildLocalRateLimit(policy *egv1a1.BackendTrafficPolicy) (*ir.RateLimit, error) {
	if policy.Spec.RateLimit.Local == nil {
		return nil, fmt.Errorf("local configuration empty for rateLimit")
	}

	local := policy.Spec.RateLimit.Local

	// Envoy local rateLimit requires a default limit to be set for a route.
	// EG uses the first rule without clientSelectors as the default route-level
	// limit. If no such rule is found, EG uses a default limit of uint32 max.
	var defaultLimit *ir.RateLimitValue
	for _, rule := range local.Rules {
		if rule.ClientSelectors == nil || len(rule.ClientSelectors) == 0 {
			if defaultLimit != nil {
				return nil, fmt.Errorf("local rateLimit can not have more than one rule without clientSelectors")
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
			return nil, fmt.Errorf("local rateLimit rule limit unit must be a multiple of the default limit unit")
		}
	}

	var err error
	var irRule *ir.RateLimitRule
	irRules := make([]*ir.RateLimitRule, 0)
	for _, rule := range local.Rules {
		// We don't process the rule without clientSelectors here because it's
		// previously used as the default route-level limit.
		if len(rule.ClientSelectors) == 0 {
			continue
		}

		irRule, err = buildRateLimitRule(rule)
		if err != nil {
			return nil, err
		}

		if irRule.CIDRMatch != nil && irRule.CIDRMatch.Distinct {
			return nil, fmt.Errorf("local rateLimit does not support distinct CIDRMatch")
		}

		for _, match := range irRule.HeaderMatches {
			if match.Distinct {
				return nil, fmt.Errorf("local rateLimit does not support distinct HeaderMatch")
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

	return rateLimit, nil
}

func (t *Translator) buildGlobalRateLimit(policy *egv1a1.BackendTrafficPolicy) (*ir.RateLimit, error) {
	if policy.Spec.RateLimit.Global == nil {
		return nil, fmt.Errorf("global configuration empty for rateLimit")
	}

	if !t.GlobalRateLimitEnabled {
		return nil, fmt.Errorf("enable Ratelimit in the EnvoyGateway config to configure global rateLimit")
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
			return nil, err
		}
	}

	return rateLimit, nil
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
			return nil, fmt.Errorf(
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
				return nil, fmt.Errorf(
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

			cidrMatch, err := parseCIDR(sourceCIDR)
			if err != nil {
				return nil, fmt.Errorf("unable to translate rateLimit: %w", err)
			}
			cidrMatch.Distinct = distinct
			irRule.CIDRMatch = cidrMatch
		}
	}
	return irRule, nil
}

func (t *Translator) buildLoadBalancer(policy *egv1a1.BackendTrafficPolicy) (*ir.LoadBalancer, error) {
	var lb *ir.LoadBalancer
	switch policy.Spec.LoadBalancer.Type {
	case egv1a1.ConsistentHashLoadBalancerType:
		consistentHash, err := t.buildConsistentHashLoadBalancer(policy)
		if err != nil {
			return nil, perr.WithMessage(err, "ConsistentHash")
		}

		lb = &ir.LoadBalancer{
			ConsistentHash: consistentHash,
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

	return lb, nil
}

func (t *Translator) buildConsistentHashLoadBalancer(policy *egv1a1.BackendTrafficPolicy) (*ir.ConsistentHash, error) {
	consistentHash := &ir.ConsistentHash{}

	if policy.Spec.LoadBalancer.ConsistentHash.TableSize != nil {
		tableSize := policy.Spec.LoadBalancer.ConsistentHash.TableSize

		if *tableSize > MaxConsistentHashTableSize || !big.NewInt(int64(*tableSize)).ProbablyPrime(0) {
			return nil, fmt.Errorf("invalid TableSize value %d", *tableSize)
		}

		consistentHash.TableSize = tableSize
	}

	switch policy.Spec.LoadBalancer.ConsistentHash.Type {
	case egv1a1.SourceIPConsistentHashType:
		consistentHash.SourceIP = ptr.To(true)
	case egv1a1.HeaderConsistentHashType:
		consistentHash.Header = &ir.Header{
			Name: policy.Spec.LoadBalancer.ConsistentHash.Header.Name,
		}
	case egv1a1.CookieConsistentHashType:
		consistentHash.Cookie = policy.Spec.LoadBalancer.ConsistentHash.Cookie
	}

	return consistentHash, nil
}

func (t *Translator) translateDNS(policy *egv1a1.BackendTrafficPolicy) *ir.DNS {
	ds := &ir.DNS{}
	if policy.Spec.DNS.RespectDNSTTL != nil {
		ds.RespectDNSTTL = policy.Spec.DNS.RespectDNSTTL
	}
	if policy.Spec.DNS.DNSRefreshRate != nil {
		ds.DNSRefreshRate = policy.Spec.DNS.DNSRefreshRate
	}
	return ds
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

func (t *Translator) buildHealthCheck(policy *egv1a1.BackendTrafficPolicy) *ir.HealthCheck {
	if policy.Spec.HealthCheck == nil {
		return nil
	}

	irhc := &ir.HealthCheck{}
	if policy.Spec.HealthCheck.Passive != nil {
		irhc.Passive = t.buildPassiveHealthCheck(policy)
	}

	if policy.Spec.HealthCheck.Active != nil {
		irhc.Active = t.buildActiveHealthCheck(policy)
	}

	return irhc
}

func (t *Translator) buildPassiveHealthCheck(policy *egv1a1.BackendTrafficPolicy) *ir.OutlierDetection {
	if policy.Spec.HealthCheck == nil || policy.Spec.HealthCheck.Passive == nil {
		return nil
	}

	hc := policy.Spec.HealthCheck.Passive
	irOD := &ir.OutlierDetection{
		Interval:                       hc.Interval,
		SplitExternalLocalOriginErrors: hc.SplitExternalLocalOriginErrors,
		ConsecutiveLocalOriginFailures: hc.ConsecutiveLocalOriginFailures,
		ConsecutiveGatewayErrors:       hc.ConsecutiveGatewayErrors,
		Consecutive5xxErrors:           hc.Consecutive5xxErrors,
		BaseEjectionTime:               hc.BaseEjectionTime,
		MaxEjectionPercent:             hc.MaxEjectionPercent,
	}
	return irOD
}

func (t *Translator) buildActiveHealthCheck(policy *egv1a1.BackendTrafficPolicy) *ir.ActiveHealthCheck {
	if policy.Spec.HealthCheck == nil || policy.Spec.HealthCheck.Active == nil {
		return nil
	}

	hc := policy.Spec.HealthCheck.Active
	irHC := &ir.ActiveHealthCheck{
		Timeout:            hc.Timeout,
		Interval:           hc.Interval,
		UnhealthyThreshold: hc.UnhealthyThreshold,
		HealthyThreshold:   hc.HealthyThreshold,
	}
	switch hc.Type {
	case egv1a1.ActiveHealthCheckerTypeHTTP:
		irHC.HTTP = t.buildHTTPActiveHealthChecker(hc.HTTP)
	case egv1a1.ActiveHealthCheckerTypeTCP:
		irHC.TCP = t.buildTCPActiveHealthChecker(hc.TCP)
	}

	return irHC
}

func (t *Translator) buildHTTPActiveHealthChecker(h *egv1a1.HTTPActiveHealthChecker) *ir.HTTPHealthChecker {
	if h == nil {
		return nil
	}

	irHTTP := &ir.HTTPHealthChecker{
		Path:   h.Path,
		Method: h.Method,
	}
	if irHTTP.Method != nil {
		*irHTTP.Method = strings.ToUpper(*irHTTP.Method)
	}

	// deduplicate http statuses
	statusSet := sets.NewInt()
	for _, r := range h.ExpectedStatuses {
		statusSet.Insert(int(r))
	}
	irStatuses := make([]ir.HTTPStatus, 0, statusSet.Len())

	for _, r := range statusSet.List() {
		irStatuses = append(irStatuses, ir.HTTPStatus(r))
	}
	irHTTP.ExpectedStatuses = irStatuses

	irHTTP.ExpectedResponse = translateActiveHealthCheckPayload(h.ExpectedResponse)
	return irHTTP
}

func (t *Translator) buildTCPActiveHealthChecker(h *egv1a1.TCPActiveHealthChecker) *ir.TCPHealthChecker {
	if h == nil {
		return nil
	}

	irTCP := &ir.TCPHealthChecker{
		Send:    translateActiveHealthCheckPayload(h.Send),
		Receive: translateActiveHealthCheckPayload(h.Receive),
	}
	return irTCP
}

func translateActiveHealthCheckPayload(p *egv1a1.ActiveHealthCheckPayload) *ir.HealthCheckPayload {
	if p == nil {
		return nil
	}

	irPayload := &ir.HealthCheckPayload{}
	switch p.Type {
	case egv1a1.ActiveHealthCheckPayloadTypeText:
		irPayload.Text = p.Text
	case egv1a1.ActiveHealthCheckPayloadTypeBinary:
		irPayload.Binary = make([]byte, len(p.Binary))
		copy(irPayload.Binary, p.Binary)
	}

	return irPayload
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

func (t *Translator) buildCircuitBreaker(policy *egv1a1.BackendTrafficPolicy) (*ir.CircuitBreaker, error) {
	var cb *ir.CircuitBreaker
	pcb := policy.Spec.CircuitBreaker

	if pcb != nil {
		cb = &ir.CircuitBreaker{}

		if pcb.MaxConnections != nil {
			if ui32, ok := int64ToUint32(*pcb.MaxConnections); ok {
				cb.MaxConnections = &ui32
			} else {
				return nil, fmt.Errorf("invalid MaxConnections value %d", *pcb.MaxConnections)
			}
		}

		if pcb.MaxParallelRequests != nil {
			if ui32, ok := int64ToUint32(*pcb.MaxParallelRequests); ok {
				cb.MaxParallelRequests = &ui32
			} else {
				return nil, fmt.Errorf("invalid MaxParallelRequests value %d", *pcb.MaxParallelRequests)
			}
		}

		if pcb.MaxPendingRequests != nil {
			if ui32, ok := int64ToUint32(*pcb.MaxPendingRequests); ok {
				cb.MaxPendingRequests = &ui32
			} else {
				return nil, fmt.Errorf("invalid MaxPendingRequests value %d", *pcb.MaxPendingRequests)
			}
		}

		if pcb.MaxParallelRetries != nil {
			if ui32, ok := int64ToUint32(*pcb.MaxParallelRetries); ok {
				cb.MaxParallelRetries = &ui32
			} else {
				return nil, fmt.Errorf("invalid MaxParallelRetries value %d", *pcb.MaxParallelRetries)
			}
		}

		if pcb.MaxRequestsPerConnection != nil {
			if ui32, ok := int64ToUint32(*pcb.MaxRequestsPerConnection); ok {
				cb.MaxRequestsPerConnection = &ui32
			} else {
				return nil, fmt.Errorf("invalid MaxRequestsPerConnection value %d", *pcb.MaxRequestsPerConnection)
			}
		}

	}

	return cb, nil
}

func (t *Translator) buildTimeout(policy *egv1a1.BackendTrafficPolicy, r *ir.HTTPRoute) (*ir.Timeout, error) {
	var (
		tto  *ir.TCPTimeout
		hto  *ir.HTTPTimeout
		terr bool
		errs error
	)

	pto := policy.Spec.Timeout

	if pto.TCP != nil && pto.TCP.ConnectTimeout != nil {
		d, err := time.ParseDuration(string(*pto.TCP.ConnectTimeout))
		if err != nil {
			terr = true
			errs = errors.Join(errs, fmt.Errorf("invalid ConnectTimeout value %s", *pto.TCP.ConnectTimeout))
		} else {
			tto = &ir.TCPTimeout{
				ConnectTimeout: ptr.To(metav1.Duration{Duration: d}),
			}
		}
	}

	if pto.HTTP != nil {
		var cit *metav1.Duration
		var mcd *metav1.Duration

		if pto.HTTP.ConnectionIdleTimeout != nil {
			d, err := time.ParseDuration(string(*pto.HTTP.ConnectionIdleTimeout))
			if err != nil {
				terr = true
				errs = errors.Join(errs, fmt.Errorf("invalid ConnectionIdleTimeout value %s", *pto.HTTP.ConnectionIdleTimeout))
			} else {
				cit = ptr.To(metav1.Duration{Duration: d})
			}
		}

		if pto.HTTP.MaxConnectionDuration != nil {
			d, err := time.ParseDuration(string(*pto.HTTP.MaxConnectionDuration))
			if err != nil {
				terr = true
				errs = errors.Join(errs, fmt.Errorf("invalid MaxConnectionDuration value %s", *pto.HTTP.MaxConnectionDuration))
			} else {
				mcd = ptr.To(metav1.Duration{Duration: d})
			}
		}

		hto = &ir.HTTPTimeout{
			ConnectionIdleTimeout: cit,
			MaxConnectionDuration: mcd,
		}
	}

	// http request timeout is translated during the gateway-api route resource translation
	// merge route timeout setting with backendtrafficpolicy timeout settings
	if terr {
		if r != nil && r.Traffic != nil && r.Traffic.Timeout != nil {
			return r.Traffic.Timeout.DeepCopy(), errs
		}
	} else {
		// http request timeout is translated during the gateway-api route resource translation
		// merge route timeout setting with backendtrafficpolicy timeout settings
		if r != nil &&
			r.Traffic != nil &&
			r.Traffic.Timeout != nil &&
			r.Traffic.Timeout.HTTP != nil &&
			r.Traffic.Timeout.HTTP.RequestTimeout != nil {
			if hto == nil {
				hto = &ir.HTTPTimeout{
					RequestTimeout: r.Traffic.Timeout.HTTP.RequestTimeout,
				}
			} else {
				hto.RequestTimeout = r.Traffic.Timeout.HTTP.RequestTimeout
			}
		}

		if hto != nil || tto != nil {
			return &ir.Timeout{
				TCP:  tto,
				HTTP: hto,
			}, nil
		}
	}

	return nil, errs
}

func int64ToUint32(in int64) (uint32, bool) {
	if in >= 0 && in <= math.MaxUint32 {
		return uint32(in), true
	}
	return 0, false
}

func (t *Translator) buildBackendConnection(policy *egv1a1.BackendTrafficPolicy) (*ir.BackendConnection, error) {
	var (
		bcIR = &ir.BackendConnection{}
		bc   = &egv1a1.BackendConnection{}
	)

	if policy.Spec.Connection != nil {
		bc = policy.Spec.Connection

		if bc.BufferLimit != nil {
			bf, ok := bc.BufferLimit.AsInt64()
			if !ok {
				return nil, fmt.Errorf("invalid BufferLimit value %s", bc.BufferLimit.String())
			}
			if bf < 0 || bf > math.MaxUint32 {
				return nil, fmt.Errorf("BufferLimit value %s is out of range", bc.BufferLimit.String())
			}

			bcIR.BufferLimitBytes = ptr.To(uint32(bf))
		}
	}

	return bcIR, nil
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

func (t *Translator) buildTCPKeepAlive(policy *egv1a1.BackendTrafficPolicy) (*ir.TCPKeepalive, error) {
	var ka *ir.TCPKeepalive
	if policy.Spec.TCPKeepalive != nil {
		pka := policy.Spec.TCPKeepalive
		ka = &ir.TCPKeepalive{}

		if pka.Probes != nil {
			ka.Probes = pka.Probes
		}

		if pka.IdleTime != nil {
			d, err := time.ParseDuration(string(*pka.IdleTime))
			if err != nil {
				return nil, fmt.Errorf("invalid IdleTime value %s", *pka.IdleTime)
			}
			ka.IdleTime = ptr.To(uint32(d.Seconds()))
		}

		if pka.Interval != nil {
			d, err := time.ParseDuration(string(*pka.Interval))
			if err != nil {
				return nil, fmt.Errorf("invalid Interval value %s", *pka.Interval)
			}
			ka.Interval = ptr.To(uint32(d.Seconds()))
		}

	}
	return ka, nil
}

func (t *Translator) buildRetry(policy *egv1a1.BackendTrafficPolicy) *ir.Retry {
	var rt *ir.Retry
	if policy.Spec.Retry != nil {
		prt := policy.Spec.Retry
		rt = &ir.Retry{}

		if prt.NumRetries != nil {
			rt.NumRetries = ptr.To(uint32(*prt.NumRetries))
		}

		if prt.RetryOn != nil {
			ro := &ir.RetryOn{}
			bro := false
			if prt.RetryOn.HTTPStatusCodes != nil {
				ro.HTTPStatusCodes = makeIrStatusSet(prt.RetryOn.HTTPStatusCodes)
				bro = true
			}

			if prt.RetryOn.Triggers != nil {
				ro.Triggers = makeIrTriggerSet(prt.RetryOn.Triggers)
				bro = true
			}

			if bro {
				rt.RetryOn = ro
			}
		}

		if prt.PerRetry != nil {
			pr := &ir.PerRetryPolicy{}
			bpr := false

			if prt.PerRetry.Timeout != nil {
				pr.Timeout = prt.PerRetry.Timeout
				bpr = true
			}

			if prt.PerRetry.BackOff != nil {
				if prt.PerRetry.BackOff.MaxInterval != nil || prt.PerRetry.BackOff.BaseInterval != nil {
					bop := &ir.BackOffPolicy{}
					if prt.PerRetry.BackOff.MaxInterval != nil {
						bop.MaxInterval = prt.PerRetry.BackOff.MaxInterval
					}

					if prt.PerRetry.BackOff.BaseInterval != nil {
						bop.BaseInterval = prt.PerRetry.BackOff.BaseInterval
					}
					pr.BackOff = bop
					bpr = true
				}
			}

			if bpr {
				rt.PerRetry = pr
			}
		}
	}

	return rt
}

func makeIrStatusSet(in []egv1a1.HTTPStatus) []ir.HTTPStatus {
	statusSet := sets.NewInt()
	for _, r := range in {
		statusSet.Insert(int(r))
	}
	irStatuses := make([]ir.HTTPStatus, 0, statusSet.Len())

	for _, r := range statusSet.List() {
		irStatuses = append(irStatuses, ir.HTTPStatus(r))
	}
	return irStatuses
}

func makeIrTriggerSet(in []egv1a1.TriggerEnum) []ir.TriggerEnum {
	triggerSet := sets.NewString()
	for _, r := range in {
		triggerSet.Insert(string(r))
	}
	irTriggers := make([]ir.TriggerEnum, 0, triggerSet.Len())

	for _, r := range triggerSet.List() {
		irTriggers = append(irTriggers, ir.TriggerEnum(r))
	}
	return irTriggers
}
