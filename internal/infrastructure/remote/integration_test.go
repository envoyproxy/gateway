//go:build integration

// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package remote

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"sync"
	"testing"
	"time"

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
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/proto/remoteinfra"
)

const (
	integrationDefaultWait = time.Second * 5
	integrationDefaultTick = time.Millisecond * 20
	integrationBufSize     = 1024 * 1024
)

// recordingRemoteInfraServer is a gRPC server implementation used by the
// integration test. It records what was received on each RPC and supports
// returning configured errors so the failure path can be exercised end to
// end.
type recordingRemoteInfraServer struct {
	remoteinfra.UnimplementedEnvoyGatewayRemoteInfrastructureProviderServer

	mu sync.Mutex

	createProxyCalls int
	deleteProxyCalls int
	createRLCalls    int
	deleteRLCalls    int

	lastCreateProxyReq *remoteinfra.CreateOrUpdateProxyInfraRequest
	lastDeleteProxyReq *remoteinfra.DeleteProxyInfraRequest

	createProxyErr error
	deleteProxyErr error
	createRLErr    error
	deleteRLErr    error
}

func (s *recordingRemoteInfraServer) CreateOrUpdateProxyInfra(_ context.Context, req *remoteinfra.CreateOrUpdateProxyInfraRequest) (*remoteinfra.CreateOrUpdateProxyInfraResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.createProxyCalls++
	s.lastCreateProxyReq = req
	if s.createProxyErr != nil {
		return nil, s.createProxyErr
	}
	return new(remoteinfra.CreateOrUpdateProxyInfraResponse{}), nil
}

func (s *recordingRemoteInfraServer) DeleteProxyInfra(_ context.Context, req *remoteinfra.DeleteProxyInfraRequest) (*remoteinfra.DeleteProxyInfraResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deleteProxyCalls++
	s.lastDeleteProxyReq = req
	if s.deleteProxyErr != nil {
		return nil, s.deleteProxyErr
	}
	return new(remoteinfra.DeleteProxyInfraResponse{}), nil
}

func (s *recordingRemoteInfraServer) CreateOrUpdateRateLimitInfra(_ context.Context, _ *remoteinfra.CreateOrUpdateRateLimitInfraRequest) (*remoteinfra.CreateOrUpdateRateLimitInfraResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.createRLCalls++
	if s.createRLErr != nil {
		return nil, s.createRLErr
	}
	return new(remoteinfra.CreateOrUpdateRateLimitInfraResponse{}), nil
}

func (s *recordingRemoteInfraServer) DeleteRateLimitInfra(_ context.Context, _ *remoteinfra.DeleteRateLimitInfraRequest) (*remoteinfra.DeleteRateLimitInfraResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deleteRLCalls++
	if s.deleteRLErr != nil {
		return nil, s.deleteRLErr
	}
	return new(remoteinfra.DeleteRateLimitInfraResponse{}), nil
}

// snapshot returns a copy of the call counts under lock so tests can read
// them safely from another goroutine.
func (s *recordingRemoteInfraServer) snapshot() (createProxy, deleteProxy, createRL, deleteRL int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.createProxyCalls, s.deleteProxyCalls, s.createRLCalls, s.deleteRLCalls
}

// startLocalRemoteInfraServer brings up a gRPC server backed by an
// in-memory bufconn listener and returns an InfraClientFactory that dials
// it. Modeled after extension/registry.NewInMemoryManager. The server is
// shut down via t.Cleanup.
func startLocalRemoteInfraServer(t *testing.T, srv *recordingRemoteInfraServer) InfraClientFactory {
	t.Helper()

	lis := bufconn.Listen(integrationBufSize)

	baseServer := grpc.NewServer()
	remoteinfra.RegisterEnvoyGatewayRemoteInfrastructureProviderServer(baseServer, srv)

	served := make(chan struct{})
	go func() {
		defer close(served)
		_ = baseServer.Serve(lis)
	}()

	t.Cleanup(func() {
		baseServer.GracefulStop()
		_ = lis.Close()
		<-served
	})

	return func(_ context.Context) (InfraClient, error) {
		conn, err := grpc.NewClient(
			"passthrough://bufconn",
			grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) {
				return lis.Dial()
			}),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, err
		}
		return new(infraClientImpl{
			namespace:          config.DefaultNamespace,
			extensionConnCache: conn,
			client:             remoteinfra.NewEnvoyGatewayRemoteInfrastructureProviderClient(conn),
		}), nil
	}
}

func newIntegrationConfig(t *testing.T) *config.Server {
	t.Helper()

	cfg, err := config.New(io.Discard, io.Discard)
	require.NoError(t, err)

	cfg.ControllerNamespace = config.DefaultNamespace
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

	return cfg
}

// newTestInfra returns an *Infra wired to the provided factory.
func newTestInfra(t *testing.T, cfg *config.Server, factory InfraClientFactory) *Infra {
	t.Helper()

	notifier := message.RunnerErrorNotifier{
		RunnerName:   "infrastructure",
		RunnerErrors: new(message.RunnerErrors{}),
	}

	infra := NewInfra(cfg, factory, notifier)
	t.Cleanup(func() {
		_ = infra.Close()
	})
	return infra
}

