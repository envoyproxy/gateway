// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package remote

import (
	"context"

	"google.golang.org/grpc"
	k8scli "sigs.k8s.io/controller-runtime/pkg/client"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	grpcExtension "github.com/envoyproxy/gateway/internal/extension"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/proto/remoteinfra"
)

const grpcServiceName = "envoygateway.remoteinfra.EnvoyGatewayRemoteInfrastructureProvider"

// InfraClient is the contract used by Infra to talk to a remote
// infrastructure provider.
type InfraClient interface {
	Close() error
	CreateOrUpdateProxyInfra(ctx context.Context, infra *ir.Infra) error
	DeleteProxyInfra(ctx context.Context, infra *ir.Infra) error
	CreateOrUpdateRateLimitInfra(ctx context.Context) error
	DeleteRateLimitInfra(ctx context.Context) error
}

// InfraClientFactory builds an InfraClient on demand. Infra invokes the
// factory lazily on the first method call that needs the client.
type InfraClientFactory func(ctx context.Context) (InfraClient, error)

// infraClientImpl is the gRPC-backed implementation of InfraClient. It maps
// IR payloads onto the structured proto contract and forwards them to a
// remote infrastructure provider.
type infraClientImpl struct {
	k8sClient          k8scli.Client
	namespace          string
	remoteService      *egv1a1.ExtensionService
	extensionConnCache *grpc.ClientConn
	client             remoteinfra.EnvoyGatewayRemoteInfrastructureProviderClient
}

// Close releases the underlying gRPC connection. It is a no-op if the
// connection was never established.
func (i *infraClientImpl) Close() error {
	if i.extensionConnCache == nil {
		return nil
	}
	return i.extensionConnCache.Close()
}

// CreateOrUpdateProxyInfra maps the IR onto the proto contract and forwards it
// to the remote provider. The provider is responsible for reconciling the
// proxy infrastructure to match.
func (i *infraClientImpl) CreateOrUpdateProxyInfra(ctx context.Context, infra *ir.Infra) error {
	pbInfra, err := infraToProto(infra)
	if err != nil {
		return err
	}

	req := new(remoteinfra.CreateOrUpdateProxyInfraRequest{Infra: pbInfra})

	_, err = i.client.CreateOrUpdateProxyInfra(ctx, req)
	return err
}

// DeleteProxyInfra maps the IR onto the proto contract and asks the remote
// provider to tear down the corresponding proxy infrastructure.
func (i *infraClientImpl) DeleteProxyInfra(ctx context.Context, infra *ir.Infra) error {
	pbInfra, err := infraToProto(infra)
	if err != nil {
		return err
	}

	req := new(remoteinfra.DeleteProxyInfraRequest{Infra: pbInfra})

	_, err = i.client.DeleteProxyInfra(ctx, req)
	return err
}

// CreateOrUpdateRateLimitInfra asks the remote provider to create or update
// the rate limit infrastructure.
func (i *infraClientImpl) CreateOrUpdateRateLimitInfra(ctx context.Context) error {
	_, err := i.client.CreateOrUpdateRateLimitInfra(ctx, new(remoteinfra.CreateOrUpdateRateLimitInfraRequest{}))
	return err
}

// DeleteRateLimitInfra asks the remote provider to tear down the rate limit
// infrastructure.
func (i *infraClientImpl) DeleteRateLimitInfra(ctx context.Context) error {
	_, err := i.client.DeleteRateLimitInfra(ctx, new(remoteinfra.DeleteRateLimitInfraRequest{}))
	return err
}

// getClientConnCache returns the cached gRPC connection, dialing the remote
// service on the first call.
func (i *infraClientImpl) getClientConnCache(ctx context.Context) (*grpc.ClientConn, error) {
	if i.extensionConnCache != nil {
		return i.extensionConnCache, nil
	}

	serverAddr := grpcExtension.GetExtensionServerAddress(i.remoteService)

	opts, err := grpcExtension.GenerateGRPCOptions(ctx, i.k8sClient, i.remoteService, nil, grpcServiceName, i.namespace)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.NewClient(serverAddr, opts...)
	if err != nil {
		return nil, err
	}

	i.extensionConnCache = conn
	return conn, nil
}

// DefaultInfraClientFactory returns an InfraClientFactory that defers
// construction of the remote gRPC client until the factory is invoked.
func DefaultInfraClientFactory(cfg *config.Server, k8sClient k8scli.Client) InfraClientFactory {
	return func(ctx context.Context) (InfraClient, error) {
		return newRemoteInfraClient(ctx, cfg, k8sClient)
	}
}

// newRemoteInfraClient returns a new InfraClient that talks to a remote
// infrastructure provider over gRPC.
func newRemoteInfraClient(ctx context.Context, cfg *config.Server, k8sClient k8scli.Client) (InfraClient, error) {
	extensionCfg := cfg.EnvoyGateway.Provider.Custom.Infrastructure.Remote.Service
	c := new(infraClientImpl{
		k8sClient:     k8sClient,
		namespace:     cfg.ControllerNamespace,
		remoteService: extensionCfg,
	})

	if err := c.getImplementationClient(ctx); err != nil {
		// Make sure we don't leak any partially-created connection.
		_ = c.Close()
		return nil, err
	}
	return c, nil
}

// getImplementationClient generates the grpc client to communicate with a remote infrastructure server.
func (i *infraClientImpl) getImplementationClient(ctx context.Context) error {
	conn, err := i.getClientConnCache(ctx)
	if err != nil {
		return err
	}

	i.client = remoteinfra.NewEnvoyGatewayRemoteInfrastructureProviderClient(conn)
	return nil
}

var _ InfraClient = (*infraClientImpl)(nil)
