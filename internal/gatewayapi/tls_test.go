// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
)

const (
	secretName      = "secret"
	secretNamespace = "test"
)

func TestValidateCipherSuites(t *testing.T) {
	testCases := []struct {
		name    string
		ciphers []string
		wantErr string
	}{
		{
			name: "openssl style names",
			ciphers: []string{
				"ECDHE-ECDSA-AES128-GCM-SHA256",
				"ECDHE-RSA-AES128-GCM-SHA256",
				"ECDHE-ECDSA-AES256-GCM-SHA384",
				"ECDHE-RSA-AES256-GCM-SHA384",
				"ECDHE-ECDSA-CHACHA20-POLY1305",
				"ECDHE-RSA-CHACHA20-POLY1305",
				"ECDHE-ECDSA-AES128-SHA",
				"ECDHE-RSA-AES128-SHA",
				"AES128-GCM-SHA256",
				"AES128-SHA",
				"ECDHE-ECDSA-AES256-SHA",
				"ECDHE-RSA-AES256-SHA",
				"AES256-GCM-SHA384",
				"AES256-SHA",
			},
		},
		{
			name: "iana aliases",
			ciphers: []string{
				"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
				"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
				"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
				"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
				"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
				"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
				"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
				"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
				"TLS_RSA_WITH_AES_128_GCM_SHA256",
				"TLS_RSA_WITH_AES_128_CBC_SHA",
				"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
				"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
				"TLS_RSA_WITH_AES_256_GCM_SHA384",
				"TLS_RSA_WITH_AES_256_CBC_SHA",
			},
		},
		{
			name:    "invalid name",
			ciphers: []string{"INVALID-CIPHER"},
			wantErr: "unsupported cipher suite: INVALID-CIPHER",
		},
		{
			name:    "unsupported iana name",
			ciphers: []string{"TLS_DHE_RSA_WITH_AES_128_GCM_SHA256"},
			wantErr: "unsupported cipher suite: TLS_DHE_RSA_WITH_AES_128_GCM_SHA256",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCipherSuites(tc.ciphers)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

// createTestSecret creates a K8s tls secret using testdata
// see for more info <https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets>
func createTestSecrets(t *testing.T, certFiles, keyFiles []string) []*corev1.Secret {
	t.Helper()

	secrets := make([]*corev1.Secret, 0, len(certFiles))
	for idx, certFile := range certFiles {
		keyFile := keyFiles[idx]

		certData, err := os.ReadFile(filepath.Join("testdata", "tls", certFile))
		require.NoError(t, err)

		keyData, err := os.ReadFile(filepath.Join("testdata", "tls", keyFile))
		require.NoError(t, err)

		secrets = append(secrets, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: secretNamespace,
			},
			Type: corev1.SecretTypeTLS,
			Data: map[string][]byte{
				corev1.TLSCertKey:       certData,
				corev1.TLSPrivateKeyKey: keyData,
			},
		})
	}
	return secrets
}

// TestValidateTLSSecretData ensures that we can properly validate the contents of a K8s tls secret.
// The test assumes the secret is valid and was able to be applied to a cluster.
func TestValidateTLSSecretsData(t *testing.T) {
	type testCase struct {
		Name               string
		CertFiles          []string
		KeyFiles           []string
		ExpectedErrMsg     string
		ExpectedErrReason  gwapiv1.ListenerConditionReason
		ExpectedValidCount int
		ExpectedCertsCount int
	}

	testCases := []testCase{
		{
			Name:               "valid-rsa-pkcs1",
			CertFiles:          []string{"rsa-cert.pem"},
			KeyFiles:           []string{"rsa-pkcs1.key"},
			ExpectedErrMsg:     "",
			ExpectedErrReason:  "",
			ExpectedValidCount: 1,
			ExpectedCertsCount: 1,
		},
		{
			Name:               "valid-rsa-pkcs8",
			CertFiles:          []string{"rsa-cert.pem"},
			KeyFiles:           []string{"rsa-pkcs8.key"},
			ExpectedErrMsg:     "",
			ExpectedErrReason:  "",
			ExpectedValidCount: 1,
			ExpectedCertsCount: 1,
		},
		{
			Name:               "valid-rsa-san-domain",
			CertFiles:          []string{"rsa-cert-san.pem"},
			KeyFiles:           []string{"rsa-pkcs8-san.key"},
			ExpectedErrMsg:     "",
			ExpectedErrReason:  "",
			ExpectedValidCount: 1,
			ExpectedCertsCount: 1,
		},
		{
			Name:               "valid-rsa-wildcard-domain",
			CertFiles:          []string{"rsa-cert-wildcard.pem"},
			KeyFiles:           []string{"rsa-pkcs1-wildcard.key"},
			ExpectedErrMsg:     "",
			ExpectedErrReason:  "",
			ExpectedValidCount: 1,
			ExpectedCertsCount: 1,
		},
		{
			Name:               "valid-rsa-duplicate-san-domain",
			CertFiles:          []string{"rsa-cert-dup-san.pem"},
			KeyFiles:           []string{"rsa-pkcs8-dup-san.key"},
			ExpectedErrMsg:     "",
			ExpectedErrReason:  "",
			ExpectedValidCount: 1,
			ExpectedCertsCount: 1,
		},
		{
			// Two distinct secrets legitimately claiming the same domain with the
			// same public key algorithm must still be rejected.
			Name:               "conflicting-rsa-algorithm-same-domain-different-secrets",
			CertFiles:          []string{"rsa-cert.pem", "rsa-cert.pem"},
			KeyFiles:           []string{"rsa-pkcs1.key", "rsa-pkcs1.key"},
			ExpectedErrMsg:     "test/secret public key algorithm must be unique, certificate domain foo.bar.com has a conflicting algorithm [RSA]",
			ExpectedErrReason:  status.ListenerReasonPartiallyInvalidCertificateRef,
			ExpectedValidCount: 1,
			ExpectedCertsCount: 1,
		},
		{
			Name:               "valid-ecdsa-p256",
			CertFiles:          []string{"ecdsa-p256-cert.pem"},
			KeyFiles:           []string{"ecdsa-p256.key"},
			ExpectedErrMsg:     "",
			ExpectedErrReason:  "",
			ExpectedValidCount: 1,
			ExpectedCertsCount: 1,
		},
		{
			Name:               "valid-ecdsa-p384",
			CertFiles:          []string{"ecdsa-p384-cert.pem"},
			KeyFiles:           []string{"ecdsa-p384.key"},
			ExpectedErrMsg:     "",
			ExpectedErrReason:  "",
			ExpectedValidCount: 1,
			ExpectedCertsCount: 1,
		},
		{
			Name:               "malformed-cert-pem-encoding",
			CertFiles:          []string{"malformed-encoding.pem"},
			KeyFiles:           []string{"rsa-pkcs8.key"},
			ExpectedErrMsg:     "test/secret must contain valid tls.crt and tls.key, unable to validate certificate in tls.crt: unable to decode pem data for certificate",
			ExpectedErrReason:  gwapiv1.ListenerReasonInvalidCertificateRef,
			ExpectedValidCount: 0,
			ExpectedCertsCount: 0,
		},
		{
			Name:               "malformed-key-pem-encoding",
			CertFiles:          []string{"rsa-cert.pem"},
			KeyFiles:           []string{"malformed-encoding.pem"},
			ExpectedErrMsg:     "test/secret must contain valid tls.crt and tls.key, unable to decode pem data in tls.key",
			ExpectedErrReason:  gwapiv1.ListenerReasonInvalidCertificateRef,
			ExpectedValidCount: 0,
			ExpectedCertsCount: 0,
		},
		{
			Name:               "malformed-cert",
			CertFiles:          []string{"malformed-cert.pem"},
			KeyFiles:           []string{"rsa-pkcs8.key"},
			ExpectedErrMsg:     "test/secret must contain valid tls.crt and tls.key, unable to validate certificate in tls.crt: x509: malformed certificate",
			ExpectedErrReason:  gwapiv1.ListenerReasonInvalidCertificateRef,
			ExpectedValidCount: 0,
			ExpectedCertsCount: 0,
		},
		{
			Name:               "malformed-pkcs8-key",
			CertFiles:          []string{"rsa-cert.pem"},
			KeyFiles:           []string{"malformed-pkcs8.key"},
			ExpectedErrMsg:     "test/secret must contain valid tls.crt and tls.key, unable to parse PKCS8 formatted private key in tls.key",
			ExpectedErrReason:  gwapiv1.ListenerReasonInvalidCertificateRef,
			ExpectedValidCount: 0,
			ExpectedCertsCount: 0,
		},
		{
			Name:               "malformed-pkcs1-key",
			CertFiles:          []string{"rsa-cert.pem"},
			KeyFiles:           []string{"malformed-pkcs1.key"},
			ExpectedErrMsg:     "test/secret must contain valid tls.crt and tls.key, unable to parse PKCS1 formatted private key in tls.key",
			ExpectedErrReason:  gwapiv1.ListenerReasonInvalidCertificateRef,
			ExpectedValidCount: 0,
			ExpectedCertsCount: 0,
		},
		{
			Name:               "malformed-ecdsa-key",
			CertFiles:          []string{"rsa-cert.pem"},
			KeyFiles:           []string{"malformed-ecdsa.key"},
			ExpectedErrMsg:     "test/secret must contain valid tls.crt and tls.key, unable to parse EC formatted private key in tls.key",
			ExpectedErrReason:  gwapiv1.ListenerReasonInvalidCertificateRef,
			ExpectedValidCount: 0,
			ExpectedCertsCount: 0,
		},
		{
			Name:               "invalid-key-type",
			CertFiles:          []string{"rsa-cert.pem"},
			KeyFiles:           []string{"invalid-key-type.key"},
			ExpectedErrMsg:     "test/secret must contain valid tls.crt and tls.key, FOO key format found in tls.key, supported formats are PKCS1, PKCS8 or EC",
			ExpectedErrReason:  gwapiv1.ListenerReasonInvalidCertificateRef,
			ExpectedValidCount: 0,
			ExpectedCertsCount: 0,
		},
		{
			Name:               "cert-key-mismatch",
			CertFiles:          []string{"rsa-cert.pem"},
			KeyFiles:           []string{"rsa-pkcs8-san.key"},
			ExpectedErrMsg:     "test/secret must contain a matching tls.crt and tls.key: tls: private key does not match public key",
			ExpectedErrReason:  gwapiv1.ListenerReasonInvalidCertificateRef,
			ExpectedValidCount: 0,
			ExpectedCertsCount: 0,
		},
		{
			Name:               "cert-key-mismatch-isolated-from-valid-secret",
			CertFiles:          []string{"ecdsa-p256-cert.pem", "rsa-cert.pem"},
			KeyFiles:           []string{"ecdsa-p256.key", "rsa-pkcs8-san.key"},
			ExpectedErrMsg:     "test/secret must contain a matching tls.crt and tls.key: tls: private key does not match public key",
			ExpectedErrReason:  status.ListenerReasonPartiallyInvalidCertificateRef,
			ExpectedValidCount: 1,
			ExpectedCertsCount: 1,
		},
		{
			Name:               "all-valid-secrets",
			CertFiles:          []string{"ecdsa-p256-cert.pem", "rsa-cert.pem"},
			KeyFiles:           []string{"ecdsa-p256.key", "rsa-pkcs1.key"},
			ExpectedErrMsg:     "",
			ExpectedErrReason:  "",
			ExpectedValidCount: 2,
			ExpectedCertsCount: 2,
		},
		{
			Name:               "partially-invalid-secrets",
			CertFiles:          []string{"ecdsa-p256-cert.pem", "rsa-cert.pem"},
			KeyFiles:           []string{"ecdsa-p256.key", "invalid-key-type.key"},
			ExpectedErrMsg:     "test/secret must contain valid tls.crt and tls.key, FOO key format found in tls.key, supported formats are PKCS1, PKCS8 or EC",
			ExpectedErrReason:  status.ListenerReasonPartiallyInvalidCertificateRef,
			ExpectedValidCount: 1,
			ExpectedCertsCount: 1,
		},
		{
			Name:               "all-invalid-secrets",
			CertFiles:          []string{"ecdsa-p256-cert.pem", "rsa-cert.pem"},
			KeyFiles:           []string{"invalid-key-type.key", "invalid-key-type.key"},
			ExpectedErrMsg:     "test/secret must contain valid tls.crt and tls.key, FOO key format found in tls.key, supported formats are PKCS1, PKCS8 or EC\ntest/secret must contain valid tls.crt and tls.key, FOO key format found in tls.key, supported formats are PKCS1, PKCS8 or EC",
			ExpectedErrReason:  gwapiv1.ListenerReasonInvalidCertificateRef,
			ExpectedValidCount: 0,
			ExpectedCertsCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			secrets := createTestSecrets(t, tc.CertFiles, tc.KeyFiles)
			require.NotNil(t, secrets)
			validSecrets, certs, err := parseCertsFromTLSSecretsData(secrets)

			require.Equal(t, len(validSecrets), tc.ExpectedValidCount)
			require.Equal(t, len(certs), tc.ExpectedCertsCount)

			if tc.ExpectedErrMsg == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, err.Error(), tc.ExpectedErrMsg)
				require.Equal(t, tc.ExpectedErrReason, err.Reason())
			}
		})
	}
}

