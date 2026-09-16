// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func TestPolicyIndexLookup(t *testing.T) {
	routeKind := gwapiv1.Kind("HTTPRoute")
	routeNN := types.NamespacedName{Namespace: "default", Name: "route-1"}
	gatewayNN := types.NamespacedName{Namespace: "default", Name: "gateway-1"}
	listenerSetNN := types.NamespacedName{Namespace: "default", Name: "listener-set-1"}
	listenerName := gwapiv1.SectionName("http")
	ruleName := gwapiv1.SectionName("rule-1")

	ruleKey := routeRuleScope(routeNN, string(routeKind), ruleName)
	routeKey := routeScope(routeNN, string(routeKind))
	listenerKey := gatewayListenerScope(gatewayNN, listenerName)
	gatewayKey := gatewayScope(gatewayNN)
	listenerSetListenerKey := listenerSetListenerScope(listenerSetNN, listenerName)
	listenerSetKey := listenerSetScope(listenerSetNN)

	cases := []struct {
		name                   string
		nilIndex               bool
		ruleEntry              *policyIndexEntry[bool]
		routeEntry             *policyIndexEntry[bool]
		listenerEntry          *policyIndexEntry[bool]
		listenerSetListenerVal *policyIndexEntry[bool]
		listenerSetVal         *policyIndexEntry[bool]
		gatewayValue           bool
		useRuleName            bool
		useListener            bool
		useListenerSet         bool
		wantValue              bool
		wantReplacesParent     bool
	}{
		{
			name:        "nil index falls through to zero value, not replacing parent",
			nilIndex:    true,
			useRuleName: true,
			useListener: true,
		},
		{
			name:        "no entry anywhere falls through to zero value, not replacing parent",
			useRuleName: true,
			useListener: true,
		},
		{
			name:               "rule-level non-zero value wins outright and replaces parent, regardless of MergeType",
			ruleEntry:          &policyIndexEntry[bool]{value: true, effective: true},
			useRuleName:        true,
			useListener:        true,
			wantValue:          true,
			wantReplacesParent: true,
		},
		{
			name:               "rule-level zero value with MergeType nil replaces parent, does not inherit",
			ruleEntry:          &policyIndexEntry[bool]{value: false, effective: true},
			gatewayValue:       true,
			useRuleName:        true,
			useListener:        true,
			wantReplacesParent: true,
		},
		{
			name:         "rule-level zero value with MergeType set falls through to gateway, not replacing parent",
			ruleEntry:    &policyIndexEntry[bool]{value: false, effective: false},
			gatewayValue: true,
			useRuleName:  true,
			useListener:  true,
			wantValue:    true,
		},
		{
			name:          "rule-level presence shields route-level entirely",
			ruleEntry:     &policyIndexEntry[bool]{value: false, effective: false},
			routeEntry:    &policyIndexEntry[bool]{value: true, effective: true},
			listenerEntry: &policyIndexEntry[bool]{value: false, effective: true},
			useRuleName:   true,
			useListener:   true,
		},
		{
			name:         "no rule-level entry falls to route-level, which falls through to gateway",
			routeEntry:   &policyIndexEntry[bool]{value: false, effective: false},
			gatewayValue: true,
			useRuleName:  true,
			useListener:  true,
			wantValue:    true,
		},
		{
			name:               "routeRuleName nil skips rule-level check even if an entry exists",
			ruleEntry:          &policyIndexEntry[bool]{value: true, effective: true},
			routeEntry:         &policyIndexEntry[bool]{value: false, effective: true},
			useListener:        true,
			wantReplacesParent: true,
		},
		{
			name:          "no rule/route entry falls to listener",
			listenerEntry: &policyIndexEntry[bool]{value: true, effective: true},
			useRuleName:   true,
			useListener:   true,
			wantValue:     true,
		},
		{
			name:          "listener entry present but zero value is still final, does not fall through",
			listenerEntry: &policyIndexEntry[bool]{value: false, effective: true},
			gatewayValue:  true,
			useRuleName:   true,
			useListener:   true,
		},
		{
			name:         "no listener entry at all falls through to gateway",
			gatewayValue: true,
			useRuleName:  true,
			useListener:  true,
			wantValue:    true,
		},
		{
			name:          "listener entry with hasValue false transparently falls through to gateway",
			listenerEntry: &policyIndexEntry[bool]{value: true, effective: false},
			gatewayValue:  true,
			useRuleName:   true,
			useListener:   true,
			wantValue:     true,
		},
		{
			name:          "listenerName nil skips listener check, falls to gateway",
			listenerEntry: &policyIndexEntry[bool]{value: true, effective: true},
			gatewayValue:  true,
			useRuleName:   true,
			wantValue:     true,
		},
		{
			name:                   "listenerSetNN set: listenerSet-listener level wins over listenerSet level",
			listenerSetListenerVal: &policyIndexEntry[bool]{value: true, effective: true},
			listenerSetVal:         &policyIndexEntry[bool]{value: false, effective: true},
			useRuleName:            true,
			useListener:            true,
			useListenerSet:         true,
			wantValue:              true,
		},
		{
			name:           "listenerSetNN set: falls to listenerSet level when no listenerSet-listener entry",
			listenerSetVal: &policyIndexEntry[bool]{value: true, effective: true},
			useRuleName:    true,
			useListener:    true,
			useListenerSet: true,
			wantValue:      true,
		},
		{
			name:           "listenerSetNN set: gateway-listener entry sharing the same listener name is ignored",
			listenerEntry:  &policyIndexEntry[bool]{value: true, effective: true},
			useRuleName:    true,
			useListener:    true,
			useListenerSet: true,
		},
		{
			name:           "listenerSetNN set: falls all the way to gateway level when neither listenerSet scope has an entry",
			gatewayValue:   true,
			useRuleName:    true,
			useListener:    true,
			useListenerSet: true,
			wantValue:      true,
		},
		{
			name:           "listenerSetNN set: listenerSet-level entry with hasValue false transparently falls through to gateway",
			listenerSetVal: &policyIndexEntry[bool]{value: true, effective: false},
			gatewayValue:   true,
			useRuleName:    true,
			useListener:    true,
			useListenerSet: true,
			wantValue:      true,
		},
		{
			name:                   "listenerSetNN set: listenerSet-listener entry with hasValue false falls through to listenerSet level",
			listenerSetListenerVal: &policyIndexEntry[bool]{value: true, effective: false},
			listenerSetVal:         &policyIndexEntry[bool]{value: true, effective: true},
			useRuleName:            true,
			useListener:            true,
			useListenerSet:         true,
			wantValue:              true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var idx *policyIndex[bool]
			if !tc.nilIndex {
				idx = newPolicyIndex[bool]()
				if tc.ruleEntry != nil {
					idx.entries[ruleKey] = *tc.ruleEntry
				}
				if tc.routeEntry != nil {
					idx.entries[routeKey] = *tc.routeEntry
				}
				if tc.listenerEntry != nil {
					idx.entries[listenerKey] = *tc.listenerEntry
				}
				if tc.listenerSetListenerVal != nil {
					idx.entries[listenerSetListenerKey] = *tc.listenerSetListenerVal
				}
				if tc.listenerSetVal != nil {
					idx.entries[listenerSetKey] = *tc.listenerSetVal
				}
				idx.entries[gatewayKey] = policyIndexEntry[bool]{value: tc.gatewayValue, effective: true}
			}

			var wantRuleName *gwapiv1.SectionName
			if tc.useRuleName {
				wantRuleName = &ruleName
			}
			var wantListenerName *gwapiv1.SectionName
			if tc.useListener {
				wantListenerName = &listenerName
			}
			var wantListenerSetNN *types.NamespacedName
			if tc.useListenerSet {
				wantListenerSetNN = &listenerSetNN
			}

			value, replacesParent := idx.Lookup(routeKind, routeNN, gatewayNN, wantListenerName, wantListenerSetNN, wantRuleName)
			require.Equal(t, tc.wantReplacesParent, replacesParent)
			require.Equal(t, tc.wantValue, value)
		})
	}
}

