// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// policyIndexEntry pairs a policy's value at some scope with effective: whether this entry
// itself is the final answer for that scope, or must transparently fall through to a broader
// parent scope instead. What decides effective differs by scope kind (see isRouteEffective and
// setGatewayListenerLevel's hasValue), but once computed, resolution reads every scope the same
// way.
type policyIndexEntry[T comparable] struct {
	value     T
	effective bool
}

// policyIndex resolves a route-rule/route/listener/listenerSet/gateway target through the same
// attachment order a policy itself uses, including MergeType-driven inheritance.
type policyIndex[T comparable] struct {
	entries map[policyScope]policyIndexEntry[T]
}

// newPolicyIndex allocates a policyIndex's map.
func newPolicyIndex[T comparable]() *policyIndex[T] {
	return &policyIndex[T]{entries: make(map[policyScope]policyIndexEntry[T])}
}

// putFirst claims scope's first-registered entry; later calls for the same scope are ignored.
func (idx *policyIndex[T]) putFirst(scope policyScope, entry policyIndexEntry[T]) {
	if _, exists := idx.entries[scope]; !exists {
		idx.entries[scope] = entry
	}
}

// isRouteEffective reports whether a route-rule/route-level policy is effective at its own
// scope: an explicit value always is, and so is an unset value whose MergeType is nil, since nil
// means the policy doesn't inherit from any parent at all.
func isRouteEffective[T comparable](value T, mergeType *egv1a1.MergeType) bool {
	var zero T
	return value != zero || mergeType == nil
}

// setRouteRuleLevel records a route-rule target's first-registered entry.
func (idx *policyIndex[T]) setRouteRuleLevel(route types.NamespacedName, routeKind string, rule gwapiv1.SectionName, value T, mergeType *egv1a1.MergeType) {
	idx.putFirst(routeRuleScope(route, routeKind, rule), policyIndexEntry[T]{value: value, effective: isRouteEffective(value, mergeType)})
}

// setRouteLevel is setRouteRuleLevel's counterpart for a route (not route-rule) target.
func (idx *policyIndex[T]) setRouteLevel(route types.NamespacedName, routeKind string, value T, mergeType *egv1a1.MergeType) {
	idx.putFirst(routeScope(route, routeKind), policyIndexEntry[T]{value: value, effective: isRouteEffective(value, mergeType)})
}

// setGatewayListenerLevel records a Gateway-listener target's first-registered entry. hasValue
// decides whether Lookup uses value directly or falls through to its parent. This needs to know
// which of three states value is in: not populated, populated with T's zero/default value, or
// populated with a non-default value - but the first two look identical by inspecting value
// alone. An Option[T] (Some(T) vs. None) would separate these for free; Go generics have no
// built-in optional, so hasValue is the caller-supplied Some/None signal instead. A
// listener-targeted policy can't set MergeType (CEL-enforced), so isRouteEffective's rule doesn't
// apply here.
func (idx *policyIndex[T]) setGatewayListenerLevel(gw types.NamespacedName, listener gwapiv1.SectionName, value T, hasValue bool) {
	idx.putFirst(gatewayListenerScope(gw, listener), policyIndexEntry[T]{value: value, effective: hasValue})
}

// setGatewayLevel records a Gateway target's first-registered value, always effective: it's the
// last level for its own scope, so there's no parent an unset value could ever wrongly block.
func (idx *policyIndex[T]) setGatewayLevel(gw types.NamespacedName, value T) {
	idx.putFirst(gatewayScope(gw), policyIndexEntry[T]{value: value, effective: true})
}

// setListenerSetListenerLevel is setGatewayListenerLevel's counterpart for a ListenerSet listener.
func (idx *policyIndex[T]) setListenerSetListenerLevel(ls types.NamespacedName, listener gwapiv1.SectionName, value T, hasValue bool) {
	idx.putFirst(listenerSetListenerScope(ls, listener), policyIndexEntry[T]{value: value, effective: hasValue})
}

// setListenerSetLevel is setGatewayLevel's counterpart for a ListenerSet.
func (idx *policyIndex[T]) setListenerSetLevel(ls types.NamespacedName, value T) {
	idx.putFirst(listenerSetScope(ls), policyIndexEntry[T]{value: value, effective: true})
}

// Lookup resolves the effective value for a route-rule/route/listener/listenerSet/gateway target.
// value alone is always correct; pinned reports whether a route-rule/route entry supplied it
// directly. listenerSetNN is nil unless the route attaches through a ListenerSet, in which case
// Gateway listener/gateway scopes are skipped: Gateway listeners and ListenerSet listeners are
// sibling scopes, not parent/child.
func (idx *policyIndex[T]) Lookup(
	routeKind gwapiv1.Kind,
	routeNN types.NamespacedName,
	gatewayNN types.NamespacedName,
	listenerName *gwapiv1.SectionName,
	listenerSetNN *types.NamespacedName,
	routeRuleName *gwapiv1.SectionName,
) (value T, pinned bool) {
	if idx == nil {
		return value, false
	}

	if routeRuleName != nil {
		key := routeRuleScope(routeNN, string(routeKind), *routeRuleName)
		if entry, found := idx.entries[key]; found {
			return idx.resolveEntry(entry, gatewayNN, listenerName, listenerSetNN)
		}
	}

	if entry, found := idx.entries[routeScope(routeNN, string(routeKind))]; found {
		return idx.resolveEntry(entry, gatewayNN, listenerName, listenerSetNN)
	}

	return idx.parentLevels(gatewayNN, listenerName, listenerSetNN), false
}

// LookupExact resolves scope's own entry, whichever level it is, with no fallback to a broader
// parent scope. effective was already decided when the entry was written, so every scope kind
// resolves the same way here.
func (idx *policyIndex[T]) LookupExact(scope policyScope) (T, bool) {
	var zero T
	if idx == nil {
		return zero, false
	}
	if entry, found := idx.entries[scope]; found && entry.effective {
		return entry.value, true
	}
	return zero, false
}

// parentLevels resolves just the listenerSet/listener/gateway levels, for Lookup/resolveEntry's
// own fallback step when no more specific route-rule/route entry applies (or one applies but
// isn't effective).
func (idx *policyIndex[T]) parentLevels(gatewayNN types.NamespacedName, listenerName *gwapiv1.SectionName, listenerSetNN *types.NamespacedName) T {
	var zero T
	if idx == nil {
		return zero
	}

	if listenerSetNN != nil {
		if listenerName != nil {
			if value, found := idx.LookupExact(listenerSetListenerScope(*listenerSetNN, *listenerName)); found {
				return value
			}
		}
		if value, found := idx.LookupExact(listenerSetScope(*listenerSetNN)); found {
			return value
		}
	} else if listenerName != nil {
		if value, found := idx.LookupExact(gatewayListenerScope(gatewayNN, *listenerName)); found {
			return value
		}
	}

	value, _ := idx.LookupExact(gatewayScope(gatewayNN))
	return value
}

// resolveEntry reports pinned: true whenever entry itself is effective, and false only when it
// falls through to a parent.
func (idx *policyIndex[T]) resolveEntry(entry policyIndexEntry[T], gatewayNN types.NamespacedName, listenerName *gwapiv1.SectionName, listenerSetNN *types.NamespacedName) (T, bool) {
	if entry.effective {
		return entry.value, true
	}
	return idx.parentLevels(gatewayNN, listenerName, listenerSetNN), false
}
