package runner

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"

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
	require.Equal(t, map[string]*ir.Xds{}, xdsIR.LoadAll())
	require.Equal(t, map[string]*ir.Infra{}, infraIR.LoadAll())

	// TODO: pass valid provider resources

	// Reset gateway slice and update with a nil gateway to trigger a delete.
	pResources.DeleteGateways()
	key := types.NamespacedName{Namespace: "test", Name: "test"}
	pResources.Gateways.Store(key, nil)
	require.Eventually(t, func() bool {
		out := xdsIR.LoadAll()
		if out == nil {
			return false
		}
		// Ensure ir is empty
		return (reflect.DeepEqual(xdsIR.LoadAll(), map[string]*ir.Xds{})) && (reflect.DeepEqual(infraIR.LoadAll(), map[string]*ir.Infra{}))
	}, time.Second*1, time.Millisecond*20)

}