func TestFilterValidCertificates(t *testing.T) {
	type testCase struct {
		Name              string
		CertFile          string
		ExpectedErrMsg    string
		ExpectedErrReason gwapiv1.ListenerConditionReason
		ExpectedDataFile  string // File containing expected filtered certificate data
	}

	testCases := []testCase{
		{
			Name:              "valid-rsa-cert",
			CertFile:          "rsa-cert.pem",
			ExpectedErrMsg:    "",
			ExpectedErrReason: "",
			ExpectedDataFile:  "rsa-cert.pem", // Same as input for valid cert
		},
		{
			Name:              "valid-ecdsa-p256-cert",
			CertFile:          "ecdsa-p256-cert.pem",
			ExpectedErrMsg:    "",
			ExpectedErrReason: "",
			ExpectedDataFile:  "ecdsa-p256-cert.pem", // Same as input
		},
		{
			Name:              "valid-ecdsa-p384-cert",
			CertFile:          "ecdsa-p384-cert.pem",
			ExpectedErrMsg:    "",
			ExpectedErrReason: "",
			ExpectedDataFile:  "ecdsa-p384-cert.pem", // Same as input
		},
		{
			Name:              "malformed-cert",
			CertFile:          "malformed-cert.pem",
			ExpectedErrMsg:    "x509: malformed certificate",
			ExpectedErrReason: gwapiv1.ListenerReasonInvalidCertificateRef,
		},
		{
			Name:              "bundle-both-valid",
			CertFile:          "bundle-both-valid.pem",
			ExpectedErrMsg:    "",
			ExpectedErrReason: "",
			ExpectedDataFile:  "bundle-both-valid.pem", // All certs are valid
		},
		{
			Name:              "bundle-first-valid",
			CertFile:          "bundle-first-valid.pem", // rsa-cert.pem + malformed-cert.pem
			ExpectedErrMsg:    "x509: malformed certificate",
			ExpectedErrReason: status.ListenerReasonPartiallyInvalidCertificateRef,
			ExpectedDataFile:  "rsa-cert.pem", // Only first cert
		},
		{
			Name:              "bundle-first-invalid",
			CertFile:          "bundle-first-invalid.pem", // malformed-cert.pem + rsa-cert.pem
			ExpectedErrMsg:    "x509: malformed certificate",
			ExpectedErrReason: status.ListenerReasonPartiallyInvalidCertificateRef,
			ExpectedDataFile:  "rsa-cert.pem", // Only second cert
		},
		{
			Name:              "bundle-both-invalid",
			CertFile:          "bundle-both-invalid.pem",
			ExpectedErrMsg:    "x509: malformed certificate\nx509: malformed certificate",
			ExpectedErrReason: gwapiv1.ListenerReasonInvalidCertificateRef,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			certData, err := os.ReadFile(filepath.Join("testdata", "tls", tc.CertFile))
			require.NoError(t, err)

			result, listenerErr := filterValidCACertificates(certData)

			if tc.ExpectedErrMsg == "" {
				require.NoError(t, listenerErr)
				require.NotNil(t, result)

				// If ExpectedDataFile is provided, compare with expected data
				if tc.ExpectedDataFile != "" {
					expectedData, err := os.ReadFile(filepath.Join("testdata", "tls", tc.ExpectedDataFile))
					require.NoError(t, err)
					require.Equal(t, expectedData, result, "filtered certificate data should match expected value")
				}
			} else {
				require.Error(t, listenerErr)
				require.Equal(t, tc.ExpectedErrMsg, listenerErr.Error())
				require.Equal(t, tc.ExpectedErrReason, listenerErr.Reason())

				// For PartiallyInvalidCertificateRef, we should still have valid data
				if tc.ExpectedErrReason == status.ListenerReasonPartiallyInvalidCertificateRef {
					require.NotNil(t, result, "result should not be nil for partially invalid certificates")
					if tc.ExpectedDataFile != "" {
						expectedData, err := os.ReadFile(filepath.Join("testdata", "tls", tc.ExpectedDataFile))
						require.NoError(t, err)
						require.Equal(t, expectedData, result, "filtered certificate data should match expected value")
					}
				} else {
					require.Nil(t, result, "result should be nil for completely invalid certificates")
				}
			}
		})
	}
}

