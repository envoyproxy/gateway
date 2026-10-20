// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestSetEnvoyProxyDeprecatedFieldsWarning(t *testing.T) {
	ancestor := &gwapiv1.ParentReference{Name: gwapiv1.ObjectName("eg")}

	t.Run("sets a warning condition when deprecated fields are used", func(t *testing.T) {
		ep := &egv1a1.EnvoyProxy{}
		SetEnvoyProxyDeprecatedFieldsWarning(ep, ancestor, map[string]DeprecatedField{
			"spec.luaValidation": {Replacement: "spec.lua.validationType"},
		})

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
		SetEnvoyProxyDeprecatedFieldsWarning(ep, ancestor, nil)
		assert.Empty(t, ep.Status.Ancestors)
	})

	t.Run("appends the warning alongside an existing Accepted condition on the same ancestor", func(t *testing.T) {
		ep := &egv1a1.EnvoyProxy{}
		UpdateEnvoyProxyStatusAccepted(ep, ancestor, egv1a1.EnvoyProxyReasonAccepted, "EnvoyProxy has been accepted.")
		SetEnvoyProxyDeprecatedFieldsWarning(ep, ancestor, map[string]DeprecatedField{
			"spec.luaValidation": {Replacement: "spec.lua.validationType"},
		})

		assert.Len(t, ep.Status.Ancestors, 1)
		conds := ep.Status.Ancestors[0].Conditions
		assert.Len(t, conds, 2)
	})

	t.Run("merges with a PatchDisabled warning on the same ancestor", func(t *testing.T) {
		ep := &egv1a1.EnvoyProxy{}
		UpdateEnvoyProxyStatusAccepted(ep, ancestor, egv1a1.EnvoyProxyReasonAccepted, "EnvoyProxy has been accepted.")
		SetEnvoyProxyDeprecatedFieldsWarning(ep, ancestor, map[string]DeprecatedField{
			"spec.luaValidation": {Replacement: "spec.lua.validationType"},
		})
		SetEnvoyProxyPatchDisabledWarning(ep, ancestor, []string{"spec.provider.kubernetes.envoyDeployment.patch"})

		require.Len(t, ep.Status.Ancestors, 1)
		cond := meta.FindStatusCondition(ep.Status.Ancestors[0].Conditions, string(egv1a1.EnvoyProxyConditionWarning))
		require.NotNil(t, cond)
		assert.Equal(t, string(egv1a1.EnvoyProxyReasonMultipleWarnings), cond.Reason)
		assert.Equal(t, "spec.luaValidation is deprecated, use spec.lua.validationType instead; "+
			"The following patch fields were ignored because patching is disabled in the EnvoyGateway configuration: "+
			"spec.provider.kubernetes.envoyDeployment.patch", cond.Message)
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
	SetEnvoyProxyDeprecatedFieldsWarning(ep, ancestor, fields)
	SetEnvoyProxyPatchDisabledWarning(ep, ancestor, []string{"spec.provider.kubernetes.envoyDeployment.patch"})

	cond := meta.FindStatusCondition(ep.Status.Ancestors[0].Conditions, string(egv1a1.EnvoyProxyConditionWarning))
	require.NotNil(t, cond)
	assert.Len(t, cond.Message, conditionMessageMaxLength)
	assert.Contains(t, cond.Message, "spec.provider.kubernetes.envoyDeployment.patch")
	assert.Equal(t, string(egv1a1.EnvoyProxyReasonMultipleWarnings), cond.Reason)
}
