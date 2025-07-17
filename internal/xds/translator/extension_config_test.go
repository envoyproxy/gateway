// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"context"
	"testing"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	resourceTypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/extension/registry"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
	"github.com/envoyproxy/gateway/proto/extension"
)

// mockExtensionServer implements the extension server interface for testing
type mockExtensionServer struct {
	extension.UnimplementedEnvoyGatewayExtensionServer
	receivedListeners []*listenerv3.Listener
	receivedRoutes    []*routev3.RouteConfiguration
}

func (m *mockExtensionServer) PostTranslateModify(ctx context.Context, req *extension.PostTranslateModifyRequest) (*extension.PostTranslateModifyResponse, error) {
	// Store what we received for verification
	m.receivedListeners = req.Listeners
	m.receivedRoutes = req.Routes

	// Return the same resources
	return &extension.PostTranslateModifyResponse{
		Clusters:  req.Clusters,
		Secrets:   req.Secrets,
		Listeners: req.Listeners,
		Routes:    req.Routes,
	}, nil
}

func TestProcessExtensionPostTranslationHookConfig(t *testing.T) {
	tests := []struct {
		name                     string
		enableListenersAndRoutes *bool
		expectListenersAndRoutes bool
	}{
		{
			name:                     "default behavior (disabled for backward compatibility)",
			enableListenersAndRoutes: nil,
			expectListenersAndRoutes: false,
		},
		{
			name:                     "explicitly enabled",
			enableListenersAndRoutes: ptr.To(true),
			expectListenersAndRoutes: true,
		},
		{
			name:                     "explicitly disabled",
			enableListenersAndRoutes: ptr.To(false),
			expectListenersAndRoutes: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock extension server
			mockServer := &mockExtensionServer{}

			// Create extension manager with the test configuration
			extManager := egv1a1.ExtensionManager{
				Hooks: &egv1a1.ExtensionHooks{
					XDSTranslator: &egv1a1.XDSTranslatorHooks{
						Post: []egv1a1.XDSTranslatorHook{egv1a1.XDSTranslation},
						Translation: &egv1a1.TranslationConfig{
							IncludeAll: tt.enableListenersAndRoutes,
						},
					},
				},
				Service: &egv1a1.ExtensionService{
					BackendEndpoint: egv1a1.BackendEndpoint{
						FQDN: &egv1a1.FQDNEndpoint{
							Hostname: "test.example.com",
							Port:     8080,
						},
					},
				},
			}

			mgr, cleanup, err := registry.NewInMemoryManager(extManager, mockServer)
			require.NoError(t, err)
			defer cleanup()

			// Create test resource table with sample resources
			tCtx := &types.ResourceVersionTable{
				XdsResources: map[string][]resourceTypes.Resource{
					resourcev3.ClusterType: {
						&clusterv3.Cluster{Name: "test-cluster"},
					},
					resourcev3.SecretType: {
						&tlsv3.Secret{Name: "test-secret"},
					},
					resourcev3.ListenerType: {
						&listenerv3.Listener{Name: "test-listener"},
					},
					resourcev3.RouteType: {
						&routev3.RouteConfiguration{Name: "test-route"},
					},
				},
			}

			// Call the function under test
			err = processExtensionPostTranslationHook(tCtx, &mgr, []*ir.UnstructuredRef{})
			require.NoError(t, err)

			// Verify the behavior based on configuration
			if tt.expectListenersAndRoutes {
				// Should have received listeners and routes
				require.Len(t, mockServer.receivedListeners, 1)
				require.Equal(t, "test-listener", mockServer.receivedListeners[0].Name)
				require.Len(t, mockServer.receivedRoutes, 1)
				require.Equal(t, "test-route", mockServer.receivedRoutes[0].Name)
			} else {
				// Should not have received listeners and routes (nil or empty)
				require.Empty(t, mockServer.receivedListeners)
				require.Empty(t, mockServer.receivedRoutes)
			}
		})
	}
}
