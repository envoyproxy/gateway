// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package crypto

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsaarni/certyaml"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// TestLoadServerTLSConfig verifies the returned tls.Config re-reads the cert from
// disk on every handshake and does not request a client certificate.
func TestLoadServerTLSConfig(t *testing.T) {
	configDir := t.TempDir()
	certFile := filepath.Join(configDir, "tls.crt")
	keyFile := filepath.Join(configDir, "tls.key")

	caCert := certyaml.Certificate{Subject: "cn=test-ca"}

	serverCertBefore := certyaml.Certificate{
		Subject:         "cn=server-before",
		SubjectAltNames: []string{"DNS:localhost"},
		Issuer:          &caCert,
	}
	serverCertAfter := certyaml.Certificate{
		Subject:         "cn=server-after",
		SubjectAltNames: []string{"DNS:localhost"},
		Issuer:          &caCert,
	}

	require.NoError(t, serverCertBefore.WritePEM(certFile, keyFile))

	tlsCfg, err := LoadServerTLSConfig(certFile, keyFile)
	require.NoError(t, err)
	require.NotNil(t, tlsCfg)

	require.NotNil(t, tlsCfg.GetConfigForClient)
	require.Equal(t, uint16(tls.VersionTLS13), tlsCfg.MinVersion)
	require.Equal(t, tls.NoClientCert, tlsCfg.ClientAuth)
	require.Empty(t, tlsCfg.Certificates)

	// Stand up a TLS server, dial it, rotate the cert on disk, dial again.
	caPool := x509.NewCertPool()
	ca, err := caCert.X509Certificate()
	require.NoError(t, err)
	caPool.AddCert(&ca)

	srv := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsCfg)))
	defer srv.GracefulStop()

	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	go func() { _ = srv.Serve(l) }()

	dial := func() *x509.Certificate {
		t.Helper()
		conn, err := tls.Dial("tcp", l.Addr().String(), &tls.Config{
			ServerName: "localhost",
			MinVersion: tls.VersionTLS13,
			RootCAs:    caPool,
			NextProtos: []string{"h2"},
		})
		require.NoError(t, err)
		defer conn.Close()
		// TLS 1.3 defers server alerts until first read; surface them now.
		_ = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		_, readErr := conn.Read(make([]byte, 1))
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			var netErr net.Error
			require.True(t, errors.As(readErr, &netErr) && netErr.Timeout(),
				"unexpected read error after handshake: %v", readErr)
		}
		require.NotEmpty(t, conn.ConnectionState().PeerCertificates)
		return conn.ConnectionState().PeerCertificates[0]
	}

	gotBefore := dial()
	expectedBefore, err := serverCertBefore.X509Certificate()
	require.NoError(t, err)
	assert.Equal(t, expectedBefore.Subject.CommonName, gotBefore.Subject.CommonName)

	require.NoError(t, serverCertAfter.WritePEM(certFile, keyFile))

	gotAfter := dial()
	expectedAfter, err := serverCertAfter.X509Certificate()
	require.NoError(t, err)
	assert.Equal(t, expectedAfter.Subject.CommonName, gotAfter.Subject.CommonName)
	assert.NotEqual(t, gotBefore.Subject.CommonName, gotAfter.Subject.CommonName)
}

func TestLoadServerTLSConfig_MissingFiles(t *testing.T) {
	_, err := LoadServerTLSConfig("/does/not/exist/tls.crt", "/does/not/exist/tls.key")
	require.Error(t, err)
}
