// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"time"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	secret "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	// Default listening port for the SDS server
	defaultPort = 18001
	// Secret names
	tlsCertSecretName    = "server_cert"
	validationSecretName = "validation_context"
)

var (
	port       = flag.Int("port", defaultPort, "gRPC port for SDS server (ignored if socket is set)")
	socketPath = flag.String("socket", "", "Unix domain socket path for SDS server")
	certFile   = flag.String("cert", "", "Path to TLS certificate file (PEM format)")
	keyFile    = flag.String("key", "", "Path to TLS private key file (PEM format)")
	caFile     = flag.String("ca", "", "Path to CA certificate file (PEM format, optional)")
)

// callbacks implements the go-control-plane callbacks interface
type callbacks struct {
	signal   chan struct{}
	fetches  int
	requests int
	cache    cachev3.SnapshotCache
	snapshot cachev3.ResourceSnapshot
}

func (cb *callbacks) Report() {
	log.Printf("Server callbacks: fetches=%d requests=%d\n", cb.fetches, cb.requests)
}

func (cb *callbacks) OnStreamOpen(ctx context.Context, id int64, typ string) error {
	log.Printf("Stream opened: id=%d type=%s\n", id, typ)
	return nil
}

func (cb *callbacks) OnStreamClosed(id int64, node *core.Node) {
	log.Printf("Stream closed: id=%d\n", id)
}

func (cb *callbacks) OnStreamRequest(id int64, req *discovery.DiscoveryRequest) error {
	cb.requests++
	log.Printf("Stream request: id=%d version=%s resources=%v type=%s\n",
		id, req.VersionInfo, req.ResourceNames, req.TypeUrl)

	// Automatically set snapshot for any node that connects
	if req.Node != nil && req.Node.Id != "" {
		if err := cb.cache.SetSnapshot(context.Background(), req.Node.Id, cb.snapshot); err != nil {
			log.Printf("Failed to set snapshot for node %s: %v", req.Node.Id, err)
		} else {
			log.Printf("Set snapshot for node: %s", req.Node.Id)
		}
	}
	return nil
}

func (cb *callbacks) OnStreamResponse(ctx context.Context, id int64, req *discovery.DiscoveryRequest, resp *discovery.DiscoveryResponse) {
	cb.fetches++
	log.Printf("Stream response: id=%d version=%s type=%s resources=%d\n",
		id, resp.VersionInfo, resp.TypeUrl, len(resp.Resources))
}

func (cb *callbacks) OnFetchRequest(ctx context.Context, req *discovery.DiscoveryRequest) error {
	cb.requests++
	log.Printf("Fetch request: version=%s resources=%v type=%s\n",
		req.VersionInfo, req.ResourceNames, req.TypeUrl)
	return nil
}

func (cb *callbacks) OnFetchResponse(req *discovery.DiscoveryRequest, resp *discovery.DiscoveryResponse) {
	cb.fetches++
	log.Printf("Fetch response: version=%s type=%s resources=%d\n",
		resp.VersionInfo, resp.TypeUrl, len(resp.Resources))
}

func (cb *callbacks) OnDeltaStreamOpen(ctx context.Context, id int64, typ string) error {
	log.Printf("Delta stream opened: id=%d type=%s\n", id, typ)
	return nil
}

func (cb *callbacks) OnDeltaStreamClosed(id int64, node *core.Node) {
	log.Printf("Delta stream closed: id=%d\n", id)
}

func (cb *callbacks) OnStreamDeltaRequest(id int64, req *discovery.DeltaDiscoveryRequest) error {
	log.Printf("Delta stream request: id=%d type=%s\n", id, req.TypeUrl)
	return nil
}

func (cb *callbacks) OnStreamDeltaResponse(id int64, req *discovery.DeltaDiscoveryRequest, resp *discovery.DeltaDiscoveryResponse) {
	log.Printf("Delta stream response: id=%d type=%s\n", id, resp.TypeUrl)
}

