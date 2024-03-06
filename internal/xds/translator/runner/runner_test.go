// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"fmt"
	"testing"
	"time"

	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
)

func TestRunner(t *testing.T) {
	// Setup
	xdsIR := new(message.XdsIR)
	xds := new(message.Xds)
	pResource := new(message.ProviderResources)
	cfg, err := config.New()
	require.NoError(t, err)
	r := New(&Config{
		Server:            *cfg,
		ProviderResources: pResource,
		XdsIR:             xdsIR,
		Xds:               xds,
	})

	ctx := context.Background()
	// Start
	err = r.Start(ctx)
	require.NoError(t, err)

	// xDS is nil at start
	require.Equal(t, map[string]*ir.Xds{}, xdsIR.LoadAll())

	// test translation
	path := "example"
	res := ir.Xds{
		HTTP: []*ir.HTTPListener{
			{
				Name:      "test",
				Address:   "0.0.0.0",
				Port:      80,
				Hostnames: []string{"example.com"},
				Routes: []*ir.HTTPRoute{
					{
						Name: "test-route",
						PathMatch: &ir.StringMatch{
							Exact: &path,
						},
						Destination: &ir.RouteDestination{
							Name: "test-dest",
							Settings: []*ir.DestinationSetting{
								{
									Endpoints: []*ir.DestinationEndpoint{
										{
											Host: "10.11.12.13",
											Port: 8080,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	xdsIR.Store("test", &res)
	require.Eventually(t, func() bool {
		out := xds.LoadAll()
		if out == nil {
			return false
		}
		if out["test"] == nil {
			return false
		}
		// Ensure an xds listener is created
		return len(out["test"].XdsResources[resourcev3.ListenerType]) == 1
	}, time.Second*5, time.Millisecond*50)

	// Delete the IR triggering an xds delete
	xdsIR.Delete("test")
	require.Eventually(t, func() bool {
		out := xds.LoadAll()
		// Ensure that xds has no key, value pairs
		return len(out) == 0
	}, time.Second*5, time.Millisecond*50)

}

func TestRunner_withExtensionManager(t *testing.T) {
	// Setup
	xdsIR := new(message.XdsIR)
	xds := new(message.Xds)
	pResource := new(message.ProviderResources)

	cfg, err := config.New()
	require.NoError(t, err)
	r := New(&Config{
		Server:            *cfg,
		ProviderResources: pResource,
		XdsIR:             xdsIR,
		Xds:               xds,
		ExtensionManager:  &extManagerMock{},
	})

	ctx := context.Background()
	// Start
	err = r.Start(ctx)
	require.NoError(t, err)

	// xDS is nil at start
	require.Equal(t, map[string]*ir.Xds{}, xdsIR.LoadAll())

	// test translation
	path := "example"
	res := ir.Xds{
		HTTP: []*ir.HTTPListener{
			{
				Name:      "test",
				Address:   "0.0.0.0",
				Port:      80,
				Hostnames: []string{"example.com"},
				Routes: []*ir.HTTPRoute{
					{
						Name: "test-route",
						PathMatch: &ir.StringMatch{
							Exact: &path,
						},
						Destination: &ir.RouteDestination{
							Name: "test-dest",
							Settings: []*ir.DestinationSetting{
								{
									Endpoints: []*ir.DestinationEndpoint{
										{
											Host: "10.11.12.13",
											Port: 8080,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	xdsIR.Store("test", &res)
	require.Eventually(t, func() bool {
		out := xds.LoadAll()
		// xDS translation is done in a best-effort manner, so event the extension
		// manager returns an error, the xDS resources should still be created.
		return len(out) == 1
	}, time.Second*5, time.Millisecond*50)
}

type extManagerMock struct {
	types.Manager
}

func (m *extManagerMock) GetPostXDSHookClient(xdsHookType v1alpha1.XDSTranslatorHook) types.XDSHookClient {
	if xdsHookType == v1alpha1.XDSHTTPListener {
		return &xdsHookClientMock{}
	}

	return nil
}

type xdsHookClientMock struct {
	types.XDSHookClient
}

func (c *xdsHookClientMock) PostHTTPListenerModifyHook(*listenerv3.Listener) (*listenerv3.Listener, error) {
	return nil, fmt.Errorf("assuming a network error during the call")
}
