// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapixv1a1 "sigs.k8s.io/gateway-api/apisx/v1alpha1"

	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
)

func (t *Translator) ProcessXListenerSets(xListenerSets []*gwapixv1a1.XListenerSet, gateways []*GatewayContext) {
	// Create a map for quick lookup of gateways
	gatewayMap := make(map[types.NamespacedName]*GatewayContext)
	for _, gw := range gateways {
		gatewayMap[types.NamespacedName{Namespace: gw.Namespace, Name: gw.Name}] = gw
	}

	for _, xls := range xListenerSets {
		t.processXListenerSet(xls, gatewayMap)
	}
}

func (t *Translator) processXListenerSet(xls *gwapixv1a1.XListenerSet, gatewayMap map[types.NamespacedName]*GatewayContext) {
	parentNamespace := NamespaceDerefOr(xls.Spec.ParentRef.Namespace, xls.Namespace)

	gatewayKey := types.NamespacedName{Namespace: parentNamespace, Name: string(xls.Spec.ParentRef.Name)}
	gatewayCtx, exists := gatewayMap[gatewayKey]

	// If the Gateway is not found (not managed by us), ignore the XListenerSet completely.
	// Just a sanity check, this should already be handled in the provider layer.
	if !exists {
		return
	}

	var (
		xlsReason gwapixv1a1.ListenerSetConditionReason
		xlsMsg    string
	)

	// If the Gateway is not accepted, mark the XListenerSet as not accepted
	if status.GatewayNotAccepted(gatewayCtx.Gateway) {
		xlsReason = gwapixv1a1.ListenerSetReasonParentNotAccepted
		xlsMsg = fmt.Sprintf("Parent Gateway %s/%s not accepted", parentNamespace, xls.Spec.ParentRef.Name)
		status.UpdateXListenerSetStatusAccepted(xls, false, xlsReason, xlsMsg)
		status.UpdateXListenerSetStatusProgrammed(xls, false, gwapixv1a1.ListenerSetReasonProgrammed, "Not Programmed")
		return
	}

	// Check if the namespace is allowed
	if !t.isXListenerSetAllowed(gatewayCtx.Gateway, xls) {
		xlsReason = gwapixv1a1.ListenerSetReasonNotAllowed
		xlsMsg = fmt.Sprintf("XListenerSet attachment from namespace %s not allowed by Gateway %s/%s", xls.Namespace, gatewayCtx.Namespace, gatewayCtx.Name)
		status.UpdateXListenerSetStatusAccepted(xls, false, xlsReason, xlsMsg)
		status.UpdateXListenerSetStatusProgrammed(xls, false, gwapixv1a1.ListenerSetReasonProgrammed, "Not Programmed")
		return
	}

	// Attach listeners to the GatewayContext
	// We do NOT update status here. It will be updated in ProcessXListenerSetStatus after listeners are processed.
	for i := range xls.Spec.Listeners {
		listener := &xls.Spec.Listeners[i]
		// Initialize listener status conditions
		xls.Status.Listeners = append(xls.Status.Listeners, gwapixv1a1.ListenerEntryStatus{
			Name:           listener.Name,
			Port:           listener.Port,
			SupportedKinds: []gwapixv1a1.RouteGroupKind{},
			AttachedRoutes: 0,
			Conditions:     []metav1.Condition{},
		})

		// Convert XListenerSet listener to Gateway listener for internal processing
		gwListener := &gwapiv1.Listener{
			Name:          listener.Name,
			Port:          listener.Port,
			Protocol:      listener.Protocol,
			TLS:           listener.TLS,
			AllowedRoutes: listener.AllowedRoutes,
			Hostname:      listener.Hostname,
		}

		listenerCtx := &ListenerContext{
			Listener:              gwListener,
			gateway:               gatewayCtx,
			xListenerSet:          xls,
			xListenerSetStatusIdx: i,
		}
		gatewayCtx.listeners = append(gatewayCtx.listeners, listenerCtx)
	}
}

