// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"os"
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
	xtypes "github.com/envoyproxy/gateway/internal/xds/types"
)

func newTestTranslator() *Translator {
	return &Translator{
		Logger: logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo),
	}
}

func newTCtxWithSystemTrustStore(t *testing.T) *xtypes.ResourceVersionTable {
	t.Helper()
	tCtx := new(xtypes.ResourceVersionTable)
	require.NoError(t, emitSystemTrustStoreSecret(tCtx, SystemTrustStoreSecretName))
	require.True(t, tCtx.SystemTrustStore)
	return tCtx
}

func TestEnsureSystemTrustStoreSecret_NoOp(t *testing.T) {
	tCtx := newTCtxWithSystemTrustStore(t)
	tr := newTestTranslator()
	tr.ensureSystemTrustStoreSecret(tCtx)
	// Secret should still be present and unchanged.
	require.NoError(t, validateSystemTrustStoreSecret(tCtx))
}

func TestEnsureSystemTrustStoreSecret_Removed(t *testing.T) {
	tCtx := newTCtxWithSystemTrustStore(t)
	// Simulate extension hook removing the secret.
	tCtx.XdsResources[resourcev3.SecretType] = nil
	tr := newTestTranslator()
	tr.ensureSystemTrustStoreSecret(tCtx)
	// Secret should be restored.
	require.NoError(t, validateSystemTrustStoreSecret(tCtx))
}

func TestEnsureSystemTrustStoreSecret_Modified(t *testing.T) {
	tCtx := newTCtxWithSystemTrustStore(t)
	// Simulate extension hook modifying the secret filename.
	for _, r := range tCtx.XdsResources[resourcev3.SecretType] {
		if s, ok := r.(*tlsv3.Secret); ok && s.Name == SystemTrustStoreSecretName {
			s.Type = &tlsv3.Secret_ValidationContext{
				ValidationContext: &tlsv3.CertificateValidationContext{
					TrustedCa: &corev3.DataSource{
						Specifier: &corev3.DataSource_Filename{Filename: "/tmp/evil-ca.crt"},
					},
				},
			}
		}
	}
	tr := newTestTranslator()
	tr.ensureSystemTrustStoreSecret(tCtx)
	// Secret should be restored to canonical form.
	require.NoError(t, validateSystemTrustStoreSecret(tCtx))
}

func TestEnsureSystemTrustStoreSecret_Duplicated(t *testing.T) {
	tCtx := newTCtxWithSystemTrustStore(t)
	// Simulate extension hook injecting a duplicate.
	duplicate := canonicalSystemTrustStoreSecret()
	_ = tCtx.AddXdsResource(resourcev3.SecretType, duplicate)
	require.Error(t, validateSystemTrustStoreSecret(tCtx)) // two copies → error
	tr := newTestTranslator()
	tr.ensureSystemTrustStoreSecret(tCtx)
	// Exactly one canonical copy should remain.
	require.NoError(t, validateSystemTrustStoreSecret(tCtx))
}

func TestEnsureSystemTrustStoreSecret_NotEmitted(t *testing.T) {
	// If the secret was never emitted (no system trust store in use), ensure is a no-op.
	tCtx := new(xtypes.ResourceVersionTable)
	tr := newTestTranslator()
	tr.ensureSystemTrustStoreSecret(tCtx)
	require.Nil(t, tCtx.XdsResources)
}
