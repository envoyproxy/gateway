// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/validation"
)

func TestSetEnvoyProxyWarnings(t *testing.T) {
	ancestor := &gwapiv1.ParentReference{Name: gwapiv1.ObjectName("eg")}

	t.Run("sets a warning condition when deprecated fields are used", func(t *testing.T) {
		ep := &egv1a1.EnvoyProxy{}
		SetEnvoyProxyWarnings(ep, ancestor, DeprecatedFieldsWarning(map[string]DeprecatedField{
			"spec.luaValidation": {Replacement: "spec.lua.validationType"},
		}), nil)

		assert.Len(t, ep.Status.Ancestors, 1)
		conds := ep.Status.Ancestors[0].Conditions
		assert.Len(t, conds, 1)
		assert.Equal(t, string(egv1a1.EnvoyProxyConditionWarning), conds[0].Type)
		assert.Equal(t, metav1.ConditionTrue, conds[0].Status)
		assert.Equal(t, string(egv1a1.EnvoyProxyReasonDeprecatedField), conds[0].Reason)
		assert.Equal(t, "spec.luaValidation is deprecated, use spec.lua.validationType instead", conds[0].Message)
	})

	t.Run("no-op when no deprecated fields are used", func(t *testing.T) {
		ep := &egv1a1.EnvoyProxy{}
		SetEnvoyProxyWarnings(ep, ancestor)
		assert.Empty(t, ep.Status.Ancestors)
	})

	t.Run("appends the warning alongside an existing Accepted condition on the same ancestor", func(t *testing.T) {
		ep := &egv1a1.EnvoyProxy{}
		UpdateEnvoyProxyStatusAccepted(ep, ancestor, egv1a1.EnvoyProxyReasonAccepted, "EnvoyProxy has been accepted.")
		SetEnvoyProxyWarnings(ep, ancestor, DeprecatedFieldsWarning(map[string]DeprecatedField{
			"spec.luaValidation": {Replacement: "spec.lua.validationType"},
		}), nil)

		assert.Len(t, ep.Status.Ancestors, 1)
		conds := ep.Status.Ancestors[0].Conditions
		assert.Len(t, conds, 2)
	})

	t.Run("merges with a PatchDisabled warning on the same ancestor", func(t *testing.T) {
		ep := &egv1a1.EnvoyProxy{}
		UpdateEnvoyProxyStatusAccepted(ep, ancestor, egv1a1.EnvoyProxyReasonAccepted, "EnvoyProxy has been accepted.")
		SetEnvoyProxyWarnings(ep, ancestor, DeprecatedFieldsWarning(map[string]DeprecatedField{
			"spec.luaValidation": {Replacement: "spec.lua.validationType"},
		}), PatchDisabledWarning([]string{"spec.provider.kubernetes.envoyDeployment.patch"}))

		require.Len(t, ep.Status.Ancestors, 1)
		cond := meta.FindStatusCondition(ep.Status.Ancestors[0].Conditions, string(egv1a1.EnvoyProxyConditionWarning))
		require.NotNil(t, cond)
		assert.Equal(t, string(egv1a1.EnvoyProxyReasonMultipleWarnings), cond.Reason)
		assert.Equal(t, "spec.luaValidation is deprecated, use spec.lua.validationType instead; "+
			"The following patch fields were ignored because patching is disabled in the EnvoyGateway configuration: "+
			"spec.provider.kubernetes.envoyDeployment.patch", cond.Message)
	})

	t.Run("replaces a warning left over from an earlier translation", func(t *testing.T) {
		ep := &egv1a1.EnvoyProxy{}
		SetEnvoyProxyWarnings(ep, ancestor, DeprecatedFieldsWarning(map[string]DeprecatedField{
			"spec.luaValidation": {Replacement: "spec.lua.validationType"},
			"spec.telemetry.foo": {Replacement: "spec.telemetry.bar"},
		}), PatchDisabledWarning([]string{"spec.provider.kubernetes.envoyDeployment.patch"}))

		// The user migrated off spec.luaValidation and the patch field.
		ep.Generation++
		SetEnvoyProxyWarnings(ep, ancestor, DeprecatedFieldsWarning(map[string]DeprecatedField{
			"spec.telemetry.foo": {Replacement: "spec.telemetry.bar"},
		}), nil)

		require.Len(t, ep.Status.Ancestors, 1)
		cond := meta.FindStatusCondition(ep.Status.Ancestors[0].Conditions, string(egv1a1.EnvoyProxyConditionWarning))
		require.NotNil(t, cond)
		assert.Equal(t, string(egv1a1.EnvoyProxyReasonDeprecatedField), cond.Reason)
		assert.Equal(t, "spec.telemetry.foo is deprecated, use spec.telemetry.bar instead", cond.Message)
		assert.Equal(t, ep.Generation, cond.ObservedGeneration)
	})

	t.Run("removes a warning left over from an earlier translation when nothing is left to warn about", func(t *testing.T) {
		ep := &egv1a1.EnvoyProxy{}
		UpdateEnvoyProxyStatusAccepted(ep, ancestor, egv1a1.EnvoyProxyReasonAccepted, "EnvoyProxy has been accepted.")
		SetEnvoyProxyWarnings(ep, ancestor, DeprecatedFieldsWarning(map[string]DeprecatedField{
			"spec.luaValidation": {Replacement: "spec.lua.validationType"},
		}), nil)
		SetEnvoyProxyWarnings(ep, ancestor)

		require.Len(t, ep.Status.Ancestors, 1)
		assert.Nil(t, meta.FindStatusCondition(ep.Status.Ancestors[0].Conditions, string(egv1a1.EnvoyProxyConditionWarning)))
		assert.NotNil(t, meta.FindStatusCondition(ep.Status.Ancestors[0].Conditions, string(egv1a1.EnvoyProxyConditionAccepted)))
	})

	t.Run("combines deprecated and invalid field warnings", func(t *testing.T) {
		ep := &egv1a1.EnvoyProxy{}
		SetEnvoyProxyWarnings(ep, ancestor, DeprecatedFieldsWarning(map[string]DeprecatedField{
			"spec.luaValidation": {Replacement: "spec.lua.validationType"},
		}), InvalidFieldsWarning([]validation.InvalidField{{Path: "filterOrder", Err: errors.New("cycle")}}))

		conds := ep.Status.Ancestors[0].Conditions
		require.Len(t, conds, 1)
		assert.Equal(t, string(egv1a1.EnvoyProxyReasonMultipleWarnings), conds[0].Reason)
		assert.Equal(t, "spec.luaValidation is deprecated, use spec.lua.validationType instead; "+
			"The following invalid fields were ignored: filterOrder: cycle", conds[0].Message)
	})

	t.Run("sets an InvalidField warning on its own", func(t *testing.T) {
		ep := &egv1a1.EnvoyProxy{}
		SetEnvoyProxyWarnings(ep, ancestor, InvalidFieldsWarning([]validation.InvalidField{{Path: "filterOrder", Err: errors.New("cycle")}}))

		conds := ep.Status.Ancestors[0].Conditions
		require.Len(t, conds, 1)
		assert.Equal(t, string(egv1a1.EnvoyProxyReasonInvalidField), conds[0].Reason)
	})

	t.Run("sets an UnsupportedField warning for fields that have no effect", func(t *testing.T) {
		ep := &egv1a1.EnvoyProxy{}
		SetEnvoyProxyWarnings(ep, ancestor, UnsupportedFieldsWarning([]string{"spec.mergeGateways"}))

		require.Len(t, ep.Status.Ancestors, 1)
		cond := meta.FindStatusCondition(ep.Status.Ancestors[0].Conditions, string(egv1a1.EnvoyProxyConditionWarning))
		require.NotNil(t, cond)
		assert.Equal(t, string(EnvoyProxyReasonUnsupportedField), cond.Reason)
		assert.Equal(t, "The following fields have no effect on an EnvoyProxy referenced by a Gateway and were ignored, "+
			"set them on the EnvoyProxy referenced by the GatewayClass instead: spec.mergeGateways", cond.Message)
	})

	t.Run("merges an UnsupportedField warning with the other warnings", func(t *testing.T) {
		ep := &egv1a1.EnvoyProxy{}
		SetEnvoyProxyWarnings(ep, ancestor, DeprecatedFieldsWarning(map[string]DeprecatedField{
			"spec.luaValidation": {Replacement: "spec.lua.validationType"},
		}), PatchDisabledWarning([]string{"spec.provider.kubernetes.envoyDeployment.patch"}),
			UnsupportedFieldsWarning([]string{"spec.mergeGateways"}))

		require.Len(t, ep.Status.Ancestors, 1)
		cond := meta.FindStatusCondition(ep.Status.Ancestors[0].Conditions, string(egv1a1.EnvoyProxyConditionWarning))
		require.NotNil(t, cond)
		assert.Equal(t, string(egv1a1.EnvoyProxyReasonMultipleWarnings), cond.Reason)
		assert.Equal(t, "spec.luaValidation is deprecated, use spec.lua.validationType instead; "+
			"The following patch fields were ignored because patching is disabled in the EnvoyGateway configuration: "+
			"spec.provider.kubernetes.envoyDeployment.patch; "+
			"The following fields have no effect on an EnvoyProxy referenced by a Gateway and were ignored, "+
			"set them on the EnvoyProxy referenced by the GatewayClass instead: spec.mergeGateways", cond.Message)
	})

	t.Run("no-op when every warning is nil", func(t *testing.T) {
		ep := &egv1a1.EnvoyProxy{}
		SetEnvoyProxyWarnings(ep, ancestor, nil, DeprecatedFieldsWarning(nil), InvalidFieldsWarning(nil), PatchDisabledWarning(nil),
			UnsupportedFieldsWarning(nil))
		assert.Empty(t, ep.Status.Ancestors)
	})
}

