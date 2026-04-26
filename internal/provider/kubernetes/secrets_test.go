// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	envoyGatewaySecret = corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-gateway",
			Namespace: "envoy-gateway-system",
		},
	}

	envoySecret = corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy",
			Namespace: "envoy-gateway-system",
		},
	}

	envoyRateLimitSecret = corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-rate-limit",
			Namespace: "envoy-gateway-system",
		},
	}

	oidcHMACSecret = corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-oidc-hmac",
			Namespace: "envoy-gateway-system",
		},
	}

	existingSecretsWithoutHMAC = []client.Object{
		&envoyGatewaySecret,
		&envoySecret,
		&envoyRateLimitSecret,
	}

	existingSecretsWithHMAC = []client.Object{
		&envoyGatewaySecret,
		&envoySecret,
		&envoyRateLimitSecret,
		&oidcHMACSecret,
	}

	SecretsToCreate = []corev1.Secret{
		envoyGatewaySecret,
		envoySecret,
		envoyRateLimitSecret,
		oidcHMACSecret,
	}
)

func TestCreateSecretsWhenUpgrade(t *testing.T) {
	t.Run("create HMAC secret when it does not exist", func(t *testing.T) {
		cli := fakeclient.NewClientBuilder().WithObjects(existingSecretsWithoutHMAC...).Build()

		created, err := CreateOrUpdateSecrets(context.Background(), cli, SecretsToCreate, false)
		require.ErrorIs(t, err, ErrSecretExists)
		require.Len(t, created, 1)
		require.Equal(t, "envoy-oidc-hmac", created[0].Name)

		err = cli.Get(context.Background(), client.ObjectKeyFromObject(&oidcHMACSecret), &corev1.Secret{})
		require.NoError(t, err)
	})

	t.Run("skip HMAC secret when it exist", func(t *testing.T) {
		cli := fakeclient.NewClientBuilder().WithObjects(existingSecretsWithHMAC...).Build()

		created, err := CreateOrUpdateSecrets(context.Background(), cli, SecretsToCreate, false)
		require.ErrorIs(t, err, ErrSecretExists)
		require.Emptyf(t, created, "expected no secrets to be created, got %v", created)
	})

	t.Run("update secrets when they exist", func(t *testing.T) {
		cli := fakeclient.NewClientBuilder().WithObjects(existingSecretsWithHMAC...).Build()

		created, err := CreateOrUpdateSecrets(context.Background(), cli, SecretsToCreate, true)
		require.NoError(t, err)
		require.Len(t, created, 4)
	})
}

// TestCreateOrUpdateSecretsBundlesCA verifies that when rotating certificates the
// old CA is bundled with the new CA so that components that haven't reloaded yet
// continue to be trusted during the transition window.
func TestCreateOrUpdateSecretsBundlesCA(t *testing.T) {
	now := time.Now()
	ca1 := makeTestCAPEM(t, now.Add(365*24*time.Hour))
	ca2 := makeTestCAPEM(t, now.Add(365*24*time.Hour))
	require.NotEqual(t, ca1, ca2)

	// Seed the cluster with a secret carrying the old CA.
	existing := newSecret(corev1.SecretTypeTLS, "envoy-gateway", "test-ns", map[string][]byte{
		caCertificateKey:        ca1,
		corev1.TLSCertKey:       []byte("old-cert"),
		corev1.TLSPrivateKeyKey: []byte("old-key"),
	})
	cli := fakeclient.NewClientBuilder().WithObjects(&existing).Build()

	// Rotate: present a new secret carrying  the new CA.
	rotated := newSecret(corev1.SecretTypeTLS, "envoy-gateway", "test-ns", map[string][]byte{
		caCertificateKey:        ca2,
		corev1.TLSCertKey:       []byte("new-cert"),
		corev1.TLSPrivateKeyKey: []byte("new-key"),
	})
	updated, err := CreateOrUpdateSecrets(context.Background(), cli, []corev1.Secret{rotated}, true)
	require.NoError(t, err)
	require.Len(t, updated, 1)

	bundle := updated[0].Data[caCertificateKey]

	// The bundle must be valid PEM accepted by x509.
	pool := x509.NewCertPool()
	require.True(t, pool.AppendCertsFromPEM(bundle), "bundle must be valid PEM")

	certs := decodePEMCerts(t, bundle)
	require.Len(t, certs, 2, "bundle must contain exactly 2 certs: new CA + old CA")
	assert.Equal(t, decodePEMCerts(t, ca2)[0].Raw, certs[0].Raw, "first cert in bundle must be the new CA")
	assert.Equal(t, decodePEMCerts(t, ca1)[0].Raw, certs[1].Raw, "second cert in bundle must be the old CA")
}

