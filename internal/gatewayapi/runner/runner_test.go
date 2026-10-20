// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"crypto/tls"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/telepresenceio/watchable"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/crypto"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/extension/registry"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/message"
	pb "github.com/envoyproxy/gateway/proto/extension"
)

func TestRunner(t *testing.T) {
	// Setup
	pResources := new(message.ProviderResources)
	xdsIR := new(message.XdsIR)
	infraIR := new(message.InfraIR)
	cfg, err := config.New(os.Stdout, os.Stderr)
	require.NoError(t, err)
	extMgr, closeFunc, err := registry.NewInMemoryManager(&egv1a1.ExtensionManager{}, &pb.UnimplementedEnvoyGatewayExtensionServer{})
	require.NoError(t, err)
	defer closeFunc()
	r := New(&Config{
		Server:            *cfg,
		ProviderResources: pResources,
		RunnerErrors:      new(message.RunnerErrors),
		XdsIR:             xdsIR,
		InfraIR:           infraIR,
		ExtensionManager:  extMgr,
	})
	// Start
	err = r.Start(t.Context())
	require.NoError(t, err)

	// IR is nil at start
	require.Equal(t, map[string]*message.XdsIRWithContext{}, xdsIR.LoadAll())
	require.Equal(t, map[string]*message.InfraIRWithContext{}, infraIR.LoadAll())

	// TODO: pass valid provider resources

	// Delete gateway from the map.
	pResources.GatewayAPIResources.Delete("test")
	require.Eventually(t, func() bool {
		out := xdsIR.LoadAll()
		if out == nil {
			return false
		}
		// Ensure ir is empty
		return maps.Equal(xdsIR.LoadAll(), map[string]*message.XdsIRWithContext{}) && maps.Equal(infraIR.LoadAll(), map[string]*message.InfraIRWithContext{})
	}, time.Second*1, time.Millisecond*20)
}

