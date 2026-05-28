// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// remote-infra is an example implementation of the Envoy Gateway remote
// infrastructure provider gRPC service. It reconciles a Deployment and
// Service for each proxy IR it receives. RateLimit RPCs are no-ops.
//
// gRPC is served over a Unix domain socket on a volume shared with the
// envoy-gateway container.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/envoyproxy/gateway-remote-infra/pb"
	"github.com/envoyproxy/gateway-remote-infra/synthesizer"
	"google.golang.org/grpc"
	"sigs.k8s.io/controller-runtime/pkg/client"
	clicfg "sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	socketPath string
	socketMode uint
)

// basicServer implements the remote infrastructure provider service. Proxy
// RPCs reconcile Kubernetes resources via the synthesizer; rate limit RPCs
// are no-ops.
type basicServer struct {
	synth synthesizer.InfraSynthesizer
	pb.UnimplementedEnvoyGatewayRemoteInfrastructureProviderServer
}

func (bs *basicServer) CreateOrUpdateProxyInfra(ctx context.Context, req *pb.CreateOrUpdateProxyInfraRequest) (*pb.CreateOrUpdateProxyInfraResponse, error) {
	ir := new(synthesizer.Infra{})
	if err := json.Unmarshal(req.GetIrBytes(), ir); err != nil {
		return nil, fmt.Errorf("failed to unmarshal IR: %w", err)
	}

	fmt.Printf("Creating proxy infra [%v]\n", ir)
	err := bs.synth.CreateOrUpdate(ctx, ir)
	return new(pb.CreateOrUpdateProxyInfraResponse{}), err
}

func (bs *basicServer) DeleteProxyInfra(ctx context.Context, req *pb.DeleteProxyInfraRequest) (*pb.DeleteProxyInfraResponse, error) {
	ir := new(synthesizer.Infra{})
	if err := json.Unmarshal(req.GetIrBytes(), ir); err != nil {
		return nil, fmt.Errorf("failed to unmarshal IR: %w", err)
	}

	fmt.Printf("Deleting proxy infra [%v]\n", ir)
	err := bs.synth.Delete(ctx, ir)
	return new(pb.DeleteProxyInfraResponse{}), err
}

func (bs *basicServer) CreateOrUpdateRateLimitInfra(_ context.Context, _ *pb.CreateOrUpdateRateLimitInfraRequest) (*pb.CreateOrUpdateRateLimitInfraResponse, error) {
	return new(pb.CreateOrUpdateRateLimitInfraResponse{}), nil
}

func (bs *basicServer) DeleteRateLimitInfra(_ context.Context, _ *pb.DeleteRateLimitInfraRequest) (*pb.DeleteRateLimitInfraResponse, error) {
	return new(pb.DeleteRateLimitInfraResponse{}), nil
}

// listenUDS binds a Unix domain socket at path with the given permission
// mode, removing any stale socket file first.
func listenUDS(path string, mode os.FileMode) (net.Listener, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create socket directory: %w", err)
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("remove stale socket %s: %w", path, err)
	}
	lis, err := net.Listen("unix", path)
	if err != nil {
		return nil, fmt.Errorf("listen on unix %s: %w", path, err)
	}
	if err := os.Chmod(path, mode); err != nil {
		_ = lis.Close()
		return nil, fmt.Errorf("chmod socket %s: %w", path, err)
	}
	return lis, nil
}

func main() {
	flag.StringVar(&socketPath, "socket-path", "/var/run/remote-infra/server.sock", "Unix domain socket path for the gRPC server")
	flag.UintVar(&socketMode, "socket-mode", 0o660, "file permission mode for the UDS socket file")
	flag.Parse()

	lis, err := listenUDS(socketPath, os.FileMode(socketMode))
	if err != nil {
		log.Fatalf("failed to listen on UDS %s: %v", socketPath, err)
	}

	gs := grpc.NewServer()

	clientConfig := clicfg.GetConfigOrDie()
	k8sClient, err := client.New(clientConfig, client.Options{})
	if err != nil {
		log.Fatalf("failed to create k8s client: %v", err)
	}

	bs := &basicServer{
		synth: synthesizer.InfraSynthesizer{KubernetesClient: k8sClient, Namespace: os.Getenv("NAMESPACE")},
	}

	pb.RegisterEnvoyGatewayRemoteInfrastructureProviderServer(gs, bs)

	log.Printf("remote-infra: serving gRPC on unix://%s", socketPath)
	if err := gs.Serve(lis); err != nil {
		log.Fatalf("gRPC server exited: %v", err)
	}
}
