// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package crypto

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// LoadTLSConfig returns TLSConfig form certificates.
func LoadTLSConfig(tlsCrt, tlsKey, caCrt string) (*tls.Config, error) {
	loadConfig := func() (*tls.Config, error) {
		cert, err := tls.LoadX509KeyPair(tlsCrt, tlsKey)
		if err != nil {
			return nil, err
		}

		// Load the CA cert.
		ca, err := os.ReadFile(caCrt)
		if err != nil {
			return nil, err
		}

		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(ca) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}

		return &tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"h2"},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    certPool,
			MinVersion:   tls.VersionTLS13,
		}, nil
	}

	// Attempt to load certificates and key to catch configuration errors early.
	if _, err := loadConfig(); err != nil {
		return nil, err
	}

	return &tls.Config{
		MinVersion: tls.VersionTLS13,
		ClientAuth: tls.RequireAndVerifyClientCert,
		Rand:       rand.Reader,
		GetConfigForClient: func(*tls.ClientHelloInfo) (*tls.Config, error) {
			return loadConfig()
		},
	}, nil
}