// setupTestRunner creates a test runner with populated stores and keyCache
func setupTestRunner(t *testing.T) (*Runner, []types.NamespacedName) {
	pResources := new(message.ProviderResources)
	xdsIR := new(message.XdsIR)
	infraIR := new(message.InfraIR)
	cfg, err := config.New(os.Stdout, os.Stderr)
	require.NoError(t, err)
	extMgr, closeFunc, err := registry.NewInMemoryManager(&egv1a1.ExtensionManager{}, &pb.UnimplementedEnvoyGatewayExtensionServer{})
	require.NoError(t, err)
	t.Cleanup(closeFunc)

	r := New(&Config{
		Server:            *cfg,
		ProviderResources: pResources,
		RunnerErrors:      new(message.RunnerErrors),
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
	r.ProviderResources.TLSRouteStatuses.Store(keys[3], &gwapiv1.TLSRouteStatus{})
	r.ProviderResources.TCPRouteStatuses.Store(keys[4], &gwapiv1.TCPRouteStatus{})
	r.ProviderResources.UDPRouteStatuses.Store(keys[5], &gwapiv1.UDPRouteStatus{})
	r.ProviderResources.UDPRouteStatuses.Store(keys[6], &gwapiv1.UDPRouteStatus{})
	r.ProviderResources.BackendStatuses.Store(keys[7], &egv1a1.BackendStatus{})
	r.ProviderResources.BackendTLSPolicyStatuses.Store(keys[8], &gwapiv1.PolicyStatus{})
	r.ProviderResources.ClientTrafficPolicyStatuses.Store(keys[9], &gwapiv1.PolicyStatus{})
	r.ProviderResources.BackendTrafficPolicyStatuses.Store(keys[10], &gwapiv1.PolicyStatus{})
	r.ProviderResources.SecurityPolicyStatuses.Store(keys[11], &gwapiv1.PolicyStatus{})
	r.ProviderResources.EnvoyExtensionPolicyStatuses.Store(keys[12], &gwapiv1.PolicyStatus{})

	// Populate keyCache to simulate normal operation where stores and keyCache are kept in sync
	r.populateKeyCache()

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
	keysToDelete.GatewayStatus[keys[0]] = ""
	keysToDelete.HTTPRouteStatus[keys[1]] = ""
	keysToDelete.TLSRouteStatus[keys[3]] = ""
	keysToDelete.UDPRouteStatus[keys[5]] = "" // Delete only one UDP route
	keysToDelete.BackendTLSPolicyStatus[keys[8]] = ""
	keysToDelete.SecurityPolicyStatus[keys[11]] = ""
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
	require.NotContains(t, r.keyCache.GatewayStatus, keys[0])
	require.NotContains(t, r.keyCache.HTTPRouteStatus, keys[1])
	require.Contains(t, r.keyCache.GRPCRouteStatus, keys[2]) // Should remain
	require.NotContains(t, r.keyCache.TLSRouteStatus, keys[3])
	require.Contains(t, r.keyCache.TCPRouteStatus, keys[4]) // Should remain
	require.NotContains(t, r.keyCache.UDPRouteStatus, keys[5])
	require.Contains(t, r.keyCache.UDPRouteStatus, keys[6]) // Should remain
	require.Contains(t, r.keyCache.BackendStatus, keys[7])  // Should remain
	require.NotContains(t, r.keyCache.BackendTLSPolicyStatus, keys[8])
	require.Contains(t, r.keyCache.ClientTrafficPolicyStatus, keys[9])   // Should remain
	require.Contains(t, r.keyCache.BackendTrafficPolicyStatus, keys[10]) // Should remain
	require.NotContains(t, r.keyCache.SecurityPolicyStatus, keys[11])
	require.Contains(t, r.keyCache.EnvoyExtensionPolicyStatus, keys[12]) // Should remain
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

func TestMergePolicyStatus(t *testing.T) {
	controllerName := "example.com/gateway"

	t.Run("nil incoming keeps existing entry", func(t *testing.T) {
		existing := aggregatedPolicyStatus{
			status: &gwapiv1.PolicyStatus{
				Ancestors: []gwapiv1.PolicyAncestorStatus{
					{
						AncestorRef:    gwapiv1.ParentReference{Name: gwapiv1.ObjectName("gw-a")},
						ControllerName: gwapiv1.GatewayController(controllerName),
					},
				},
			},
			generation: 2,
		}

		got := mergePolicyStatus(existing, nil, 10, "")
		require.Equal(t, existing, got)
	})

	t.Run("nil existing takes incoming", func(t *testing.T) {
		incoming := &gwapiv1.PolicyStatus{
			Ancestors: []gwapiv1.PolicyAncestorStatus{
				{
					AncestorRef:    gwapiv1.ParentReference{Name: gwapiv1.ObjectName("gw-a")},
					ControllerName: gwapiv1.GatewayController(controllerName),
				},
			},
		}

		got := mergePolicyStatus(aggregatedPolicyStatus{}, incoming, 7, "")
		require.Same(t, incoming, got.status)
		require.Equal(t, int64(7), got.generation)
	})

	t.Run("appends ancestors and tracks max generation", func(t *testing.T) {
		first := &gwapiv1.PolicyStatus{
			Ancestors: []gwapiv1.PolicyAncestorStatus{
				{
					AncestorRef:    gwapiv1.ParentReference{Name: gwapiv1.ObjectName("gw-a")},
					ControllerName: gwapiv1.GatewayController(controllerName),
				},
			},
		}
		second := &gwapiv1.PolicyStatus{
			Ancestors: []gwapiv1.PolicyAncestorStatus{
				{
					AncestorRef:    gwapiv1.ParentReference{Name: gwapiv1.ObjectName("gw-b")},
					ControllerName: gwapiv1.GatewayController(controllerName),
				},
			},
		}

		entry := mergePolicyStatus(aggregatedPolicyStatus{}, first, 3, "")
		entry = mergePolicyStatus(entry, second, 9, "")

		require.Len(t, entry.status.Ancestors, 2)
		require.Equal(t, gwapiv1.ObjectName("gw-a"), entry.status.Ancestors[0].AncestorRef.Name)
		require.Equal(t, gwapiv1.ObjectName("gw-b"), entry.status.Ancestors[1].AncestorRef.Name)
		require.Equal(t, int64(9), entry.generation)

		entry = mergePolicyStatus(entry, &gwapiv1.PolicyStatus{}, 4, "")
		require.Equal(t, int64(9), entry.generation)
	})

	t.Run("self-merge prevention", func(t *testing.T) {
		status := &gwapiv1.PolicyStatus{
			Ancestors: []gwapiv1.PolicyAncestorStatus{
				{
					AncestorRef:    gwapiv1.ParentReference{Name: gwapiv1.ObjectName("gw-a")},
					ControllerName: gwapiv1.GatewayController(controllerName),
				},
			},
		}

		// First merge - status becomes both aggregated and incoming
		entry := mergePolicyStatus(aggregatedPolicyStatus{}, status, 1, "")
		// Second merge with same status - should not duplicate
		entry = mergePolicyStatus(entry, status, 2, "")

		// Should still have only one ancestor, not duplicated
		require.Len(t, entry.status.Ancestors, 1)
		require.Equal(t, gwapiv1.ObjectName("gw-a"), entry.status.Ancestors[0].AncestorRef.Name)
	})
}

func TestMergeEnvoyProxyStatus(t *testing.T) {
	t.Run("nil incoming keeps existing entry", func(t *testing.T) {
		existing := &egv1a1.EnvoyProxyStatus{
			Ancestors: []egv1a1.EnvoyProxyAncestorStatus{
				{
					AncestorRef: gwapiv1.ParentReference{Name: gwapiv1.ObjectName("gc-a")},
				},
			},
		}

		got := mergeEnvoyProxyStatus(existing, nil)
		require.Same(t, existing, got)
	})

	t.Run("nil existing takes incoming", func(t *testing.T) {
		incoming := &egv1a1.EnvoyProxyStatus{
			Ancestors: []egv1a1.EnvoyProxyAncestorStatus{
				{
					AncestorRef: gwapiv1.ParentReference{Name: gwapiv1.ObjectName("gc-a")},
				},
			},
		}

		got := mergeEnvoyProxyStatus(nil, incoming)
		require.Same(t, incoming, got)
	})

	t.Run("appends ancestors", func(t *testing.T) {
		first := &egv1a1.EnvoyProxyStatus{
			Ancestors: []egv1a1.EnvoyProxyAncestorStatus{
				{
					AncestorRef: gwapiv1.ParentReference{Name: gwapiv1.ObjectName("gc-a")},
				},
			},
		}
		second := &egv1a1.EnvoyProxyStatus{
			Ancestors: []egv1a1.EnvoyProxyAncestorStatus{
				{
					AncestorRef: gwapiv1.ParentReference{Name: gwapiv1.ObjectName("gw-b")},
				},
			},
		}

		got := mergeEnvoyProxyStatus(first, second)

		require.Len(t, got.Ancestors, 2)
		require.Equal(t, gwapiv1.ObjectName("gc-a"), got.Ancestors[0].AncestorRef.Name)
		require.Equal(t, gwapiv1.ObjectName("gw-b"), got.Ancestors[1].AncestorRef.Name)
	})

	t.Run("self-merge prevention", func(t *testing.T) {
		status := &egv1a1.EnvoyProxyStatus{
			Ancestors: []egv1a1.EnvoyProxyAncestorStatus{
				{
					AncestorRef: gwapiv1.ParentReference{Name: gwapiv1.ObjectName("gc-a")},
				},
			},
		}

		// Merge status with itself - should not duplicate
		got := mergeEnvoyProxyStatus(status, status)

		// Should still have only one ancestor, not duplicated
		require.Len(t, got.Ancestors, 1)
		require.Equal(t, gwapiv1.ObjectName("gc-a"), got.Ancestors[0].AncestorRef.Name)
	})
}

func TestMergeRouteStatus(t *testing.T) {
	controllerName := "example.com/gateway"

	t.Run("nil incoming keeps existing entry", func(t *testing.T) {
		existing := aggregatedRouteStatus{
			status: &gwapiv1.RouteStatus{
				Parents: []gwapiv1.RouteParentStatus{
					{
						ParentRef:      gwapiv1.ParentReference{Name: gwapiv1.ObjectName("gw-a")},
						ControllerName: gwapiv1.GatewayController(controllerName),
					},
				},
			},
			generation: 2,
		}

		got := mergeAggregatedRouteStatus(existing, nil, 10, "")
		require.Equal(t, existing, got)
	})

	t.Run("nil existing takes incoming", func(t *testing.T) {
		incoming := &gwapiv1.RouteStatus{
			Parents: []gwapiv1.RouteParentStatus{
				{
					ParentRef:      gwapiv1.ParentReference{Name: gwapiv1.ObjectName("gw-a")},
					ControllerName: gwapiv1.GatewayController(controllerName),
				},
			},
		}

		got := mergeAggregatedRouteStatus(aggregatedRouteStatus{}, incoming, 7, "")
		require.Same(t, incoming, got.status)
		require.Equal(t, int64(7), got.generation)
	})

	t.Run("appends parents and tracks max generation", func(t *testing.T) {
		first := &gwapiv1.RouteStatus{
			Parents: []gwapiv1.RouteParentStatus{
				{
					ParentRef:      gwapiv1.ParentReference{Name: gwapiv1.ObjectName("gw-a")},
					ControllerName: gwapiv1.GatewayController(controllerName),
				},
			},
		}
		second := &gwapiv1.RouteStatus{
			Parents: []gwapiv1.RouteParentStatus{
				{
					ParentRef:      gwapiv1.ParentReference{Name: gwapiv1.ObjectName("gw-b")},
					ControllerName: gwapiv1.GatewayController(controllerName),
				},
			},
		}

		entry := mergeAggregatedRouteStatus(aggregatedRouteStatus{}, first, 3, "")
		entry = mergeAggregatedRouteStatus(entry, second, 9, "")

		require.Len(t, entry.status.Parents, 2)
		require.Equal(t, gwapiv1.ObjectName("gw-a"), entry.status.Parents[0].ParentRef.Name)
		require.Equal(t, gwapiv1.ObjectName("gw-b"), entry.status.Parents[1].ParentRef.Name)
		require.Equal(t, int64(9), entry.generation)

		entry = mergeAggregatedRouteStatus(entry, &gwapiv1.RouteStatus{}, 4, "")
		require.Equal(t, int64(9), entry.generation)
	})
}

func TestLoadTLSConfig_HostMode(t *testing.T) {
	// Create temporary directory structure for certs using t.TempDir()
	configHome := t.TempDir()
	certsDir := filepath.Join(configHome, "certs", "envoy-gateway")
	hmacDir := filepath.Join(configHome, "certs", "envoy-oidc-hmac")
	require.NoError(t, os.MkdirAll(certsDir, 0o750))
	require.NoError(t, os.MkdirAll(hmacDir, 0o750))

	// Create test certificates using internal/crypto package
	cfg, err := config.New(os.Stdout, os.Stderr)
	require.NoError(t, err)

	// Generate certificates with default provider (crypto.GenerateCerts only supports Kubernetes)
	certs, err := crypto.GenerateCerts(cfg)
	require.NoError(t, err)

	// Write certificates to temp directory
	caFile := filepath.Join(certsDir, "ca.crt")
	certFile := filepath.Join(certsDir, "tls.crt")
	keyFile := filepath.Join(certsDir, "tls.key")
	hmacFile := filepath.Join(hmacDir, "hmac-secret")

	require.NoError(t, os.WriteFile(caFile, certs.CACertificate, 0o600))
	require.NoError(t, os.WriteFile(certFile, certs.EnvoyGatewayCertificate, 0o600))
	require.NoError(t, os.WriteFile(keyFile, certs.EnvoyGatewayPrivateKey, 0o600))
	require.NoError(t, os.WriteFile(hmacFile, certs.OIDCHMACSecret, 0o600))

	// Configure host mode with custom configHome (certs are stored in configHome)
	// MUST be set BEFORE creating Runner since Config{Server: *cfg} makes a copy
	cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{
		Type: egv1a1.ProviderTypeCustom,
		Custom: &egv1a1.EnvoyGatewayCustomProvider{
			Infrastructure: &egv1a1.EnvoyGatewayInfrastructureProvider{
				Type: egv1a1.InfrastructureProviderTypeHost,
				Host: &egv1a1.EnvoyGatewayHostInfrastructureProvider{
					ConfigHome: &configHome,
				},
			},
		},
	}

	r := &Runner{
		Config: Config{
			Server: *cfg,
		},
	}

	// Test loadTLSConfig with host mode
	tlsConfig, salt, err := r.loadTLSConfig(context.Background())
	require.NoError(t, err)
	require.NotNil(t, tlsConfig)
	require.NotNil(t, salt)

	// Verify HMAC secret was loaded
	require.Equal(t, certs.OIDCHMACSecret, salt)

	// Verify TLS config properties
	// crypto.LoadTLSConfig uses GetConfigForClient callback to load certs on demand
	require.NotNil(t, tlsConfig.GetConfigForClient)
	require.Equal(t, tls.RequireAndVerifyClientCert, tlsConfig.ClientAuth)
	require.Equal(t, uint16(tls.VersionTLS13), tlsConfig.MinVersion)
}

// writeHostModeCerts writes the Envoy Gateway TLS certificates into a temporary
// host-mode config home and returns a config pointing at it. The HMAC secret is
// returned rather than written, so callers decide when it becomes available.
func writeHostModeCerts(t *testing.T) (cfg *config.Server, hmacFile string, hmacSecret []byte) {
	t.Helper()
	configHome := t.TempDir()
	certsDir := filepath.Join(configHome, "certs", "envoy-gateway")
	hmacDir := filepath.Join(configHome, "certs", "envoy-oidc-hmac")
	require.NoError(t, os.MkdirAll(certsDir, 0o750))
	require.NoError(t, os.MkdirAll(hmacDir, 0o750))

	cfg, err := config.New(os.Stdout, os.Stderr)
	require.NoError(t, err)
	certs, err := crypto.GenerateCerts(cfg)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(certsDir, "ca.crt"), certs.CACertificate, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(certsDir, "tls.crt"), certs.EnvoyGatewayCertificate, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(certsDir, "tls.key"), certs.EnvoyGatewayPrivateKey, 0o600))

	cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{
		Type: egv1a1.ProviderTypeCustom,
		Custom: &egv1a1.EnvoyGatewayCustomProvider{
			Infrastructure: &egv1a1.EnvoyGatewayInfrastructureProvider{
				Type: egv1a1.InfrastructureProviderTypeHost,
				Host: &egv1a1.EnvoyGatewayHostInfrastructureProvider{
					ConfigHome: &configHome,
				},
			},
		},
	}
	return cfg, filepath.Join(hmacDir, "hmac-secret"), certs.OIDCHMACSecret
}

// TestStartWasmCacheRetriesUntilTLSConfigLoads verifies that a transient failure to
// load the HMAC secret delays the Wasm cache instead of disabling it until restart.
func TestStartWasmCacheRetriesUntilTLSConfigLoads(t *testing.T) {
	// The Wasm cache directory is created under the home directory in host mode.
	t.Setenv("HOME", t.TempDir())
	cfg, hmacFile, hmacSecret := writeHostModeCerts(t)

	r := New(&Config{Server: *cfg})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		r.startWasmCache(ctx)
		close(done)
	}()

	// The HMAC secret is missing at startup and becomes available shortly after.
	time.Sleep(200 * time.Millisecond)
	require.NoError(t, os.WriteFile(hmacFile, hmacSecret, 0o600))

	require.Eventually(t, func() bool {
		return r.getWasmCache() != nil
	}, 5*time.Second, 50*time.Millisecond, "wasm cache was not initialized after the HMAC secret became available")

	cancel()
	<-done
}

