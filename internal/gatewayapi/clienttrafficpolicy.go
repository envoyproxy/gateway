// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"sort"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	gwv1b1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/status"
	"github.com/envoyproxy/gateway/internal/utils/ptr"
)

const (
	// Use an invalid string to represent all sections (listeners) within a Gateway
	AllSections = "/"
)

func hasSectionName(policy *egv1a1.ClientTrafficPolicy) bool {
	return policy.Spec.TargetRef.SectionName != nil
}

func ProcessClientTrafficPolicies(clientTrafficPolicies []*egv1a1.ClientTrafficPolicy,
	gateways []*GatewayContext,
	xdsIR XdsIRMap) []*egv1a1.ClientTrafficPolicy {
	var res []*egv1a1.ClientTrafficPolicy

	// Sort based on timestamp
	sort.Slice(clientTrafficPolicies, func(i, j int) bool {
		return clientTrafficPolicies[i].CreationTimestamp.Before(&(clientTrafficPolicies[j].CreationTimestamp))
	})

	policyMap := make(map[types.NamespacedName]sets.Set[string])

	// Translate
	// 1. First translate Policies with a sectionName set
	// 2. Then loop again and translate the policies without a sectionName
	// TODO: Import sort order to ensure policy with same section always appear
	// before policy with no section so below loops can be flattened into 1.

	for _, policy := range clientTrafficPolicies {
		if hasSectionName(policy) {
			policy := policy.DeepCopy()
			res = append(res, policy)

			gateway := resolveCTPolicyTargetRef(policy, gateways)

			// Negative statuses have already been assigned so its safe to skip
			if gateway == nil {
				continue
			}

			// Check for conflicts
			key := types.NamespacedName{
				Name:      gateway.Name,
				Namespace: gateway.Namespace,
			}

			// Check if another policy targeting the same section exists
			section := string(*(policy.Spec.TargetRef.SectionName))
			s, ok := policyMap[key]
			if ok && s.Has(section) {
				message := "Unable to target section, another ClientTrafficPolicy has already attached to it"

				status.SetClientTrafficPolicyCondition(policy,
					gwv1a2.PolicyConditionAccepted,
					metav1.ConditionFalse,
					gwv1a2.PolicyReasonConflicted,
					message,
				)
				continue
			}

			// Add section to policy map
			if s == nil {
				policyMap[key] = sets.New[string]()
			}
			policyMap[key].Insert(section)

			// Translate for listener matching section name
			for _, l := range gateway.listeners {
				if string(l.Name) == section {
					translateClientTrafficPolicyForListener(&policy.Spec, l, xdsIR)
					break
				}
			}
			// Set Accepted=True
			status.SetClientTrafficPolicyCondition(policy,
				gwv1a2.PolicyConditionAccepted,
				metav1.ConditionTrue,
				gwv1a2.PolicyReasonAccepted,
				"ClientTrafficPolicy has been accepted.",
			)
		}
	}

	// Policy with no section set (targeting all sections)
	for _, policy := range clientTrafficPolicies {
		if !hasSectionName(policy) {

			policy := policy.DeepCopy()
			res = append(res, policy)

			gateway := resolveCTPolicyTargetRef(policy, gateways)

			// Negative statuses have already been assigned so its safe to skip
			if gateway == nil {
				continue
			}

			// Check for conflicts
			key := types.NamespacedName{
				Name:      gateway.Name,
				Namespace: gateway.Namespace,
			}
			s, ok := policyMap[key]
			// Check if another policy targeting the same Gateway exists
			if ok && s.Has(AllSections) {
				message := "Unable to target Gateway, another ClientTrafficPolicy has already attached to it"

				status.SetClientTrafficPolicyCondition(policy,
					gwv1a2.PolicyConditionAccepted,
					metav1.ConditionFalse,
					gwv1a2.PolicyReasonConflicted,
					message,
				)

				continue

			}

			// Check if another policy targeting the same Gateway exists
			if ok && (s.Len() > 0) {
				// Maintain order here to ensure status/string does not change with same data
				sections := s.UnsortedList()
				sort.Strings(sections)
				message := fmt.Sprintf("There are existing ClientTrafficPolicies that are overriding these sections %v", sections)

				status.SetClientTrafficPolicyCondition(policy,
					egv1a1.PolicyConditionOverridden,
					metav1.ConditionTrue,
					egv1a1.PolicyReasonOverridden,
					message,
				)
			}

			// Add section to policy map
			if s == nil {
				policyMap[key] = sets.New[string]()
			}
			policyMap[key].Insert(AllSections)

			// Translate sections that have not yet been targeted
			for _, l := range gateway.listeners {
				// Skip if section has already been targeted
				if s != nil && s.Has(string(l.Name)) {
					continue
				}

				translateClientTrafficPolicyForListener(&policy.Spec, l, xdsIR)
			}

			// Set Accepted=True
			status.SetClientTrafficPolicyCondition(policy,
				gwv1a2.PolicyConditionAccepted,
				metav1.ConditionTrue,
				gwv1a2.PolicyReasonAccepted,
				"ClientTrafficPolicy has been accepted.",
			)
		}
	}

	return res
}

