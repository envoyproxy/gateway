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
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"google.golang.org/grpc"
	"sigs.k8s.io/controller-runtime/pkg/client"
	clicfg "sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/envoyproxy/gateway-remote-infra/pb"
	"github.com/envoyproxy/gateway-remote-infra/synthesizer"
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
	ir := infraFromProto(req.GetInfra())

	fmt.Printf("Creating proxy infra [%v]\n", ir)
	err := bs.synth.CreateOrUpdate(ctx, ir)
	return new(pb.CreateOrUpdateProxyInfraResponse{}), err
}

func (bs *basicServer) DeleteProxyInfra(ctx context.Context, req *pb.DeleteProxyInfraRequest) (*pb.DeleteProxyInfraResponse, error) {
	ir := infraFromProto(req.GetInfra())

	fmt.Printf("Deleting proxy infra [%v]\n", ir)
	err := bs.synth.Delete(ctx, ir)
	return new(pb.DeleteProxyInfraResponse{}), err
}

// infraFromProto maps the structured proto IR onto the pared-down
// synthesizer.Infra used by this example. Fields the synthesizer does not
// consume (proxy config, addresses, resolved metric sinks, and the owner
// reference within metadata) are intentionally dropped, demonstrating the
// forward-compatible posture a provider should adopt: read only what you
// understand.
func infraFromProto(in *pb.Infra) *synthesizer.Infra {
	if in == nil || in.GetProxy() == nil {
		return new(synthesizer.Infra{})
	}

	p := in.GetProxy()
	proxy := &synthesizer.ProxyInfra{
		Name:      p.GetName(),
		Namespace: p.GetNamespace(),
	}

	if md := p.GetMetadata(); md != nil {
		proxy.Metadata = &synthesizer.InfraMetadata{
			Annotations: md.GetAnnotations(),
			Labels:      md.GetLabels(),
		}
	}

	for _, l := range p.GetListeners() {
		if l == nil {
			continue
		}
		listener := &synthesizer.ProxyListener{Name: l.GetName()}
		for _, port := range l.GetPorts() {
			listener.Ports = append(listener.Ports, synthesizer.ListenerPort{
				Name:          port.GetName(),
				Protocol:      port.GetProtocol(),
				ServicePort:   port.GetServicePort(),
				ContainerPort: port.GetContainerPort(),
			})
		}
		proxy.Listeners = append(proxy.Listeners, listener)
	}

	return &synthesizer.Infra{Proxy: proxy}
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
