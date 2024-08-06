// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

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
		h2        *ir.HTTP2Settings
		err, errs error
	)

	// Build IR
	if policy.Spec.RateLimit != nil {
		if rl, err = t.buildRateLimit(policy); err != nil {
			err = perr.WithMessage(err, "RateLimit")
			errs = errors.Join(errs, err)
		}
	}
	if lb, err = buildLoadBalancer(policy.Spec.ClusterSettings); err != nil {
		err = perr.WithMessage(err, "LoadBalancer")
		errs = errors.Join(errs, err)
	}
	pp = buildProxyProtocol(policy.Spec.ClusterSettings)
	hc = buildHealthCheck(policy.Spec.ClusterSettings)
	if cb, err = buildCircuitBreaker(policy.Spec.ClusterSettings); err != nil {
		err = perr.WithMessage(err, "CircuitBreaker")
		errs = errors.Join(errs, err)
	}
	if policy.Spec.FaultInjection != nil {
		fi = t.buildFaultInjection(policy)
	}
	if ka, err = buildTCPKeepAlive(policy.Spec.ClusterSettings); err != nil {
		err = perr.WithMessage(err, "TCPKeepalive")
		errs = errors.Join(errs, err)
	}
	if policy.Spec.Retry != nil {
		rt = t.buildRetry(policy)
	}
	if to, err = buildTimeout(policy.Spec.ClusterSettings, nil); err != nil {
		err = perr.WithMessage(err, "Timeout")
		errs = errors.Join(errs, err)
	}

	if bc, err = buildBackendConnection(policy.Spec.ClusterSettings); err != nil {
		err = perr.WithMessage(err, "BackendConnection")
		errs = errors.Join(errs, err)
	}

	if h2, err = buildIRHTTP2Settings(policy.Spec.HTTP2); err != nil {
		err = perr.WithMessage(err, "HTTP2")
		errs = errors.Join(errs, err)
	}

	ds = translateDNS(policy.Spec.ClusterSettings)

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
						HTTP2:             h2,
						DNS:               ds,
					}

					// Update the Host field in HealthCheck, now that we have access to the Route Hostname.
					r.Traffic.HealthCheck.SetHTTPHostIfAbsent(r.Hostname)

					// Some timeout setting originate from the route.
					if to, err = buildTimeout(policy.Spec.ClusterSettings, r); err == nil {
						r.Traffic.Timeout = to
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
		h2        *ir.HTTP2Settings
		err, errs error
	)

	// Build IR
	if policy.Spec.RateLimit != nil {
		if rl, err = t.buildRateLimit(policy); err != nil {
			err = perr.WithMessage(err, "RateLimit")
			errs = errors.Join(errs, err)
		}
	}
	if lb, err = buildLoadBalancer(policy.Spec.ClusterSettings); err != nil {
		err = perr.WithMessage(err, "LoadBalancer")
		errs = errors.Join(errs, err)
	}
	pp = buildProxyProtocol(policy.Spec.ClusterSettings)
	hc = buildHealthCheck(policy.Spec.ClusterSettings)
	if cb, err = buildCircuitBreaker(policy.Spec.ClusterSettings); err != nil {
		err = perr.WithMessage(err, "CircuitBreaker")
		errs = errors.Join(errs, err)
	}
	if policy.Spec.FaultInjection != nil {
		fi = t.buildFaultInjection(policy)
	}
	if ka, err = buildTCPKeepAlive(policy.Spec.ClusterSettings); err != nil {
		err = perr.WithMessage(err, "TCPKeepalive")
		errs = errors.Join(errs, err)
	}
	if policy.Spec.Retry != nil {
		rt = t.buildRetry(policy)
	}
	if ct, err = buildTimeout(policy.Spec.ClusterSettings, nil); err != nil {
		err = perr.WithMessage(err, "Timeout")
		errs = errors.Join(errs, err)
	}
	if h2, err = buildIRHTTP2Settings(policy.Spec.HTTP2); err != nil {
		err = perr.WithMessage(err, "HTTP2")
		errs = errors.Join(errs, err)
	}

	ds = translateDNS(policy.Spec.ClusterSettings)

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
			// only set attributes which weren't already set by a more
			// specific policy
			setIfNil(&r.LoadBalancer, lb)
			setIfNil(&r.ProxyProtocol, pp)
			setIfNil(&r.HealthCheck, hc)
			setIfNil(&r.CircuitBreaker, cb)
			setIfNil(&r.TCPKeepalive, ka)
			setIfNil(&r.Timeout, ct)
			setIfNil(&r.DNS, ds)
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

		// only set attributes which weren't already set by a more
		// specific policy
		setIfNil(&route.LoadBalancer, lb)
		setIfNil(&route.Timeout, ct)
		setIfNil(&route.DNS, ds)
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
				HTTP2:          h2,
				DNS:            ds,
			}

			// Update the Host field in HealthCheck, now that we have access to the Route Hostname.
			r.Traffic.HealthCheck.SetHTTPHostIfAbsent(r.Hostname)

			if ct, err = buildTimeout(policy.Spec.ClusterSettings, r); err == nil {
				r.Traffic.Timeout = ct
			}

			if policy.Spec.UseClientProtocol != nil {
				setIfNil(&r.UseClientProtocol, policy.Spec.UseClientProtocol)
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
