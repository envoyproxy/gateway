// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
)

// validateTLSSecretData ensures the cert and key provided in a secret
// is not malformed and can be properly parsed
func validateTLSSecretsData(secrets []*corev1.Secret) ([]*x509.Certificate, error) {
	var publicKeyAlgorithm string
	var parseErr error
	certs := make([]*x509.Certificate, 0, len(secrets))

	pkaSecretSet := make(map[string]string)
	for _, secret := range secrets {
		certData := secret.Data[corev1.TLSCertKey]

		if err := validateCertificate(certData); err != nil {
			return nil, fmt.Errorf("%s/%s must contain valid %s and %s, unable to validate certificate in %s: %w", secret.Namespace, secret.Name, corev1.TLSCertKey, corev1.TLSPrivateKeyKey, corev1.TLSCertKey, err)
		}

		certBlock, _ := pem.Decode(certData)
		if certBlock == nil {
			return nil, fmt.Errorf("%s/%s must contain valid %s and %s, unable to decode pem data in %s", secret.Namespace, secret.Name, corev1.TLSCertKey, corev1.TLSPrivateKeyKey, corev1.TLSCertKey)
		}

		cert, err := x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("%s/%s must contain valid %s and %s, unable to parse certificate in %s: %w", secret.Namespace, secret.Name, corev1.TLSCertKey, corev1.TLSPrivateKeyKey, corev1.TLSCertKey, err)
		}
		publicKeyAlgorithm = cert.PublicKeyAlgorithm.String()

		keyData := secret.Data[corev1.TLSPrivateKeyKey]

		keyBlock, _ := pem.Decode(keyData)
		if keyBlock == nil {
			return nil, fmt.Errorf("%s/%s must contain valid %s and %s, unable to decode pem data in %s", secret.Namespace, secret.Name, corev1.TLSCertKey, corev1.TLSPrivateKeyKey, corev1.TLSPrivateKeyKey)
		}

		// SNI and SAN/Cert Domain mismatch is allowed
		// Consider converting this into a warning once
		// https://github.com/envoyproxy/gateway/issues/6717 is in

		// Extract certificate domains (SANs or CN) for uniqueness checking
		var certDomains []string
		if len(cert.DNSNames) > 0 {
			certDomains = cert.DNSNames
		} else if cert.Subject.CommonName != "" {
			certDomains = []string{cert.Subject.CommonName}
		}

		// Check uniqueness for each domain in the certificate with this algorithm
		for _, domain := range certDomains {
			pkaSecretKey := fmt.Sprintf("%s/%s", publicKeyAlgorithm, domain)

			// Check whether the public key algorithm and certificate domain are unique
			if _, ok := pkaSecretSet[pkaSecretKey]; ok {
				return nil, fmt.Errorf("%s/%s public key algorithm must be unique, certificate domain %s has a conflicting algorithm [%s]",
					secret.Namespace, secret.Name, domain, publicKeyAlgorithm)
			}
			pkaSecretSet[pkaSecretKey] = domain
		}

		switch keyBlock.Type {
		case "PRIVATE KEY":
			_, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
			if err != nil {
				parseErr = fmt.Errorf("%s/%s must contain valid %s and %s, unable to parse PKCS8 formatted private key in %s", secret.Namespace, secret.Name, corev1.TLSCertKey, corev1.TLSPrivateKeyKey, corev1.TLSPrivateKeyKey)
			}
		case "RSA PRIVATE KEY":
			_, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
			if err != nil {
				parseErr = fmt.Errorf("%s/%s must contain valid %s and %s, unable to parse PKCS1 formatted private key in %s", secret.Namespace, secret.Name, corev1.TLSCertKey, corev1.TLSPrivateKeyKey, corev1.TLSPrivateKeyKey)
			}
		case "EC PRIVATE KEY":
			_, err := x509.ParseECPrivateKey(keyBlock.Bytes)
			if err != nil {
				parseErr = fmt.Errorf("%s/%s must contain valid %s and %s, unable to parse EC formatted private key in %s", secret.Namespace, secret.Name, corev1.TLSCertKey, corev1.TLSPrivateKeyKey, corev1.TLSPrivateKeyKey)
			}
		default:
			return nil, fmt.Errorf("%s/%s must contain valid %s and %s, %s key format found in %s, supported formats are PKCS1, PKCS8 or EC", secret.Namespace, secret.Name, corev1.TLSCertKey, corev1.TLSPrivateKeyKey, keyBlock.Type, corev1.TLSPrivateKeyKey)
		}
		certs = append(certs, cert)
	}

	return certs, parseErr
}

func validateCertificate(data []byte) error {
	block, _ := pem.Decode(data)
	if block == nil {
		return fmt.Errorf("pem decode failed")
	}
	certs, err := x509.ParseCertificates(block.Bytes)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, cert := range certs {
		if now.After(cert.NotAfter) {
			return fmt.Errorf("certificate is expired")
		}
		if now.Before(cert.NotBefore) {
			return fmt.Errorf("certificate is not yet valid")
		}
	}
	return nil
}
