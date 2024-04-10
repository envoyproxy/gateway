// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsaarni/certyaml"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
)

func TestTLSConfig(t *testing.T) {
	// Create trusted CA, server and client certs.
	trustedCACert := certyaml.Certificate{
		Subject: "cn=trusted-ca",
	}
	egCertBeforeRotation := certyaml.Certificate{
		Subject:         "cn=eg-before-rotation",
		SubjectAltNames: []string{"DNS:localhost"},
		Issuer:          &trustedCACert,
	}
	egCertAfterRotation := certyaml.Certificate{
		Subject:         "cn=eg-after-rotation",
		SubjectAltNames: []string{"DNS:localhost"},
		Issuer:          &trustedCACert,
	}
	trustedEnvoyCert := certyaml.Certificate{
		Subject: "cn=trusted-envoy",
		Issuer:  &trustedCACert,
	}

	// Create another CA and a client cert to test that untrusted clients are denied.
	untrustedCACert := certyaml.Certificate{
		Subject: "cn=untrusted-ca",
	}
	untrustedClientCert := certyaml.Certificate{
		Subject: "cn=untrusted-client",
		Issuer:  &untrustedCACert,
	}

	caCertPool := x509.NewCertPool()
	ca, err := trustedCACert.X509Certificate()
	require.NoError(t, err)
	caCertPool.AddCert(&ca)

	tests := map[string]struct {
		serverCredentials *certyaml.Certificate
		clientCredentials *certyaml.Certificate
		expectError       bool
	}{
		"successful TLS connection established": {
			serverCredentials: &egCertBeforeRotation,
			clientCredentials: &trustedEnvoyCert,
			expectError:       false,
		},
		"rotating server credentials returns new server cert": {
			serverCredentials: &egCertAfterRotation,
			clientCredentials: &trustedEnvoyCert,
			expectError:       false,
		},
		"rotating server credentials again to ensure rotation can be repeated": {
			serverCredentials: &egCertBeforeRotation,
			clientCredentials: &trustedEnvoyCert,
			expectError:       false,
		},
		"fail to connect with client certificate which is not signed by correct CA": {
			serverCredentials: &egCertBeforeRotation,
			clientCredentials: &untrustedClientCert,
			expectError:       true,
		},
	}

	// Create temporary directory to store certificates and key for the server.
	configDir, err := os.MkdirTemp("", "eg-testdata-")
	require.NoError(t, err)
	defer os.RemoveAll(configDir)

	caFile := filepath.Join(configDir, "ca.crt")
	certFile := filepath.Join(configDir, "tls.crt")
	keyFile := filepath.Join(configDir, "tls.key")

	// Initial set of credentials must be written into temp directory before
	// starting the tests to avoid error at server startup.
	err = trustedCACert.WritePEM(caFile, keyFile)
	require.NoError(t, err)
	err = egCertBeforeRotation.WritePEM(certFile, keyFile)
	require.NoError(t, err)

	// Start a dummy server.
	logger := logging.DefaultLogger(v1alpha1.LogLevelInfo)

	cfg := &Config{
		Server: config.Server{
			Logger: logger,
		},
	}
	r := New(cfg)
	g := grpc.NewServer(grpc.Creds(credentials.NewTLS(r.tlsConfig(certFile, keyFile, caFile))))
	if g == nil {
		t.Error("failed to create server")
	}

	address := "localhost:8001"
	l, err := net.Listen("tcp", address)
	require.NoError(t, err)

	go func() {
		err := g.Serve(l)
		require.NoError(t, err)
	}()
	defer g.GracefulStop()

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			// Store certificate and key to temp dir used by serveContext.
			err = tc.serverCredentials.WritePEM(certFile, keyFile)
			require.NoError(t, err)
			clientCert, _ := tc.clientCredentials.TLSCertificate()
			receivedCert, err := tryConnect(address, clientCert, caCertPool)
			gotError := err != nil
			if gotError != tc.expectError {
				t.Errorf("Unexpected result when connecting to the server: %s", err)
			}
			if err == nil {
				expectedCert, _ := tc.serverCredentials.X509Certificate()
				assert.Equal(t, &expectedCert, receivedCert)
			}
		})
	}
}

// tryConnect tries to establish TLS connection to the server.
// If successful, return the server certificate.
func tryConnect(address string, clientCert tls.Certificate, caCertPool *x509.CertPool) (*x509.Certificate, error) {
	clientConfig := &tls.Config{
		ServerName:   "localhost",
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caCertPool,
	}
	conn, err := tls.Dial("tcp", address, clientConfig)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	err = peekError(conn)
	if err != nil {
		return nil, err
	}

	return conn.ConnectionState().PeerCertificates[0], nil
}

// peekError is a workaround for TLS 1.3: due to shortened handshake, TLS alert
// from server is received at first read from the socket. To receive alert for
// bad certificate, this function tries to read one byte.
// Adapted from https://golang.org/src/crypto/tls/handshake_client_test.go
func peekError(conn net.Conn) error {
	_ = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	_, err := conn.Read(make([]byte, 1))
	if err != nil {
		var netErr net.Error
		if !errors.As(netErr, &netErr) || !netErr.Timeout() {
			return err
		}
	}
	return nil
}

func TestServeXdsServerListenFailed(t *testing.T) {
	// Occupy the address to make listening failed
	addr := net.JoinHostPort(XdsServerAddress, strconv.Itoa(bootstrap.DefaultXdsServerPort))
	l, err := net.Listen("tcp", addr)
	require.NoError(t, err)
	defer l.Close()

	cfg, _ := config.New()
	r := New(&Config{
		Server: *cfg,
	})
	r.Logger = r.Logger.WithName(r.Name()).WithValues("runner", r.Name())
	// Don't crash in this function
	r.serveXdsServer(context.Background())
}
