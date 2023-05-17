// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"unicode/utf8"

	corev1 "k8s.io/api/core/v1"
)

// validateTLSSecretData ensures the cert and key provided in a secret
// is not malformed and can be properly parsed
func validateTLSSecretData(secret *corev1.Secret, domain string) error {
	certData := secret.Data[corev1.TLSCertKey]

	certBlock, _ := pem.Decode(certData)
	if certBlock == nil {
		return fmt.Errorf("unable to decode pem data in %s", corev1.TLSCertKey)
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("unable to parse certificate in %s: %w", corev1.TLSCertKey, err)
	}

	keyData := secret.Data[corev1.TLSPrivateKeyKey]

	keyBlock, _ := pem.Decode(keyData)
	if keyBlock == nil {
		return fmt.Errorf("unable to decode pem data in %s", corev1.TLSPrivateKeyKey)
	}
	if domain != "" {
		err = verifyHostname(cert, domain)
		if err != nil {
			return fmt.Errorf("hostname %s does not match Common Name or DNS Names in the certificate %s", domain, corev1.TLSCertKey)
		}
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
		return fmt.Errorf("%s key format found in %s, supported formats are PKCS1, PKCS8 or EC", keyBlock.Type, corev1.TLSPrivateKeyKey)
	}

	return parseErr
}

// https://golang.google.cn/doc/go1.15#commonname
func verifyHostname(cert *x509.Certificate, host string) error {
	lowered := toLowerCaseASCII(host)
	if len(cert.DNSNames) > 0 {
		for _, match := range cert.DNSNames {
			if matchHostnames(toLowerCaseASCII(match), lowered) {
				return nil
			}
		}
	} else if matchHostnames(toLowerCaseASCII(cert.Subject.CommonName), lowered) {
		return nil
	}

	return x509.HostnameError{Certificate: cert, Host: host}
}

func matchHostnames(pattern, host string) bool {
	host = strings.TrimSuffix(host, ".")
	pattern = strings.TrimSuffix(pattern, ".")

	if len(pattern) == 0 || len(host) == 0 {
		return false
	}

	patternParts := strings.Split(pattern, ".")
	hostParts := strings.Split(host, ".")

	if len(patternParts) != len(hostParts) {
		return false
	}

	for i, patternPart := range patternParts {
		if i == 0 && patternPart == "*" {
			continue
		}
		if patternPart != hostParts[i] {
			return false
		}
	}

	return true
}

// toLowerCaseASCII returns a lower-case version of in. See RFC 6125 6.4.1. We use
// an explicitly ASCII function to avoid any sharp corners resulting from
// performing Unicode operations on DNS labels.
func toLowerCaseASCII(in string) string {
	// If the string is already lower-case then there's nothing to do.
	isAlreadyLowerCase := true
	for _, c := range in {
		if c == utf8.RuneError {
			// If we get a UTF-8 error then there might be
			// upper-case ASCII bytes in the invalid sequence.
			isAlreadyLowerCase = false
			break
		}
		if 'A' <= c && c <= 'Z' {
			isAlreadyLowerCase = false
			break
		}
	}

	if isAlreadyLowerCase {
		return in
	}

	out := []byte(in)
	for i, c := range out {
		if 'A' <= c && c <= 'Z' {
			out[i] += 'a' - 'A'
		}
	}
	return string(out)
}
