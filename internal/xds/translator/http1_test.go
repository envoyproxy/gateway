// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestHTTP1ProtocolOptionsHeaderKeyFormat(t *testing.T) {
	t.Run("unset returns nil", func(t *testing.T) {
		require.Nil(t, http1ProtocolOptions(nil))
		require.Nil(t, http1ProtocolOptions(&ir.HTTP1Settings{}))
	})

	t.Run("proper case words", func(t *testing.T) {
		got := http1ProtocolOptions(&ir.HTTP1Settings{
			HeaderKeyFormat: egv1a1.HTTP1HeaderKeyFormatProperCaseWords,
		})
		require.NotNil(t, got)
		require.NotNil(t, got.GetHeaderKeyFormat().GetProperCaseWords())
		require.Nil(t, got.GetHeaderKeyFormat().GetStatefulFormatter())
	})

	t.Run("preserve header case", func(t *testing.T) {
		got := http1ProtocolOptions(&ir.HTTP1Settings{PreserveHeaderCase: true})
		require.NotNil(t, got)
		require.Nil(t, got.GetHeaderKeyFormat().GetProperCaseWords())
		require.Equal(t, "preserve_case", got.GetHeaderKeyFormat().GetStatefulFormatter().GetName())
	})
}

func TestBuildHTTP1SettingsHeaderKeyFormat(t *testing.T) {
	t.Run("unset has no header key format", func(t *testing.T) {
		got := buildHTTP1Settings(&ir.HTTP1Settings{})
		require.Nil(t, got.GetHeaderKeyFormat())
	})

	t.Run("proper case words", func(t *testing.T) {
		got := buildHTTP1Settings(&ir.HTTP1Settings{
			HeaderKeyFormat: egv1a1.HTTP1HeaderKeyFormatProperCaseWords,
		})
		require.NotNil(t, got.GetHeaderKeyFormat().GetProperCaseWords())
	})

	t.Run("preserve header case", func(t *testing.T) {
		got := buildHTTP1Settings(&ir.HTTP1Settings{PreserveHeaderCase: true})
		require.Equal(t, "preserve_case", got.GetHeaderKeyFormat().GetStatefulFormatter().GetName())
	})
}