func TestAppendDedupPEMCertsWithSeen(t *testing.T) {
	rsaCert, err := os.ReadFile(filepath.Join("testdata", "tls", "rsa-cert.pem"))
	require.NoError(t, err)

	ecdsaCert, err := os.ReadFile(filepath.Join("testdata", "tls", "ecdsa-p256-cert.pem"))
	require.NoError(t, err)

	testCases := []struct {
		name     string
		dst      []byte
		src      []byte
		expected []byte
	}{
		{
			name:     "append distinct cert to empty dst",
			dst:      []byte{},
			src:      rsaCert,
			expected: rsaCert,
		},
		{
			name:     "append distinct cert to non-empty dst",
			dst:      rsaCert,
			src:      ecdsaCert,
			expected: append(append([]byte{}, rsaCert...), ecdsaCert...),
		},
		{
			name:     "duplicate src cert already in dst is skipped",
			dst:      rsaCert,
			src:      rsaCert,
			expected: rsaCert,
		},
		{
			name:     "only new cert appended when src contains one known and one new",
			dst:      rsaCert,
			src:      append(append([]byte{}, rsaCert...), ecdsaCert...),
			expected: append(append([]byte{}, rsaCert...), ecdsaCert...),
		},
		{
			name:     "duplicate within src itself only appended once",
			dst:      []byte{},
			src:      append(append([]byte{}, rsaCert...), rsaCert...),
			expected: rsaCert,
		},
		{
			name:     "empty src leaves dst unchanged",
			dst:      rsaCert,
			src:      []byte{},
			expected: rsaCert,
		},
		{
			name:     "both empty returns empty",
			dst:      []byte{},
			src:      []byte{},
			expected: []byte{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := appendDedupPEMCertsWithSeen(tc.dst, tc.src, make(map[[sha256.Size]byte]struct{}))
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestBuildListenerTLSParametersDedupCACerts(t *testing.T) {
	caCertPEM, err := os.ReadFile(filepath.Join("testdata", "tls", "rsa-cert.pem"))
	require.NoError(t, err)

	ns := "envoy-gateway"

	// both secrets contain the identical CA PEM
	makeCASecret := func(name string) *corev1.Secret {
		return &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
			Data:       map[string][]byte{CACertKey: caCertPEM},
		}
	}

	policy := &egv1a1.ClientTrafficPolicy{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "test-policy"},
		Spec: egv1a1.ClientTrafficPolicySpec{
			TLS: &egv1a1.ClientTLSSettings{
				ClientValidation: &egv1a1.ClientValidationContext{
					CACertificateRefs: []gwapiv1.SecretObjectReference{
						{Name: "ca-secret-1", Namespace: new(gwapiv1.Namespace(ns))},
						{Name: "ca-secret-2", Namespace: new(gwapiv1.Namespace(ns))},
					},
				},
			},
		},
	}

	resources := &resource.Resources{
		Secrets: []*corev1.Secret{
			makeCASecret("ca-secret-1"),
			makeCASecret("ca-secret-2"),
		},
		ReferenceGrants: nil,
	}

	translator := &Translator{
		TranslatorContext: &TranslatorContext{},
	}
	translator.SetSecrets(resources.Secrets)

	// seed irTLSConfig with a server cert so buildListenerTLSParameters doesn't
	// return early
	irTLSConfig := &ir.TLSConfig{
		Certificates: []ir.TLSCertificate{
			{Name: "dummy"},
		},
	}

	result, err := translator.buildListenerTLSParameters(policy, irTLSConfig, resources)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.CACertificate)

	// PEM blocks must be 1.
	pemCount := 0
	rest := result.CACertificate.Certificate
	for len(rest) > 0 {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		pemCount++
	}
	require.Equal(t, 1, pemCount,
		"expected exactly 1 PEM block after deduplication, got %d — "+
			"appendDedupPEMCertsWithSeen may not be used in buildListenerTLSParameters", pemCount)
}

// buildServingChain builds a serving certificate chain of [leaf, intermediate]
// (leaf first) with the given validity windows, plus the leaf's matching key.
func buildServingChain(t *testing.T, leafNotAfter, caNotAfter time.Time) (chainPEM, keyPEM []byte) {
	t.Helper()
	now := time.Now()

	caKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)
	caTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test Intermediate"},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              caNotAfter,
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	require.NoError(t, err)
	caCert, err := x509.ParseCertificate(caDER)
	require.NoError(t, err)

	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	leafTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "test.example.com"},
		DNSNames:     []string{"test.example.com"},
		NotBefore:    now.Add(-2 * time.Hour),
		NotAfter:     leafNotAfter,
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTmpl, caCert, &leafKey.PublicKey, caKey)
	require.NoError(t, err)

	leafPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: leafDER})
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})
	chainPEM = append(append([]byte{}, leafPEM...), caPEM...)

	keyDER, err := x509.MarshalECPrivateKey(leafKey)
	require.NoError(t, err)
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	return chainPEM, keyPEM
}

