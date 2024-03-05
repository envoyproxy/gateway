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
	v1 "sigs.k8s.io/gateway-api/apis/v1"

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

func TestDeletableStatus(t *testing.T) {
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

	// No status is set
	ds1 := r.getDeletableStatus()
	require.Empty(t, len(ds1.GatewayStatusKeys))

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
	}
	statuses := []v1.GatewayStatus{
		{
			Addresses: []v1.GatewayStatusAddress{
				{
					Value: "192.168.0.1",
				},
			},
		},
		{
			Addresses: []v1.GatewayStatusAddress{
				{
					Value: "192.168.0.2",
				},
			},
		},
		{
			Addresses: []v1.GatewayStatusAddress{
				{
					Value: "192.168.0.3",
				},
			},
		},
	}
	r.ProviderResources.GatewayStatuses.Store(keys[0], &statuses[0])
	r.ProviderResources.GatewayStatuses.Store(keys[1], &statuses[1])
	r.ProviderResources.GatewayStatuses.Store(keys[2], &statuses[2])

	// getDeletableStatus: Checks that the keys are successfully stored
	ds2 := r.getDeletableStatus()
	require.True(t, ds2.GatewayStatusKeys[keys[0]])
	require.True(t, ds2.GatewayStatusKeys[keys[1]])
	require.True(t, ds2.GatewayStatusKeys[keys[2]])
	require.Equal(t, 3, r.ProviderResources.GatewayStatuses.Len())

	// deleteStatusKeys: Delete the third key by removing the first
	// and second key from deletables
	delete(ds2.GatewayStatusKeys, keys[0])
	delete(ds2.GatewayStatusKeys, keys[1])
	r.deleteStatusKeys(ds2)
	require.Empty(t, len(ds2.GatewayStatusKeys))
	require.Equal(t, 2, r.ProviderResources.GatewayStatuses.Len())

	// deleteAllStatusKeys: Delete remaining keys
	r.deleteAllStatusKeys()
	require.Equal(t, 0, r.ProviderResources.GatewayStatuses.Len())
}