func TestMergeBoundedWarningMessages(t *testing.T) {
	t.Run("short messages are merged as before", func(t *testing.T) {
		assert.Equal(t, "a; b", mergeBoundedWarningMessages("a", "b"))
		assert.Equal(t, "a", mergeBoundedWarningMessages("a", "a"))
	})

	t.Run("a full existing message is cut short to keep the next warning", func(t *testing.T) {
		existing := truncateConditionMessage(strings.Repeat("x", conditionMessageMaxLength+100))
		next := "The following patch fields were ignored because patching is disabled"
		merged := mergeBoundedWarningMessages(existing, next)
		require.Len(t, merged, conditionMessageMaxLength)
		assert.True(t, strings.HasSuffix(merged, conditionMessageTruncationSuffix+"; "+next))
		assert.Equal(t, merged, truncateConditionMessage(merged))
	})
}

func TestSetEnvoyProxyWarningKeepsLaterWarningWhenFull(t *testing.T) {
	ancestor := &gwapiv1.ParentReference{Name: gwapiv1.ObjectName("eg")}
	ep := &egv1a1.EnvoyProxy{}
	fields := make(map[string]DeprecatedField)
	for i := range 2500 {
		fields[fmt.Sprintf("spec.telemetry.accessLog.settings[%d].sinks[0].openTelemetry.host", i)] = DeprecatedField{Replacement: "backendRefs"}
	}
	SetEnvoyProxyWarnings(ep, ancestor, DeprecatedFieldsWarning(fields), PatchDisabledWarning([]string{"spec.provider.kubernetes.envoyDeployment.patch"}))

	cond := meta.FindStatusCondition(ep.Status.Ancestors[0].Conditions, string(egv1a1.EnvoyProxyConditionWarning))
	require.NotNil(t, cond)
	assert.Len(t, cond.Message, conditionMessageMaxLength)
	assert.Contains(t, cond.Message, "spec.provider.kubernetes.envoyDeployment.patch")
	assert.Equal(t, string(egv1a1.EnvoyProxyReasonMultipleWarnings), cond.Reason)
}