// TestStartWasmCacheStopsRetryingOnContextDone verifies that startWasmCache returns
// once the context is done if the HMAC secret never becomes available.
func TestStartWasmCacheStopsRetryingOnContextDone(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg, _, _ := writeHostModeCerts(t)

	r := New(&Config{Server: *cfg})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		r.startWasmCache(ctx)
		close(done)
	}()

	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("startWasmCache did not return after the context was cancelled")
	}
	require.Nil(t, r.getWasmCache())
	select {
	case <-r.wasmCacheReady:
		t.Fatal("wasmCacheReady was closed although the wasm cache was not started")
	default:
	}
}

func TestRetranslateOnWasmCacheReady(t *testing.T) {
	newSnapshot := func(val *resource.ControllerResourcesContext) watchable.Snapshot[string, *resource.ControllerResourcesContext] {
		return watchable.Snapshot[string, *resource.ControllerResourcesContext]{
			State: map[string]*resource.ControllerResourcesContext{"test": val},
			Updates: []watchable.Update[string, *resource.ControllerResourcesContext]{
				{Key: "test", Value: val},
			},
		}
	}
	receive := func(t *testing.T, out <-chan watchable.Snapshot[string, *resource.ControllerResourcesContext]) watchable.Snapshot[string, *resource.ControllerResourcesContext] {
		t.Helper()
		select {
		case snapshot, ok := <-out:
			require.True(t, ok, "channel closed unexpectedly")
			return snapshot
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for a snapshot")
		}
		return watchable.Snapshot[string, *resource.ControllerResourcesContext]{}
	}
	cfg, err := config.New(os.Stdout, os.Stderr)
	require.NoError(t, err)

	t.Run("replays the latest state once the cache is ready", func(t *testing.T) {
		r := New(&Config{Server: *cfg})
		sub := make(chan watchable.Snapshot[string, *resource.ControllerResourcesContext])
		out := r.retranslateOnWasmCacheReady(sub)

		val := &resource.ControllerResourcesContext{
			Resources: &resource.ControllerResources{
				{GatewayClass: &gwapiv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "eg"}}},
			},
			StoredAt: time.Now(),
		}
		sub <- newSnapshot(val)
		require.Equal(t, newSnapshot(val), receive(t, out))

		close(r.wasmCacheReady)
		replay := receive(t, out)
		require.Len(t, replay.Updates, 1)
		update := replay.Updates[0]
		require.Equal(t, "test", update.Key)
		require.False(t, update.Delete)
		require.NotSame(t, val, update.Value)
		require.Equal(t, val.Resources, update.Value.Resources)
		require.True(t, update.Value.StoredAt.IsZero())

		// Later snapshots are forwarded unchanged.
		sub <- newSnapshot(val)
		require.Equal(t, newSnapshot(val), receive(t, out))
		close(sub)
		_, ok := <-out
		require.False(t, ok)
	})

	t.Run("does not replay before the first snapshot", func(t *testing.T) {
		r := New(&Config{Server: *cfg})
		sub := make(chan watchable.Snapshot[string, *resource.ControllerResourcesContext])
		close(r.wasmCacheReady)
		out := r.retranslateOnWasmCacheReady(sub)

		val := &resource.ControllerResourcesContext{StoredAt: time.Now()}
		sub <- newSnapshot(val)
		require.Equal(t, newSnapshot(val), receive(t, out))
		close(sub)
		_, ok := <-out
		require.False(t, ok)
	})
}

