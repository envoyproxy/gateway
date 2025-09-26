// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/extension/registry"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
	pb "github.com/envoyproxy/gateway/proto/extension"
)

func TestRunner(t *testing.T) {
	// Setup
	pResources := new(message.ProviderResources)
	xdsIR := new(message.XdsIR)
	infraIR := new(message.InfraIR)
	cfg, err := config.New(os.Stdout)
	require.NoError(t, err)
	extMgr, closeFunc, err := registry.NewInMemoryManager(&egv1a1.ExtensionManager{}, &pb.UnimplementedEnvoyGatewayExtensionServer{})
	require.NoError(t, err)
	defer closeFunc()
	r := New(&Config{
		Server:            *cfg,
		ProviderResources: pResources,
		XdsIR:             xdsIR,
		InfraIR:           infraIR,
		ExtensionManager:  extMgr,
	})
	ctx := context.Background()
	// Start
	err = r.Start(ctx)
	require.NoError(t, err)

	// IR is nil at start
	require.Equal(t, map[string]*ir.Xds{}, xdsIR.LoadAll())
	require.Equal(t, map[string]*ir.Infra{}, infraIR.LoadAll())

	// TODO: pass valid provider resources

	// Delete gateway from the map.
	pResources.GatewayAPIResources.Delete("test")
	require.Eventually(t, func() bool {
		out := xdsIR.LoadAll()
		if out == nil {
			return false
		}
		// Ensure ir is empty
		return (reflect.DeepEqual(xdsIR.LoadAll(), map[string]*ir.Xds{})) && (reflect.DeepEqual(infraIR.LoadAll(), map[string]*ir.Infra{}))
	}, time.Second*1, time.Millisecond*20)
}

// setupTestRunner creates a test runner with populated stores and keyCache
func setupTestRunner(t *testing.T) (*Runner, []types.NamespacedName) {
	pResources := new(message.ProviderResources)
	xdsIR := new(message.XdsIR)
	infraIR := new(message.InfraIR)
	cfg, err := config.New(os.Stdout)
	require.NoError(t, err)
	extMgr, closeFunc, err := registry.NewInMemoryManager(&egv1a1.ExtensionManager{}, &pb.UnimplementedEnvoyGatewayExtensionServer{})
	require.NoError(t, err)
	t.Cleanup(closeFunc)

	r := New(&Config{
		Server:            *cfg,
		ProviderResources: pResources,
		XdsIR:             xdsIR,
		InfraIR:           infraIR,
		ExtensionManager:  extMgr,
	})

	// Store test data in IR and status stores
	r.InfraIR.Store("test-ir-1", nil)
	r.InfraIR.Store("test-ir-2", nil)
	r.XdsIR.Store("test-ir-1", nil)
	r.XdsIR.Store("test-ir-2", nil)

	keys := []types.NamespacedName{
		{Name: "gateway1", Namespace: "test-namespace"},
		{Name: "httproute1", Namespace: "test-namespace"},
		{Name: "grpcroute1", Namespace: "test-namespace"},
		{Name: "tlsroute1", Namespace: "test-namespace"},
		{Name: "tcproute1", Namespace: "test-namespace"},
		{Name: "udproute1", Namespace: "test-namespace"},
		{Name: "udproute2", Namespace: "test-namespace"},
		{Name: "backend1", Namespace: "test-namespace"},
		{Name: "backendtls1", Namespace: "test-namespace"},
		{Name: "clientpolicy1", Namespace: "test-namespace"},
		{Name: "backendpolicy1", Namespace: "test-namespace"},
		{Name: "security1", Namespace: "test-namespace"},
		{Name: "envoyext1", Namespace: "test-namespace"},
	}

	// Store various status types
	r.ProviderResources.GatewayStatuses.Store(keys[0], &gwapiv1.GatewayStatus{})
	r.ProviderResources.HTTPRouteStatuses.Store(keys[1], &gwapiv1.HTTPRouteStatus{})
	r.ProviderResources.GRPCRouteStatuses.Store(keys[2], &gwapiv1.GRPCRouteStatus{})
	r.ProviderResources.TLSRouteStatuses.Store(keys[3], &gwapiv1a2.TLSRouteStatus{})
	r.ProviderResources.TCPRouteStatuses.Store(keys[4], &gwapiv1a2.TCPRouteStatus{})
	r.ProviderResources.UDPRouteStatuses.Store(keys[5], &gwapiv1a2.UDPRouteStatus{})
	r.ProviderResources.UDPRouteStatuses.Store(keys[6], &gwapiv1a2.UDPRouteStatus{})
	r.ProviderResources.BackendStatuses.Store(keys[7], &egv1a1.BackendStatus{})
	r.ProviderResources.BackendTLSPolicyStatuses.Store(keys[8], &gwapiv1a2.PolicyStatus{})
	r.ProviderResources.ClientTrafficPolicyStatuses.Store(keys[9], &gwapiv1a2.PolicyStatus{})
	r.ProviderResources.BackendTrafficPolicyStatuses.Store(keys[10], &gwapiv1a2.PolicyStatus{})
	r.ProviderResources.SecurityPolicyStatuses.Store(keys[11], &gwapiv1a2.PolicyStatus{})
	r.ProviderResources.EnvoyExtensionPolicyStatuses.Store(keys[12], &gwapiv1a2.PolicyStatus{})

	return r, keys
}

