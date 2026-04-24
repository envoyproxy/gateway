// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package registry

import (
	"context"
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	extTypes "github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/proto/extension"
)

// NewInMemoryCompositeManager builds a CompositeManager with one namedManager per
// supplied ExtensionManager config. All entries share a single in-process gRPC server
// (bufconn) so tests exercise the composite code path without needing distinct servers.
// Callers tear down the bufconn/server by calling CleanupHookConns() on the returned
// Manager (idempotent — safe to call multiple times).
func NewInMemoryCompositeManager(
	exts []*egv1a1.ExtensionManager,
	server extension.EnvoyGatewayExtensionServer,
) (extTypes.Manager, error) {
	if server == nil {
		return nil, fmt.Errorf("in-memory composite manager must be passed a server")
	}
	if len(exts) == 0 {
		return nil, fmt.Errorf("in-memory composite manager requires at least one extension")
	}

	buffer := 10 * 1024 * 1024
	lis := bufconn.Listen(buffer)

	baseServer := grpc.NewServer()
	extension.RegisterEnvoyGatewayExtensionServer(baseServer, server)
	go func() {
		_ = baseServer.Serve(lis)
	}()

	conn, err := grpc.DialContext(context.Background(), "",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		baseServer.Stop()
		lis.Close()
		return nil, err
	}

	// All entries share a single bufconn/server, so the cleanup must run at most
	// once even though it's wired to every entry's cleanupHookConn (and
	// CompositeManager.CleanupHookConns invokes every entry's callback).
	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			_ = conn.Close()
			baseServer.Stop()
			lis.Close()
		})
	}

	named := make([]namedManager, 0, len(exts))
	for _, ext := range exts {
		mgr := &Manager{
			extension:          *ext,
			extensionConnCache: conn,
		}
		resourceGKSet, policyGKSet := buildManagerGKSets(ext)
		named = append(named, namedManager{
			name:            ext.Name,
			manager:         mgr,
			resourceGKSet:   resourceGKSet,
			policyGKSet:     policyGKSet,
			cleanupHookConn: cleanup,
		})
	}

	return NewCompositeManager(named), nil
}
