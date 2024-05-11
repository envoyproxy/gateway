// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// SetProgrammedForEnvoyPatchPolicy sets programmed conditions for each ancestor reference in policy status if it is unset.
func SetProgrammedForEnvoyPatchPolicy(s *gwv1a2.PolicyStatus) {
	// Return early if Programmed condition is already set
	for _, ancestor := range s.Ancestors {
		for _, c := range ancestor.Conditions {
			if c.Type == string(egv1a1.PolicyConditionProgrammed) {
				return
			}
			if c.Type == string(gwv1a2.PolicyConditionAccepted) && c.Status == metav1.ConditionFalse {
				return
			}
		}
	}

	message := "Patches have been successfully applied."
	cond := newCondition(string(egv1a1.PolicyConditionProgrammed), metav1.ConditionTrue, string(egv1a1.PolicyReasonProgrammed), message, time.Now(), 0)
	for i := range s.Ancestors {
		s.Ancestors[i].Conditions = MergeConditions(s.Ancestors[i].Conditions, cond)
	}
}

func SetTranslationErrorForEnvoyPatchPolicy(s *gwv1a2.PolicyStatus, errMsg string) {
	cond := newCondition(string(egv1a1.PolicyConditionProgrammed), metav1.ConditionFalse, string(egv1a1.PolicyReasonInvalid), errMsg, time.Now(), 0)
	for i := range s.Ancestors {
		s.Ancestors[i].Conditions = MergeConditions(s.Ancestors[i].Conditions, cond)
	}
}

func SetResourceNotFoundErrorForEnvoyPatchPolicy(s *gwv1a2.PolicyStatus, notFoundResources []string) {
	message := "Unable to find xds resources: " + strings.Join(notFoundResources, ",")
	cond := newCondition(string(egv1a1.PolicyConditionProgrammed), metav1.ConditionFalse, string(egv1a1.PolicyReasonResourceNotFound), message, time.Now(), 0)
	for i := range s.Ancestors {
		s.Ancestors[i].Conditions = MergeConditions(s.Ancestors[i].Conditions, cond)
	}
}
