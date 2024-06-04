// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"
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
)

func (t *Translator) ProcessEnvoyExtensionPolicies(envoyExtensionPolicies []*egv1a1.EnvoyExtensionPolicy,
	gateways []*GatewayContext,
	routes []RouteContext,
	resources *Resources,
	xdsIR XdsIRMap,
) []*egv1a1.EnvoyExtensionPolicy {
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

	handledPolicies := make(map[types.NamespacedName]*egv1a1.EnvoyExtensionPolicy)

	// Translate
	// 1. First translate Policies targeting xRoutes
	// 2. Finally, the policies targeting Gateways

	// Process the policies targeting xRoutes
	for _, currPolicy := range envoyExtensionPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := currPolicy.Spec.GetTargetRefs()
		for _, currTarget := range targetRefs {
			if currTarget.Kind != KindGateway {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy.DeepCopy()
					res = append(res, policy)
					handledPolicies[policyName] = policy
				}

				// Negative statuses have already been assigned so its safe to skip
				route, resolveErr := resolveEEPolicyRouteTargetRef(policy, currTarget, routeMap)
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
	}

	// Process the policies targeting Gateways
	for _, currPolicy := range envoyExtensionPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := currPolicy.Spec.GetTargetRefs()
		for _, currTarget := range targetRefs {
			if currTarget.Kind == KindGateway {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy.DeepCopy()
					res = append(res, policy)
					handledPolicies[policyName] = policy
				}

				// Negative statuses have already been assigned so its safe to skip
				gateway, resolveErr := resolveEEPolicyGatewayTargetRef(policy, currTarget, gatewayMap)
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
				if err := t.translateEnvoyExtensionPolicyForGateway(policy, currTarget, gateway, xdsIR, resources); err != nil {
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
	}

	return res
}

func resolveEEPolicyGatewayTargetRef(policy *egv1a1.EnvoyExtensionPolicy, target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName, gateways map[types.NamespacedName]*policyGatewayTargetContext) (*GatewayContext, *status.PolicyResolveError) {
	targetNs := policy.Namespace

	// Check if the gateway exists
	key := types.NamespacedName{
		Name:      string(target.Name),
		Namespace: targetNs,
	}
	gateway, ok := gateways[key]

	// Gateway not found
	if !ok {
		return nil, nil
	}

	// Ensure Policy and target are in the same namespace
	if policy.Namespace != targetNs {
		message := fmt.Sprintf("Namespace:%s TargetRef.Namespace:%s, EnvoyExtensionPolicy can only target a resource in the same namespace.",
			policy.Namespace, targetNs)

		return gateway.GatewayContext, &status.PolicyResolveError{
			Reason:  gwapiv1a2.PolicyReasonInvalid,
			Message: message,
		}
	}

	// Check if another policy targeting the same Gateway exists
	if gateway.attached {
		message := "Unable to target Gateway, another EnvoyExtensionPolicy has already attached to it"

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

func resolveEEPolicyRouteTargetRef(policy *egv1a1.EnvoyExtensionPolicy, target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName, routes map[policyTargetRouteKey]*policyRouteTargetContext) (RouteContext, *status.PolicyResolveError) {
	targetNs := policy.Namespace

	// Check if the route exists
	key := policyTargetRouteKey{
		Kind:      string(target.Kind),
		Name:      string(target.Name),
		Namespace: targetNs,
	}

	route, ok := routes[key]
	// Route not found
	if !ok {
		return nil, nil
	}

	// Ensure Policy and target are in the same namespace
	if policy.Namespace != targetNs {
		message := fmt.Sprintf("Namespace:%s TargetRef.Namespace:%s, EnvoyExtensionPolicy can only target a resource in the same namespace.",
			policy.Namespace, targetNs)

		return route.RouteContext, &status.PolicyResolveError{
			Reason:  gwapiv1a2.PolicyReasonInvalid,
			Message: message,
		}
	}

	// Check if another policy targeting the same xRoute exists
	if route.attached {
		message := fmt.Sprintf("Unable to target %s, another EnvoyExtensionPolicy has already attached to it",
			string(target.Kind))

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

func (t *Translator) translateEnvoyExtensionPolicyForRoute(policy *egv1a1.EnvoyExtensionPolicy, route RouteContext,
	xdsIR XdsIRMap, resources *Resources,
) error {
	var (
		extProcs  []ir.ExtProc
		wasms     []ir.Wasm
		err, errs error
	)

	if wasms, err = t.buildWasms(policy); err != nil {
		err = perr.WithMessage(err, "WASMs")
		errs = errors.Join(errs, err)
	}

	// Early return if got any errors
	if errs != nil {
		return errs
	}

	// Apply IR to all relevant routes
	prefix := irRoutePrefix(route)
	parentRefs := GetParentReferences(route)
	for _, p := range parentRefs {
		parentRefCtx := GetRouteParentContext(route, p)
		if parentRefCtx != nil && parentRefCtx.gateway != nil {
			if extProcs, err = t.buildExtProcs(policy, resources, parentRefCtx.gateway.envoyProxy); err != nil {
				err = perr.WithMessage(err, "ExtProcs")
				errs = errors.Join(errs, err)
				extProcs = nil
			}
			for _, listener := range parentRefCtx.listeners {
				irKey := t.getIRKey(listener.gateway)
				irListener := xdsIR[irKey].GetHTTPListener(irListenerName(listener))
				if irListener != nil {
					for _, r := range irListener.Routes {
						if strings.HasPrefix(r.Name, prefix) {
							r.ExtProcs = extProcs
							r.Wasms = wasms
						}
					}
				}
			}
		}
	}

	return nil
}

func (t *Translator) translateEnvoyExtensionPolicyForGateway(
	policy *egv1a1.EnvoyExtensionPolicy,
	target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName,
	gateway *GatewayContext,
	xdsIR XdsIRMap,
	resources *Resources,
) error {
	var (
		extProcs  []ir.ExtProc
		wasms     []ir.Wasm
		err, errs error
	)

	if extProcs, err = t.buildExtProcs(policy, resources, gateway.envoyProxy); err != nil {
		err = perr.WithMessage(err, "ExtProcs")
		errs = errors.Join(errs, err)
	}
	if wasms, err = t.buildWasms(policy); err != nil {
		err = perr.WithMessage(err, "WASMs")
		errs = errors.Join(errs, err)
	}

	// Early return if got any errors
	if errs != nil {
		return errs
	}

	irKey := t.getIRKey(gateway.Gateway)
	// Should exist since we've validated this
	x := xdsIR[irKey]

	policyTarget := irStringKey(policy.Namespace, string(target.Name))

	for _, http := range x.HTTP {
		gatewayName := http.Name[0:strings.LastIndex(http.Name, "/")]
		if t.MergeGateways && gatewayName != policyTarget {
			continue
		}

		// A Policy targeting the most specific scope(xRoute) wins over a policy
		// targeting a lesser specific scope(Gateway).
		for _, r := range http.Routes {
			// if already set - there's a route level policy, so skip
			if r.ExtProcs != nil ||
				r.Wasms != nil {
				continue
			}

			if r.ExtProcs == nil {
				r.ExtProcs = extProcs
			}
			if r.Wasms == nil {
				r.Wasms = wasms
			}
		}
	}

	return nil
}

func (t *Translator) buildExtProcs(policy *egv1a1.EnvoyExtensionPolicy, resources *Resources, envoyProxy *egv1a1.EnvoyProxy) ([]ir.ExtProc, error) {
	var extProcIRList []ir.ExtProc

	if policy == nil {
		return nil, nil
	}

	for idx, ep := range policy.Spec.ExtProc {
		name := irConfigNameForEEP(policy, idx)
		extProcIR, err := t.buildExtProc(name, utils.NamespacedName(policy), ep, idx, resources, envoyProxy)
		if err != nil {
			return nil, err
		}
		extProcIRList = append(extProcIRList, *extProcIR)
	}
	return extProcIRList, nil
}

func (t *Translator) buildExtProc(
	name string,
	policyNamespacedName types.NamespacedName,
	extProc egv1a1.ExtProc,
	extProcIdx int,
	resources *Resources,
	envoyProxy *egv1a1.EnvoyProxy,
) (*ir.ExtProc, error) {
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
			egv1a1.KindEnvoyExtensionPolicy,
			resources); err != nil {
			return nil, err
		}

		ds, err = t.processExtServiceDestination(
			&extProc.BackendRefs[i].BackendObjectReference,
			policyNamespacedName,
			egv1a1.KindEnvoyExtensionPolicy,
			ir.GRPC,
			resources,
			envoyProxy,
		)
		if err != nil {
			return nil, err
		}

		dsl = append(dsl, ds)
	}

	rd := ir.RouteDestination{
		Name:     irIndexedExtServiceDestinationName(policyNamespacedName, egv1a1.KindEnvoyExtensionPolicy, extProcIdx),
		Settings: dsl,
	}

	if extProc.BackendRefs[0].Port != nil {
		authority = fmt.Sprintf(
			"%s.%s:%d",
			extProc.BackendRefs[0].Name,
			NamespaceDerefOr(extProc.BackendRefs[0].Namespace, policyNamespacedName.Namespace),
			*extProc.BackendRefs[0].Port)
	} else {
		authority = fmt.Sprintf(
			"%s.%s",
			extProc.BackendRefs[0].Name,
			NamespaceDerefOr(extProc.BackendRefs[0].Namespace, policyNamespacedName.Namespace))
	}

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

	if extProc.ProcessingMode != nil {
		if extProc.ProcessingMode.Request != nil {
			extProcIR.RequestHeaderProcessing = true
			if extProc.ProcessingMode.Request.Body != nil {
				extProcIR.RequestBodyProcessingMode = ptr.To(ir.ExtProcBodyProcessingMode(*extProc.ProcessingMode.Request.Body))
			}
		}

		if extProc.ProcessingMode.Response != nil {
			extProcIR.ResponseHeaderProcessing = true
			if extProc.ProcessingMode.Response.Body != nil {
				extProcIR.ResponseBodyProcessingMode = ptr.To(ir.ExtProcBodyProcessingMode(*extProc.ProcessingMode.Response.Body))
			}
		}
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

func (t *Translator) buildWasms(policy *egv1a1.EnvoyExtensionPolicy) ([]ir.Wasm, error) {
	var wasmIRList []ir.Wasm

	if policy == nil {
		return nil, nil
	}

	for idx, wasm := range policy.Spec.Wasm {
		name := irConfigNameForEEP(policy, idx)
		wasmIR, err := t.buildWasm(name, wasm)
		if err != nil {
			return nil, err
		}
		wasmIRList = append(wasmIRList, *wasmIR)
	}
	return wasmIRList, nil
}

func (t *Translator) buildWasm(name string, wasm egv1a1.Wasm) (*ir.Wasm, error) {
	var (
		failOpen     = false
		httpWasmCode *ir.HTTPWasmCode
	)

	if wasm.FailOpen != nil {
		failOpen = *wasm.FailOpen
	}

	switch wasm.Code.Type {
	case egv1a1.HTTPWasmCodeSourceType:
		httpWasmCode = &ir.HTTPWasmCode{
			URL:    wasm.Code.HTTP.URL,
			SHA256: wasm.Code.SHA256,
		}
	case egv1a1.ImageWasmCodeSourceType:
		return nil, fmt.Errorf("OCI image Wasm code source is not supported yet")
	default:
		// should never happen because of kubebuilder validation, just a sanity check
		return nil, fmt.Errorf("unsupported Wasm code source type %q", wasm.Code.Type)
	}

	wasmIR := &ir.Wasm{
		Name:         name,
		RootID:       wasm.RootID,
		WasmName:     wasm.Name,
		Config:       wasm.Config,
		FailOpen:     failOpen,
		HTTPWasmCode: httpWasmCode,
	}

	return wasmIR, nil
}
