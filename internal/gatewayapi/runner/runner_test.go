package runner

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
)

func TestRunner(t *testing.T) {
	// Setup
	pResources := new(message.ProviderResources)
	xdsIR := new(message.XdsIR)
	infraIR := new(message.InfraIR)
	cfg, err := config.NewDefaultServer()
	require.NoError(t, err)
	r := New(&Config{
		Server:            *cfg,
		ProviderResources: pResources,
		XdsIR:             xdsIR,
		InfraIR:           infraIR,
	})
	ctx := context.Background()
	// Start
	err = r.Start(ctx)
	require.NoError(t, err)

	// IR is nil at start
	require.Equal(t, (*ir.Xds)(nil), xdsIR.Get())
	require.Equal(t, (*ir.Infra)(nil), infraIR.Get())

	// TODO: pass valid provider resources

	// Reset gatewayclass slice and update with a nil gatewayclass to trigger a delete
	pResources.DeleteGatewayClasses()
	pResources.GatewayClasses.Store("test", nil)
	require.Eventually(t, func() bool {
		out := xdsIR.Get()
		if out == nil {
			return false
		}
		// Ensure ir is empty
		return (reflect.DeepEqual(*xdsIR.Get(), ir.Xds{})) && (reflect.DeepEqual(*infraIR.Get(), ir.Infra{Proxy: nil}))
	}, time.Second*1, time.Millisecond*20)

}
