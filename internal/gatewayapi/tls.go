// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

// validateTLSSecretData ensures the cert and key provided in a secret
// is not malformed and can be properly parsed. It also returns the public key
// encryption algorithm used to encrypt secret information.
func validateTLSSecretData(secret *corev1.Secret) (string, error) {
	var publicKeyAlgorithm string
	certData := secret.Data[corev1.TLSCertKey]

	certBlock, _ := pem.Decode(certData)
	if certBlock == nil {
		return publicKeyAlgorithm, fmt.Errorf("unable to decode pem data in %s", corev1.TLSCertKey)
	}

	certificate, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return publicKeyAlgorithm, fmt.Errorf("unable to parse certificate in %s: %w", corev1.TLSCertKey, err)
	}

	publicKeyAlgorithm = certificate.PublicKeyAlgorithm.String()

	keyData := secret.Data[corev1.TLSPrivateKeyKey]

	keyBlock, _ := pem.Decode(keyData)
	if keyBlock == nil {
		return publicKeyAlgorithm, fmt.Errorf("unable to decode pem data in %s", corev1.TLSPrivateKeyKey)
	}

	var parseErr error

	switch keyBlock.Type {
	case "PRIVATE KEY":
		_, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
		if err != nil {
			parseErr = fmt.Errorf("unable to parse PKCS8 formatted private key in %s", corev1.TLSPrivateKeyKey)
		}
	case "RSA PRIVATE KEY":
		_, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		if err != nil {
			parseErr = fmt.Errorf("unable to parse PKCS1 formatted private key in %s", corev1.TLSPrivateKeyKey)
		}
	case "EC PRIVATE KEY":
		_, err := x509.ParseECPrivateKey(keyBlock.Bytes)
		if err != nil {
			parseErr = fmt.Errorf("unable to parse EC formatted private key in %s", corev1.TLSPrivateKeyKey)
		}
	default:
		return publicKeyAlgorithm, fmt.Errorf("%s key format found in %s, supported formats are PKCS1, PKCS8 or EC", keyBlock.Type, corev1.TLSPrivateKeyKey)
	}

	return publicKeyAlgorithm, parseErr
}
