// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/status"
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

			gateway := getGatewayTargetRef(policy, gateways)

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

			translateClientTrafficPolicy(policy, xdsIR)

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

			gateway := getGatewayTargetRef(policy, gateways)

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
				// Mantain order here to ensure status/string doesnt change with same data
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

			translateClientTrafficPolicy(policy, xdsIR)

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

func getGatewayTargetRef(policy *egv1a1.ClientTrafficPolicy, gateways []*GatewayContext) *GatewayContext {
	targetNs := policy.Spec.TargetRef.Namespace

	// Ensure Namespace is set
	if targetNs == nil {
		status.SetClientTrafficPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionFalse,
			gwv1a2.PolicyReasonInvalid,
			"TargetRef.Namespace must be set",
		)
		return nil
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
		for _, l := range gateway.Spec.Listeners {
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

func translateClientTrafficPolicy(policy *egv1a1.ClientTrafficPolicy, xdsIR XdsIRMap) {
	// TODO
}
