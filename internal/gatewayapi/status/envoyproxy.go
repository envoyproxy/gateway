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

// EnvoyProxyReasonUnsupportedField is used with the "Warning" condition when a field set on an
// EnvoyProxy has no effect where the EnvoyProxy is referenced, e.g. spec.mergeGateways on an
// EnvoyProxy referenced by a Gateway.
const EnvoyProxyReasonUnsupportedField egv1a1.EnvoyProxyConditionReason = "UnsupportedField"

// SetEnvoyProxyWarnings sets the Warning condition on the EnvoyProxy ancestor status from the
// warnings found in the current translation: the deprecated fields in use, the Kubernetes
// resource `patch` fields that were ignored because patching is disabled in the EnvoyGateway
// configuration, and the fields that have no effect for this ancestor. An EnvoyProxy has a single
// Warning condition per ancestor, so these are merged into one message, and the reason becomes
// EnvoyProxyReasonMultipleWarnings when more than one is present.
// Any Warning condition left over from an earlier translation is replaced, or removed when there
// is nothing to warn about, so a warning goes away once the field is no longer used.
func SetEnvoyProxyWarnings(ep *egv1a1.EnvoyProxy, ancestor *gwapiv1.ParentReference,
	deprecatedFields map[string]DeprecatedField, patchedFields, unsupportedFields []string,
) {
	if ep == nil || ancestor == nil {
		return
	}

	type warning struct {
		reason  egv1a1.EnvoyProxyConditionReason
		message string
	}
	var warnings []warning
	if len(deprecatedFields) > 0 {
		warnings = append(warnings, warning{egv1a1.EnvoyProxyReasonDeprecatedField, buildDeprecationWarningMessage(deprecatedFields)})
	}
	if len(patchedFields) > 0 {
		warnings = append(warnings, warning{egv1a1.EnvoyProxyReasonPatchDisabled, fmt.Sprintf(
			"The following patch fields were ignored because patching is disabled in the EnvoyGateway configuration: %s",
			strings.Join(patchedFields, ", "))})
	}
	if len(unsupportedFields) > 0 {
		warnings = append(warnings, warning{EnvoyProxyReasonUnsupportedField, fmt.Sprintf(
			"The following fields have no effect on an EnvoyProxy referenced by a Gateway and were ignored, "+
				"set them on the EnvoyProxy referenced by the GatewayClass instead: %s",
			strings.Join(unsupportedFields, ", "))})
	}

	var (
		reason egv1a1.EnvoyProxyConditionReason
		msg    string
	)
	switch len(warnings) {
	case 0:
	case 1:
		reason, msg = warnings[0].reason, warnings[0].message
	default:
		reason = egv1a1.EnvoyProxyReasonMultipleWarnings
		// Merge the later warnings first, so that if the message is too long only the first
		// warning is cut short and every later one is kept in full.
		rest := warnings[1].message
		for _, w := range warnings[2:] {
			rest = mergePolicyWarningMessages(rest, w.message)
		}
		msg = mergeBoundedWarningMessages(warnings[0].message, rest)
	}

	for i := range ep.Status.Ancestors {
		item := &ep.Status.Ancestors[i]
		if !ancestorRefsEqual(&item.AncestorRef, ancestor) {
			continue
		}
		if msg == "" {
			meta.RemoveStatusCondition(&item.Conditions, string(egv1a1.EnvoyProxyConditionWarning))
			return
		}
		cond := newCondition(string(egv1a1.EnvoyProxyConditionWarning), metav1.ConditionTrue, string(reason), msg, ep.Generation)
		item.Conditions = MergeConditions(item.Conditions, cond)
		return
	}

	if msg == "" {
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
