// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestSetEnvoyProxyDeprecatedFieldsWarning(t *testing.T) {
	ancestor := &gwapiv1.ParentReference{Name: gwapiv1.ObjectName("eg")}

	t.Run("sets a warning condition when deprecated fields are used", func(t *testing.T) {
		ep := &egv1a1.EnvoyProxy{}
		SetEnvoyProxyDeprecatedFieldsWarning(ep, ancestor, map[string]string{
			"spec.luaValidation": "spec.lua.validationType",
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
		SetEnvoyProxyDeprecatedFieldsWarning(ep, ancestor, map[string]string{
			"spec.luaValidation": "spec.lua.validationType",
		})

		assert.Len(t, ep.Status.Ancestors, 1)
		conds := ep.Status.Ancestors[0].Conditions
		assert.Len(t, conds, 2)
	})
}
