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
	v1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	secretName      = "secret"
	secretNamespace = "test"
)

// createTestSecret creates a K8s tls secret using testdata
// see for more info <https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets>
func createTestSecrets(t *testing.T, certFile, keyFile string) []*corev1.Secret {
	t.Helper()

	certData, err := os.ReadFile(filepath.Join("testdata", "tls", certFile))
	require.NoError(t, err)

	keyData, err := os.ReadFile(filepath.Join("testdata", "tls", keyFile))
	require.NoError(t, err)

	return []*corev1.Secret{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: secretNamespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       certData,
			corev1.TLSPrivateKeyKey: keyData,
		},
	}}
}

/*
TestValidateTLSSecretData ensures that we can properly validate the contents of a K8s tls secret.
The test assumes the secret is valid and was able to be applied to a cluster.

The following commands were used to generate test key/cert pairs
using openssl (LibreSSL 3.3.6)

# RSA

	openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout rsa-pkcs8.key -out rsa-cert.pem -subj "/CN=foo.bar.com"`
	openssl rsa -in rsa-pkcs8.key -out rsa-pkcs1.key

# RSA with SAN extension

	openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout rsa-pkcs8-san.key -out rsa-cert-san.pem -subj "/CN=Test Inc" -addext "subjectAltName = DNS:foo.bar.com"
	openssl rsa -in rsa-pkcs8-san.key -out rsa-pkcs1-san.key

# RSA with wildcard SAN domain

	openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout rsa-pkcs8-wildcard.key -out rsa-cert-wildcard.pem -subj "/CN=Test Inc" -addext "subjectAltName = DNS:*, DNS:*.example.com"
	openssl rsa -in rsa-pkcs8-wildcard.key -out rsa-pkcs1-wildcard.key

# ECDSA-p256

	openssl ecparam -name prime256v1 -genkey -noout -out ecdsa-p256.key
	openssl req -new -x509 -days 365 -key ecdsa-p256.key -out ecdsa-p256-cert.pem -subj "/CN=foo.bar.com"

# ECDSA-p384

	openssl ecparam -name secp384r1 -genkey -noout -out ecdsa-p384.key
	openssl req -new -x509 -days 365 -key ecdsa-p384.key -out ecdsa-p384-cert.pem -subj "/CN=foo.bar.com"
*/
func TestValidateTLSSecretsData(t *testing.T) {
	type testCase struct {
		Name        string
		CertFile    string
		KeyFile     string
		Domain      v1.Hostname
		ExpectedErr error
	}

	testCases := []testCase{
		{
			Name:        "valid-rsa-pkcs1",
			CertFile:    "rsa-cert.pem",
			KeyFile:     "rsa-pkcs1.key",
			Domain:      "*",
			ExpectedErr: nil,
		},
		{
			Name:        "valid-rsa-pkcs8",
			CertFile:    "rsa-cert.pem",
			KeyFile:     "rsa-pkcs8.key",
			Domain:      "*",
			ExpectedErr: nil,
		},
		{
			Name:        "valid-rsa-san-domain",
			CertFile:    "rsa-cert-san.pem",
			KeyFile:     "rsa-pkcs8-san.key",
			Domain:      "foo.bar.com",
			ExpectedErr: nil,
		},
		{
			Name:        "valid-rsa-wildcard-domain",
			CertFile:    "rsa-cert-wildcard.pem",
			KeyFile:     "rsa-pkcs1-wildcard.key",
			Domain:      "foo.bar.com",
			ExpectedErr: nil,
		},
		{
			Name:        "valid-ecdsa-p256",
			CertFile:    "ecdsa-p256-cert.pem",
			KeyFile:     "ecdsa-p256.key",
			Domain:      "*",
			ExpectedErr: nil,
		},
		{
			Name:        "valid-ecdsa-p384",
			CertFile:    "ecdsa-p384-cert.pem",
			KeyFile:     "ecdsa-p384.key",
			Domain:      "*",
			ExpectedErr: nil,
		},
		{
			Name:        "malformed-cert-pem-encoding",
			CertFile:    "malformed-encoding.pem",
			KeyFile:     "rsa-pkcs8.key",
			Domain:      "*",
			ExpectedErr: errors.New("test/secret must contain valid tls.crt and tls.key, unable to decode pem data in tls.crt"),
		},
		{
			Name:        "malformed-key-pem-encoding",
			CertFile:    "rsa-cert.pem",
			KeyFile:     "malformed-encoding.pem",
			Domain:      "*",
			ExpectedErr: errors.New("test/secret must contain valid tls.crt and tls.key, unable to decode pem data in tls.key"),
		},
		{
			Name:        "malformed-cert",
			CertFile:    "malformed-cert.pem",
			KeyFile:     "rsa-pkcs8.key",
			Domain:      "*",
			ExpectedErr: errors.New("test/secret must contain valid tls.crt and tls.key, unable to parse certificate in tls.crt: x509: malformed certificate"),
		},
		{
			Name:        "malformed-pkcs8-key",
			CertFile:    "rsa-cert.pem",
			KeyFile:     "malformed-pkcs8.key",
			Domain:      "*",
			ExpectedErr: errors.New("test/secret must contain valid tls.crt and tls.key, unable to parse PKCS8 formatted private key in tls.key"),
		},
		{
			Name:        "malformed-pkcs1-key",
			CertFile:    "rsa-cert.pem",
			KeyFile:     "malformed-pkcs1.key",
			Domain:      "*",
			ExpectedErr: errors.New("test/secret must contain valid tls.crt and tls.key, unable to parse PKCS1 formatted private key in tls.key"),
		},
		{
			Name:        "malformed-ecdsa-key",
			CertFile:    "rsa-cert.pem",
			KeyFile:     "malformed-ecdsa.key",
			Domain:      "*",
			ExpectedErr: errors.New("test/secret must contain valid tls.crt and tls.key, unable to parse EC formatted private key in tls.key"),
		},
		{
			Name:        "invalid-key-type",
			CertFile:    "rsa-cert.pem",
			KeyFile:     "invalid-key-type.key",
			Domain:      "*",
			ExpectedErr: errors.New("test/secret must contain valid tls.crt and tls.key, FOO key format found in tls.key, supported formats are PKCS1, PKCS8 or EC"),
		},
		{
			Name:        "invalid-domain-cert",
			CertFile:    "rsa-cert-san.pem",
			KeyFile:     "rsa-pkcs8-san.key",
			Domain:      "*.example.com",
			ExpectedErr: errors.New("test/secret must contain valid tls.crt and tls.key, hostname *.example.com does not match Common Name or DNS Names in the certificate tls.crt"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			secrets := createTestSecrets(t, tc.CertFile, tc.KeyFile)
			require.NotNil(t, secrets)
			err := validateTLSSecretsData(secrets, &tc.Domain)
			if tc.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.ExpectedErr.Error())
			}
		})
	}
}
