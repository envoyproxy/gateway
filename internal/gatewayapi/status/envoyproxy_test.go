// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/api/v1alpha1/validation"
)

func TestSetEnvoyProxyWarnings(t *testing.T) {
	ancestor := &gwapiv1.ParentReference{Name: gwapiv1.ObjectName("eg")}
	deprecated := DeprecatedFieldsWarning(map[string]string{"spec.luaValidation": "spec.lua.validationType"})
	invalid := InvalidFieldsWarning([]validation.InvalidField{{Path: "filterOrder", Err: errors.New("cycle")}})

	t.Run("sets a warning condition with the reason of the single warning", func(t *testing.T) {
		ep := &egv1a1.EnvoyProxy{}
		SetEnvoyProxyWarnings(ep, ancestor, deprecated, nil)

		require.Len(t, ep.Status.Ancestors, 1)
		conds := ep.Status.Ancestors[0].Conditions
		require.Len(t, conds, 1)
		assert.Equal(t, string(egv1a1.EnvoyProxyConditionWarning), conds[0].Type)
		assert.Equal(t, metav1.ConditionTrue, conds[0].Status)
		assert.Equal(t, string(egv1a1.EnvoyProxyReasonDeprecatedField), conds[0].Reason)
		assert.Equal(t, "spec.luaValidation is deprecated, use spec.lua.validationType instead", conds[0].Message)
	})

	t.Run("combines warnings with different reasons", func(t *testing.T) {
		ep := &egv1a1.EnvoyProxy{}
		SetEnvoyProxyWarnings(ep, ancestor, deprecated, invalid)

		conds := ep.Status.Ancestors[0].Conditions
		require.Len(t, conds, 1)
		assert.Equal(t, string(egv1a1.EnvoyProxyReasonMultipleWarnings), conds[0].Reason)
		assert.Equal(t, "spec.luaValidation is deprecated, use spec.lua.validationType instead; "+
			"The following invalid fields were ignored: filterOrder: cycle", conds[0].Message)
	})

	t.Run("no-op when there is no warning and no ancestor status", func(t *testing.T) {
		ep := &egv1a1.EnvoyProxy{}
		SetEnvoyProxyWarnings(ep, ancestor, nil, DeprecatedFieldsWarning(nil), InvalidFieldsWarning(nil), PatchDisabledWarning(nil))
		assert.Empty(t, ep.Status.Ancestors)
	})

	t.Run("keeps the Accepted condition on the same ancestor", func(t *testing.T) {
		ep := &egv1a1.EnvoyProxy{}
		UpdateEnvoyProxyStatusAccepted(ep, ancestor, egv1a1.EnvoyProxyReasonAccepted, "EnvoyProxy has been accepted.")
		SetEnvoyProxyWarnings(ep, ancestor, invalid)

		assert.Len(t, ep.Status.Ancestors[0].Conditions, 2)
	})

	t.Run("clears a stale warning once there is no warning", func(t *testing.T) {
		ep := &egv1a1.EnvoyProxy{}
		UpdateEnvoyProxyStatusAccepted(ep, ancestor, egv1a1.EnvoyProxyReasonAccepted, "EnvoyProxy has been accepted.")
		SetEnvoyProxyWarnings(ep, ancestor, invalid)
		SetEnvoyProxyWarnings(ep, ancestor)

		conds := ep.Status.Ancestors[0].Conditions
		require.Len(t, conds, 1)
		assert.Equal(t, string(egv1a1.EnvoyProxyConditionAccepted), conds[0].Type)
	})
}
