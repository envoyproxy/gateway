// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubejwt

import (
	"context"
	"fmt"
	"io"
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
)

func newTestInterceptor(t *testing.T) *JWTAuthInterceptor {
	t.Helper()
	logger := logging.DefaultLogger(io.Discard, egv1a1.LogLevelInfo)
	return &JWTAuthInterceptor{
		logger: logger.WithName("test"),
	}
}

// mockServerStream implements grpc.ServerStream for testing.
type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context { return m.ctx }
func (m *mockServerStream) RecvMsg(_ any) error      { return nil }

// authTestCase defines a single test case for the authenticate method.
// The same cases are run against both the stream (RecvMsg) and unary interceptor paths.
type authTestCase struct {
	name    string
	msg     any
	ctx     context.Context
	wantErr string
}

func authTestCases() []authTestCase {
	return []authTestCase{
		{
			name:    "unknown message type is rejected (fail-closed)",
			msg:     &discoveryv3.DiscoveryResponse{},
			ctx:     context.Background(),
			wantErr: "unexpected message type",
		},
		{
			name:    "nil message is rejected",
			msg:     nil,
			ctx:     context.Background(),
			wantErr: "unexpected message type",
		},
		{
			name:    "DeltaDiscoveryRequest with nil node is rejected",
			msg:     &discoveryv3.DeltaDiscoveryRequest{},
			ctx:     context.Background(),
			wantErr: "missing node ID",
		},
		{
			name:    "DiscoveryRequest with nil node is rejected",
			msg:     &discoveryv3.DiscoveryRequest{},
			ctx:     context.Background(),
			wantErr: "missing node ID",
		},
		{
			name: "DeltaDiscoveryRequest with empty node ID is rejected",
			msg: &discoveryv3.DeltaDiscoveryRequest{
				Node: &corev3.Node{Id: ""},
			},
			ctx:     context.Background(),
			wantErr: "missing node ID",
		},
		{
			name: "DiscoveryRequest with empty node ID is rejected",
			msg: &discoveryv3.DiscoveryRequest{
				Node: &corev3.Node{Id: ""},
			},
			ctx:     context.Background(),
			wantErr: "missing node ID",
		},
		{
			name: "DeltaDiscoveryRequest without metadata is rejected",
			msg: &discoveryv3.DeltaDiscoveryRequest{
				Node: &corev3.Node{Id: "pod-1"},
			},
			ctx:     context.Background(),
			wantErr: "missing metadata",
		},
		{
			name: "DiscoveryRequest without metadata is rejected",
			msg: &discoveryv3.DiscoveryRequest{
				Node: &corev3.Node{Id: "pod-1"},
			},
			ctx:     context.Background(),
			wantErr: "missing metadata",
		},
		{
			name: "DeltaDiscoveryRequest without auth header is rejected",
			msg: &discoveryv3.DeltaDiscoveryRequest{
				Node: &corev3.Node{Id: "pod-1"},
			},
			ctx:     metadata.NewIncomingContext(context.Background(), metadata.MD{}),
			wantErr: "missing authorization token",
		},
		{
			name: "DiscoveryRequest without auth header is rejected",
			msg: &discoveryv3.DiscoveryRequest{
				Node: &corev3.Node{Id: "pod-1"},
			},
			ctx:     metadata.NewIncomingContext(context.Background(), metadata.MD{}),
			wantErr: "missing authorization token",
		},
	}
}

func TestAuthenticate_Stream(t *testing.T) {
	interceptor := newTestInterceptor(t)

	for _, tt := range authTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockServerStream{ctx: tt.ctx}
			ws := &wrappedStream{
				ServerStream: mock,
				ctx:          tt.ctx,
				interceptor:  interceptor,
				validated:    false,
			}

			err := ws.RecvMsg(tt.msg)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestAuthenticate_Unary(t *testing.T) {
	interceptor := newTestInterceptor(t)
	unary := interceptor.Unary()
	handler := func(_ context.Context, _ any) (any, error) {
		return nil, nil
	}

	for _, tt := range authTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			_, err := unary(tt.ctx, tt.msg, &grpc.UnaryServerInfo{}, handler)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestExtractNodeID(t *testing.T) {
	tests := []struct {
		name    string
		msg     any
		want    string
		wantErr string
	}{
		{
			name: "DeltaDiscoveryRequest with valid node ID",
			msg:  &discoveryv3.DeltaDiscoveryRequest{Node: &corev3.Node{Id: "pod-1"}},
			want: "pod-1",
		},
		{
			name: "DiscoveryRequest with valid node ID",
			msg:  &discoveryv3.DiscoveryRequest{Node: &corev3.Node{Id: "pod-2"}},
			want: "pod-2",
		},
		{
			name:    "DeltaDiscoveryRequest with nil node",
			msg:     &discoveryv3.DeltaDiscoveryRequest{},
			wantErr: "missing node ID",
		},
		{
			name:    "DeltaDiscoveryRequest with empty node ID",
			msg:     &discoveryv3.DeltaDiscoveryRequest{Node: &corev3.Node{Id: ""}},
			wantErr: "missing node ID",
		},
		{
			name:    "DiscoveryRequest with nil node",
			msg:     &discoveryv3.DiscoveryRequest{},
			wantErr: "missing node ID",
		},
		{
			name:    "DiscoveryRequest with empty node ID",
			msg:     &discoveryv3.DiscoveryRequest{Node: &corev3.Node{Id: ""}},
			wantErr: "missing node ID",
		},
		{
			name:    "unknown message type is rejected (fail-closed)",
			msg:     "not-a-discovery-request",
			wantErr: "unexpected message type",
		},
		{
			name:    "nil message is rejected",
			msg:     nil,
			wantErr: "unexpected message type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractNodeID(tt.msg)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRecvMsg(t *testing.T) {
	interceptor := newTestInterceptor(t)

	mock := &mockServerStream{ctx: context.Background()}
	ws := &wrappedStream{
		ServerStream: mock,
		ctx:          context.Background(),
		interceptor:  interceptor,
		validated:    true,
	}

	// Passes when already validated, even without metadata.
	err := ws.RecvMsg(&discoveryv3.DeltaDiscoveryRequest{
		Node: &corev3.Node{Id: "pod-1"},
	})
	require.NoError(t, err)
}

func TestStreamInterceptor(t *testing.T) {
	tests := []struct {
		name          string
		handler       func(t *testing.T) grpc.StreamHandler
		wantErr       error
		checkWrapped  bool
		handlerCalled bool
	}{
		{
			name: "wraps stream with authentication",
			handler: func(t *testing.T) grpc.StreamHandler {
				return func(_ any, stream grpc.ServerStream) error {
					_, ok := stream.(*wrappedStream)
					assert.True(t, ok, "stream should be wrapped")
					return nil
				}
			},
			checkWrapped:  true,
			handlerCalled: true,
		},
		{
			name: "propagates handler error",
			handler: func(_ *testing.T) grpc.StreamHandler {
				return func(_ any, _ grpc.ServerStream) error {
					return fmt.Errorf("handler error")
				}
			},
			wantErr:       fmt.Errorf("handler error"),
			handlerCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := newTestInterceptor(t)
			streamInterceptor := interceptor.Stream()
			mock := &mockServerStream{ctx: context.Background()}

			var called bool
			wrappedHandler := func(srv any, stream grpc.ServerStream) error {
				called = true
				return tt.handler(t)(srv, stream)
			}

			err := streamInterceptor(nil, mock, &grpc.StreamServerInfo{}, wrappedHandler)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.handlerCalled, called)
		})
	}
}
