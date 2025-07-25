// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
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
	cfg, err := config.New(os.Stdout)
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
				CoreListenerDetails: ir.CoreListenerDetails{
					Name:    "test",
					Address: "0.0.0.0",
					Port:    80,
				},
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

func TestRunner_withExtensionManager_FailOpen(t *testing.T) {
	// Setup
	xdsIR := new(message.XdsIR)
	xds := new(message.Xds)
	pResource := new(message.ProviderResources)

	cfg, err := config.New(os.Stdout)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	extMgr := &extManagerMock{}
	extMgr.ShouldFailOpen = true

	r := New(&Config{
		Server:            *cfg,
		ProviderResources: pResource,
		XdsIR:             xdsIR,
		Xds:               xds,
		ExtensionManager:  extMgr,
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
				CoreListenerDetails: ir.CoreListenerDetails{
					Name:    "test",
					Address: "0.0.0.0",
					Port:    80,
				},
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
		// Since the extension manager is configured to fail open, in an event of an error
		// from the extension manager hooks, xds update should be published.
		return len(out) == 1
	}, time.Second*5, time.Millisecond*50)
}

func TestRunner_withExtensionManager_FailClosed(t *testing.T) {
	// Setup
	xdsIR := new(message.XdsIR)
	xds := new(message.Xds)
	pResource := new(message.ProviderResources)

	cfg, err := config.New(os.Stdout)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	extMgr := &extManagerMock{}

	r := New(&Config{
		Server:            *cfg,
		ProviderResources: pResource,
		XdsIR:             xdsIR,
		Xds:               xds,
		ExtensionManager:  extMgr,
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
				CoreListenerDetails: ir.CoreListenerDetails{
					Name:    "test",
					Address: "0.0.0.0",
					Port:    80,
				},
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
	require.Never(t, func() bool {
		out := xds.LoadAll()
		// Since the extension manager is configured to fail closed,  in an event of an error
		// from the extension manager hooks, xds update should not be published.
		return len(out) > 0
	}, time.Second*5, time.Millisecond*50)
}

type extManagerMock struct {
	types.Manager
	ShouldFailOpen bool
}

func (m *extManagerMock) GetPostXDSHookClient(xdsHookType egv1a1.XDSTranslatorHook) (types.XDSHookClient, error) {
	if xdsHookType == egv1a1.XDSHTTPListener {
		return &xdsHookClientMock{}, nil
	}

	return nil, nil
}

func (m *extManagerMock) FailOpen() bool {
	return m.ShouldFailOpen
}

type xdsHookClientMock struct {
	types.XDSHookClient
}

func (c *xdsHookClientMock) PostHTTPListenerModifyHook(*listenerv3.Listener, []*unstructured.Unstructured) (*listenerv3.Listener, error) {
	return nil, fmt.Errorf("assuming a network error during the call")
}
