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
	"github.com/envoyproxy/gateway/internal/validation"
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

// EnvoyProxyWarning is one of the reasons an EnvoyProxy is warned about in its Warning condition.
type EnvoyProxyWarning struct {
	Reason  egv1a1.EnvoyProxyConditionReason
	Message string
}

// DeprecatedFieldsWarning returns the warning for the deprecated fields used by an EnvoyProxy,
// or nil when there is none.
func DeprecatedFieldsWarning(deprecatedFields map[string]DeprecatedField) *EnvoyProxyWarning {
	if len(deprecatedFields) == 0 {
		return nil
	}
	return &EnvoyProxyWarning{
		Reason:  egv1a1.EnvoyProxyReasonDeprecatedField,
		Message: buildDeprecationWarningMessage(deprecatedFields),
	}
}

// PatchDisabledWarning returns the warning for the Kubernetes resource `patch` fields that were
// ignored because patching is disabled in the EnvoyGateway configuration, or nil when there is none.
func PatchDisabledWarning(patchedFields []string) *EnvoyProxyWarning {
	if len(patchedFields) == 0 {
		return nil
	}
	return &EnvoyProxyWarning{
		Reason: egv1a1.EnvoyProxyReasonPatchDisabled,
		Message: fmt.Sprintf(
			"The following patch fields were ignored because patching is disabled in the EnvoyGateway configuration: %s",
			strings.Join(patchedFields, ", ")),
	}
}

// InvalidFieldsWarning returns the warning for the invalid fields of an EnvoyProxy that were
// ignored, or nil when there is none.
func InvalidFieldsWarning(invalidFields []validation.InvalidField) *EnvoyProxyWarning {
	if len(invalidFields) == 0 {
		return nil
	}
	msgs := make([]string, 0, len(invalidFields))
	for _, f := range invalidFields {
		msgs = append(msgs, f.Error())
	}
	return &EnvoyProxyWarning{
		Reason:  egv1a1.EnvoyProxyReasonInvalidField,
		Message: fmt.Sprintf("The following invalid fields were ignored: %s", strings.Join(msgs, "; ")),
	}
}

// EnvoyProxyReasonUnsupportedField is used with the "Warning" condition when a field set on an
// EnvoyProxy has no effect where the EnvoyProxy is referenced, e.g. spec.mergeGateways on an
// EnvoyProxy referenced by a Gateway.
const EnvoyProxyReasonUnsupportedField egv1a1.EnvoyProxyConditionReason = "UnsupportedField"

// UnsupportedFieldsWarning returns the warning for the fields of an EnvoyProxy referenced by a
// Gateway that only take effect on an EnvoyProxy referenced by a GatewayClass, or nil when there
// is none.
func UnsupportedFieldsWarning(unsupportedFields []string) *EnvoyProxyWarning {
	if len(unsupportedFields) == 0 {
		return nil
	}
	return &EnvoyProxyWarning{
		Reason: EnvoyProxyReasonUnsupportedField,
		Message: fmt.Sprintf(
			"The following fields have no effect on an EnvoyProxy referenced by a Gateway and were ignored, "+
				"set them on the EnvoyProxy referenced by the GatewayClass instead: %s",
			strings.Join(unsupportedFields, ", ")),
	}
}

// SetEnvoyProxyWarnings sets the Warning condition on the EnvoyProxy ancestor status to the
// given warnings, ignoring nil ones. A single condition holds them all, since conditions are
// keyed by type: its reason is the warning's own reason when there is one, and
// EnvoyProxyReasonMultipleWarnings otherwise. Any Warning condition left over from an earlier
// translation is replaced, or removed when there is nothing to warn about, so a warning goes
// away once the field is no longer used.
func SetEnvoyProxyWarnings(ep *egv1a1.EnvoyProxy, ancestor *gwapiv1.ParentReference, warnings ...*EnvoyProxyWarning) {
	if ep == nil || ancestor == nil {
		return
	}

	var (
		reason egv1a1.EnvoyProxyConditionReason
		msg    string
	)
	for _, w := range warnings {
		if w == nil {
			continue
		}
		if msg == "" {
			reason = w.Reason
			msg = w.Message
		} else {
			reason = egv1a1.EnvoyProxyReasonMultipleWarnings
			msg = mergeBoundedWarningMessages(msg, w.Message)
		}
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
