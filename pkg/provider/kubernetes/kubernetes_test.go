package kubernetes

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/envoyproxy/gateway/pkg/envoygateway/config"
)

func TestProvider(t *testing.T) {
	// Setup the test environment.
	testEnv, err := startEnv()
	require.NoError(t, err)

	// Setup and start the kube provider.
	cfg, err := config.NewDefaultServer()
	require.NoError(t, err)
	provider, err := New(cfg)
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
}

func startEnv() (*envtest.Environment, error) {
	log.SetLogger(zap.New(zap.WriteTo(os.Stderr), zap.UseDevMode(true)))
	env := &envtest.Environment{}
	_, err := env.Start()
	if err != nil {
		return nil, err
	}
	return env, nil
}