func resolveCTPolicyTargetRef(policy *egv1a1.ClientTrafficPolicy, gateways []*GatewayContext) *GatewayContext {
	targetNs := policy.Spec.TargetRef.Namespace
	// If empty, default to namespace of policy
	if targetNs == nil {
		targetNs = ptr.To(gwv1b1.Namespace(policy.Namespace))
	}

	// Ensure policy can only target a Gateway
	if policy.Spec.TargetRef.Group != gwv1b1.GroupName || policy.Spec.TargetRef.Kind != KindGateway {
		message := fmt.Sprintf("TargetRef.Group:%s TargetRef.Kind:%s, only TargetRef.Group:%s and TargetRef.Kind:%s is supported.",
			policy.Spec.TargetRef.Group, policy.Spec.TargetRef.Kind, gwv1b1.GroupName, KindGateway)

		status.SetClientTrafficPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwv1a2.PolicyReasonInvalid,
			message,
		)
		return nil
	}

	// Ensure Policy and target Gateway are in the same namespace
	if policy.Namespace != string(*targetNs) {

		message := fmt.Sprintf("Namespace:%s TargetRef.Namespace:%s, ClientTrafficPolicy can only target a Gateway in the same namespace.",
			policy.Namespace, *targetNs)
		status.SetClientTrafficPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwv1a2.PolicyReasonInvalid,
			message,
		)
		return nil
	}

	// Find the Gateway
	var gateway *GatewayContext
	for _, g := range gateways {
		if g.Name == string(policy.Spec.TargetRef.Name) && g.Namespace == string(*targetNs) {
			gateway = g
			break
		}
	}

	// Gateway not found
	if gateway == nil {
		message := fmt.Sprintf("Gateway:%s not found.", policy.Spec.TargetRef.Name)

		status.SetClientTrafficPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwv1a2.PolicyReasonTargetNotFound,
			message,
		)
		return nil
	}

	// If sectionName is set, make sure its valid
	if policy.Spec.TargetRef.SectionName != nil {
		found := false
		for _, l := range gateway.listeners {
			if l.Name == *(policy.Spec.TargetRef.SectionName) {
				found = true
				break
			}
		}
		if !found {
			message := fmt.Sprintf("SectionName(Listener):%s not found.", *(policy.Spec.TargetRef.SectionName))
			status.SetClientTrafficPolicyCondition(policy,
				gwv1a2.PolicyConditionAccepted,
				metav1.ConditionFalse,
				gwv1a2.PolicyReasonTargetNotFound,
				message,
			)
			return nil
		}
	}

	return gateway
}

func translateClientTrafficPolicyForListener(policySpec *egv1a1.ClientTrafficPolicySpec, l *ListenerContext, xdsIR XdsIRMap) {
	// Find IR
	irKey := irStringKey(l.gateway.Namespace, l.gateway.Name)
	// It must exist since we've already finished processing the gateways
	gwXdsIR := xdsIR[irKey]

	// Find Listener IR
	// TODO: Support TLSRoute and TCPRoute once
	// https://github.com/envoyproxy/gateway/issues/1635 is completed

	irListenerName := irHTTPListenerName(l)
	var httpIR *ir.HTTPListener
	for _, http := range gwXdsIR.HTTP {
		if http.Name == irListenerName {
			httpIR = http
			break
		}
	}

	// IR must exist since we're past validation
	if httpIR != nil {
		// Translate TCPKeepalive
		translateListenerTCPKeepalive(policySpec.TCPKeepalive, httpIR)
	}
}

func translateListenerTCPKeepalive(tcpKeepAlive *egv1a1.TCPKeepalive, httpIR *ir.HTTPListener) {
	// Return early if not set
	if tcpKeepAlive == nil {
		return
	}

	irTCPKeepalive := &ir.TCPKeepalive{}

	if tcpKeepAlive.Probes != nil {
		irTCPKeepalive.Probes = tcpKeepAlive.Probes
	}

	if tcpKeepAlive.IdleTime != nil {
		d, err := time.ParseDuration(string(*tcpKeepAlive.IdleTime))
		if err != nil {
			return
		}
		irTCPKeepalive.IdleTime = ptr.To(uint32(d.Seconds()))
	}

	if tcpKeepAlive.Interval != nil {
		d, err := time.ParseDuration(string(*tcpKeepAlive.Interval))
		if err != nil {
			return
		}
		irTCPKeepalive.Interval = ptr.To(uint32(d.Seconds()))
	}

	httpIR.TCPKeepalive = irTCPKeepalive
}
