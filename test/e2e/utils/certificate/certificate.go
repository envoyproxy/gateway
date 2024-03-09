// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

// This utility is copied from: https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/utils/kubernetes/certificate.go
// and adapted to support creation of certificates that are signed with a CA

package certificate

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// ensure auth plugins are loaded
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	rsaBits  = 2048
	validFor = 365 * 24 * time.Hour
)

// MustCreateSelfSignedCAConfigmapAndCertSecret creates a self-signed CA certficiate and stores it a configma
// it also creates an SSL certificate and stores it in a secret
func MustCreateSelfSignedCAConfigmapAndCertSecret(t *testing.T, namespace, secretName string, hosts []string) (*corev1.Secret, *corev1.ConfigMap) {
	require.Greater(t, len(hosts), 0, "require a non-empty hosts for Subject Alternate Name values")

	var caCert, serverKey, serverCert bytes.Buffer

	require.NoError(t, generateCertAndCA(hosts, &caCert, &serverKey, &serverCert), "failed to generate RSA certificate")

	secretData := map[string][]byte{
		corev1.TLSCertKey:       serverCert.Bytes(),
		corev1.TLSPrivateKeyKey: serverKey.Bytes(),
	}

	newSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      secretName,
		},
		Type: corev1.SecretTypeTLS,
		Data: secretData,
	}

	configmapData := map[string]string{
		"ca.crt": caCert.String(),
	}

	newConfigmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      secretName,
		},
		Data: configmapData,
	}

	return newSecret, newConfigmap
}

// generateCertAndCA generates a basic self signed certificate valid for a year
func generateCertAndCA(hosts []string, caCertOut, keyOut, certOut io.Writer) error {
	// first create CA certificate
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Create a template for the CA certificate
	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:         "Envoy Gateway CA",
			OrganizationalUnit: []string{"Gateway"},
			Organization:       []string{"EnvoyProxy"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0), // Valid for 1 year
		BasicConstraintsValid: true,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}

	// Create a self-signed CA certificate using the private key and template
	caDERBytes, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return err
	}

	caCert, err := x509.ParseCertificate(caDERBytes)
	if err != nil {
		return err
	}

	if err := pem.Encode(caCertOut, &pem.Block{Type: "CERTIFICATE", Bytes: caDERBytes}); err != nil {
		return fmt.Errorf("failed creating cert: %w", err)
	}

	// now create leaf certificate signed with CA's private key
	leafPrivateKey, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "default",
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, caCert, &leafPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed creating cert: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(leafPrivateKey)}); err != nil {
		return fmt.Errorf("failed creating key: %w", err)
	}

	return nil
}
