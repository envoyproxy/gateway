// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	secretservicev3 "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

func TestSDSServerServesTLSSecretOverUnixSocket(t *testing.T) {
	// Given a server configured with certificate material on disk.
	tempDir := t.TempDir()
	socketFile, err := os.CreateTemp("", "sds-*.sock")
	require.NoError(t, err)
	socketPath := socketFile.Name()
	require.NoError(t, socketFile.Close())
	require.NoError(t, os.Remove(socketPath))
	t.Cleanup(func() { _ = os.Remove(socketPath) })
	certificate := []byte("test certificate")
	privateKey := []byte("test private key")
	certificatePath := filepath.Join(tempDir, "tls.crt")
	privateKeyPath := filepath.Join(tempDir, "tls.key")
	require.NoError(t, os.WriteFile(certificatePath, certificate, 0o600))
	require.NoError(t, os.WriteFile(privateKeyPath, privateKey, 0o600))
	cfg := config{
		socketPath:      socketPath,
		secretName:      "listener-certificate",
		certificatePath: certificatePath,
		privateKeyPath:  privateKeyPath,
	}
	grpcServer, listener, err := newSDSServer(t.Context(), cfg)
	require.NoError(t, err)
	t.Cleanup(grpcServer.Stop)
	t.Cleanup(func() { require.NoError(t, listener.Close()) })
	go func() { _ = grpcServer.Serve(listener) }()

	connection, err := grpc.NewClient(
		"passthrough:///sds-test-server",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", cfg.socketPath)
		}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, connection.Close()) })
	stream, err := secretservicev3.NewSecretDiscoveryServiceClient(connection).StreamSecrets(t.Context())
	require.NoError(t, err)

	// When the listener certificate is requested through the SDS stream.
	require.NoError(t, stream.Send(&discoveryv3.DiscoveryRequest{
		Node:          &corev3.Node{Id: "envoy-test-node"},
		ResourceNames: []string{cfg.secretName},
		TypeUrl:       resourcev3.SecretType,
	}))
	response, err := stream.Recv()
	require.NoError(t, err)

	// Then the named secret contains the configured certificate and key.
	require.Len(t, response.Resources, 1)
	secret := &tlsv3.Secret{}
	require.NoError(t, response.Resources[0].UnmarshalTo(secret))
	require.Equal(t, cfg.secretName, secret.Name)
	require.True(t, proto.Equal(inlineBytes(certificate), secret.GetTlsCertificate().CertificateChain))
	require.True(t, proto.Equal(inlineBytes(privateKey), secret.GetTlsCertificate().PrivateKey))
}