// TestWasmPolicyRecoversOnceWasmCacheStarts verifies that an EnvoyExtensionPolicy
// rejected because the Wasm cache was not started yet is translated again, and
// accepted, once the cache starts, without any further resource change.
func TestWasmPolicyRecoversOnceWasmCacheStarts(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg, hmacFile, hmacSecret := writeHostModeCerts(t)

	wasmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// The smallest valid Wasm module: the magic number and the version.
		_, _ = w.Write([]byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00})
	}))
	defer wasmServer.Close()

	pResources := new(message.ProviderResources)
	r := New(&Config{
		Server:            *cfg,
		ProviderResources: pResources,
		RunnerErrors:      new(message.RunnerErrors),
		XdsIR:             new(message.XdsIR),
		InfraIR:           new(message.InfraIR),
	})
	ctx, cancel := context.WithCancel(context.Background())
	require.NoError(t, r.Start(ctx))
	defer func() {
		cancel()
		require.NoError(t, r.Close())
	}()

	policyKey := types.NamespacedName{Namespace: "default", Name: "wasm"}
	pResources.GatewayAPIResources.Store(cfg.EnvoyGateway.Gateway.ControllerName, &resource.ControllerResourcesContext{
		Context: ctx,
		Resources: &resource.ControllerResources{{
			GatewayClass: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{Name: "eg"},
				Spec:       gwapiv1.GatewayClassSpec{ControllerName: gwapiv1.GatewayController(cfg.EnvoyGateway.Gateway.ControllerName)},
			},
			Gateways: []*gwapiv1.Gateway{{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "eg"},
				Spec: gwapiv1.GatewaySpec{
					GatewayClassName: "eg",
					Listeners:        []gwapiv1.Listener{{Name: "http", Protocol: gwapiv1.HTTPProtocolType, Port: 80}},
				},
			}},
			EnvoyExtensionPolicies: []*egv1a1.EnvoyExtensionPolicy{{
				ObjectMeta: metav1.ObjectMeta{Namespace: policyKey.Namespace, Name: policyKey.Name},
				Spec: egv1a1.EnvoyExtensionPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.GroupName,
								Kind:  "Gateway",
								Name:  "eg",
							},
						}},
					},
					Wasm: []egv1a1.Wasm{{
						Code: &egv1a1.WasmCodeSource{
							Type: egv1a1.HTTPWasmCodeSourceType,
							HTTP: &egv1a1.HTTPWasmCodeSource{URL: wasmServer.URL + "/filter.wasm"},
						},
					}},
				},
			}},
		}},
	})

	policyMessage := func() string {
		status, ok := pResources.EnvoyExtensionPolicyStatuses.Load(policyKey)
		if !ok || len(status.Ancestors) == 0 {
			return ""
		}
		for _, cond := range status.Ancestors[0].Conditions {
			if cond.Type == string(gwapiv1.PolicyConditionAccepted) {
				return string(cond.Status) + ": " + cond.Message
			}
		}
		return ""
	}

	// The HMAC secret is missing, so the policy is rejected.
	require.Eventually(t, func() bool {
		return strings.Contains(policyMessage(), "wasm cache is not initialized")
	}, 5*time.Second, 50*time.Millisecond)

	// The HMAC secret becomes available, and the policy is accepted with no
	// change to the resources.
	require.NoError(t, os.WriteFile(hmacFile, hmacSecret, 0o600))
	require.Eventually(t, func() bool {
		return strings.HasPrefix(policyMessage(), string(metav1.ConditionTrue))
	}, 10*time.Second, 50*time.Millisecond, "policy status: %s", policyMessage())
}