// loadCertificateFromFile loads a certificate from a PEM file
func loadCertificateFromFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file %s: %w", filePath, err)
	}

	// Verify it's valid PEM
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from %s", filePath)
	}

	return data, nil
}

// loadPrivateKeyFromFile loads a private key from a PEM file
func loadPrivateKeyFromFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file %s: %w", filePath, err)
	}

	// Verify it's valid PEM
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from %s", filePath)
	}

	return data, nil
}

// generateSelfSignedCert generates a self-signed certificate and private key for testing
func generateSelfSignedCert() (certPEM, keyPEM []byte, err error) {
	// Generate RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour) // Valid for 1 year

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Envoy Gateway Test"},
			CommonName:   "sds-test.example.com",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"sds-test.example.com", "*.example.com", "localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	// Create self-signed certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Encode certificate to PEM
	certPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Encode private key to PEM
	keyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return certPEM, keyPEM, nil
}

// createTLSCertificateSecret creates an Envoy TLS certificate secret
func createTLSCertificateSecret(name string, certPEM, keyPEM []byte) (*auth.Secret, error) {
	tlsCertificate := &auth.TlsCertificate{
		CertificateChain: &core.DataSource{
			Specifier: &core.DataSource_InlineBytes{
				InlineBytes: certPEM,
			},
		},
		PrivateKey: &core.DataSource{
			Specifier: &core.DataSource_InlineBytes{
				InlineBytes: keyPEM,
			},
		},
	}

	return &auth.Secret{
		Name: name,
		Type: &auth.Secret_TlsCertificate{
			TlsCertificate: tlsCertificate,
		},
	}, nil
}

// createValidationContextSecret creates an Envoy validation context secret
func createValidationContextSecret(name string, caPEM []byte) (*auth.Secret, error) {
	validationContext := &auth.CertificateValidationContext{
		TrustedCa: &core.DataSource{
			Specifier: &core.DataSource_InlineBytes{
				InlineBytes: caPEM,
			},
		},
	}

	return &auth.Secret{
		Name: name,
		Type: &auth.Secret_ValidationContext{
			ValidationContext: validationContext,
		},
	}, nil
}

// makeSnapshot creates a snapshot with secrets
func makeSnapshot(version string, certPEM, keyPEM, caPEM []byte) (cachev3.ResourceSnapshot, error) {
	log.Printf("Creating snapshot with provided certificates")

	// Create TLS certificate secret
	tlsSecret, err := createTLSCertificateSecret(tlsCertSecretName, certPEM, keyPEM)
	if err != nil {
		return &cachev3.Snapshot{}, fmt.Errorf("failed to create TLS secret: %w", err)
	}

	tlsSecretAny, err := anypb.New(tlsSecret)
	if err != nil {
		return &cachev3.Snapshot{}, fmt.Errorf("failed to marshal TLS secret: %w", err)
	}

	// Use CA if provided, otherwise use cert as CA
	if len(caPEM) == 0 {
		caPEM = certPEM
	}

	// Create validation context secret
	validationSecret, err := createValidationContextSecret(validationSecretName, caPEM)
	if err != nil {
		return &cachev3.Snapshot{}, fmt.Errorf("failed to create validation secret: %w", err)
	}

	validationSecretAny, err := anypb.New(validationSecret)
	if err != nil {
		return &cachev3.Snapshot{}, fmt.Errorf("failed to marshal validation secret: %w", err)
	}

	secrets := []types.Resource{
		&auth.Secret{},
		&auth.Secret{},
	}

	// Unmarshal back to ensure proper type
	if err := tlsSecretAny.UnmarshalTo(secrets[0]); err != nil {
		return &cachev3.Snapshot{}, fmt.Errorf("failed to unmarshal TLS secret: %w", err)
	}

	if err := validationSecretAny.UnmarshalTo(secrets[1]); err != nil {
		return &cachev3.Snapshot{}, fmt.Errorf("failed to unmarshal validation secret: %w", err)
	}

	// Create snapshot
	snapshot, err := cachev3.NewSnapshot(
		version,
		map[resource.Type][]types.Resource{
			resource.SecretType: secrets,
		},
	)
	if err != nil {
		return &cachev3.Snapshot{}, fmt.Errorf("failed to create snapshot: %w", err)
	}

	return snapshot, nil
}

