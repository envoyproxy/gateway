// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/extension/testutils"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
)

func TestRunner(t *testing.T) {
	// Setup
	pResources := new(message.ProviderResources)
	xdsIR := new(message.XdsIR)
	infraIR := new(message.InfraIR)
	cfg, err := config.New()
	require.NoError(t, err)
	r := New(&Config{
		Server:            *cfg,
		ProviderResources: pResources,
		XdsIR:             xdsIR,
		InfraIR:           infraIR,
		ExtensionManager:  testutils.NewManager(egv1a1.ExtensionManager{}),
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

func TestGetIRKeysToDelete(t *testing.T) {
	testCases := []struct {
		name    string
		curKeys []string
		newKeys []string
		delKeys []string
	}{
		{
			name:    "empty",
			curKeys: []string{},
			newKeys: []string{},
			delKeys: []string{},
		},
		{name: "no new keys",
			curKeys: []string{"one", "two"},
			newKeys: []string{},
			delKeys: []string{"one", "two"},
		},
		{
			name:    "no cur keys",
			curKeys: []string{},
			newKeys: []string{"one", "two"},
			delKeys: []string{},
		},
		{
			name:    "equal",
			curKeys: []string{"one", "two"},
			newKeys: []string{"two", "one"},
			delKeys: []string{},
		},
		{
			name:    "mix",
			curKeys: []string{"one", "two"},
			newKeys: []string{"two", "three"},
			delKeys: []string{"one"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.ElementsMatch(t, tc.delKeys, getIRKeysToDelete(tc.curKeys, tc.newKeys))
		})
	}
}

func TestDeleteStatusKeys(t *testing.T) {
	// Setup
	pResources := new(message.ProviderResources)
	xdsIR := new(message.XdsIR)
	infraIR := new(message.InfraIR)
	cfg, err := config.New()
	require.NoError(t, err)
	r := New(&Config{
		Server:            *cfg,
		ProviderResources: pResources,
		XdsIR:             xdsIR,
		InfraIR:           infraIR,
		ExtensionManager:  testutils.NewManager(egv1a1.ExtensionManager{}),
	})
	ctx := context.Background()

	// Start
	err = r.Start(ctx)
	require.NoError(t, err)

	// A new status gets stored
	keys := []types.NamespacedName{
		{
			Name:      "test1",
			Namespace: "test-namespace",
		},
		{
			Name:      "test2",
			Namespace: "test-namespace",
		},
		{
			Name:      "test3",
			Namespace: "test-namespace",
		},
		{
			Name:      "test4",
			Namespace: "test-namespace",
		},
		{
			Name:      "test5",
			Namespace: "test-namespace",
		},
		{
			Name:      "test6",
			Namespace: "test-namespace",
		},
		{
			Name:      "test7",
			Namespace: "test-namespace",
		},
	}

	r.ProviderResources.GatewayStatuses.Store(keys[0], &gwapiv1.GatewayStatus{})
	r.ProviderResources.HTTPRouteStatuses.Store(keys[1], &gwapiv1.HTTPRouteStatus{})
	r.ProviderResources.GRPCRouteStatuses.Store(keys[2], &gwapiv1a2.GRPCRouteStatus{})
	r.ProviderResources.TLSRouteStatuses.Store(keys[3], &gwapiv1a2.TLSRouteStatus{})
	r.ProviderResources.TCPRouteStatuses.Store(keys[4], &gwapiv1a2.TCPRouteStatus{})
	r.ProviderResources.UDPRouteStatuses.Store(keys[5], &gwapiv1a2.UDPRouteStatus{})
	r.ProviderResources.UDPRouteStatuses.Store(keys[6], &gwapiv1a2.UDPRouteStatus{})

	// Checks that the keys are successfully stored to DeletableStatus and watchable maps
	ds := r.getAllStatuses()

	require.True(t, ds.GatewayStatusKeys[keys[0]])
	require.True(t, ds.HTTPRouteStatusKeys[keys[1]])
	require.True(t, ds.GRPCRouteStatusKeys[keys[2]])
	require.True(t, ds.TLSRouteStatusKeys[keys[3]])
	require.True(t, ds.TCPRouteStatusKeys[keys[4]])
	require.True(t, ds.UDPRouteStatusKeys[keys[5]])
	require.True(t, ds.UDPRouteStatusKeys[keys[6]])

	require.Equal(t, 1, r.ProviderResources.GatewayStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.HTTPRouteStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.GRPCRouteStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.TLSRouteStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.TCPRouteStatuses.Len())
	require.Equal(t, 2, r.ProviderResources.UDPRouteStatuses.Len())

	// Delete all keys except the last UDPRouteStatus key
	delete(ds.UDPRouteStatusKeys, keys[6])
	r.deleteStatusKeys(ds)

	require.Equal(t, 0, r.ProviderResources.GatewayStatuses.Len())
	require.Equal(t, 0, r.ProviderResources.HTTPRouteStatuses.Len())
	require.Equal(t, 0, r.ProviderResources.GRPCRouteStatuses.Len())
	require.Equal(t, 0, r.ProviderResources.TLSRouteStatuses.Len())
	require.Equal(t, 0, r.ProviderResources.TCPRouteStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.UDPRouteStatuses.Len())
}

func TestDeleteAllStatusKeys(t *testing.T) {
	// Setup
	pResources := new(message.ProviderResources)
	xdsIR := new(message.XdsIR)
	infraIR := new(message.InfraIR)
	cfg, err := config.New()
	require.NoError(t, err)
	r := New(&Config{
		Server:            *cfg,
		ProviderResources: pResources,
		XdsIR:             xdsIR,
		InfraIR:           infraIR,
		ExtensionManager:  testutils.NewManager(egv1a1.ExtensionManager{}),
	})
	ctx := context.Background()

	// Start
	err = r.Start(ctx)
	require.NoError(t, err)

	// A new status gets stored
	keys := []types.NamespacedName{
		{
			Name:      "test1",
			Namespace: "test-namespace",
		},
		{
			Name:      "test2",
			Namespace: "test-namespace",
		},
		{
			Name:      "test3",
			Namespace: "test-namespace",
		},
		{
			Name:      "test4",
			Namespace: "test-namespace",
		},
		{
			Name:      "test5",
			Namespace: "test-namespace",
		},
		{
			Name:      "test6",
			Namespace: "test-namespace",
		},
	}

	r.ProviderResources.GatewayStatuses.Store(keys[0], &gwapiv1.GatewayStatus{})
	r.ProviderResources.HTTPRouteStatuses.Store(keys[1], &gwapiv1.HTTPRouteStatus{})
	r.ProviderResources.GRPCRouteStatuses.Store(keys[2], &gwapiv1a2.GRPCRouteStatus{})
	r.ProviderResources.TLSRouteStatuses.Store(keys[3], &gwapiv1a2.TLSRouteStatus{})
	r.ProviderResources.TCPRouteStatuses.Store(keys[4], &gwapiv1a2.TCPRouteStatus{})
	r.ProviderResources.UDPRouteStatuses.Store(keys[5], &gwapiv1a2.UDPRouteStatus{})

	// Checks that the keys are successfully stored to DeletableStatus and watchable maps
	ds := r.getAllStatuses()

	require.True(t, ds.GatewayStatusKeys[keys[0]])
	require.True(t, ds.HTTPRouteStatusKeys[keys[1]])
	require.True(t, ds.GRPCRouteStatusKeys[keys[2]])
	require.True(t, ds.TLSRouteStatusKeys[keys[3]])
	require.True(t, ds.TCPRouteStatusKeys[keys[4]])
	require.True(t, ds.UDPRouteStatusKeys[keys[5]])

	require.Equal(t, 1, r.ProviderResources.GatewayStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.HTTPRouteStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.GRPCRouteStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.TLSRouteStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.TCPRouteStatuses.Len())
	require.Equal(t, 1, r.ProviderResources.UDPRouteStatuses.Len())

	// Delete all keys
	r.deleteAllStatusKeys()
	require.Equal(t, 0, r.ProviderResources.GatewayStatuses.Len())
	require.Equal(t, 0, r.ProviderResources.HTTPRouteStatuses.Len())
	require.Equal(t, 0, r.ProviderResources.GRPCRouteStatuses.Len())
	require.Equal(t, 0, r.ProviderResources.TLSRouteStatuses.Len())
	require.Equal(t, 0, r.ProviderResources.TCPRouteStatuses.Len())
	require.Equal(t, 0, r.ProviderResources.UDPRouteStatuses.Len())
}