func TestSetEnvoyProxyWarningsKeepsUnsupportedFieldWarningWhenTruncating(t *testing.T) {
	ancestor := &gwapiv1.ParentReference{Name: gwapiv1.ObjectName("eg")}
	invalid := make([]validation.InvalidField, 0, 2000)
	for i := range 2000 {
		invalid = append(invalid, validation.InvalidField{
			Path: fmt.Sprintf("telemetry.accessLog.settings[%d].sinks[0]", i),
			Err:  errors.New("unsupported sink type"),
		})
	}
	unsupportedMsg := "The following fields have no effect on an EnvoyProxy referenced by a Gateway and were ignored, " +
		"set them on the EnvoyProxy referenced by the GatewayClass instead: spec.mergeGateways"

	ep := &egv1a1.EnvoyProxy{}
	SetEnvoyProxyWarnings(ep, ancestor,
		PatchDisabledWarning([]string{"spec.provider.kubernetes.envoyDeployment.patch"}),
		InvalidFieldsWarning(invalid),
		UnsupportedFieldsWarning([]string{"spec.mergeGateways"}))

	require.Len(t, ep.Status.Ancestors, 1)
	cond := meta.FindStatusCondition(ep.Status.Ancestors[0].Conditions, string(egv1a1.EnvoyProxyConditionWarning))
	require.NotNil(t, cond)
	assert.Equal(t, string(egv1a1.EnvoyProxyReasonMultipleWarnings), cond.Reason)
	assert.LessOrEqual(t, len(cond.Message), conditionMessageMaxLength)
	assert.Contains(t, cond.Message, conditionMessageTruncationSuffix)
	assert.True(t, strings.HasSuffix(cond.Message, "; "+unsupportedMsg), "the UnsupportedField warning must be kept in full")
}