// verifyInitialState checks that all stores and keyCache are properly populated
func verifyInitialState(t *testing.T, r *Runner) {
	require.Equal(t, 2, r.InfraIR.Len())
	require.Equal(t, 2, r.XdsIR.Len())
	require.Equal(t, 1, r.ProviderResources.GatewayStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.HTTPRouteStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.GRPCRouteStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.TLSRouteStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.TCPRouteStatuses.Len())
	require.Equal(t, 2, r.ProviderResources.UDPRouteStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.BackendStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.BackendTLSPolicyStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.ClientTrafficPolicyStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.BackendTrafficPolicyStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.SecurityPolicyStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.EnvoyExtensionPolicyStatuses.Len())
}

func TestDeleteKeys(t *testing.T) {
	r, keys := setupTestRunner(t)
	verifyInitialState(t, r)

	// Create KeyCache with subset of keys to delete (selective deletion)
	keysToDelete := newKeyCache()
	keysToDelete.IR["test-ir-1"] = true // Delete only one IR key
	keysToDelete.GatewayStatus[keys[0]] = true
	keysToDelete.HTTPRouteStatus[keys[1]] = true
	keysToDelete.TLSRouteStatus[keys[3]] = true
	keysToDelete.UDPRouteStatus[keys[5]] = true // Delete only one UDP route
	keysToDelete.BackendTLSPolicyStatus[keys[8]] = true
	keysToDelete.SecurityPolicyStatus[keys[11]] = true
	// Leave some keys to verify selective deletion works

	// Test selective deletion
	r.deleteKeys(keysToDelete)

	// Verify selective deletion worked
	require.Equal(t, 1, r.InfraIR.Len()) // One IR key should remain
	require.Equal(t, 1, r.XdsIR.Len())   // One XDS key should remain
	require.Equal(t, 0, r.ProviderResources.GatewayStatuses.Len())
	require.Equal(t, 0, r.ProviderResources.HTTPRouteStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.GRPCRouteStatuses.Len()) // Should remain
	require.Equal(t, 0, r.ProviderResources.TLSRouteStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.TCPRouteStatuses.Len()) // Should remain
	require.Equal(t, 1, r.ProviderResources.UDPRouteStatuses.Len()) // One should remain
	require.Equal(t, 1, r.ProviderResources.BackendStatuses.Len())  // Should remain
	require.Equal(t, 0, r.ProviderResources.BackendTLSPolicyStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.ClientTrafficPolicyStatuses.Len())  // Should remain
	require.Equal(t, 1, r.ProviderResources.BackendTrafficPolicyStatuses.Len()) // Should remain
	require.Equal(t, 0, r.ProviderResources.SecurityPolicyStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.EnvoyExtensionPolicyStatuses.Len()) // Should remain

	// Verify keyCache was updated correctly
	require.False(t, r.keyCache.IR["test-ir-1"])
	require.True(t, r.keyCache.IR["test-ir-2"]) // Should remain
	require.False(t, r.keyCache.GatewayStatus[keys[0]])
	require.False(t, r.keyCache.HTTPRouteStatus[keys[1]])
	require.True(t, r.keyCache.GRPCRouteStatus[keys[2]]) // Should remain
	require.False(t, r.keyCache.TLSRouteStatus[keys[3]])
	require.True(t, r.keyCache.TCPRouteStatus[keys[4]]) // Should remain
	require.False(t, r.keyCache.UDPRouteStatus[keys[5]])
	require.True(t, r.keyCache.UDPRouteStatus[keys[6]]) // Should remain
	require.True(t, r.keyCache.BackendStatus[keys[7]])  // Should remain
	require.False(t, r.keyCache.BackendTLSPolicyStatus[keys[8]])
	require.True(t, r.keyCache.ClientTrafficPolicyStatus[keys[9]])   // Should remain
	require.True(t, r.keyCache.BackendTrafficPolicyStatus[keys[10]]) // Should remain
	require.False(t, r.keyCache.SecurityPolicyStatus[keys[11]])
	require.True(t, r.keyCache.EnvoyExtensionPolicyStatus[keys[12]]) // Should remain
}

