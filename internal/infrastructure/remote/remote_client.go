// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package remote

import (
	"context"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	grpcExtension "github.com/envoyproxy/gateway/internal/extension"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/proto/remoteinfra"
	"google.golang.org/grpc"
	k8scli "sigs.k8s.io/controller-runtime/pkg/client"
)

const grpcServiceName = "envoygateway.remoteinfra.EnvoyGatewayRemoteInfrastructureProvider"

type InfraClient interface {
	Close() error
	CreateOrUpdateProxyInfra(ctx context.Context, infra *ir.Infra) error
	DeleteProxyInfra(ctx context.Context, infra *ir.Infra) error
	CreateOrUpdateRateLimitInfra(ctx context.Context) error
	DeleteRateLimitInfra(ctx context.Context) error
}

type InfraClientImpl struct {
	k8sClient          k8scli.Client
	namespace          string
	remoteService      *egv1a1.ExtensionService
	extensionConnCache *grpc.ClientConn
	client             remoteinfra.EnvoyGatewayRemoteInfrastructureProviderClient
}

func (i *InfraClientImpl) Close() error {
	if i.extensionConnCache == nil {
		return nil
	}
	return i.extensionConnCache.Close()
}

func (i *InfraClientImpl) CreateOrUpdateProxyInfra(ctx context.Context, infra *ir.Infra) error {
	bs := []byte(infra.JSONString())

	req := new(remoteinfra.CreateOrUpdateProxyInfraRequest{IrBytes: bs})

	_, err := i.client.CreateOrUpdateProxyInfra(ctx, req)
	return err
}

func (i *InfraClientImpl) DeleteProxyInfra(ctx context.Context, infra *ir.Infra) error {
	bs := []byte(infra.JSONString())

	req := new(remoteinfra.DeleteProxyInfraRequest{IrBytes: bs})

	_, err := i.client.DeleteProxyInfra(ctx, req)
	return err
}

func (i *InfraClientImpl) CreateOrUpdateRateLimitInfra(ctx context.Context) error {
	_, err := i.client.CreateOrUpdateRateLimitInfra(ctx, new(remoteinfra.CreateOrUpdateRateLimitInfraRequest{}))
	return err
}

func (i *InfraClientImpl) DeleteRateLimitInfra(ctx context.Context) error {
	_, err := i.client.DeleteRateLimitInfra(ctx, new(remoteinfra.DeleteRateLimitInfraRequest{}))
	return err
}

func (i *InfraClientImpl) getClientConnCache(ctx context.Context) (*grpc.ClientConn, error) {
	if i.extensionConnCache != nil {
		return i.extensionConnCache, nil
	}

	serverAddr := grpcExtension.GetExtensionServerAddress(i.remoteService)

	opts, err := grpcExtension.GenerateGRPCOptions(ctx, i.k8sClient, i.remoteService, nil, grpcServiceName, i.namespace)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		return nil, err
	}

	i.extensionConnCache = conn
	return conn, nil
}

// newRemoteInfraClient returns a new Manager
func newRemoteInfraClient(cfg *config.Server, k8sClient k8scli.Client) (InfraClient, error) {
	extensionCfg := cfg.EnvoyGateway.Provider.Custom.Infrastructure.Remote.Service
	cfg.Logger.Info("extensionCfg", "config", extensionCfg)
	c := &InfraClientImpl{
		k8sClient:     k8sClient,
		remoteService: extensionCfg,
	}

	err := c.getImplementationClient(context.Background())
	if err != nil {
		return nil, err
	}
	return c, nil
}

// getImplementationClient generates the grpc client to communicate with a remote infrastructure server.
func (i *InfraClientImpl) getImplementationClient(ctx context.Context) error {
	conn, err := i.getClientConnCache(ctx)
	if err != nil {
		return err
	}

	i.client = remoteinfra.NewEnvoyGatewayRemoteInfrastructureProviderClient(conn)
	return nil
}