func countPEMCertBlocks(data []byte) int {
	n := 0
	for rest := data; len(rest) > 0; {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			n++
		}
	}
	return n
}

// TestValidateServingCertificateChain ensures a serving chain is validated as a
// whole: any expired/not-yet-valid/malformed member rejects the entire chain
// (it is never partially filtered), while a fully valid chain is returned intact.
func TestValidateServingCertificateChain(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-time.Minute)

	t.Run("all-valid-chain-kept-intact", func(t *testing.T) {
		chain, _ := buildServingChain(t, future, future)
		out, lerr := validateServingCertificateChain(chain)
		require.Nil(t, lerr)
		require.Equal(t, 2, countPEMCertBlocks(out), "full chain (leaf + intermediate) must be preserved")
	})

	t.Run("expired-leaf-rejects-whole-chain", func(t *testing.T) {
		chain, _ := buildServingChain(t, past, future) // leaf expired, intermediate valid
		out, lerr := validateServingCertificateChain(chain)
		require.Nil(t, out)
		require.Error(t, lerr)
		require.Equal(t, gwapiv1.ListenerReasonInvalidCertificateRef, lerr.Reason())
		require.Contains(t, lerr.Error(), "has expired")
	})

	t.Run("malformed-cert-rejects-whole-chain", func(t *testing.T) {
		malformed := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("not a der certificate")})
		out, lerr := validateServingCertificateChain(malformed)
		require.Nil(t, out)
		require.Error(t, lerr)
		require.Equal(t, gwapiv1.ListenerReasonInvalidCertificateRef, lerr.Reason())
	})

	t.Run("empty-data", func(t *testing.T) {
		out, lerr := validateServingCertificateChain(nil)
		require.Nil(t, out)
		require.Error(t, lerr)
		require.Equal(t, gwapiv1.ListenerReasonInvalidCertificateRef, lerr.Reason())
	})
}

