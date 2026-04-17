// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package registry

import (
	"context"
	"fmt"
	"net"

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
// Returns the composite Manager, a cleanup func that tears down the gRPC server and
// connection, and any construction error.
func NewInMemoryCompositeManager(
	exts []*egv1a1.ExtensionManager,
	server extension.EnvoyGatewayExtensionServer,
) (extTypes.Manager, func(), error) {
	if server == nil {
		return nil, nil, fmt.Errorf("in-memory composite manager must be passed a server")
	}
	if len(exts) == 0 {
		return nil, nil, fmt.Errorf("in-memory composite manager requires at least one extension")
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
		return nil, nil, err
	}

	named := make([]namedManager, 0, len(exts))
	for _, ext := range exts {
		mgr := &Manager{
			extension:          *ext,
			extensionConnCache: conn,
		}
		resourceGVKSet, policyGVKSet := buildManagerGVKSets(ext)
		named = append(named, namedManager{
			name:           ext.Name,
			manager:        mgr,
			resourceGVKSet: resourceGVKSet,
			policyGVKSet:   policyGVKSet,
		})
	}

	cleanup := func() {
		_ = conn.Close()
		baseServer.Stop()
		lis.Close()
	}

	return NewCompositeManager(named), cleanup, nil
}
