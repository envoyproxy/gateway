// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1" // nolint:gosec
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

const (
	// DefaultEnvoyGatewayDNSPrefix defines the default Envoy Gateway DNS prefix.
	DefaultEnvoyGatewayDNSPrefix = config.EnvoyGatewayServiceName

	// DefaultEnvoyDNSPrefix defines the default Envoy DNS prefix.
	DefaultEnvoyDNSPrefix = "*"

	// DefaultCertificateLifetime holds the default certificate lifetime (in days).
	DefaultCertificateLifetime = 365 * 5

	// keySize sets the RSA key size to 2048 bits. This is minimum recommended size
	// for RSA keys.
	keySize = 2048
)

// Configuration holds config parameters used for generating certificates.
type Configuration struct {
	// Provider defines the desired cert provider and provider-specific
	// configuration.
	Provider *CertProvider
}

func (c *Configuration) getProvider() {
	if c.Provider == nil {
		c.Provider = &CertProvider{
			Type: ProviderTypeEnvoyGateway,
		}
	}
}

// CertProvider defines the provider of certificates.
type CertProvider struct {
	// Type is the type of provider to use for managing certificates.
	Type ProviderType `json:"type"`
}

// ProviderType defines the types of supported certificate providers.
type ProviderType string

const (
	// ProviderTypeEnvoyGateway defines the "EnvoyGateway" provider.
	// EnvoyGateway implements a self-signed CA and generates server
	// certs for Envoy Gateway and Envoy.
	ProviderTypeEnvoyGateway ProviderType = "EnvoyGateway"
)

// Certificates contains a set of Certificates as []byte each holding
// the CA Cert along with Envoy Gateway & Envoy certificates.
type Certificates struct {
	CACertificate             []byte
	EnvoyGatewayCertificate   []byte
	EnvoyGatewayPrivateKey    []byte
	EnvoyCertificate          []byte
	EnvoyPrivateKey           []byte
	EnvoyRateLimitCertificate []byte
	EnvoyRateLimitPrivateKey  []byte
	OIDCHMACSecret            []byte
}

// certificateRequest defines a certificate request.
type certificateRequest struct {
	caCertPEM  []byte
	caKeyPEM   []byte
	expiry     time.Time
	commonName string
	altNames   []string
}

// GenerateCerts generates a CA Certificate along with certificates for Envoy Gateway
// and Envoy returning them as a *Certificates struct or error if encountered.
func GenerateCerts(cfg *config.Server) (*Certificates, error) {
	certCfg := new(Configuration)

	certCfg.getProvider()
	switch certCfg.Provider.Type {
	case ProviderTypeEnvoyGateway:
		now := time.Now()
		expiry := now.Add(24 * time.Duration(DefaultCertificateLifetime) * time.Hour)
		caCertPEM, caKeyPEM, err := newCA(DefaultEnvoyGatewayDNSPrefix, expiry)
		if err != nil {
			return nil, err
		}

		var egDNSNames, envoyDNSNames []string
		egProvider := cfg.EnvoyGateway.GetEnvoyGatewayProvider().Type
		switch egProvider {
		case v1alpha1.ProviderTypeKubernetes:
			egDNSNames = kubeServiceNames(DefaultEnvoyGatewayDNSPrefix, cfg.Namespace, cfg.DNSDomain)
			envoyDNSNames = append(envoyDNSNames, fmt.Sprintf("*.%s", cfg.Namespace))
		default:
			// Kubernetes is the only supported Envoy Gateway provider.
			return nil, fmt.Errorf("unsupported provider type %v", egProvider)
		}

		egCertReq := &certificateRequest{
			caCertPEM:  caCertPEM,
			caKeyPEM:   caKeyPEM,
			expiry:     expiry,
			commonName: DefaultEnvoyGatewayDNSPrefix,
			altNames:   egDNSNames,
		}

		egCert, egKey, err := newCert(egCertReq)
		if err != nil {
			return nil, err
		}

		envoyCertReq := &certificateRequest{
			caCertPEM:  caCertPEM,
			caKeyPEM:   caKeyPEM,
			expiry:     expiry,
			commonName: DefaultEnvoyDNSPrefix,
			altNames:   envoyDNSNames,
		}

		envoyCert, envoyKey, err := newCert(envoyCertReq)
		if err != nil {
			return nil, err
		}

		envoyRateLimitCertReq := &certificateRequest{
			caCertPEM:  caCertPEM,
			caKeyPEM:   caKeyPEM,
			expiry:     expiry,
			commonName: DefaultEnvoyDNSPrefix,
			altNames:   envoyDNSNames,
		}

		envoyRateLimitCert, envoyRateLimitKey, err := newCert(envoyRateLimitCertReq)
		if err != nil {
			return nil, err
		}

		oidcHMACSecret, err := generateHMACSecret()
		if err != nil {
			return nil, err
		}

		return &Certificates{
			CACertificate:             caCertPEM,
			EnvoyGatewayCertificate:   egCert,
			EnvoyGatewayPrivateKey:    egKey,
			EnvoyCertificate:          envoyCert,
			EnvoyPrivateKey:           envoyKey,
			EnvoyRateLimitCertificate: envoyRateLimitCert,
			EnvoyRateLimitPrivateKey:  envoyRateLimitKey,
			OIDCHMACSecret:            oidcHMACSecret,
		}, nil
	default:
		// Envoy Gateway, e.g. self-signed CA, is the only supported certificate provider.
		return nil, fmt.Errorf("unsupported certificate provider type %v", certCfg.Provider.Type)
	}
}