// TestParseCertsExpiredLeafChainRejected is the #9225 / #9473 regression: a
// legitimate secret (leaf and key match) whose chain has an expired leaf is
// rejected outright at translation, so the corrupted chain never reaches Envoy
// (which would otherwise NACK the whole SDS push with KEY_VALUES_MISMATCH).
func TestParseCertsExpiredLeafChainRejected(t *testing.T) {
	now := time.Now()
	chain, key := buildServingChain(t, now.Add(-time.Minute), now.Add(24*time.Hour))

	// The secret itself is legitimate: the leaf and key match.
	_, err := tls.X509KeyPair(chain, key)
	require.NoError(t, err, "leaf and key are expected to match; the chain is only expired")

	secrets := []*corev1.Secret{{
		ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: secretNamespace},
		Type:       corev1.SecretTypeTLS,
		Data:       map[string][]byte{corev1.TLSCertKey: chain, corev1.TLSPrivateKeyKey: key},
	}}

	validSecrets, certs, listenerErr := parseCertsFromTLSSecretsData(secrets)
	require.Empty(t, validSecrets, "expired serving chain must not be served")
	require.Empty(t, certs)
	require.Error(t, listenerErr)
	require.Equal(t, gwapiv1.ListenerReasonInvalidCertificateRef, listenerErr.Reason())
	require.Contains(t, listenerErr.Error(), "has expired")
}
