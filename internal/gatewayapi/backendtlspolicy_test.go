// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestUpstreamCASecretNameIsolation(t *testing.T) {
	t.Run("listener certificate", func(t *testing.T) {
		name := upstreamCASecretName([]gwapiv1.ObjectReference{{Kind: "ClusterTrustBundle", Name: "root-ca"}}, "default")
		listenerName := irTLSListenerConfigName(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Namespace: "clustertrustbundle", Name: "root-ca"},
		})
		require.Equal(t, "upstream-ca:clustertrustbundle/root-ca", name)
		require.NotEqual(t, listenerName, name)
	})

	t.Run("client validation CA", func(t *testing.T) {
		for _, kind := range []gwapiv1.Kind{"ConfigMap", "Secret"} {
			name := upstreamCASecretName([]gwapiv1.ObjectReference{{Kind: kind, Name: "ca.crt"}}, "backend")
			downstreamName := irTLSCACertName(strings.ToLower(string(kind)), "backend")
			require.NotEqual(t, downstreamName, name)
		}
	})

	t.Run("bounded name retains prefix", func(t *testing.T) {
		name := upstreamCASecretName([]gwapiv1.ObjectReference{{
			Kind: "ConfigMap", Name: gwapiv1.ObjectName(strings.Repeat("a", 253)),
		}}, "default")
		require.True(t, strings.HasPrefix(name, "upstream-ca:"))
		require.Len(t, name, maxUpstreamCASecretNameBytes)
	})
}
