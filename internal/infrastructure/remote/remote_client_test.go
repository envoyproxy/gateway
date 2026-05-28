// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package remote

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/proto/remoteinfra"
)

// fakeRemoteInfraServer is a stub implementation of the gRPC server used to
// observe what the infraClientImpl sends and to control responses.
type fakeRemoteInfraServer struct {
	remoteinfra.UnimplementedEnvoyGatewayRemoteInfrastructureProviderServer

	mu               sync.Mutex
	createProxyReq   *remoteinfra.CreateOrUpdateProxyInfraRequest
	deleteProxyReq   *remoteinfra.DeleteProxyInfraRequest
	createRLReq      *remoteinfra.CreateOrUpdateRateLimitInfraRequest
	deleteRLReq      *remoteinfra.DeleteRateLimitInfraRequest
	createProxyErr   error
	deleteProxyErr   error
	createRLErr      error
	deleteRLErr      error
	createProxyCalls int
	deleteProxyCalls int
	createRLCalls    int
	deleteRLCalls    int
}

func (s *fakeRemoteInfraServer) CreateOrUpdateProxyInfra(_ context.Context, req *remoteinfra.CreateOrUpdateProxyInfraRequest) (*remoteinfra.CreateOrUpdateProxyInfraResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.createProxyCalls++
	s.createProxyReq = req
	if s.createProxyErr != nil {
		return nil, s.createProxyErr
	}
	return new(remoteinfra.CreateOrUpdateProxyInfraResponse{}), nil
}

func (s *fakeRemoteInfraServer) DeleteProxyInfra(_ context.Context, req *remoteinfra.DeleteProxyInfraRequest) (*remoteinfra.DeleteProxyInfraResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deleteProxyCalls++
	s.deleteProxyReq = req
	if s.deleteProxyErr != nil {
		return nil, s.deleteProxyErr
	}
	return new(remoteinfra.DeleteProxyInfraResponse{}), nil
}

func (s *fakeRemoteInfraServer) CreateOrUpdateRateLimitInfra(_ context.Context, req *remoteinfra.CreateOrUpdateRateLimitInfraRequest) (*remoteinfra.CreateOrUpdateRateLimitInfraResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.createRLCalls++
	s.createRLReq = req
	if s.createRLErr != nil {
		return nil, s.createRLErr
	}
	return new(remoteinfra.CreateOrUpdateRateLimitInfraResponse{}), nil
}

func (s *fakeRemoteInfraServer) DeleteRateLimitInfra(_ context.Context, req *remoteinfra.DeleteRateLimitInfraRequest) (*remoteinfra.DeleteRateLimitInfraResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deleteRLCalls++
	s.deleteRLReq = req
	if s.deleteRLErr != nil {
		return nil, s.deleteRLErr
	}
	return new(remoteinfra.DeleteRateLimitInfraResponse{}), nil
}

// newInMemoryInfraClient stands up an in-process gRPC server backed by
// bufconn and returns an infraClientImpl wired to it. Modeled after
// extension/registry.NewInMemoryManager.
func newInMemoryInfraClient(t *testing.T, srv remoteinfra.EnvoyGatewayRemoteInfrastructureProviderServer) *infraClientImpl {
	t.Helper()

	const buffer = 1024 * 1024
	lis := bufconn.Listen(buffer)

	baseServer := grpc.NewServer()
	remoteinfra.RegisterEnvoyGatewayRemoteInfrastructureProviderServer(baseServer, srv)
	go func() {
		_ = baseServer.Serve(lis)
	}()

	conn, err := grpc.NewClient(
		"passthrough://bufconn",
		grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = conn.Close()
		baseServer.Stop()
		_ = lis.Close()
	})

	return new(infraClientImpl{
		namespace:          "envoy-gateway-system",
		extensionConnCache: conn,
		client:             remoteinfra.NewEnvoyGatewayRemoteInfrastructureProviderClient(conn),
	})
}