func TestPolicyIndexLookupExact(t *testing.T) {
	routeNN := types.NamespacedName{Namespace: "default", Name: "route-1"}
	gatewayNN := types.NamespacedName{Namespace: "default", Name: "gateway-1"}
	listenerSetNN := types.NamespacedName{Namespace: "default", Name: "listener-set-1"}
	listenerName := gwapiv1.SectionName("http")
	ruleName := gwapiv1.SectionName("rule-1")

	ruleKey := routeRuleScope(routeNN, resource.KindHTTPRoute, ruleName)
	routeKey := routeScope(routeNN, resource.KindHTTPRoute)
	listenerKey := gatewayListenerScope(gatewayNN, listenerName)
	gatewayKey := gatewayScope(gatewayNN)
	listenerSetListenerKey := listenerSetListenerScope(listenerSetNN, listenerName)
	listenerSetKey := listenerSetScope(listenerSetNN)

	cases := []struct {
		name               string
		scope              policyScope
		entry              *policyIndexEntry[bool]
		extraScope         policyScope
		extraEntry         *policyIndexEntry[bool]
		wantValue          bool
		wantReplacesParent bool
	}{
		{
			name:               "route-rule scope with a non-zero value pins outright",
			scope:              ruleKey,
			entry:              &policyIndexEntry[bool]{value: true, effective: true},
			wantValue:          true,
			wantReplacesParent: true,
		},
		{
			name:               "route-rule scope with a zero value and MergeType nil still pins",
			scope:              ruleKey,
			entry:              &policyIndexEntry[bool]{value: false, effective: true},
			wantReplacesParent: true,
		},
		{
			name:  "route-rule scope with a zero value and MergeType set does not pin",
			scope: ruleKey,
			entry: &policyIndexEntry[bool]{value: false, effective: false},
		},
		{
			name:  "route-rule scope with no entry does not pin",
			scope: ruleKey,
		},
		{
			name:               "route scope with a non-zero value pins outright",
			scope:              routeKey,
			entry:              &policyIndexEntry[bool]{value: true, effective: true},
			wantValue:          true,
			wantReplacesParent: true,
		},
		{
			name:               "listener scope with hasValue true pins",
			scope:              listenerKey,
			entry:              &policyIndexEntry[bool]{value: true, effective: true},
			wantValue:          true,
			wantReplacesParent: true,
		},
		{
			name:  "listener scope with hasValue false does not pin",
			scope: listenerKey,
			entry: &policyIndexEntry[bool]{value: true, effective: false},
		},
		{
			name:               "gateway scope pins whenever an entry exists",
			scope:              gatewayKey,
			entry:              &policyIndexEntry[bool]{value: true, effective: true},
			wantValue:          true,
			wantReplacesParent: true,
		},
		{
			name:               "gateway scope ignores a coexisting listener-scope entry",
			scope:              gatewayKey,
			entry:              &policyIndexEntry[bool]{value: true, effective: true},
			extraScope:         listenerKey,
			extraEntry:         &policyIndexEntry[bool]{value: false, effective: true},
			wantValue:          true,
			wantReplacesParent: true,
		},
		{
			name:               "listenerSet listener scope with hasValue true pins",
			scope:              listenerSetListenerKey,
			entry:              &policyIndexEntry[bool]{value: true, effective: true},
			wantValue:          true,
			wantReplacesParent: true,
		},
		{
			name:  "listenerSet listener scope with hasValue false does not pin",
			scope: listenerSetListenerKey,
			entry: &policyIndexEntry[bool]{value: true, effective: false},
		},
		{
			name:               "listenerSet scope pins whenever an entry exists",
			scope:              listenerSetKey,
			entry:              &policyIndexEntry[bool]{value: true, effective: true},
			wantValue:          true,
			wantReplacesParent: true,
		},
		{
			name:               "listenerSet scope ignores a coexisting listenerSet-listener-scope entry",
			scope:              listenerSetKey,
			entry:              &policyIndexEntry[bool]{value: true, effective: true},
			extraScope:         listenerSetListenerKey,
			extraEntry:         &policyIndexEntry[bool]{value: false, effective: true},
			wantValue:          true,
			wantReplacesParent: true,
		},
		{
			name:               "gateway and listenerSet scopes sharing a NamespacedName don't collide",
			scope:              gatewayKey,
			entry:              &policyIndexEntry[bool]{value: true, effective: true},
			extraScope:         listenerSetScope(gatewayNN),
			extraEntry:         &policyIndexEntry[bool]{value: false, effective: true},
			wantValue:          true,
			wantReplacesParent: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			idx := newPolicyIndex[bool]()
			if tc.entry != nil {
				idx.entries[tc.scope] = *tc.entry
			}
			if tc.extraEntry != nil {
				idx.entries[tc.extraScope] = *tc.extraEntry
			}

			value, replacesParent := idx.LookupExact(tc.scope)
			require.Equal(t, tc.wantReplacesParent, replacesParent)
			require.Equal(t, tc.wantValue, value)
		})
	}
}

