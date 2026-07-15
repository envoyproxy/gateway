// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"maps"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
)

// validCipherSuites contains the list of supported TLS cipher suites.
// The source of truth for these ciphers is the Envoy documentation:
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/transport_sockets/tls/v3/common.proto#extensions-transport-sockets-tls-v3-tlsparameters
//
// Envoy accepts IANA cipher suite names by mapping them to OpenSSL names before
// strict validation. See:
// https://github.com/envoyproxy/envoy/blob/main/compat/openssl/source/iana_2_ossl_names.cc
// https://github.com/envoyproxy/envoy/blob/main/compat/openssl/source/SSL_CTX_set_strict_cipher_list.cc
var validCipherSuites = sets.New(
	"ECDHE-ECDSA-AES128-GCM-SHA256",
	"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
	"ECDHE-RSA-AES128-GCM-SHA256",
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
	"ECDHE-ECDSA-AES256-GCM-SHA384",
	"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
	"ECDHE-RSA-AES256-GCM-SHA384",
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
	"ECDHE-ECDSA-CHACHA20-POLY1305",
	"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
	"ECDHE-RSA-CHACHA20-POLY1305",
	"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
	"ECDHE-ECDSA-AES128-SHA",
	"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
	"ECDHE-RSA-AES128-SHA",
	"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
	"AES128-GCM-SHA256",
	"TLS_RSA_WITH_AES_128_GCM_SHA256",
	"AES128-SHA",
	"TLS_RSA_WITH_AES_128_CBC_SHA",
	"ECDHE-ECDSA-AES256-SHA",
	"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
	"ECDHE-RSA-AES256-SHA",
	"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
	"AES256-GCM-SHA384",
	"TLS_RSA_WITH_AES_256_GCM_SHA384",
	"AES256-SHA",
	"TLS_RSA_WITH_AES_256_CBC_SHA",
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

		// A serving certificate chain is ordered and must stay intact: unlike a CA
		// bundle (see filterValidCACertificates), an expired or malformed member is
		// not dropped but rejects the whole secret, so a corrupted chain never
		// reaches Envoy. The failure is isolated to the referencing listener.
		canonicalChain, listenerErr := validateServingCertificateChain(certData)
		if listenerErr != nil {
			errs = append(errs, fmt.Errorf("%s/%s must contain valid tls.crt and tls.key, unable to validate certificate in tls.crt: %s",
				secret.Namespace, secret.Name, listenerErr.Error()))
			continue
		}

		certBlock, _ := pem.Decode(canonicalChain)
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

		// Ensure the private key actually corresponds to the public key in the
		// certificate. Validating the cert and key independently is not enough:
		// a mismatched pair (e.g. a stale/rotated Secret) parses fine locally but
		// is rejected by BoringSSL at the xDS layer with KEY_VALUES_MISMATCH. With
		// mergeGateways enabled that NACKs the whole SDS push, breaking TLS for all
		// Gateways sharing the proxy. Reject the bad Secret here so it never reaches
		// Envoy and only the affected listener degrades.
		if _, err := tls.X509KeyPair(canonicalChain, pem.EncodeToMemory(keyBlock)); err != nil {
			errs = append(errs, fmt.Errorf("%s/%s must contain a matching tls.crt and tls.key: %w",
				secret.Namespace, secret.Name, err))
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

		normalizedSecret := *secret
		normalizedSecret.Data = maps.Clone(secret.Data)
		normalizedSecret.Data[corev1.TLSCertKey] = canonicalChain
		normalizedSecret.Data[corev1.TLSPrivateKeyKey] = pem.EncodeToMemory(keyBlock)
		validSecrets = append(validSecrets, &normalizedSecret)

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

// validateCertBlock parses the certificate(s) in a single PEM block and returns
// an error if the block is not a valid certificate or if any certificate in it
// is expired or not yet valid (outside its NotBefore/NotAfter window).
func validateCertBlock(block *pem.Block) error {
	certs, err := x509.ParseCertificates(block.Bytes)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, cert := range certs {
		if now.After(cert.NotAfter) {
			return fmt.Errorf("certificate %s has expired since %v", cert.Subject.CommonName, cert.NotAfter)
		}
		if now.Before(cert.NotBefore) {
			return fmt.Errorf("certificate %s will be valid after %v", cert.Subject.CommonName, cert.NotBefore)
		}
	}
	return nil
}

// validateServingCertificateChain validates every certificate in a serving
// certificate chain and returns the canonically re-encoded chain. Unlike a CA
// bundle (see filterValidCACertificates), a serving chain is ordered and must stay
// intact: if ANY certificate (leaf or intermediate) is expired, not yet valid,
// or malformed, the whole secret is rejected. Dropping a member — e.g. an
// expired leaf — would leave the private key matching a certificate that is no
// longer served, which Envoy/BoringSSL rejects as KEY_VALUES_MISMATCH,
// stalling the whole (merged) xDS config (#9225, #9473).
func validateServingCertificateChain(data []byte) ([]byte, status.ListenerError) {
	if len(data) == 0 {
		return nil, status.NewListenerStatusError(
			fmt.Errorf("no certificate data provided"),
			gwapiv1.ListenerReasonInvalidCertificateRef,
		)
	}

	out := make([]byte, 0, len(data))
	rest := data
	for len(rest) > 0 {
		block, remaining := pem.Decode(rest)
		if block == nil {
			break
		}
		rest = remaining

		// A serving chain must stay intact: reject the whole secret on the first
		// invalid member instead of dropping it (cf. filterValidCACertificates).
		if err := validateCertBlock(block); err != nil {
			return nil, status.NewListenerStatusError(err, gwapiv1.ListenerReasonInvalidCertificateRef)
		}
		out = append(out, pem.EncodeToMemory(block)...)
	}

	if len(out) == 0 {
		return nil, status.NewListenerStatusError(
			fmt.Errorf("unable to decode pem data for certificate"),
			gwapiv1.ListenerReasonInvalidCertificateRef,
		)
	}
	return out, nil
}

// filterValidCACertificates filters out expired or not-yet-valid certificates from
// a CA bundle. It accepts CA bundles (multiple independent PEM blocks) and
// returns only the valid certificates; dropping an expired CA from a bundle of
// trust anchors is safe.
// A certificate is considered valid if the current time is within its NotBefore and NotAfter period.
//
// Return a status.ListenerError with InvalidCertificateRef Condition if no valid certificates are found in the provided data,
// Return a status.ListenerError with PartiallyInvalidCertificateRef Condition if some certificates are invalid but also valid certificates exist.
func filterValidCACertificates(data []byte) ([]byte, status.ListenerError) {
	if len(data) == 0 {
		return nil, status.NewListenerStatusError(
			fmt.Errorf("no certificate data provided"),
			gwapiv1.ListenerReasonInvalidCertificateRef,
		)
	}

	var errs []error
	validData := make([]byte, 0, len(data))

	// Process each PEM block; a CA bundle is a set of independent trust anchors,
	// so drop an invalid (malformed/expired) CA and keep the rest.
	rest := data
	for len(rest) > 0 {
		block, remaining := pem.Decode(rest)
		if block == nil {
			break
		}
		rest = remaining

		if err := validateCertBlock(block); err != nil {
			errs = append(errs, err)
			continue
		}
		validData = append(validData, pem.EncodeToMemory(block)...)
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
