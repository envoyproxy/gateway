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
	"strings"
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
	// Default node ID
	defaultNodeID = "sds-test-node"
	// Default common name
	defaultCommonName = "sds-test.example.com"
	// Default DNS names
	defaultDNSNames = "sds-test.example.com,*.example.com,localhost"
	// Secret names
	tlsCertSecretName    = "server_cert"
	validationSecretName = "validation_context"
)

var (
	port       = flag.Int("port", defaultPort, "gRPC port for SDS server")
	nodeID     = flag.String("node", defaultNodeID, "Node ID for envoy")
	commonName = flag.String("cn", defaultCommonName, "Common Name for the certificate")
	dnsNames   = flag.String("dns", defaultDNSNames, "Comma-separated list of DNS names for the certificate")
)

// callbacks implements the go-control-plane callbacks interface
type callbacks struct {
	signal   chan struct{}
	fetches  int
	requests int
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

// generateSelfSignedCert generates a self-signed certificate and private key
func generateSelfSignedCert(cn string, dnsNames []string) (certPEM, keyPEM []byte, err error) {
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
			CommonName:   cn,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              dnsNames,
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
func makeSnapshot(version string, cn string, dnsNames []string) (cachev3.ResourceSnapshot, error) {
	// Generate a self-signed certificate
	certPEM, keyPEM, err := generateSelfSignedCert(cn, dnsNames)
	if err != nil {
		return &cachev3.Snapshot{}, fmt.Errorf("failed to generate certificate: %w", err)
	}

	log.Printf("Generated certificate for CN=%s, DNS names=%v", cn, dnsNames)

	// Create TLS certificate secret
	tlsSecret, err := createTLSCertificateSecret(tlsCertSecretName, certPEM, keyPEM)
	if err != nil {
		return &cachev3.Snapshot{}, fmt.Errorf("failed to create TLS secret: %w", err)
	}

	// Create validation context secret (using the same cert as CA for testing)
	validationSecret, err := createValidationContextSecret(validationSecretName, certPEM)
	if err != nil {
		return &cachev3.Snapshot{}, fmt.Errorf("failed to create validation secret: %w", err)
	}

	// Marshal secrets to Any
	tlsSecretAny, err := anypb.New(tlsSecret)
	if err != nil {
		return &cachev3.Snapshot{}, fmt.Errorf("failed to marshal TLS secret: %w", err)
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
	flag.Parse()

	// Parse DNS names from comma-separated string
	dnsNamesList := []string{}
	if *dnsNames != "" {
		for _, name := range strings.Split(*dnsNames, ",") {
			name = strings.TrimSpace(name)
			if name != "" {
				dnsNamesList = append(dnsNamesList, name)
			}
		}
	}

	log.Printf("Starting Envoy SDS server on port %d", *port)
	log.Printf("Certificate CN: %s", *commonName)
	log.Printf("Certificate DNS names: %v", dnsNamesList)

	// Create a cache
	cache := cachev3.NewSnapshotCache(false, cachev3.IDHash{}, nil)

	// Create snapshot with secrets
	snapshot, err := makeSnapshot("v1", *commonName, dnsNamesList)
	if err != nil {
		log.Fatalf("Failed to create snapshot: %v", err)
	}

	// Set snapshot for the node
	if err := cache.SetSnapshot(context.Background(), *nodeID, snapshot); err != nil {
		log.Fatalf("Failed to set snapshot: %v", err)
	}

	log.Printf("Snapshot created with version v1")
	log.Printf("Available secrets: %s, %s", tlsCertSecretName, validationSecretName)

	// Create callbacks
	cb := &callbacks{
		signal:   make(chan struct{}),
		fetches:  0,
		requests: 0,
	}

	// Create xDS server
	ctx := context.Background()
	srv := serverv3.NewServer(ctx, cache, cb)

	// Configure gRPC server
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions,
		grpc.MaxConcurrentStreams(1000),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    30 * time.Second,
			Timeout: 5 * time.Second,
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

	// Listen and serve
	addr := fmt.Sprintf(":%d", *port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	log.Printf("SDS server listening on %s", addr)
	log.Printf("Node ID: %s", *nodeID)
	log.Println("Ready to serve SDS requests")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
