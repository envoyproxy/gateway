// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func UpdateEnvoyProxyStatusAccepted(ep *egv1a1.EnvoyProxy, ancestor *gwapiv1.ParentReference,
	reason egv1a1.EnvoyProxyConditionReason, msg string,
) {
	if ep == nil || ancestor == nil {
		return
	}
	status := metav1.ConditionTrue
	if reason != egv1a1.EnvoyProxyReasonAccepted {
		status = metav1.ConditionFalse
	}

	cond := newCondition(string(egv1a1.EnvoyProxyConditionAccepted), status,
		string(reason), msg, ep.Generation)

	for i := range ep.Status.Ancestors {
		item := ep.Status.Ancestors[i]
		if ancestorRefsEqual(&item.AncestorRef, ancestor) {
			ep.Status.Ancestors[i].Conditions = MergeConditions(item.Conditions, cond)
			return
		}
	}

	// ancestor not found, append a new one
	ep.Status.Ancestors = append(ep.Status.Ancestors, egv1a1.EnvoyProxyAncestorStatus{
		AncestorRef: *ancestor,
		Conditions: []metav1.Condition{
			cond,
		},
	})
}

func SetEnvoyProxyDeprecatedFieldsWarning(ep *egv1a1.EnvoyProxy, ancestor *gwapiv1.ParentReference, deprecatedFields map[string]DeprecatedField) {
	if len(deprecatedFields) == 0 {
		return
	}
	setEnvoyProxyWarning(ep, ancestor, egv1a1.EnvoyProxyReasonDeprecatedField, buildDeprecationWarningMessage(deprecatedFields))
}

// SetEnvoyProxyPatchDisabledWarning sets a Warning condition on the EnvoyProxy
// ancestor status listing the Kubernetes resource `patch` fields that were
// ignored because patching is disabled in the EnvoyGateway configuration.
// It is a no-op when patchedFields is empty.
func SetEnvoyProxyPatchDisabledWarning(ep *egv1a1.EnvoyProxy, ancestor *gwapiv1.ParentReference, patchedFields []string) {
	if len(patchedFields) == 0 {
		return
	}

	msg := fmt.Sprintf(
		"The following patch fields were ignored because patching is disabled in the EnvoyGateway configuration: %s",
		strings.Join(patchedFields, ", "))
	setEnvoyProxyWarning(ep, ancestor, egv1a1.EnvoyProxyReasonPatchDisabled, msg)
}

// setEnvoyProxyWarning sets the Warning condition on the EnvoyProxy ancestor status. An EnvoyProxy
// has a single Warning condition per ancestor, so a warning set earlier is merged with this one
// rather than replaced, and the reason becomes EnvoyProxyReasonMultipleWarnings when they differ.
func setEnvoyProxyWarning(ep *egv1a1.EnvoyProxy, ancestor *gwapiv1.ParentReference, reason egv1a1.EnvoyProxyConditionReason, msg string) {
	if ep == nil || ancestor == nil || msg == "" {
		return
	}

	for i := range ep.Status.Ancestors {
		item := &ep.Status.Ancestors[i]
		if !ancestorRefsEqual(&item.AncestorRef, ancestor) {
			continue
		}
		if existing := meta.FindStatusCondition(item.Conditions, string(egv1a1.EnvoyProxyConditionWarning)); existing != nil {
			msg = mergeBoundedWarningMessages(existing.Message, msg)
			if existing.Reason != "" && existing.Reason != string(reason) {
				reason = egv1a1.EnvoyProxyReasonMultipleWarnings
			}
		}
		cond := newCondition(string(egv1a1.EnvoyProxyConditionWarning), metav1.ConditionTrue, string(reason), msg, ep.Generation)
		item.Conditions = MergeConditions(item.Conditions, cond)
		return
	}

	// ancestor not found, append a new one
	ep.Status.Ancestors = append(ep.Status.Ancestors, egv1a1.EnvoyProxyAncestorStatus{
		AncestorRef: *ancestor,
		Conditions: []metav1.Condition{
			newCondition(string(egv1a1.EnvoyProxyConditionWarning), metav1.ConditionTrue, string(reason), msg, ep.Generation),
		},
	})
}

// mergeBoundedWarningMessages merges next into the existing warning message like
// mergePolicyWarningMessages. If the result would be longer than a condition message can be, the
// existing message is cut short to make room for next, so the newer warning is not truncated off
// the end. If next alone is too long, the merged message is returned as is and newCondition
// truncates it.
func mergeBoundedWarningMessages(existing, next string) string {
	const sep = "; "
	merged := mergePolicyWarningMessages(existing, next)
	if len(merged) <= conditionMessageMaxLength || merged == existing || merged == next {
		return merged
	}
	room := conditionMessageMaxLength - len(sep) - len(next) - len(conditionMessageTruncationSuffix)
	if room <= 0 {
		return merged
	}
	return existing[:room] + conditionMessageTruncationSuffix + sep + next
}
