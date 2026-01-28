// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"k8s.io/apimachinery/pkg/util/sets"
)

var validCipherSuites = sets.New(
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
)

// parseCertsFromTLSSecretsData parses the cert and key provided in a secret
// and ensures that they are not malformed and can be properly parsed.
// this function returns a list of valid secrets and certificates.
func parseCertsFromTLSSecretsData(secrets []*corev1.Secret) ([]*corev1.Secret, []*x509.Certificate, status.ListenerError) {
	var (
		publicKeyAlgorithm string
		errs               []error
	)

	validSecrets := make([]*corev1.Secret, 0, len(secrets))
	certs := make([]*x509.Certificate, 0, len(secrets))

	pkaSecretSet := make(map[string]string)
	for _, secret := range secrets {
		certData := secret.Data[corev1.TLSCertKey]

		validData, listenerErr := filterValidCertificates(certData)
		if listenerErr != nil {
			if listenerErr.Reason() == gwapiv1.ListenerReasonInvalidCertificateRef {
				errs = append(errs, fmt.Errorf("%s/%s must contain valid tls.crt and tls.key, unable to validate certificate in tls.crt: %s",
					secret.Namespace, secret.Name, listenerErr.Error()))
				continue
			} else if listenerErr.Reason() == status.ListenerReasonPartiallyInvalidCertificateRef {
				errs = append(errs, fmt.Errorf("%s/%s has some invalid certificates: %s",
					secret.Namespace, secret.Name, listenerErr.Error()))
			}
		}

		certBlock, _ := pem.Decode(validData)
		if certBlock == nil {
			errs = append(errs, fmt.Errorf("%s/%s must contain valid %s and %s, unable to decode pem data in %s",
				secret.Namespace, secret.Name, corev1.TLSCertKey, corev1.TLSPrivateKeyKey, corev1.TLSCertKey))
			continue
		}

		cert, err := x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s/%s must contain valid %s and %s, unable to parse certificate in %s: %w",
				secret.Namespace, secret.Name, corev1.TLSCertKey, corev1.TLSPrivateKeyKey, corev1.TLSCertKey, err))
			continue
		}
		publicKeyAlgorithm = cert.PublicKeyAlgorithm.String()

		keyData := secret.Data[corev1.TLSPrivateKeyKey]

		keyBlock, _ := pem.Decode(keyData)
		if keyBlock == nil {
			errs = append(errs, fmt.Errorf("%s/%s must contain valid %s and %s, unable to decode pem data in %s",
				secret.Namespace, secret.Name, corev1.TLSCertKey, corev1.TLSPrivateKeyKey, corev1.TLSPrivateKeyKey))
			continue
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
		hasConflictDomainAlgorithm := false
		for _, domain := range certDomains {
			pkaSecretKey := fmt.Sprintf("%s/%s", publicKeyAlgorithm, domain)

			// Check whether the public key algorithm and certificate domain are unique
			if _, ok := pkaSecretSet[pkaSecretKey]; ok {
				errs = append(errs, fmt.Errorf("%s/%s public key algorithm must be unique, certificate domain %s has a conflicting algorithm [%s]",
					secret.Namespace, secret.Name, domain, publicKeyAlgorithm))
				hasConflictDomainAlgorithm = true
				break
			}
			pkaSecretSet[pkaSecretKey] = domain
		}
		if hasConflictDomainAlgorithm {
			continue
		}

		switch keyBlock.Type {
		case "PRIVATE KEY":
			_, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s/%s must contain valid %s and %s, unable to parse PKCS8 formatted private key in %s",
					secret.Namespace, secret.Name, corev1.TLSCertKey, corev1.TLSPrivateKeyKey, corev1.TLSPrivateKeyKey))
				continue
			}
		case "RSA PRIVATE KEY":
			_, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s/%s must contain valid %s and %s, unable to parse PKCS1 formatted private key in %s",
					secret.Namespace, secret.Name, corev1.TLSCertKey, corev1.TLSPrivateKeyKey, corev1.TLSPrivateKeyKey))
				continue
			}
		case "EC PRIVATE KEY":
			_, err := x509.ParseECPrivateKey(keyBlock.Bytes)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s/%s must contain valid %s and %s, unable to parse EC formatted private key in %s",
					secret.Namespace, secret.Name, corev1.TLSCertKey, corev1.TLSPrivateKeyKey, corev1.TLSPrivateKeyKey))
				continue
			}
		default:
			errs = append(errs, fmt.Errorf("%s/%s must contain valid %s and %s, %s key format found in %s, supported formats are PKCS1, PKCS8 or EC",
				secret.Namespace, secret.Name, corev1.TLSCertKey, corev1.TLSPrivateKeyKey, keyBlock.Type, corev1.TLSPrivateKeyKey))
			continue
		}

		validSecrets = append(validSecrets, secret)
		certs = append(certs, cert)
	}

	// If there are validation errors, determine the appropriate listener reason based on whether any valid certificates were found.
	// If no valid certs exist, the listener cannot process traffic normally, so this is treated as a InvalidCertificateRef.
	// If some valid certs exist, this is treated as a PartiallyInvalidCertificateRef to notify cert error to user.
	if len(errs) > 0 {
		if len(certs) == 0 {
			return nil, nil, status.NewListenerStatusError(
				errors.Join(errs...),
				gwapiv1.ListenerReasonInvalidCertificateRef,
			)
		} else {
			return validSecrets, certs, status.NewListenerStatusError(
				errors.Join(errs...),
				status.ListenerReasonPartiallyInvalidCertificateRef,
			)
		}
	}
	return validSecrets, certs, nil
}