func main() {
	log.Printf("Starting Envoy SDS server")
	flag.Parse()

	var certPEM, keyPEM, caPEM []byte
	var err error

	// Load certificates from files or generate self-signed for testing
	if *certFile != "" && *keyFile != "" {
		log.Printf("Loading certificate from files: cert=%s, key=%s", *certFile, *keyFile)

		certPEM, err = loadCertificateFromFile(*certFile)
		if err != nil {
			log.Fatalf("Failed to load certificate: %v", err)
		}

		keyPEM, err = loadPrivateKeyFromFile(*keyFile)
		if err != nil {
			log.Fatalf("Failed to load private key: %v", err)
		}

		// Load CA certificate if provided
		if *caFile != "" {
			log.Printf("Loading CA certificate from: %s", *caFile)
			caPEM, err = loadCertificateFromFile(*caFile)
			if err != nil {
				log.Fatalf("Failed to load CA certificate: %v", err)
			}
		}

		log.Printf("Successfully loaded certificates from file system")
	} else {
		log.Printf("No certificate files provided, generating self-signed certificate for testing")
		certPEM, keyPEM, err = generateSelfSignedCert()
		if err != nil {
			log.Fatalf("Failed to generate self-signed certificate: %v", err)
		}
		log.Printf("Generated self-signed certificate for CN=sds-test.example.com")
	}

	// Create a cache
	cache := cachev3.NewSnapshotCache(false, cachev3.IDHash{}, nil)

	// Create snapshot with secrets
	snapshot, err := makeSnapshot("v1", certPEM, keyPEM, caPEM)
	if err != nil {
		log.Fatalf("Failed to create snapshot: %v", err)
	}

	log.Printf("Snapshot created with version v1")
	log.Printf("Available secrets: %s, %s", tlsCertSecretName, validationSecretName)
	log.Printf("SDS server will serve secrets to any connecting node")

	// Create callbacks with cache and snapshot
	cb := &callbacks{
		signal:   make(chan struct{}),
		fetches:  0,
		requests: 0,
		cache:    cache,
		snapshot: snapshot,
	}

	// Create xDS server
	ctx := context.Background()
	srv := serverv3.NewServer(ctx, cache, cb)

	// Configure gRPC server
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions,
		grpc.MaxConcurrentStreams(1000),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time: 5 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)

	grpcServer := grpc.NewServer(grpcOptions...)

	// Register SDS service
	secret.RegisterSecretDiscoveryServiceServer(grpcServer, srv)
	// Also register ADS for clients that use it
	discovery.RegisterAggregatedDiscoveryServiceServer(grpcServer, srv)

	var lis net.Listener
	if *socketPath != "" {
		// Listen and serve on Unix domain socket
		// Remove existing socket file if it exists
		if err := os.RemoveAll(*socketPath); err != nil {
			log.Printf("Warning: failed to remove existing socket: %v", err)
		}

		lis, err = net.Listen("unix", *socketPath)
		if err != nil {
			log.Fatalf("Failed to listen on socket %s: %v", *socketPath, err)
		}

		// Ensure socket has proper permissions
		if err := os.Chmod(*socketPath, 0666); err != nil {
			log.Printf("Warning: failed to chmod socket: %v", err)
		}

		log.Printf("SDS server listening on Unix socket: %s", *socketPath)
	} else {
		// Listen and serve on TCP
		addr := fmt.Sprintf(":%d", *port)
		lis, err = net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("Failed to listen on %s: %v", addr, err)
		}

		log.Printf("SDS server listening on %s", addr)
	}

	log.Println("Ready to serve SDS requests")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
