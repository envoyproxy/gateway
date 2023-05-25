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
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

// validateTLSSecretData ensures the cert and key provided in a secret
// is not malformed and can be properly parsed
func validateTLSSecretsData(secrets []*corev1.Secret, host *v1beta1.Hostname) error {
	var publicKeyAlgorithm string
	var parseErr error

	pkaSecretSet := make(map[string]string)
	for _, secret := range secrets {
		certData := secret.Data[corev1.TLSCertKey]

		certBlock, _ := pem.Decode(certData)
		if certBlock == nil {
			return fmt.Errorf("unable to decode pem data in %s", corev1.TLSCertKey)
		}

		cert, err := x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			return fmt.Errorf("unable to parse certificate in %s: %w", corev1.TLSCertKey, err)
		}
		publicKeyAlgorithm = cert.PublicKeyAlgorithm.String()

		keyData := secret.Data[corev1.TLSPrivateKeyKey]

		keyBlock, _ := pem.Decode(keyData)
		if keyBlock == nil {
			return fmt.Errorf("unable to decode pem data in %s", corev1.TLSPrivateKeyKey)
		}

		matchedFQDN, err := verifyHostname(cert, host)
		if err != nil {
			return fmt.Errorf("hostname %s does not match Common Name or DNS Names in the certificate %s", string(*host), corev1.TLSCertKey)
		}
		pkaSecretKey := fmt.Sprintf("%s/%s", publicKeyAlgorithm, matchedFQDN)

		// Check whether the public key algorithm and matched certificate FQDN in the referenced secrets are unique.
		if certFQDN, ok := pkaSecretSet[pkaSecretKey]; ok {
			return fmt.Errorf("secret %s/%s public key algorithm must be unique. Cerificate FQDN %s has a conficting algorithm [%s]",
				secret.Namespace, secret.Name, certFQDN, publicKeyAlgorithm)

		}
		pkaSecretSet[pkaSecretKey] = getCertFQDN(cert)

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
			return fmt.Errorf("%s key format found in %s, supported formats are PKCS1, PKCS8 or EC", keyBlock.Type, corev1.TLSPrivateKeyKey)
		}
	}

	return parseErr
}

func getCertFQDN(cert *x509.Certificate) string {
	fqdn := ""
	if len(cert.DNSNames) > 0 {
		for _, name := range cert.DNSNames {
			fqdn += name + ","
		}
	} else {
		fqdn = cert.Subject.CommonName
	}
	return fqdn
}

// verifyHostname checks if the listener Hostname matches any domain in the certificate, returns a list of matched hosts.
func verifyHostname(cert *x509.Certificate, host *v1beta1.Hostname) ([]string, error) {
	var matchedHosts []string

	if len(cert.DNSNames) > 0 {
		matchedHosts = computeHosts(cert.DNSNames, host)
	} else {
		matchedHosts = computeHosts([]string{cert.Subject.CommonName}, host)
	}

	if len(matchedHosts) > 0 {
		return matchedHosts, nil
	}

	return nil, x509.HostnameError{Certificate: cert, Host: string(*host)}
}
