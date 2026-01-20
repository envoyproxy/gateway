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

	t.Logger.Info("Processing XListenerSet 1", "namespace", xls.Namespace, "name", xls.Name)

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

	t.Logger.Info("Processing XListenerSet 2", "namespace", xls.Namespace, "name", xls.Name)

	// Check if the namespace is allowed
	if !t.isXListenerSetAllowed(gatewayCtx.Gateway, xls) {
		xlsReason = gwapixv1a1.ListenerSetReasonNotAllowed
		xlsMsg = fmt.Sprintf("XListenerSet attachment from namespace %s not allowed by Gateway %s/%s", xls.Namespace, gatewayCtx.Namespace, gatewayCtx.Name)
		status.UpdateXListenerSetStatusAccepted(xls, false, xlsReason, xlsMsg)
		status.UpdateXListenerSetStatusProgrammed(xls, false, gwapixv1a1.ListenerSetReasonProgrammed, "Not Programmed")
		return
	}

	t.Logger.Info("Processing XListenerSet 3", "namespace", xls.Namespace, "name", xls.Name)

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
		t.Logger.Info("Processing XListenerSet 3.1", "namespace", xls.Namespace, "name", xls.Name)
	}

	t.Logger.Info("Processing XListenerSet 4", "namespace", xls.Namespace, "name", xls.Name)
}

// ProcessXListenerSetStatus computes the status of XListenerSets after their listeners have been processed.
func (t *Translator) ProcessXListenerSetStatus(xListenerSets []*gwapixv1a1.XListenerSet) {
	for _, xls := range xListenerSets {
		// If Accepted condition is already set to False, it means it failed during attachment (parent not found/accepted or not allowed).
		// We skip re-processing.
		// If Accepted is not set (or True from previous reconciliation?), we assume it was attached and we need to validate listeners.
		// Note: Since we re-process resources from scratch, Status.Conditions should be from the fresh object or we need to check if we just updated it.
		// My implementation of processXListenerSet only updates status on failure.
		// So if we don't find a False Accepted condition, we proceed.

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
			// Check if listener is accepted and not conflicted
			accepted := false
			conflicted := false
			for _, cond := range lStatus.Conditions {
				if cond.Type == string(gwapixv1a1.ListenerEntryConditionAccepted) && cond.Status == metav1.ConditionTrue {
					accepted = true
				}
				if cond.Type == string(gwapixv1a1.ListenerEntryConditionConflicted) && cond.Status == metav1.ConditionTrue {
					conflicted = true
				}
			}

			if accepted && !conflicted {
				anyListenerValid = true
			} else {
				allListenersValid = false
			}
		}

		xlsAccepted := true
		xlsReason := gwapixv1a1.ListenerSetReasonAccepted
		xlsMsg := "Attached to Gateway"

		if !allListenersValid {
			// If not all listeners are valid, we still Accept the ListenerSet (partial success), but warn.
			// If NONE are valid, we also Accept but with invalid listeners reason?
			// The spec says: "This can be the reason when 'Accepted' is 'True' or 'False'..."
			// We choose True to allow partial success updates if supported, or at least indicate the Set itself is valid.
			xlsReason = gwapixv1a1.ListenerSetReasonListenersNotValid
			xlsMsg = "Some listeners are invalid or conflicted"
			if !anyListenerValid {
				// If no listeners are valid, should we set Accepted=False?
				// Gateway API usually prefers Accepted=True if the resource structure is valid.
				xlsMsg = "All listeners are invalid or conflicted"
			}
		}

		status.UpdateXListenerSetStatusAccepted(xls, xlsAccepted, xlsReason, xlsMsg)
		status.UpdateXListenerSetStatusProgrammed(xls, xlsAccepted, gwapixv1a1.ListenerSetReasonProgrammed, "Programmed")
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
