package runner

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

func TestRunner(t *testing.T) {
	// Setup
	xdsIR := new(message.XdsIR)
	xds := new(message.Xds)
	cfg, err := config.NewDefaultServer()
	require.NoError(t, err)
	r := New(&Config{
		Server: *cfg,
		XdsIR:  xdsIR,
		Xds:    xds,
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
						DestinationCluster: &ir.DestinationCluster{
							ClusterName: "cluster_test-route",
							Destinations: []*ir.RouteDestination{
								{
									Host: "1.2.3.4",
									Port: 50000,
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