// filterValidCertificates filters out expired or not-yet-valid certificates from PEM encoded data.
// It accepts certificate bundles (multiple PEM blocks) and returns only the valid certificates.
// A certificate is considered valid if the current time is within its NotBefore and NotAfter period.
//
// Return a status.ListenerError with InvalidCertificateRef Condition if no valid certificates are found in the provided data,
// Return a status.ListenerError with PartiallyInvalidCertificateRef Condition if some certificates are invalid but also valid certificates exist.
func filterValidCertificates(data []byte) ([]byte, status.ListenerError) {
	if len(data) == 0 {
		return nil, status.NewListenerStatusError(
			fmt.Errorf("no certificate data provided"),
			gwapiv1.ListenerReasonInvalidCertificateRef,
		)
	}

	now := time.Now()
	var errs []error
	validData := make([]byte, 0, len(data))

	// Process each PEM block in the data
	rest := data
	for len(rest) > 0 {
		block, remaining := pem.Decode(rest)
		if block == nil {
			break
		}
		rest = remaining

		// Parse all certificates in this PEM block
		certs, err := x509.ParseCertificates(block.Bytes)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		// Validate all certificates in this PEM block
		blockValid := true
		for _, cert := range certs {
			if now.After(cert.NotAfter) {
				errs = append(errs, fmt.Errorf("certificate %s has expired since %v", cert.Subject.CommonName, cert.NotAfter))
				blockValid = false
				break
			}
			if now.Before(cert.NotBefore) {
				errs = append(errs, fmt.Errorf("certificate %s will be valid after %v", cert.Subject.CommonName, cert.NotBefore))
				blockValid = false
				break
			}
		}
		// Only include this PEM block if all certificates in it are valid
		if blockValid {
			validData = append(validData, pem.EncodeToMemory(block)...)
		}
	}

	if len(validData) == 0 {
		if len(errs) > 0 {
			return nil, status.NewListenerStatusError(
				errors.Join(errs...),
				gwapiv1.ListenerReasonInvalidCertificateRef,
			)
		}
		// No errors but no valid PEM blocks found - PEM decoding failed
		return nil, status.NewListenerStatusError(
			fmt.Errorf("unable to decode pem data for certificate"),
			gwapiv1.ListenerReasonInvalidCertificateRef,
		)
	}
	if len(errs) > 0 {
		return validData, status.NewListenerStatusError(
			errors.Join(errs...),
			status.ListenerReasonPartiallyInvalidCertificateRef,
		)
	}
	return validData, nil
}

// validateCrl validates a CRL in PEM encoded data.
func validateCrl(data []byte) error {
	block, _ := pem.Decode(data)
	if block == nil {
		return fmt.Errorf("unable to decode pem data for CRL")
	}
	crl, err := x509.ParseRevocationList(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CRL: %w", err)
	}
	now := time.Now()
	if !crl.NextUpdate.IsZero() && now.After(crl.NextUpdate) {
		return fmt.Errorf("CRL is expired (next update was due at %v)", crl.NextUpdate)
	}
	if now.Before(crl.ThisUpdate) {
		return fmt.Errorf("CRL is not yet valid (this update starts at %v)", crl.ThisUpdate)
	}
	return nil
}

// validateCipherSuites validates the cipher suites provided in the TLS settings.
func validateCipherSuites(ciphers []string) error {
	for _, cipher := range ciphers {
		if !validCipherSuites.Has(cipher) {
			return fmt.Errorf("unsupported cipher suite: %s", cipher)
		}
	}
	return nil
}