// TestRemoteInfraIntegration drives the Infra wrapper end to end against an
// in-memory gRPC server hosted in the same process via bufconn. Each
// subtest exercises one of the four RPCs exposed by the remote
// infrastructure provider.
func TestRemoteInfraIntegration(t *testing.T) {
	srv := new(recordingRemoteInfraServer{})
	factory := startLocalRemoteInfraServer(t, srv)

	cfg := newIntegrationConfig(t)
	infra := newTestInfra(t, cfg, factory)

	t.Run("create_or_update_proxy_infra", func(t *testing.T) {
		input := new(ir.Infra{
			Proxy: new(ir.ProxyInfra{
				Name:      "test-proxy",
				Namespace: "envoy-gateway-system",
			}),
		})

		require.NoError(t, infra.CreateOrUpdateProxyInfra(t.Context(), input))

		// Eventually because the gRPC handler runs in a separate goroutine
		// on the server side. In practice this is observed synchronously,
		// but we don't want to assume that.
		require.Eventually(t, func() bool {
			c, _, _, _ := srv.snapshot()
			return c == 1
		}, integrationDefaultWait, integrationDefaultTick,
			"timed out waiting for server to observe CreateOrUpdateProxyInfra")

		srv.mu.Lock()
		got := srv.lastCreateProxyReq
		srv.mu.Unlock()
		require.NotNil(t, got)

		// IR is sent as JSON bytes matching the IR's own JSONString output.
		assert.Equal(t, []byte(input.JSONString()), got.IrBytes)

		var roundTripped ir.Infra
		require.NoError(t, json.Unmarshal(got.IrBytes, &roundTripped))
		require.NotNil(t, roundTripped.Proxy)
		assert.Equal(t, "test-proxy", roundTripped.Proxy.Name)
		assert.Equal(t, "envoy-gateway-system", roundTripped.Proxy.Namespace)
	})

	t.Run("delete_proxy_infra", func(t *testing.T) {
		input := new(ir.Infra{
			Proxy: new(ir.ProxyInfra{
				Name:      "test-proxy",
				Namespace: "envoy-gateway-system",
			}),
		})

		require.NoError(t, infra.DeleteProxyInfra(t.Context(), input))

		require.Eventually(t, func() bool {
			_, d, _, _ := srv.snapshot()
			return d == 1
		}, integrationDefaultWait, integrationDefaultTick,
			"timed out waiting for server to observe DeleteProxyInfra")

		srv.mu.Lock()
		got := srv.lastDeleteProxyReq
		srv.mu.Unlock()
		require.NotNil(t, got)
		assert.Equal(t, []byte(input.JSONString()), got.IrBytes)
	})

	t.Run("create_or_update_rate_limit_infra", func(t *testing.T) {
		require.NoError(t, infra.CreateOrUpdateRateLimitInfra(t.Context()))

		require.Eventually(t, func() bool {
			_, _, c, _ := srv.snapshot()
			return c == 1
		}, integrationDefaultWait, integrationDefaultTick,
			"timed out waiting for server to observe CreateOrUpdateRateLimitInfra")
	})

	t.Run("delete_rate_limit_infra", func(t *testing.T) {
		require.NoError(t, infra.DeleteRateLimitInfra(t.Context()))

		require.Eventually(t, func() bool {
			_, _, _, d := srv.snapshot()
			return d == 1
		}, integrationDefaultWait, integrationDefaultTick,
			"timed out waiting for server to observe DeleteRateLimitInfra")
	})

	t.Run("client_is_built_lazily_and_reused", func(t *testing.T) {
		// At this point the four prior subtests have already forced lazy
		// construction. Issue another call and verify the call counters
		// continue to advance, which confirms the same client is reused
		// across calls (no flake from a fresh dial).
		before, _, _, _ := srv.snapshot()
		require.NoError(t, infra.CreateOrUpdateProxyInfra(t.Context(), new(ir.Infra{
			Proxy: new(ir.ProxyInfra{Name: "p2", Namespace: "ns2"}),
		})))
		require.Eventually(t, func() bool {
			now, _, _, _ := srv.snapshot()
			return now == before+1
		}, integrationDefaultWait, integrationDefaultTick)
	})
}

// TestRemoteInfraIntegration_ServerErrorPropagates verifies that an error
// returned by the remote provider surfaces back to the caller of Infra,
// preserving the gRPC status code.
func TestRemoteInfraIntegration_ServerErrorPropagates(t *testing.T) {
	srv := new(recordingRemoteInfraServer{
		createProxyErr: status.Error(codes.Internal, "boom"),
	})
	factory := startLocalRemoteInfraServer(t, srv)

	cfg := newIntegrationConfig(t)
	infra := newTestInfra(t, cfg, factory)

	err := infra.CreateOrUpdateProxyInfra(t.Context(), new(ir.Infra{
		Proxy: new(ir.ProxyInfra{Name: "p", Namespace: "ns"}),
	}))
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok, "expected gRPC status error, got %T: %v", err, err)
	assert.Equal(t, codes.Internal, st.Code())
}
