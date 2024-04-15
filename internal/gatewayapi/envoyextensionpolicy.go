// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"sort"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"

	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/status"
	"github.com/envoyproxy/gateway/internal/utils"
)

func (t *Translator) ProcessEnvoyExtensionPolicies(envoyExtensionPolicies []*egv1a1.EnvoyExtensionPolicy,
	gateways []*GatewayContext,
	routes []RouteContext,
	resources *Resources,
	xdsIR XdsIRMap) []*egv1a1.EnvoyExtensionPolicy {
	var res []*egv1a1.EnvoyExtensionPolicy

	// Sort based on timestamp
	sort.Slice(envoyExtensionPolicies, func(i, j int) bool {
		return envoyExtensionPolicies[i].CreationTimestamp.Before(&(envoyExtensionPolicies[j].CreationTimestamp))
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

	// Translate
	// 1. First translate Policies targeting xRoutes
	// 2. Finally, the policies targeting Gateways

	// Process the policies targeting xRoutes
	for _, policy := range envoyExtensionPolicies {
		if policy.Spec.TargetRef.Kind != KindGateway {
			policy := policy.DeepCopy()
			res = append(res, policy)

			// Negative statuses have already been assigned so its safe to skip
			route, resolveErr := resolveEEPolicyRouteTargetRef(policy, routeMap)
			if route == nil {
				continue
			}

			// Find the Gateway that the route belongs to and add it to the
			// gatewayRouteMap and ancestor list, which will be used to check
			// policy overrides and populate its ancestor status.
			parentRefs := GetParentReferences(route)
			ancestorRefs := make([]gwv1a2.ParentReference, 0, len(parentRefs))
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
			if err := t.translateEnvoyExtensionPolicyForRoute(policy, route, xdsIR, resources); err != nil {
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

	// Process the policies targeting Gateways
	for _, policy := range envoyExtensionPolicies {
		if policy.Spec.TargetRef.Kind == KindGateway {
			policy := policy.DeepCopy()
			res = append(res, policy)

			// Negative statuses have already been assigned so its safe to skip
			gateway, resolveErr := resolveEEPolicyGatewayTargetRef(policy, gatewayMap)
			if gateway == nil {
				continue
			}

			// Find its ancestor reference by resolved gateway, even with resolve error
			gatewayNN := utils.NamespacedName(gateway)
			ancestorRefs := []gwv1a2.ParentReference{
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
			if err := t.translateEnvoyExtensionPolicyForGateway(policy, gateway, xdsIR, resources); err != nil {
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
				message := fmt.Sprintf("This policy is being overridden by other envoyExtensionPolicies for these routes: %v", routes)

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

	return res
}

func resolveEEPolicyGatewayTargetRef(policy *egv1a1.EnvoyExtensionPolicy, gateways map[types.NamespacedName]*policyGatewayTargetContext) (*GatewayContext, *status.PolicyResolveError) {
	targetNs := policy.Spec.TargetRef.Namespace
	// If empty, default to namespace of policy
	if targetNs == nil {
		targetNs = ptr.To(gwv1b1.Namespace(policy.Namespace))
	}

	// Check if the gateway exists
	key := types.NamespacedName{
		Name:      string(policy.Spec.TargetRef.Name),
		Namespace: string(*targetNs),
	}
	gateway, ok := gateways[key]

	// Gateway not found
	if !ok {
		return nil, nil
	}

	// Ensure Policy and target are in the same namespace
	if policy.Namespace != string(*targetNs) {
		message := fmt.Sprintf("Namespace:%s TargetRef.Namespace:%s, EnvoyExtensionPolicy can only target a resource in the same namespace.",
			policy.Namespace, *targetNs)

		return gateway.GatewayContext, &status.PolicyResolveError{
			Reason:  gwv1a2.PolicyReasonInvalid,
			Message: message,
		}
	}

	// Check if another policy targeting the same Gateway exists
	if gateway.attached {
		message := "Unable to target Gateway, another EnvoyExtensionPolicy has already attached to it"

		return gateway.GatewayContext, &status.PolicyResolveError{
			Reason:  gwv1a2.PolicyReasonConflicted,
			Message: message,
		}
	}

	// Set context and save
	gateway.attached = true
	gateways[key] = gateway

	return gateway.GatewayContext, nil
}

func resolveEEPolicyRouteTargetRef(policy *egv1a1.EnvoyExtensionPolicy, routes map[policyTargetRouteKey]*policyRouteTargetContext) (RouteContext, *status.PolicyResolveError) {
	targetNs := policy.Spec.TargetRef.Namespace
	// If empty, default to namespace of policy
	if targetNs == nil {
		targetNs = ptr.To(gwv1b1.Namespace(policy.Namespace))
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
		return nil, nil
	}

	// Ensure Policy and target are in the same namespace
	if policy.Namespace != string(*targetNs) {
		message := fmt.Sprintf("Namespace:%s TargetRef.Namespace:%s, EnvoyExtensionPolicy can only target a resource in the same namespace.",
			policy.Namespace, *targetNs)

		return route.RouteContext, &status.PolicyResolveError{
			Reason:  gwv1a2.PolicyReasonInvalid,
			Message: message,
		}
	}

	// Check if another policy targeting the same xRoute exists
	if route.attached {
		message := fmt.Sprintf("Unable to target %s, another EnvoyExtensionPolicy has already attached to it",
			string(policy.Spec.TargetRef.Kind))

		return route.RouteContext, &status.PolicyResolveError{
			Reason:  gwv1a2.PolicyReasonConflicted,
			Message: message,
		}
	}

	// Set context and save
	route.attached = true
	routes[key] = route

	return route.RouteContext, nil
}

func (t *Translator) translateEnvoyExtensionPolicyForRoute(policy *egv1a1.EnvoyExtensionPolicy, route RouteContext,
	xdsIR XdsIRMap, resources *Resources) error {
	// Apply IR to all relevant routes
	prefix := irRoutePrefix(route)
	for _, ir := range xdsIR {
		for _, http := range ir.HTTP {
			for _, r := range http.Routes {
				// Apply if there is a match
				if strings.HasPrefix(r.Name, prefix) {
					if extProcs, err := t.buildExtProcs(policy, resources); err == nil {
						r.ExtProcs = extProcs
					} else {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (t *Translator) buildExtProcs(policy *egv1a1.EnvoyExtensionPolicy, resources *Resources) ([]ir.ExtProc, error) {
	var extProcIRList []ir.ExtProc

	if policy == nil {
		return nil, nil
	}

	if len(policy.Spec.ExtProc) > 0 {
		for idx, ep := range policy.Spec.ExtProc {
			name := irConfigNameForEEP(policy, idx)
			extProcIR, err := t.buildExtProc(name, utils.NamespacedName(policy), ep, idx, resources)
			if err != nil {
				return nil, err
			}
			extProcIRList = append(extProcIRList, *extProcIR)
		}
	}
	return extProcIRList, nil
}

func (t *Translator) translateEnvoyExtensionPolicyForGateway(policy *egv1a1.EnvoyExtensionPolicy,
	gateway *GatewayContext, xdsIR XdsIRMap, resources *Resources) error {

	irKey := t.getIRKey(gateway.Gateway)
	// Should exist since we've validated this
	ir := xdsIR[irKey]

	policyTarget := irStringKey(
		string(ptr.Deref(policy.Spec.TargetRef.Namespace, gwv1a2.Namespace(policy.Namespace))),
		string(policy.Spec.TargetRef.Name),
	)

	extProcs, err := t.buildExtProcs(policy, resources)
	if err != nil {
		return err
	}

	for _, http := range ir.HTTP {
		gatewayName := http.Name[0:strings.LastIndex(http.Name, "/")]
		if t.MergeGateways && gatewayName != policyTarget {
			continue
		}

		// A Policy targeting the most specific scope(xRoute) wins over a policy
		// targeting a lesser specific scope(Gateway).
		for _, r := range http.Routes {
			// if already set - there's a route level policy, so skip
			if r.ExtProcs == nil {
				r.ExtProcs = extProcs
			}
		}
	}

	return nil
}

func (t *Translator) buildExtProc(
	name string,
	policyNamespacedName types.NamespacedName,
	extProc egv1a1.ExtProc,
	extProcIdx int,
	resources *Resources) (*ir.ExtProc, error) {
	var (
		ds        *ir.DestinationSetting
		authority string
		err       error
	)

	var dsl []*ir.DestinationSetting
	for i := range extProc.BackendRefs {
		if err = t.validateExtServiceBackendReference(
			&extProc.BackendRefs[i].BackendObjectReference,
			policyNamespacedName.Namespace,
			resources); err != nil {
			return nil, err
		}

		ds, err = t.processExtServiceDestination(
			&extProc.BackendRefs[i].BackendObjectReference,
			policyNamespacedName,
			egv1a1.KindEnvoyExtensionPolicy,
			ir.GRPC,
			resources)

		if err != nil {
			return nil, err
		}

		dsl = append(dsl, ds)
	}

	rd := ir.RouteDestination{
		Name:     irIndexedExtServiceDestinationName(policyNamespacedName, egv1a1.KindEnvoyExtensionPolicy, extProcIdx),
		Settings: dsl,
	}

	authority = fmt.Sprintf(
		"%s.%s:%d",
		extProc.BackendRefs[0].Name,
		NamespaceDerefOr(extProc.BackendRefs[0].Namespace, policyNamespacedName.Namespace),
		*extProc.BackendRefs[0].Port)

	extProcIR := &ir.ExtProc{
		Name:        name,
		Destination: rd,
		Authority:   authority,
	}

	if extProc.MessageTimeout != nil {
		d, err := time.ParseDuration(string(*extProc.MessageTimeout))
		if err != nil {
			return nil, fmt.Errorf("invalid ExtProc MessageTimeout value %v", extProc.MessageTimeout)
		}
		extProcIR.MessageTimeout = ptr.To(metav1.Duration{Duration: d})
	}

	if extProc.FailOpen != nil {
		extProcIR.FailOpen = extProc.FailOpen
	}

	return extProcIR, err
}

func irConfigNameForEEP(policy *egv1a1.EnvoyExtensionPolicy, idx int) string {
	return fmt.Sprintf(
		"%s/%s/%d",
		strings.ToLower(egv1a1.KindEnvoyExtensionPolicy),
		utils.NamespacedName(policy).String(),
		idx)
}