func TestNewRemoteInfraClient(t *testing.T) {
	t.Run("populates_fields_from_config", func(t *testing.T) {
		// grpc.NewClient is lazy, so newRemoteInfraClient succeeds even though
		// the configured address has no listener.
		cfg, err := config.New(io.Discard, io.Discard)
		require.NoError(t, err)
		cfg.EnvoyGateway.Provider = new(egv1a1.EnvoyGatewayProvider{
			Type: egv1a1.ProviderTypeCustom,
			Custom: new(egv1a1.EnvoyGatewayCustomProvider{
				Resource: egv1a1.EnvoyGatewayResourceProvider{
					Type: egv1a1.ResourceProviderTypeKubernetes,
				},
				Infrastructure: new(egv1a1.EnvoyGatewayInfrastructureProvider{
					Type: egv1a1.InfrastructureProviderTypeRemote,
					Remote: new(egv1a1.EnvoyGatewayRemoteInfrastructureProvider{
						Service: new(egv1a1.ExtensionService{
							BackendEndpoint: egv1a1.BackendEndpoint{
								IP: new(egv1a1.IPEndpoint{Address: "127.0.0.1", Port: 1}),
							},
						}),
					}),
				}),
			}),
		})
		cfg.ControllerNamespace = "envoy-gateway-system"

		c, err := newRemoteInfraClient(context.Background(), cfg, nil)
		require.NoError(t, err)
		require.NotNil(t, c)

		impl, ok := c.(*infraClientImpl)
		require.True(t, ok)
		assert.Equal(t, "envoy-gateway-system", impl.namespace,
			"namespace should be sourced from cfg.ControllerNamespace")
		assert.NotNil(t, impl.remoteService)
		assert.NotNil(t, impl.client)
		assert.NotNil(t, impl.extensionConnCache)

		require.NoError(t, c.Close())
	})
}

func TestInfraClientImpl_Close(t *testing.T) {
	t.Run("closes_open_connection", func(t *testing.T) {
		impl := newInMemoryInfraClient(t, new(fakeRemoteInfraServer{}))
		require.NoError(t, impl.Close())
	})

	t.Run("nil_connection_is_a_noop", func(t *testing.T) {
		impl := new(infraClientImpl{})
		require.NoError(t, impl.Close())
	})
}

func TestInfraClientImpl_CreateOrUpdateProxyInfra(t *testing.T) {
	t.Run("success_sends_serialized_ir", func(t *testing.T) {
		srv := new(fakeRemoteInfraServer{})
		impl := newInMemoryInfraClient(t, srv)

		input := new(ir.Infra{Proxy: new(ir.ProxyInfra{Name: "proxy", Namespace: "ns"})})
		require.NoError(t, impl.CreateOrUpdateProxyInfra(context.Background(), input))

		srv.mu.Lock()
		defer srv.mu.Unlock()
		assert.Equal(t, 1, srv.createProxyCalls)
		require.NotNil(t, srv.createProxyReq)

		// The IR is sent as JSON bytes matching the IR's own JSONString output.
		assert.Equal(t, []byte(input.JSONString()), srv.createProxyReq.IrBytes)

		var got ir.Infra
		require.NoError(t, json.Unmarshal(srv.createProxyReq.IrBytes, &got))
		require.NotNil(t, got.Proxy)
		assert.Equal(t, "proxy", got.Proxy.Name)
		assert.Equal(t, "ns", got.Proxy.Namespace)
	})

	t.Run("server_error_propagates", func(t *testing.T) {
		srv := new(fakeRemoteInfraServer{
			createProxyErr: status.Error(codes.Internal, "boom"),
		})
		impl := newInMemoryInfraClient(t, srv)

		err := impl.CreateOrUpdateProxyInfra(context.Background(), new(ir.Infra{Proxy: new(ir.ProxyInfra{})}))
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok, "expected gRPC status error, got %T: %v", err, err)
		assert.Equal(t, codes.Internal, st.Code())
	})

	t.Run("context_cancellation_propagates", func(t *testing.T) {
		impl := newInMemoryInfraClient(t, new(fakeRemoteInfraServer{}))

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := impl.CreateOrUpdateProxyInfra(ctx, new(ir.Infra{Proxy: new(ir.ProxyInfra{})}))
		require.Error(t, err)
		assert.True(t,
			errors.Is(err, context.Canceled) || status.Code(err) == codes.Canceled,
			"expected canceled error, got: %v", err)
	})
}