func TestDeleteAllKeys(t *testing.T) {
	r, _ := setupTestRunner(t)
	verifyInitialState(t, r)

	// Test deleteAllKeys functionality
	r.deleteAllKeys()

	// Verify everything is deleted
	require.Equal(t, 0, r.InfraIR.Len())
	require.Equal(t, 0, r.XdsIR.Len())
	require.Equal(t, 0, r.ProviderResources.GatewayStatuses.Len())
	require.Equal(t, 0, r.ProviderResources.HTTPRouteStatuses.Len())
	require.Equal(t, 0, r.ProviderResources.GRPCRouteStatuses.Len())
	require.Equal(t, 0, r.ProviderResources.TLSRouteStatuses.Len())
	require.Equal(t, 0, r.ProviderResources.TCPRouteStatuses.Len())
	require.Equal(t, 0, r.ProviderResources.UDPRouteStatuses.Len())
	require.Equal(t, 0, r.ProviderResources.BackendStatuses.Len())
	require.Equal(t, 0, r.ProviderResources.BackendTLSPolicyStatuses.Len())
	require.Equal(t, 0, r.ProviderResources.ClientTrafficPolicyStatuses.Len())
	require.Equal(t, 0, r.ProviderResources.BackendTrafficPolicyStatuses.Len())
	require.Equal(t, 0, r.ProviderResources.SecurityPolicyStatuses.Len())
	require.Equal(t, 0, r.ProviderResources.EnvoyExtensionPolicyStatuses.Len())

	// Verify keyCache is reset
	require.Empty(t, r.keyCache.IR)
	require.Empty(t, r.keyCache.GatewayStatus)
	require.Empty(t, r.keyCache.HTTPRouteStatus)
	require.Empty(t, r.keyCache.GRPCRouteStatus)
	require.Empty(t, r.keyCache.TLSRouteStatus)
	require.Empty(t, r.keyCache.TCPRouteStatus)
	require.Empty(t, r.keyCache.UDPRouteStatus)
	require.Empty(t, r.keyCache.BackendStatus)
	require.Empty(t, r.keyCache.BackendTLSPolicyStatus)
	require.Empty(t, r.keyCache.ClientTrafficPolicyStatus)
	require.Empty(t, r.keyCache.BackendTrafficPolicyStatus)
	require.Empty(t, r.keyCache.SecurityPolicyStatus)
	require.Empty(t, r.keyCache.EnvoyExtensionPolicyStatus)
}