// newCert generates a new keypair based on the given the request.
// The return values are cert, key, err.
func newCert(request *certificateRequest) ([]byte, []byte, error) {
	caKeyPair, err := tls.X509KeyPair(request.caCertPEM, request.caKeyPEM)
	if err != nil {
		return nil, nil, err
	}
	caCert, err := x509.ParseCertificate(caKeyPair.Certificate[0])
	if err != nil {
		return nil, nil, err
	}
	caKey, ok := caKeyPair.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("CA private key has unexpected type %T", caKeyPair.PrivateKey)
	}

	newKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot generate key: %w", err)
	}

	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: newSerial(now),
		Subject: pkix.Name{
			CommonName: request.commonName,
		},
		NotBefore:    now.UTC().AddDate(0, 0, -1),
		NotAfter:     request.expiry.UTC(),
		SubjectKeyId: bigIntHash(newKey.N),
		KeyUsage: x509.KeyUsageDigitalSignature |
			x509.KeyUsageDataEncipherment |
			x509.KeyUsageKeyEncipherment |
			x509.KeyUsageContentCommitment,
		DNSNames: request.altNames,
	}
	newCert, err := x509.CreateCertificate(rand.Reader, template, caCert, &newKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}

	newKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(newKey),
	})
	newCertPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: newCert,
	})
	return newCertPEM, newKeyPEM, nil

}

// newCA generates a new CA, given the CA's CN and an expiry time.
// The return order is cacert, cakey, error.
func newCA(cn string, expiry time.Time) ([]byte, []byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now()
	serial := newSerial(now)
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   cn,
			SerialNumber: serial.String(),
		},
		NotBefore:             now.UTC().AddDate(0, 0, -1),
		NotAfter:              expiry.UTC(),
		SubjectKeyId:          bigIntHash(key.N),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}
	certPEMData := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})
	keyPEMData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	return certPEMData, keyPEMData, nil
}

func newSerial(now time.Time) *big.Int {
	return big.NewInt(int64(now.Nanosecond()))
}

// bigIntHash generates a SubjectKeyId by hashing the modulus of the private
// key. This isn't one of the methods listed in RFC 5280 4.2.1.2, but that also
// notes that other methods are acceptable.
//
// gosec makes a blanket claim that SHA-1 is unacceptable, which is
// false here. The core Go method of generations the SubjectKeyId (see
// https://github.com/golang/go/issues/26676) also uses SHA-1, as recommended
// by RFC 5280.
func bigIntHash(n *big.Int) []byte {
	h := sha1.New()    // nolint:gosec
	h.Write(n.Bytes()) // nolint:errcheck
	return h.Sum(nil)
}

func kubeServiceNames(service, namespace, dnsName string) []string {
	return []string{
		service,
		fmt.Sprintf("%s.%s", service, namespace),
		fmt.Sprintf("%s.%s.svc", service, namespace),
		fmt.Sprintf("%s.%s.svc.%s", service, namespace, dnsName),
	}
}

func generateHMACSecret() ([]byte, error) {
	// Set the desired length of the secret key in bytes
	keyLength := 32

	// Create a byte slice to hold the random bytes
	key := make([]byte, keyLength)

	// Read random bytes from the cryptographically secure random number generator
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hmack secret key: %w", err)
	}

	return key, nil
}
