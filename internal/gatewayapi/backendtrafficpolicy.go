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
	"strconv"
	"strings"

	perr "github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/utils/ratelimit"
	"github.com/envoyproxy/gateway/internal/utils/regex"
)

const (
	MaxConsistentHashTableSize = 5000011 // https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/cluster.proto#config-cluster-v3-cluster-maglevlbconfig
)

func (t *Translator) ProcessBackendTrafficPolicies(resources *resource.Resources,
	gateways []*GatewayContext,
	routes []RouteContext,
	xdsIR resource.XdsIRMap,
) []*egv1a1.BackendTrafficPolicy {
	res := make([]*egv1a1.BackendTrafficPolicy, 0, len(resources.BackendTrafficPolicies))

	backendTrafficPolicies := resources.BackendTrafficPolicies
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

	gatewayPolicyMap := make(map[types.NamespacedName]*egv1a1.BackendTrafficPolicy)
	gatewayPolicyMerged := make(map[types.NamespacedName]sets.Set[string])

	// Translate
	// 1. First translate Policies targeting xRoutes
	// 2. Finally, the policies targeting Gateways

	// Build gateway policy map, which is needed when processing the policies targeting xRoutes.
	// TODO: This loop is similar to the one 'Process the policies targeting Gateways', we may want to
	// merge them into one if possible.
	for _, currPolicy := range backendTrafficPolicies {
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, gateways)
		for _, currTarget := range targetRefs {
			if currTarget.Kind == resource.KindGateway {
				// Check if the gateway exists
				key := types.NamespacedName{
					Name:      string(currTarget.Name),
					Namespace: currPolicy.Namespace,
				}
				gateway, ok := gatewayMap[key]
				if !ok {
					continue
				}

				// Check if policy for gateway exists
				gatewayNN := utils.NamespacedName(gateway)
				_, policyExists := gatewayPolicyMap[gatewayNN]
				if policyExists {
					continue
				}
				gatewayPolicyMap[gatewayNN] = currPolicy
			}
		}
	}

	// Process the policies targeting xRoutes
	for _, currPolicy := range backendTrafficPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, routes)
		for _, currTarget := range targetRefs {
			if currTarget.Kind != resource.KindGateway {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy.DeepCopy()
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}

				// Negative statuses have already been assigned so it's safe to skip
				route, resolveErr := resolveBTPolicyRouteTargetRef(policy, currTarget, routeMap)
				if route == nil {
					continue
				}

				// Find the Gateway that the route belongs to and add it to the
				// gatewayRouteMap and ancestor list, which will be used to check
				// policy overrides and populate its ancestor status.
				parentRefs := GetParentReferences(route)
				ancestorRefs := make([]gwapiv1a2.ParentReference, 0, len(parentRefs))
				routeParents := sets.New[types.NamespacedName]()
				for _, p := range parentRefs {
					if p.Kind == nil || *p.Kind == resource.KindGateway {
						namespace := route.GetNamespace()
						if p.Namespace != nil {
							namespace = string(*p.Namespace)
						}
						gwNN := types.NamespacedName{
							Namespace: namespace,
							Name:      string(p.Name),
						}
						routeParents.Insert(gwNN)

						key := gwNN.String()
						if _, ok := gatewayRouteMap[key]; !ok {
							gatewayRouteMap[key] = make(sets.Set[string])
						}
						gatewayRouteMap[key].Insert(utils.NamespacedName(route).String())

						// Do need a section name since the policy is targeting to a route
						ancestorRefs = append(ancestorRefs, getAncestorRefForPolicy(gwNN, p.SectionName))
					}
				}

				// Set conditions for resolve error, then skip current xRoute
				if resolveErr != nil {
					status.SetResolveErrorForPolicyAncestors(&policy.Status,
						ancestorRefs,
						t.GatewayControllerName,
						policy.Generation,
						resolveErr,
					)

					continue
				}

				if policy.Spec.MergeType == nil {
					// Set conditions for translation error if it got any
					if err := t.translateBackendTrafficPolicyForRoute(policy, route, xdsIR, resources, nil); err != nil {
						status.SetTranslationErrorForPolicyAncestors(&policy.Status,
							ancestorRefs,
							t.GatewayControllerName,
							policy.Generation,
							status.Error2ConditionMsg(err),
						)
					}
				} else {
					// merge with parent target policy if exists
					for _, gwNN := range routeParents.UnsortedList() {
						ancestorRef := getAncestorRefForPolicy(gwNN, nil)
						// find policy for Gateway
						gwPolicy := gatewayPolicyMap[gwNN]
						if gwPolicy == nil {
							// not found, fall back to the current policy
							if err := t.translateBackendTrafficPolicyForRoute(policy, route, xdsIR, resources, &gwNN); err != nil {
								status.SetConditionForPolicyAncestor(&policy.Status,
									ancestorRef,
									t.GatewayControllerName,
									gwapiv1a2.PolicyConditionAccepted, metav1.ConditionFalse,
									egv1a1.PolicyReasonInvalid,
									status.Error2ConditionMsg(err),
									policy.Generation,
								)
							}

							continue
						}

						// merge with parent policy
						if err := t.translateBackendTrafficPolicyForRouteWithMerge(policy, gwNN, gwPolicy, route, xdsIR, resources); err != nil {
							status.SetConditionForPolicyAncestor(&policy.Status,
								ancestorRef,
								t.GatewayControllerName,
								gwapiv1a2.PolicyConditionAccepted, metav1.ConditionFalse,
								egv1a1.PolicyReasonInvalid,
								status.Error2ConditionMsg(err),
								policy.Generation,
							)
							continue
						}

						if _, ok := gatewayPolicyMerged[gwNN]; !ok {
							gatewayPolicyMerged[gwNN] = make(sets.Set[string])
						}
						gatewayPolicyMerged[gwNN].Insert(utils.NamespacedName(route).String())

						status.SetConditionForPolicyAncestor(&policy.Status,
							ancestorRef,
							t.GatewayControllerName,
							egv1a1.PolicyConditionMerged,
							metav1.ConditionTrue,
							egv1a1.PolicyReasonMerged,
							fmt.Sprintf("Merged with policy %s/%s", gwPolicy.Namespace, gwPolicy.Name),
							policy.Generation,
						)
					}
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
			if currTarget.Kind == resource.KindGateway {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy.DeepCopy()
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}

				// Negative statuses have already been assigned so it's safe to skip
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
				if err := t.translateBackendTrafficPolicyForGateway(policy, currTarget, gateway, xdsIR, resources); err != nil {
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
				mergedRoutes, routeMerged := gatewayPolicyMerged[gatewayNN]
				if routeMerged && len(mergedRoutes) > 0 {
					// if merge happened, skip the overridden check
					gatewayMergedRoutes := mergedRoutes.UnsortedList()
					sort.Strings(gatewayMergedRoutes)

					status.SetConditionForPolicyAncestors(&policy.Status,
						ancestorRefs,
						t.GatewayControllerName,
						egv1a1.PolicyConditionMerged,
						metav1.ConditionTrue,
						egv1a1.PolicyReasonMerged,
						fmt.Sprintf("This policy is being merged by other backendTrafficPolicies for these routes: %v", gatewayMergedRoutes),
						policy.Generation,
					)
				}

				if r, ok := gatewayRouteMap[gatewayNN.String()]; ok {
					// Maintain order here to ensure status/string does not change with the same data
					var gatewayRoutes []string
					if len(mergedRoutes) > 0 {
						gatewayRoutes = r.Difference(mergedRoutes).UnsortedList()
					} else {
						gatewayRoutes = r.UnsortedList()
					}
					sort.Strings(gatewayRoutes)
					if len(gatewayRoutes) == 0 {
						continue
					}

					status.SetConditionForPolicyAncestors(&policy.Status,
						ancestorRefs,
						t.GatewayControllerName,
						egv1a1.PolicyConditionOverridden,
						metav1.ConditionTrue,
						egv1a1.PolicyReasonOverridden,
						fmt.Sprintf("This policy is being overridden by other backendTrafficPolicies for these routes: %v", gatewayRoutes),
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

func (t *Translator) translateBackendTrafficPolicyForRoute(
	policy *egv1a1.BackendTrafficPolicy,
	route RouteContext,
	xdsIR resource.XdsIRMap,
	resources *resource.Resources,
	gatewayNN *types.NamespacedName,
) error {
	tf, errs := t.buildTrafficFeatures(policy, resources)
	if tf == nil {
		// should not happen
		return nil
	}

	// Apply IR to all relevant routes
	for key, x := range xdsIR {
		// if gatewayNN is not nil, only apply to the specific gateway
		if gatewayNN != nil && key != t.IRKey(*gatewayNN) {
			// Skip if not the gateway wanted
			continue
		}
		applyTrafficFeatureToRoute(route, tf, errs, policy, x)
	}

	return errs
}

func (t *Translator) translateBackendTrafficPolicyForRouteWithMerge(
	policy *egv1a1.BackendTrafficPolicy,
	gatewayNN types.NamespacedName, gwPolicy *egv1a1.BackendTrafficPolicy,
	route RouteContext, xdsIR resource.XdsIRMap, resources *resource.Resources,
) error {
	mergedPolicy, err := mergeBackendTrafficPolicy(policy, gwPolicy)
	if err != nil {
		return fmt.Errorf("error merging policies: %w", err)
	}

	// Build traffic features from the merged policy
	tf, errs := t.buildTrafficFeatures(mergedPolicy, resources)
	if tf == nil {
		// should not happen
		return nil
	}

	// Since GlobalRateLimit merge relies on IR auto-generated key: (<policy-ns>/<policy-name>/rule/<rule-index>)
	// We can't simply merge the BTP's using utils.Merge() we need to specifically merge the GlobalRateLimit.Rules using IR fields.
	// Since ir.TrafficFeatures is not a built-in Kubernetes API object with defined merging strategies and it does not support a deep merge (for lists/maps).

	// Handle rate limit merging cases:
	// 1. Both policies have rate limits - merge them
	// 2. Only gateway policy has rate limits - preserve gateway policy's rule names
	// 3. Only route policy has rate limits - use route policy's rule names (default behavior)
	if policy.Spec.RateLimit != nil && gwPolicy.Spec.RateLimit != nil {
		// Case 1: Both policies have rate limits - merge them
		tfGW, _ := t.buildTrafficFeatures(gwPolicy, resources)
		tfRoute, _ := t.buildTrafficFeatures(policy, resources)

		if tfGW != nil && tfRoute != nil &&
			tfGW.RateLimit != nil && tfRoute.RateLimit != nil {

			mergedRL, err := utils.Merge(tfGW.RateLimit, tfRoute.RateLimit, *policy.Spec.MergeType)
			if err != nil {
				return fmt.Errorf("error merging rate limits: %w", err)
			}
			// Replace the rate limit in the merged features if successful
			tf.RateLimit = mergedRL
		}
	} else if policy.Spec.RateLimit == nil && gwPolicy.Spec.RateLimit != nil {
		// Case 2: Only gateway policy has rate limits - preserve gateway policy's rule names
		tfGW, _ := t.buildTrafficFeatures(gwPolicy, resources)
		if tfGW != nil && tfGW.RateLimit != nil {
			// Use the gateway policy's rate limit with its original rule names
			tf.RateLimit = tfGW.RateLimit
		}
	}
	// Case 3: Only route policy has rate limits or neither has rate limits - use default behavior (tf already built from merged policy)

	x, ok := xdsIR[t.IRKey(gatewayNN)]
	if !ok {
		// should not happen.
		return nil
	}
	applyTrafficFeatureToRoute(route, tf, errs, mergedPolicy, x)

	return nil
}

func applyTrafficFeatureToRoute(route RouteContext,
	tf *ir.TrafficFeatures, errs error,
	policy *egv1a1.BackendTrafficPolicy, x *ir.Xds,
) {
	prefix := irRoutePrefix(route)
	for _, tcp := range x.TCP {
		for _, r := range tcp.Routes {
			if strings.HasPrefix(r.Destination.Name, prefix) {
				r.LoadBalancer = tf.LoadBalancer
				r.ProxyProtocol = tf.ProxyProtocol
				r.HealthCheck = tf.HealthCheck
				r.CircuitBreaker = tf.CircuitBreaker
				r.TCPKeepalive = tf.TCPKeepalive
				r.Timeout = tf.Timeout
				r.BackendConnection = tf.BackendConnection
				r.DNS = tf.DNS
			}
		}
	}

	for _, udp := range x.UDP {
		if udp.Route != nil {
			r := udp.Route
			if strings.HasPrefix(r.Destination.Name, prefix) {
				r.LoadBalancer = tf.LoadBalancer
				r.DNS = tf.DNS
			}
		}
	}

	for _, http := range x.HTTP {
		for _, r := range http.Routes {
			// Apply if there is a match
			if strings.HasPrefix(r.Name, prefix) {
				if errs != nil {
					// Return a 500 direct response
					r.DirectResponse = &ir.CustomResponse{
						StatusCode: ptr.To(uint32(500)),
					}
					continue
				}

				r.Traffic = tf.DeepCopy()

				if localTo, err := buildClusterSettingsTimeout(policy.Spec.ClusterSettings); err == nil {
					r.Traffic.Timeout = localTo
				}

				// Update the Host field in HealthCheck, now that we have access to the Route Hostname.
				r.Traffic.HealthCheck.SetHTTPHostIfAbsent(r.Hostname)

				if policy.Spec.UseClientProtocol != nil {
					r.UseClientProtocol = policy.Spec.UseClientProtocol
				}
			}
		}
	}
}

func mergeBackendTrafficPolicy(routePolicy, gwPolicy *egv1a1.BackendTrafficPolicy) (*egv1a1.BackendTrafficPolicy, error) {
	if routePolicy.Spec.MergeType == nil || gwPolicy == nil {
		return routePolicy.DeepCopy(), nil
	}

	return utils.Merge[*egv1a1.BackendTrafficPolicy](gwPolicy, routePolicy, *routePolicy.Spec.MergeType)
}

func (t *Translator) buildTrafficFeatures(policy *egv1a1.BackendTrafficPolicy, resources *resource.Resources) (*ir.TrafficFeatures, error) {
	var (
		rl          *ir.RateLimit
		lb          *ir.LoadBalancer
		pp          *ir.ProxyProtocol
		hc          *ir.HealthCheck
		cb          *ir.CircuitBreaker
		fi          *ir.FaultInjection
		to          *ir.Timeout
		ka          *ir.TCPKeepalive
		rt          *ir.Retry
		bc          *ir.BackendConnection
		ds          *ir.DNS
		h2          *ir.HTTP2Settings
		ro          *ir.ResponseOverride
		rb          *ir.RequestBuffer
		cp          []*ir.Compression
		httpUpgrade []ir.HTTPUpgradeConfig
		err, errs   error
	)

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

	if rt, err = buildRetry(policy.Spec.Retry); err != nil {
		err = perr.WithMessage(err, "Retry")
		errs = errors.Join(errs, err)
	}

	if to, err = buildClusterSettingsTimeout(policy.Spec.ClusterSettings); err != nil {
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

	if ro, err = buildResponseOverride(policy, resources); err != nil {
		err = perr.WithMessage(err, "ResponseOverride")
		errs = errors.Join(errs, err)
	}

	if rb, err = buildRequestBuffer(policy.Spec.RequestBuffer); err != nil {
		err = perr.WithMessage(err, "RequestBuffer")
		errs = errors.Join(errs, err)
	}

	cp = buildCompression(policy.Spec.Compression)
	httpUpgrade = buildHTTPProtocolUpgradeConfig(policy.Spec.HTTPUpgrade)

	ds = translateDNS(policy.Spec.ClusterSettings)

	return &ir.TrafficFeatures{
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
		Timeout:           to,
		ResponseOverride:  ro,
		RequestBuffer:     rb,
		Compression:       cp,
		HTTPUpgrade:       httpUpgrade,
		Telemetry:         policy.Spec.Telemetry,
	}, errs
}

func (t *Translator) translateBackendTrafficPolicyForGateway(
	policy *egv1a1.BackendTrafficPolicy, target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName,
	gateway *GatewayContext, xdsIR resource.XdsIRMap, resources *resource.Resources,
) error {
	tf, errs := t.buildTrafficFeatures(policy, resources)
	if tf == nil {
		// should not happen
		return errs
	}

	// Apply IR to all the routes within the specific Gateway
	// If the feature is already set, then skip it, since it must have
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
			setIfNil(&r.LoadBalancer, tf.LoadBalancer)
			setIfNil(&r.ProxyProtocol, tf.ProxyProtocol)
			setIfNil(&r.HealthCheck, tf.HealthCheck)
			setIfNil(&r.CircuitBreaker, tf.CircuitBreaker)
			setIfNil(&r.TCPKeepalive, tf.TCPKeepalive)
			setIfNil(&r.Timeout, tf.Timeout)
			setIfNil(&r.DNS, tf.DNS)
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
		setIfNil(&route.LoadBalancer, tf.LoadBalancer)
		setIfNil(&route.DNS, tf.DNS)
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

			if errs != nil {
				// Return a 500 direct response
				r.DirectResponse = &ir.CustomResponse{
					StatusCode: ptr.To(uint32(500)),
				}
				continue
			}

			r.Traffic = tf.DeepCopy()

			// Update the Host field in HealthCheck, now that we have access to the Route Hostname.
			r.Traffic.HealthCheck.SetHTTPHostIfAbsent(r.Hostname)

			if ct, err := buildClusterSettingsTimeout(policy.Spec.ClusterSettings); err == nil {
				r.Traffic.Timeout = ct
			}

			if policy.Spec.UseClientProtocol != nil {
				setIfNil(&r.UseClientProtocol, policy.Spec.UseClientProtocol)
			}
		}
	}

	return errs
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
		if len(rule.ClientSelectors) == 0 {
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
	defaultLimitUnit := ratelimit.UnitToSeconds(egv1a1.RateLimitUnit(defaultLimit.Unit))
	for _, rule := range local.Rules {
		ruleLimitUint := ratelimit.UnitToSeconds(rule.Limit.Unit)
		if defaultLimitUnit == 0 || ruleLimitUint%defaultLimitUnit != 0 {
			return nil, fmt.Errorf("local rateLimit rule limit unit must be a multiple of the default limit unit")
		}
	}

	var err error
	var irRule *ir.RateLimitRule
	irRules := make([]*ir.RateLimitRule, 0)
	for i, rule := range local.Rules {
		// We don't process the rule without clientSelectors here because it's
		// previously used as the default route-level limit.
		if len(rule.ClientSelectors) == 0 {
			continue
		}

		irRule, err = buildRateLimitRule(rule)
		if err != nil {
			return nil, err
		}
		// Set the Name field as <policy-ns>/<policy-name>/rule/<rule-index>
		irRule.Name = irRuleName(policy.Namespace, policy.Name, i)
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
		// Set the Name field as <policy-ns>/<policy-name>/rule/<rule-index>
		irRules[i].Name = irRuleName(policy.Namespace, policy.Name, i)
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
		Shared:        rule.Shared,
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
					Name:   header.Name,
					Exact:  header.Value,
					Invert: header.Invert,
				}
				irRule.HeaderMatches = append(irRule.HeaderMatches, m)
			case *header.Type == egv1a1.HeaderMatchRegularExpression && header.Value != nil:
				if err := regex.Validate(*header.Value); err != nil {
					return nil, err
				}
				m := &ir.StringMatch{
					Name:      header.Name,
					SafeRegex: header.Value,
					Invert:    header.Invert,
				}
				irRule.HeaderMatches = append(irRule.HeaderMatches, m)
			case *header.Type == egv1a1.HeaderMatchDistinct && header.Value == nil:
				if header.Invert != nil && *header.Invert {
					return nil, fmt.Errorf("unable to translate rateLimit." +
						"Invert is not applicable for distinct header match type")
				}
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

	if cost := rule.Cost; cost != nil {
		if cost.Request != nil {
			irRule.RequestCost = translateRateLimitCost(cost.Request)
		}
		if cost.Response != nil {
			irRule.ResponseCost = translateRateLimitCost(cost.Response)
		}
	}
	return irRule, nil
}

func translateRateLimitCost(cost *egv1a1.RateLimitCostSpecifier) *ir.RateLimitCost {
	ret := &ir.RateLimitCost{}
	if cost.Number != nil {
		ret.Number = cost.Number
	}
	if cost.Metadata != nil {
		ret.Format = ptr.To(fmt.Sprintf("%%DYNAMIC_METADATA(%s:%s)%%",
			cost.Metadata.Namespace, cost.Metadata.Key))
	}
	return ret
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

func buildRequestBuffer(spec *egv1a1.RequestBuffer) (*ir.RequestBuffer, error) {
	if spec == nil {
		return nil, nil
	}

	if _, ok := spec.Limit.AsInt64(); !ok {
		return nil, fmt.Errorf("limit must be convertible to an int64")
	}

	return &ir.RequestBuffer{
		Limit: spec.Limit,
	}, nil
}

func buildResponseOverride(policy *egv1a1.BackendTrafficPolicy, resources *resource.Resources) (*ir.ResponseOverride, error) {
	if len(policy.Spec.ResponseOverride) == 0 {
		return nil, nil
	}

	rules := make([]ir.ResponseOverrideRule, 0, len(policy.Spec.ResponseOverride))
	for index, ro := range policy.Spec.ResponseOverride {
		match := ir.CustomResponseMatch{
			StatusCodes: make([]ir.StatusCodeMatch, 0, len(ro.Match.StatusCodes)),
		}

		for _, code := range ro.Match.StatusCodes {
			if code.Type != nil && *code.Type == egv1a1.StatusCodeValueTypeRange {
				match.StatusCodes = append(match.StatusCodes, ir.StatusCodeMatch{
					Range: &ir.StatusCodeRange{
						Start: code.Range.Start,
						End:   code.Range.End,
					},
				})
			} else {
				match.StatusCodes = append(match.StatusCodes, ir.StatusCodeMatch{
					Value: code.Value,
				})
			}
		}

		if ro.Redirect != nil {
			redirect := &ir.Redirect{
				Scheme: ro.Redirect.Scheme,
			}
			if ro.Redirect.Path != nil {
				redirect.Path = &ir.HTTPPathModifier{
					FullReplace:        ro.Redirect.Path.ReplaceFullPath,
					PrefixMatchReplace: ro.Redirect.Path.ReplacePrefixMatch,
				}
			}
			if ro.Redirect.Hostname != nil {
				redirect.Hostname = ptr.To(string(*ro.Redirect.Hostname))
			}
			if ro.Redirect.Port != nil {
				redirect.Port = ptr.To(uint32(*ro.Redirect.Port))
			}
			if ro.Redirect.StatusCode != nil {
				redirect.StatusCode = ptr.To(int32(*ro.Redirect.StatusCode))
			}

			rules = append(rules, ir.ResponseOverrideRule{
				Name:     defaultResponseOverrideRuleName(policy, index),
				Match:    match,
				Redirect: redirect,
			})
		} else {
			response := &ir.CustomResponse{
				ContentType: ro.Response.ContentType,
			}

			if ro.Response.StatusCode != nil {
				response.StatusCode = ptr.To(uint32(*ro.Response.StatusCode))
			}

			var err error
			response.Body, err = getCustomResponseBody(ro.Response.Body, resources, policy.Namespace)
			if err != nil {
				return nil, err
			}

			rules = append(rules, ir.ResponseOverrideRule{
				Name:     defaultResponseOverrideRuleName(policy, index),
				Match:    match,
				Response: response,
			})
		}
	}
	return &ir.ResponseOverride{
		Name:  irConfigName(policy),
		Rules: rules,
	}, nil
}

func checkResponseBodySize(b *string) error {
	// Make this configurable in the future
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route.proto.html#max_direct_response_body_size_bytes
	maxDirectResponseSize := 4096
	lenB := len(*b)
	if lenB > maxDirectResponseSize {
		return fmt.Errorf("response.body size %d greater than the max size %d", lenB, maxDirectResponseSize)
	}

	return nil
}

func getCustomResponseBody(body *egv1a1.CustomResponseBody, resources *resource.Resources, policyNs string) (*string, error) {
	if body != nil && body.Type != nil && *body.Type == egv1a1.ResponseValueTypeValueRef {
		cm := resources.GetConfigMap(policyNs, string(body.ValueRef.Name))
		if cm != nil {
			b, dataOk := cm.Data["response.body"]
			switch {
			case dataOk:
				if err := checkResponseBodySize(&b); err != nil {
					return nil, err
				}
				return &b, nil
			case len(cm.Data) > 0: // Fallback to the first key if response.body is not found
				for _, value := range cm.Data {
					b = value
					break
				}
				if err := checkResponseBodySize(&b); err != nil {
					return nil, err
				}
				return &b, nil
			default:
				return nil, fmt.Errorf("can't find the key response.body in the referenced configmap %s", body.ValueRef.Name)
			}

		} else {
			return nil, fmt.Errorf("can't find the referenced configmap %s", body.ValueRef.Name)
		}
	} else if body != nil && body.Inline != nil {
		if err := checkResponseBodySize(body.Inline); err != nil {
			return nil, err
		}
		return body.Inline, nil
	}

	return nil, nil
}

func defaultResponseOverrideRuleName(policy *egv1a1.BackendTrafficPolicy, index int) string {
	return fmt.Sprintf(
		"%s/responseoverride/rule/%s",
		irConfigName(policy),
		strconv.Itoa(index))
}

func buildCompression(compression []*egv1a1.Compression) []*ir.Compression {
	if compression == nil {
		return nil
	}
	irCompression := make([]*ir.Compression, 0, len(compression))
	for _, c := range compression {
		irCompression = append(irCompression, &ir.Compression{
			Type: c.Type,
		})
	}

	return irCompression
}

func buildHTTPProtocolUpgradeConfig(cfgs []*egv1a1.ProtocolUpgradeConfig) []ir.HTTPUpgradeConfig {
	if len(cfgs) == 0 {
		return nil
	}

	result := make([]ir.HTTPUpgradeConfig, 0, len(cfgs))
	for _, cfg := range cfgs {
		upgrade := ir.HTTPUpgradeConfig{
			Type: cfg.Type,
		}
		if cfg.Connect != nil {
			upgrade.Connect = &ir.ConnectConfig{
				Terminate: ptr.Deref(cfg.Connect.Terminate, false),
			}
		}
		result = append(result, upgrade)
	}

	return result
}
