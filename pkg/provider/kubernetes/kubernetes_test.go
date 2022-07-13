package kubernetes

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/pkg/envoygateway/config"
)

const (
	defaultWait = time.Second * 10
	defaultTick = time.Millisecond * 20
)

func TestProvider(t *testing.T) {
	// Setup the test environment.
	testEnv, cliCfg, err := startEnv()
	require.NoError(t, err)

	// Setup and start the kube provider.
	svr, err := config.NewDefaultServer()
	require.NoError(t, err)
	provider, err := New(cliCfg, svr)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(ctrl.SetupSignalHandler())
	go func() {
		require.NoError(t, provider.Start(ctx))
	}()

	// Stop the kube provider.
	defer func() {
		cancel()
		require.NoError(t, testEnv.Stop())
	}()

	testcases := map[string]func(context.Context, *testing.T, client.Client){
		"gateway controller name": testGatewayClassReconciler,
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			tc(ctx, t, provider.manager.GetClient())
		})
	}
}

func startEnv() (*envtest.Environment, *rest.Config, error) {
	log.SetLogger(zap.New(zap.WriteTo(os.Stderr), zap.UseDevMode(true)))
	crd := filepath.Join("..", "testdata", "in")
	env := &envtest.Environment{
		CRDDirectoryPaths: []string{crd},
	}
	cfg, err := env.Start()
	if err != nil {
		return nil, nil, err
	}
	return env, cfg, nil
}

func testGatewayClassReconciler(ctx context.Context, t *testing.T, cli client.Client) {
	gc := &gwapiv1a2.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-gc",
		},
		Spec: gwapiv1a2.GatewayClassSpec{
			ControllerName: gwapiv1a2.GatewayController(v1alpha1.GatewayControllerName),
		},
	}
	require.NoError(t, cli.Create(ctx, gc))

	defer func() {
		require.NoError(t, cli.Delete(ctx, gc))
	}()

	require.Eventually(t, func() bool {
		return cli.Get(ctx, types.NamespacedName{Name: gc.Name}, gc) == nil
	}, defaultWait, defaultTick)
	assert.Equal(t, gc.ObjectMeta.Generation, int64(1))
}
