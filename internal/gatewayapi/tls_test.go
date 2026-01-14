// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
)

const (
	secretName      = "secret"
	secretNamespace = "test"
)

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
		Name             string
		CertFile         string
		ExpectedErr      error
		ExpectedDataFile string // File containing expected filtered certificate data
	}

	testCases := []testCase{
		{
			Name:             "valid-rsa-cert",
			CertFile:         "rsa-cert.pem",
			ExpectedErr:      nil,
			ExpectedDataFile: "rsa-cert.pem", // Same as input for valid cert
		},
		{
			Name:             "valid-ecdsa-p256-cert",
			CertFile:         "ecdsa-p256-cert.pem",
			ExpectedErr:      nil,
			ExpectedDataFile: "ecdsa-p256-cert.pem", // Same as input
		},
		{
			Name:             "valid-ecdsa-p384-cert",
			CertFile:         "ecdsa-p384-cert.pem",
			ExpectedErr:      nil,
			ExpectedDataFile: "ecdsa-p384-cert.pem", // Same as input
		},
		{
			Name:        "malformed-cert",
			CertFile:    "malformed-cert.pem",
			ExpectedErr: errors.New("x509: malformed certificate"),
		},
		{
			Name:             "bundle-both-valid",
			CertFile:         "bundle-both-valid.pem",
			ExpectedErr:      nil,
			ExpectedDataFile: "bundle-both-valid.pem", // All certs are valid
		},
		{
			Name:             "bundle-first-valid",
			CertFile:         "bundle-first-valid.pem", // rsa-cert.pem + malformed-cert.pem
			ExpectedErr:      nil,
			ExpectedDataFile: "rsa-cert.pem", // Only first cert
		},
		{
			Name:             "bundle-first-invalid",
			CertFile:         "bundle-first-invalid.pem", // malformed-cert.pem + rsa-cert.pem
			ExpectedErr:      nil,
			ExpectedDataFile: "rsa-cert.pem", // Only second cert
		},
		{
			Name:        "bundle-both-invalid",
			CertFile:    "bundle-both-invalid.pem",
			ExpectedErr: errors.New("x509: malformed certificate\nx509: malformed certificate"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			certData, err := os.ReadFile(filepath.Join("testdata", "tls", tc.CertFile))
			require.NoError(t, err)

			result, err := filterValidCertificates(certData)

			if tc.ExpectedErr == nil {
				require.NoError(t, err)
				require.NotNil(t, result)

				// If ExpectedDataFile is provided, compare with expected data
				if tc.ExpectedDataFile != "" {
					expectedData, err := os.ReadFile(filepath.Join("testdata", "tls", tc.ExpectedDataFile))
					require.NoError(t, err)
					require.Equal(t, expectedData, result, "filtered certificate data should match expected value")
				}
			} else {
				require.EqualError(t, err, tc.ExpectedErr.Error())
				require.Nil(t, result)
			}
		})
	}
}
