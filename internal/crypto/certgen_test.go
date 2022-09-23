package crypto

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
)

func TestGenerateCerts(t *testing.T) {
	type testcase struct {
		envoyGateway            *v1alpha1.EnvoyGateway
		certConfig              *Configuration
		wantEnvoyGatewayDNSName string
		wantEnvoyDNSName        string
	}

	run := func(t *testing.T, name string, tc testcase) {
		t.Helper()

		t.Run(name, func(t *testing.T) {
			t.Helper()

			got, err := GenerateCerts(tc.envoyGateway)
			require.NoError(t, err)

			roots := x509.NewCertPool()
			ok := roots.AppendCertsFromPEM(got.CACertificate)
			require.Truef(t, ok, "Failed to set up CA cert for testing, maybe it's an invalid PEM")

			currentTime := time.Now()

			err = verifyCert(got.EnvoyGatewayCertificate, roots, tc.wantEnvoyGatewayDNSName, currentTime)
			assert.NoErrorf(t, err, "Validating %s failed", name)

			err = verifyCert(got.EnvoyCertificate, roots, tc.wantEnvoyDNSName, currentTime)
			assert.NoErrorf(t, err, "Validating %s failed", name)
		})
	}

	run(t, "no configuration - use defaults", testcase{
		certConfig:              &Configuration{},
		wantEnvoyGatewayDNSName: "envoy-gateway",
		wantEnvoyDNSName:        "*.envoy-gateway-system",
	})
}

func TestGeneratedValidKubeCerts(t *testing.T) {
	now := time.Now()
	expiry := now.Add(24 * 365 * time.Hour)

	caCert, caKey, err := newCA("envoy-gateway", expiry)
	require.NoErrorf(t, err, "Failed to generate CA cert")

	egCertReq := &certificateRequest{
		caCertPEM:  caCert,
		caKeyPEM:   caKey,
		expiry:     expiry,
		commonName: "envoy-gateway",
		altNames:   kubeServiceNames("envoy-gateway", "envoy-gateway-system", "cluster.local"),
	}
	egCert, _, err := newCert(egCertReq)
	require.NoErrorf(t, err, "Failed to generate Envoy Gateway cert")

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(caCert)
	require.Truef(t, ok, "Failed to set up CA cert for testing, maybe it's an invalid PEM")

	envoyCertReq := &certificateRequest{
		caCertPEM:  caCert,
		caKeyPEM:   caKey,
		expiry:     expiry,
		commonName: "envoy",
		altNames:   kubeServiceNames("envoy", "envoy-gateway-system", "cluster.local"),
	}
	envoyCert, _, err := newCert(envoyCertReq)
	require.NoErrorf(t, err, "Failed to generate Envoy cert")

	tests := []struct {
		name    string
		cert    []byte
		dnsName string
	}{
		{
			name:    "envoy gateway cert",
			cert:    egCert,
			dnsName: "envoy-gateway",
		},
		{
			name:    "envoy cert",
			cert:    envoyCert,
			dnsName: "envoy",
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			err := verifyCert(tc.cert, roots, tc.dnsName, now)
			assert.NoErrorf(t, err, "Validating %s failed", tc.name)
		})
	}

}

func verifyCert(certPEM []byte, roots *x509.CertPool, dnsname string, currentTime time.Time) error {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return fmt.Errorf("failed to decode %s certificate from PEM form", dnsname)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}

	opts := x509.VerifyOptions{
		DNSName:     dnsname,
		Roots:       roots,
		CurrentTime: currentTime,
	}
	if _, err = cert.Verify(opts); err != nil {
		return fmt.Errorf("certificate verification failed: %s", err)
	}

	return nil
}
