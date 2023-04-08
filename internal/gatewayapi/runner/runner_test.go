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

	egv1a1cfg "github.com/envoyproxy/gateway/api/config/v1alpha1"
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
		ExtensionManager:  testutils.NewManager(egv1a1cfg.Extension{}),
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
