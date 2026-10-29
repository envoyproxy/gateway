// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"fmt"
	"slices"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/api/v1alpha1/validation"
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
func DeprecatedFieldsWarning(deprecatedFields map[string]string) *EnvoyProxyWarning {
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

// SetEnvoyProxyWarnings sets the Warning condition on the EnvoyProxy ancestor status to the
// given warnings, ignoring nil ones. A single condition holds them all, since conditions are
// keyed by type: its reason is the warning's own reason when there is one, and
// EnvoyProxyReasonMultipleWarnings otherwise. When there is no warning, a Warning condition left
// from a previous translation is removed.
func SetEnvoyProxyWarnings(ep *egv1a1.EnvoyProxy, ancestor *gwapiv1.ParentReference, warnings ...*EnvoyProxyWarning) {
	if ep == nil || ancestor == nil {
		return
	}

	var reason egv1a1.EnvoyProxyConditionReason
	var msgs []string
	for _, w := range warnings {
		if w == nil {
			continue
		}
		if reason == "" {
			reason = w.Reason
		} else {
			reason = egv1a1.EnvoyProxyReasonMultipleWarnings
		}
		msgs = append(msgs, w.Message)
	}

	if len(msgs) == 0 {
		for i := range ep.Status.Ancestors {
			if ancestorRefsEqual(&ep.Status.Ancestors[i].AncestorRef, ancestor) {
				ep.Status.Ancestors[i].Conditions = slices.DeleteFunc(ep.Status.Ancestors[i].Conditions, func(c metav1.Condition) bool {
					return c.Type == string(egv1a1.EnvoyProxyConditionWarning)
				})
				return
			}
		}
		return
	}

	cond := newCondition(string(egv1a1.EnvoyProxyConditionWarning), metav1.ConditionTrue,
		string(reason), strings.Join(msgs, "; "), ep.Generation)

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
