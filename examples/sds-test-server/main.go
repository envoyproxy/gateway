// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	secretservicev3 "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	cachetypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	logv3 "github.com/envoyproxy/go-control-plane/pkg/log"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"google.golang.org/grpc"
)

const snapshotKey = "sds-test-server"

type config struct {
	socketPath      string
	secretName      string
	certificatePath string
	privateKeyPath  string
}

type staticNodeHash struct{}

func (staticNodeHash) ID(*corev3.Node) string {
	return snapshotKey
}

func main() {
	var cfg config
	flag.StringVar(&cfg.socketPath, "socket-path", "", "path to the SDS Unix domain socket")
	flag.StringVar(&cfg.secretName, "secret-name", "", "name of the SDS TLS secret")
	flag.StringVar(&cfg.certificatePath, "cert-path", "", "path to the PEM certificate chain")
	flag.StringVar(&cfg.privateKeyPath, "key-path", "", "path to the PEM private key")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg); err != nil {
		log.Fatalf("SDS server failed: %v", err)
	}
}

func run(ctx context.Context, cfg config) error {
	grpcServer, listener, err := newSDSServer(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() {
		_ = listener.Close()
		_ = os.Remove(cfg.socketPath)
	}()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- grpcServer.Serve(listener)
	}()

	select {
	case <-ctx.Done():
		grpcServer.Stop()
		return nil
	case err := <-serveErr:
		return fmt.Errorf("serve SDS requests: %w", err)
	}
}

func newSDSServer(ctx context.Context, cfg config) (*grpc.Server, net.Listener, error) {
	certificate, err := os.ReadFile(cfg.certificatePath)
	if err != nil {
		return nil, nil, fmt.Errorf("read certificate: %w", err)
	}
	privateKey, err := os.ReadFile(cfg.privateKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read private key: %w", err)
	}

	secret := &tlsv3.Secret{
		Name: cfg.secretName,
		Type: &tlsv3.Secret_TlsCertificate{
			TlsCertificate: &tlsv3.TlsCertificate{
				CertificateChain: inlineBytes(certificate),
				PrivateKey:       inlineBytes(privateKey),
			},
		},
	}
	snapshot, err := cachev3.NewSnapshot("1", map[resourcev3.Type][]cachetypes.Resource{
		resourcev3.SecretType: {secret},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("create SDS snapshot: %w", err)
	}
	if err := snapshot.Consistent(); err != nil {
		return nil, nil, fmt.Errorf("validate SDS snapshot: %w", err)
	}

	snapshotCache := cachev3.NewSnapshotCache(false, staticNodeHash{}, logv3.NewDefaultLogger())
	if err := snapshotCache.SetSnapshot(ctx, snapshotKey, snapshot); err != nil {
		return nil, nil, fmt.Errorf("store SDS snapshot: %w", err)
	}

	if err := os.Remove(cfg.socketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, nil, fmt.Errorf("remove stale socket: %w", err)
	}
	listener, err := net.Listen("unix", cfg.socketPath)
	if err != nil {
		return nil, nil, fmt.Errorf("listen on SDS socket: %w", err)
	}
	if err := os.Chmod(cfg.socketPath, 0o666); err != nil {
		_ = listener.Close()
		return nil, nil, fmt.Errorf("set SDS socket permissions: %w", err)
	}

	grpcServer := grpc.NewServer()
	secretservicev3.RegisterSecretDiscoveryServiceServer(
		grpcServer,
		serverv3.NewServer(ctx, snapshotCache, nil),
	)
	return grpcServer, listener, nil
}

func inlineBytes(data []byte) *corev3.DataSource {
	return &corev3.DataSource{
		Specifier: &corev3.DataSource_InlineBytes{InlineBytes: data},
	}
}