// TestBundleCACerts covers the bundleCACerts helper directly.
func TestBundleCACerts(t *testing.T) {
	now := time.Now()
	ca1 := makeTestCAPEM(t, now.Add(365*24*time.Hour))
	ca2 := makeTestCAPEM(t, now.Add(365*24*time.Hour))
	expired := makeTestCAPEM(t, now.Add(-time.Second)) // already expired

	t.Run("identical CAs return the original bytes unchanged", func(t *testing.T) {
		result := bundleCACerts(ca1, ca1)
		assert.Equal(t, ca1, result)
	})

	t.Run("different CAs are concatenated new-first", func(t *testing.T) {
		result := bundleCACerts(ca2, ca1)
		certs := decodePEMCerts(t, result)
		require.Len(t, certs, 2)
		assert.Equal(t, decodePEMCerts(t, ca2)[0].Raw, certs[0].Raw, "new CA must be first")
		assert.Equal(t, decodePEMCerts(t, ca1)[0].Raw, certs[1].Raw, "old CA must be second")
	})

	t.Run("cert already present in newCA is not duplicated", func(t *testing.T) {
		// Bundle containing ca2+ca1; applying bundleCACerts(ca2, bundle) must
		// not add ca2 again.
		bundle := bundleCACerts(ca2, ca1)
		result := bundleCACerts(ca2, bundle)
		certs := decodePEMCerts(t, result)
		require.Len(t, certs, 2, "ca2 that is already in newCA must not be re-appended")
	})

	t.Run("expired cert from old bundle is excluded", func(t *testing.T) {
		result := bundleCACerts(ca1, expired)
		certs := decodePEMCerts(t, result)
		require.Len(t, certs, 1, "expired cert must not be included in the bundle")
	})

	t.Run("bundle never exceeds two CAs across multiple rotations", func(t *testing.T) {
		ca3 := makeTestCAPEM(t, now.Add(365*24*time.Hour))

		// Rotation 1: CA1 -> CA2
		after1 := bundleCACerts(ca2, ca1)
		require.Len(t, decodePEMCerts(t, after1), 2)

		// Rotation 2: CA2 -> CA3 (oldCA is the after1 bundle: CA2+CA1)
		// Only CA2 (the head of after1) should be carried forward; CA1 is dropped.
		after2 := bundleCACerts(ca3, after1)
		certs := decodePEMCerts(t, after2)
		require.Len(t, certs, 2, "bundle must not grow beyond 2 CAs after a second rotation")
		assert.Equal(t, decodePEMCerts(t, ca3)[0].Raw, certs[0].Raw, "new CA must be first")
		assert.Equal(t, decodePEMCerts(t, ca2)[0].Raw, certs[1].Raw, "only the immediately-previous CA is carried forward")
	})
}

// makeTestCAPEM generates a minimal self-signed CA certificate with the given
// validity window and returns it as a PEM-encoded CERTIFICATE block.
func makeTestCAPEM(t *testing.T, notAfter time.Time) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	now := time.Now()
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              notAfter,
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
}

// decodePEMCerts returns all x509 certificates decoded from a PEM bundle.
func decodePEMCerts(t *testing.T, data []byte) []*x509.Certificate {
	t.Helper()
	var certs []*x509.Certificate
	for rest := data; len(rest) > 0; {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		require.NoError(t, err)
		certs = append(certs, cert)
	}
	return certs
}