func TestInfraClientImpl_DeleteProxyInfra(t *testing.T) {
	t.Run("success_sends_serialized_ir", func(t *testing.T) {
		srv := new(fakeRemoteInfraServer{})
		impl := newInMemoryInfraClient(t, srv)

		input := new(ir.Infra{Proxy: new(ir.ProxyInfra{Name: "proxy", Namespace: "ns"})})
		require.NoError(t, impl.DeleteProxyInfra(context.Background(), input))

		srv.mu.Lock()
		defer srv.mu.Unlock()
		assert.Equal(t, 1, srv.deleteProxyCalls)
		require.NotNil(t, srv.deleteProxyReq)
		assert.Equal(t, []byte(input.JSONString()), srv.deleteProxyReq.IrBytes)
	})

	t.Run("server_error_propagates", func(t *testing.T) {
		srv := new(fakeRemoteInfraServer{
			deleteProxyErr: status.Error(codes.NotFound, "missing"),
		})
		impl := newInMemoryInfraClient(t, srv)

		err := impl.DeleteProxyInfra(context.Background(), new(ir.Infra{Proxy: new(ir.ProxyInfra{})}))
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok, "expected gRPC status error, got %T: %v", err, err)
		assert.Equal(t, codes.NotFound, st.Code())
	})
}

func TestInfraClientImpl_CreateOrUpdateRateLimitInfra(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := new(fakeRemoteInfraServer{})
		impl := newInMemoryInfraClient(t, srv)

		require.NoError(t, impl.CreateOrUpdateRateLimitInfra(context.Background()))

		srv.mu.Lock()
		defer srv.mu.Unlock()
		assert.Equal(t, 1, srv.createRLCalls)
		require.NotNil(t, srv.createRLReq)
	})

	t.Run("server_error_propagates", func(t *testing.T) {
		srv := new(fakeRemoteInfraServer{
			createRLErr: status.Error(codes.PermissionDenied, "denied"),
		})
		impl := newInMemoryInfraClient(t, srv)

		err := impl.CreateOrUpdateRateLimitInfra(context.Background())
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok, "expected gRPC status error, got %T: %v", err, err)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})
}

func TestInfraClientImpl_DeleteRateLimitInfra(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := new(fakeRemoteInfraServer{})
		impl := newInMemoryInfraClient(t, srv)

		require.NoError(t, impl.DeleteRateLimitInfra(context.Background()))

		srv.mu.Lock()
		defer srv.mu.Unlock()
		assert.Equal(t, 1, srv.deleteRLCalls)
		require.NotNil(t, srv.deleteRLReq)
	})

	t.Run("server_error_propagates", func(t *testing.T) {
		srv := new(fakeRemoteInfraServer{
			deleteRLErr: status.Error(codes.Unimplemented, "not implemented"),
		})
		impl := newInMemoryInfraClient(t, srv)

		err := impl.DeleteRateLimitInfra(context.Background())
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok, "expected gRPC status error, got %T: %v", err, err)
		assert.Equal(t, codes.Unimplemented, st.Code())
	})
}

func TestGetClientConnCache(t *testing.T) {
	t.Run("returns_cached_conn_on_subsequent_calls", func(t *testing.T) {
		impl := newInMemoryInfraClient(t, new(fakeRemoteInfraServer{}))

		first := impl.extensionConnCache
		require.NotNil(t, first)

		conn, err := impl.getClientConnCache(context.Background())
		require.NoError(t, err)
		require.NotNil(t, conn)
		assert.Same(t, first, conn, "expected cached connection to be reused")
	})
}