// ProcessXListenerSetStatus computes the status of XListenerSets after their listeners have been processed.
func (t *Translator) ProcessXListenerSetStatus(xListenerSets []*gwapixv1a1.XListenerSet) {
	for _, xls := range xListenerSets {
		// If Accepted condition is already set to False, it means it failed during attachment (parent not found/accepted or not allowed).
		// We skip re-processing.
		alreadyRejected := false
		for _, cond := range xls.Status.Conditions {
			if cond.Type == string(gwapixv1a1.ListenerSetConditionAccepted) && cond.Status == metav1.ConditionFalse {
				alreadyRejected = true
				break
			}
		}
		if alreadyRejected {
			continue
		}

		// Calculate status based on listeners
		allListenersValid := true
		anyListenerValid := false

		for _, lStatus := range xls.Status.Listeners {
			accepted := false
			for _, cond := range lStatus.Conditions {
				if cond.Type == string(gwapixv1a1.ListenerEntryConditionAccepted) && cond.Status == metav1.ConditionTrue {
					accepted = true
					break
				}
			}
			anyListenerValid = anyListenerValid || accepted
			allListenersValid = allListenersValid && accepted
		}

		var (
			xlsAccepted         bool
			xlsReason           gwapixv1a1.ListenerSetConditionReason
			xlsProgrammedReason gwapixv1a1.ListenerSetConditionReason
			xlsMsg              string
		)

		switch {
		case allListenersValid:
			xlsAccepted = true
			xlsReason = gwapixv1a1.ListenerSetReasonAccepted
			xlsProgrammedReason = gwapixv1a1.ListenerSetReasonProgrammed
			xlsMsg = "All listeners are valid"
		case anyListenerValid: // TODO: implement PartiallyInvalid conditions when Gateway API supports it
			xlsAccepted = true
			xlsReason = gwapixv1a1.ListenerSetReasonListenersNotValid
			xlsProgrammedReason = gwapixv1a1.ListenerSetReasonProgrammed
			xlsMsg = "Some listeners are invalid"
		default:
			xlsAccepted = false
			xlsReason = gwapixv1a1.ListenerSetReasonListenersNotValid
			xlsProgrammedReason = gwapixv1a1.ListenerSetReasonInvalid
			xlsMsg = "All listeners are invalid"
		}

		status.UpdateXListenerSetStatusAccepted(xls, xlsAccepted, xlsReason, xlsMsg)
		status.UpdateXListenerSetStatusProgrammed(xls, xlsAccepted, xlsProgrammedReason, xlsMsg)
	}
}

func (t *Translator) isXListenerSetAllowed(gateway *gwapiv1.Gateway, xls *gwapixv1a1.XListenerSet) bool {
	// If AllowedListeners is not set, attachment is not allowed (default is None)
	if gateway.Spec.AllowedListeners == nil || gateway.Spec.AllowedListeners.Namespaces == nil || gateway.Spec.AllowedListeners.Namespaces.From == nil {
		return false
	}

	from := *gateway.Spec.AllowedListeners.Namespaces.From

	switch from {
	case gwapiv1.NamespacesFromAll:
		return true
	case gwapiv1.NamespacesFromSame:
		return gateway.Namespace == xls.Namespace
	case gwapiv1.NamespacesFromSelector:
		selectorVal := gateway.Spec.AllowedListeners.Namespaces.Selector
		if selectorVal == nil {
			return false
		}
		selector, err := metav1.LabelSelectorAsSelector(selectorVal)
		if err != nil {
			t.Logger.Error(err, "invalid label selector in AllowedListeners", "gateway", gateway.Name)
			return false
		}
		// We need to look up the namespace of the XListenerSet to check labels
		// translatorContext has NamespaceMap
		ns := t.GetNamespace(xls.Namespace)
		if ns != nil {
			return selector.Matches(labels.Set(ns.Labels))
		}
		return false
	}

	return false
}
