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

	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
)

func (t *Translator) ProcessListenerSets(listenerSets []*gwapiv1.ListenerSet, gateways []*GatewayContext) {
	// Create a map for quick lookup of gateways
	gatewayMap := make(map[types.NamespacedName]*GatewayContext)
	for _, gw := range gateways {
		gatewayMap[types.NamespacedName{Namespace: gw.Namespace, Name: gw.Name}] = gw
	}

	for _, ls := range listenerSets {
		t.processListenerSet(ls, gatewayMap)
	}
}

func (t *Translator) processListenerSet(ls *gwapiv1.ListenerSet, gatewayMap map[types.NamespacedName]*GatewayContext) {
	parentNamespace := NamespaceDerefOr(ls.Spec.ParentRef.Namespace, ls.Namespace)

	gatewayKey := types.NamespacedName{Namespace: parentNamespace, Name: string(ls.Spec.ParentRef.Name)}
	gatewayCtx, exists := gatewayMap[gatewayKey]

	// If the Gateway is not found (not managed by us), ignore the ListenerSet completely.
	// Just a sanity check, this should already be handled in the provider layer.
	if !exists {
		return
	}

	var (
		xlsReason gwapiv1.ListenerSetConditionReason
		xlsMsg    string
	)

	// If the Gateway is not accepted, mark the ListenerSet as not accepted
	if status.GatewayNotAccepted(gatewayCtx.Gateway) {
		xlsReason = gwapiv1.ListenerSetReasonParentNotAccepted
		xlsMsg = fmt.Sprintf("Parent Gateway %s/%s not accepted", parentNamespace, ls.Spec.ParentRef.Name)
		status.UpdateListenerSetStatusAccepted(ls, false, xlsReason, xlsMsg)
		status.UpdateListenerSetStatusProgrammed(ls, false, gwapiv1.ListenerSetReasonProgrammed, "Not Programmed")
		return
	}

	// Check if the namespace is allowed
	if !t.isListenerSetAllowed(gatewayCtx.Gateway, ls) {
		xlsReason = gwapiv1.ListenerSetReasonNotAllowed
		xlsMsg = fmt.Sprintf("ListenerSet attachment from namespace %s not allowed by Gateway %s/%s", ls.Namespace, gatewayCtx.Namespace, gatewayCtx.Name)
		status.UpdateListenerSetStatusAccepted(ls, false, xlsReason, xlsMsg)
		status.UpdateListenerSetStatusProgrammed(ls, false, gwapiv1.ListenerSetReasonNotAllowed, "Not Programmed")
		return
	}

	// Attach listeners to the GatewayContext
	// We do NOT update status here. It will be updated in ProcessListenerSetStatus after listeners are processed.
	for i := range ls.Spec.Listeners {
		listener := &ls.Spec.Listeners[i]
		// Initialize listener status conditions
		ls.Status.Listeners = append(ls.Status.Listeners, gwapiv1.ListenerEntryStatus{
			Name:           listener.Name,
			SupportedKinds: []gwapiv1.RouteGroupKind{},
			AttachedRoutes: 0,
			Conditions:     []metav1.Condition{},
		})

		// Convert ListenerSet listener to Gateway listener for internal processing
		gwListener := &gwapiv1.Listener{
			Name:          listener.Name,
			Port:          gwapiv1.PortNumber(listener.Port), //nolint
			Protocol:      listener.Protocol,
			TLS:           listener.TLS,
			AllowedRoutes: listener.AllowedRoutes,
			Hostname:      listener.Hostname,
		}

		listenerCtx := &ListenerContext{
			Listener:             gwListener,
			gateway:              gatewayCtx,
			listenerSet:          ls,
			listenerSetStatusIdx: i,
		}
		gatewayCtx.listeners = append(gatewayCtx.listeners, listenerCtx)
	}
	gatewayCtx.IncreaseAttachedListenerSets()
}

// ProcessListenerSetStatus computes the status of ListenerSets after their listeners have been processed.
func (t *Translator) ProcessListenerSetStatus(listenerSets []*gwapiv1.ListenerSet) {
	for _, ls := range listenerSets {
		// If Accepted condition is already set to False, it means it failed during attachment (parent not found/accepted or not allowed).
		// We skip re-processing.
		alreadyRejected := false
		for _, cond := range ls.Status.Conditions {
			if cond.Type == string(gwapiv1.ListenerSetConditionAccepted) && cond.Status == metav1.ConditionFalse {
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

		for _, lStatus := range ls.Status.Listeners {
			accepted := false
			for _, cond := range lStatus.Conditions {
				if cond.Type == string(gwapiv1.ListenerEntryConditionAccepted) && cond.Status == metav1.ConditionTrue {
					accepted = true
					break
				}
			}
			anyListenerValid = anyListenerValid || accepted
			allListenersValid = allListenersValid && accepted
		}

		var (
			lsAccepted         bool
			lsReason           gwapiv1.ListenerSetConditionReason
			lsProgrammedReason gwapiv1.ListenerSetConditionReason
			lsMsg              string
		)

		switch {
		case allListenersValid:
			lsAccepted = true
			lsReason = gwapiv1.ListenerSetReasonAccepted
			lsProgrammedReason = gwapiv1.ListenerSetReasonProgrammed
			lsMsg = "All listeners are valid"
		case anyListenerValid: // TODO: implement PartiallyInvalid conditions when Gateway API supports it
			lsAccepted = true
			lsReason = gwapiv1.ListenerSetReasonListenersNotValid
			lsProgrammedReason = gwapiv1.ListenerSetReasonProgrammed
			lsMsg = "Some listeners are invalid"
		default:
			lsAccepted = false
			lsReason = gwapiv1.ListenerSetReasonListenersNotValid
			lsProgrammedReason = gwapiv1.ListenerSetReasonListenersNotValid
			lsMsg = "All listeners are invalid"
		}

		status.UpdateListenerSetStatusAccepted(ls, lsAccepted, lsReason, lsMsg)
		status.UpdateListenerSetStatusProgrammed(ls, lsAccepted, lsProgrammedReason, lsMsg)
	}
}

func (t *Translator) isListenerSetAllowed(gateway *gwapiv1.Gateway, ls *gwapiv1.ListenerSet) bool {
	// If AllowedListeners is not set, attachment is not allowed (default is None)
	if gateway.Spec.AllowedListeners == nil || gateway.Spec.AllowedListeners.Namespaces == nil || gateway.Spec.AllowedListeners.Namespaces.From == nil {
		return false
	}

	from := *gateway.Spec.AllowedListeners.Namespaces.From

	switch from {
	case gwapiv1.NamespacesFromAll:
		return true
	case gwapiv1.NamespacesFromSame:
		return gateway.Namespace == ls.Namespace
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
		// We need to look up the namespace of the ListenerSet to check labels
		// translatorContext has NamespaceMap
		ns := t.GetNamespace(ls.Namespace)
		if ns != nil {
			return selector.Matches(labels.Set(ns.Labels))
		}
		return false
	}

	return false
}