func TestPolicyIndexLookupPointerType(t *testing.T) {
	routeKind := gwapiv1.Kind("HTTPRoute")
	routeNN := types.NamespacedName{Namespace: "default", Name: "route-1"}
	gatewayNN := types.NamespacedName{Namespace: "default", Name: "gateway-1"}
	listenerName := gwapiv1.SectionName("http")
	ruleName := gwapiv1.SectionName("rule-1")
	ruleKey := routeRuleScope(routeNN, string(routeKind), ruleName)
	gatewayKey := gatewayScope(gatewayNN)

	endpoint := egv1a1.RoutingType("Endpoint")
	idx := newPolicyIndex[*egv1a1.RoutingType]()
	idx.entries[ruleKey] = policyIndexEntry[*egv1a1.RoutingType]{value: nil, effective: true}
	idx.entries[gatewayKey] = policyIndexEntry[*egv1a1.RoutingType]{value: &endpoint, effective: true}

	value, _ := idx.Lookup(routeKind, routeNN, gatewayNN, &listenerName, nil, &ruleName)
	require.Nil(t, value)
}

func TestPolicyIndexSetters(t *testing.T) {
	routeNN := types.NamespacedName{Namespace: "default", Name: "route-1"}
	gatewayNN := types.NamespacedName{Namespace: "default", Name: "gateway-1"}
	listenerSetNN := types.NamespacedName{Namespace: "default", Name: "listener-set-1"}
	ruleName := gwapiv1.SectionName("rule-1")
	listenerName := gwapiv1.SectionName("http")
	strategicMerge := egv1a1.StrategicMerge

	t.Run("setRouteRuleLevel keeps the first entry recorded for a key", func(t *testing.T) {
		idx := newPolicyIndex[bool]()
		idx.setRouteRuleLevel(routeNN, resource.KindHTTPRoute, ruleName, true, &strategicMerge)
		idx.setRouteRuleLevel(routeNN, resource.KindHTTPRoute, ruleName, false, nil)
		require.Equal(t, policyIndexEntry[bool]{value: true, effective: true}, idx.entries[routeRuleScope(routeNN, resource.KindHTTPRoute, ruleName)])
	})

	t.Run("setRouteLevel keeps the first entry recorded for a key", func(t *testing.T) {
		idx := newPolicyIndex[bool]()
		idx.setRouteLevel(routeNN, resource.KindHTTPRoute, true, nil)
		idx.setRouteLevel(routeNN, resource.KindHTTPRoute, false, &strategicMerge)
		require.Equal(t, policyIndexEntry[bool]{value: true, effective: true}, idx.entries[routeScope(routeNN, resource.KindHTTPRoute)])
	})

	t.Run("setGatewayListenerLevel keeps the first entry recorded for a key", func(t *testing.T) {
		idx := newPolicyIndex[bool]()
		idx.setGatewayListenerLevel(gatewayNN, listenerName, true, true)
		idx.setGatewayListenerLevel(gatewayNN, listenerName, false, true)
		require.Equal(t, policyIndexEntry[bool]{value: true, effective: true}, idx.entries[gatewayListenerScope(gatewayNN, listenerName)])
	})

	t.Run("setGatewayListenerLevel with hasValue false still claims the slot", func(t *testing.T) {
		idx := newPolicyIndex[bool]()
		idx.setGatewayListenerLevel(gatewayNN, listenerName, true, false)
		require.Equal(t, policyIndexEntry[bool]{value: true, effective: false}, idx.entries[gatewayListenerScope(gatewayNN, listenerName)])
	})

	t.Run("setGatewayListenerLevel with hasValue false blocks a later, hasValue true call from a different policy", func(t *testing.T) {
		idx := newPolicyIndex[bool]()
		idx.setGatewayListenerLevel(gatewayNN, listenerName, true, false)
		idx.setGatewayListenerLevel(gatewayNN, listenerName, true, true)
		require.Equal(t, policyIndexEntry[bool]{value: true, effective: false}, idx.entries[gatewayListenerScope(gatewayNN, listenerName)])
	})

	t.Run("setGatewayLevel keeps the first value recorded for a key", func(t *testing.T) {
		idx := newPolicyIndex[bool]()
		idx.setGatewayLevel(gatewayNN, true)
		idx.setGatewayLevel(gatewayNN, false)
		require.Equal(t, policyIndexEntry[bool]{value: true, effective: true}, idx.entries[gatewayScope(gatewayNN)])
	})

	t.Run("setListenerSetListenerLevel keeps the first entry recorded for a key", func(t *testing.T) {
		idx := newPolicyIndex[bool]()
		idx.setListenerSetListenerLevel(listenerSetNN, listenerName, true, true)
		idx.setListenerSetListenerLevel(listenerSetNN, listenerName, false, true)
		require.Equal(t, policyIndexEntry[bool]{value: true, effective: true}, idx.entries[listenerSetListenerScope(listenerSetNN, listenerName)])
	})

	t.Run("setListenerSetLevel keeps the first value recorded for a key", func(t *testing.T) {
		idx := newPolicyIndex[bool]()
		idx.setListenerSetLevel(listenerSetNN, true, true)
		idx.setListenerSetLevel(listenerSetNN, false, true)
		require.Equal(t, policyIndexEntry[bool]{value: true, effective: true}, idx.entries[listenerSetScope(listenerSetNN)])
	})

	t.Run("setListenerSetLevel with hasValue false still claims the slot", func(t *testing.T) {
		idx := newPolicyIndex[bool]()
		idx.setListenerSetLevel(listenerSetNN, true, false)
		require.Equal(t, policyIndexEntry[bool]{value: true, effective: false}, idx.entries[listenerSetScope(listenerSetNN)])
	})

	t.Run("setListenerSetLevel with hasValue false blocks a later, hasValue true call from a different policy", func(t *testing.T) {
		idx := newPolicyIndex[bool]()
		idx.setListenerSetLevel(listenerSetNN, true, false)
		idx.setListenerSetLevel(listenerSetNN, true, true)
		require.Equal(t, policyIndexEntry[bool]{value: true, effective: false}, idx.entries[listenerSetScope(listenerSetNN)])
	})
}
