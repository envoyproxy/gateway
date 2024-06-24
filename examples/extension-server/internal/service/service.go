package service

import (
	"fmt"
	"os"

	"github.com/exampleorg/envoygateway-extension/internal/buildflags"
	luascript_controller "github.com/exampleorg/envoygateway-extension/internal/controllers/luascript"
	"github.com/exampleorg/envoygateway-extension/internal/ir"
	xds_hooks "github.com/exampleorg/envoygateway-extension/internal/xds-hooks"
	zerologadapter "github.com/exampleorg/envoygateway-extension/internal/zerolog-adapter"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"golang.org/x/sync/errgroup"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlHealthz "sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

const (
	HealthProbeAddress = "0.0.0.0:8081"
	MetricsAddress     = "0.0.0.0:8080"
)

type ExtensionService struct {
}

func NewExtensionService() *ExtensionService {
	return &ExtensionService{}
}

func (s *ExtensionService) Start() error {
	// Setup a context that will get canceled by shutdown signals
	ctx := ctrl.SetupSignalHandler()
	g, gCtx := errgroup.WithContext(ctx)

	setupLogger()
	log.Info().Msgf("starting example Envoy Gateway extension service version: %s built for %s/%s...",
		buildflags.GetVersion(),
		buildflags.GetOS(),
		buildflags.GetArch(),
	)

	// Setup IR resources to keep an internal store of all the resources in the kluster and manage updates to them
	// The controller will take the manager to manage the IR while the xDS hooks server can just use the managed resources directly
	resources := ir.NewIR()
	irManager := ir.NewIRManager(resources)

	// Start the server that will respond to the xDS modification hooks from Envoy Gateway
	xdsHooksServer := xds_hooks.NewHooksServer(resources)
	g.Go(func() error {
		return xdsHooksServer.Start(gCtx)
	})

	// Start the controller to watch our kubernetes custom resources a general recommended pattern is one controller per custom resource type
	// rather than a single controller that manages everything, otherwise it is very easy to fall into infinite reconcile loops and generally speaking
	// performance tends to be worse with a single monolithic controller that watches many different resources
	g.Go(func() error {
		clusterConfig, err := ctrl.GetConfig()
		if err != nil {
			return err
		}

		// Adapt zerolog to logr so we can make the logs from controller runtime match our global logging config
		zeroLogLogr := zerologadapter.NewLogr(log.Logger)
		ctrl.SetLogger(zeroLogLogr)
		ctrlManagerOptions := manager.Options{
			Scheme:                 GetScheme(),
			Logger:                 zeroLogLogr,
			LeaderElection:         false,
			HealthProbeBindAddress: HealthProbeAddress,
			Metrics:                server.Options{BindAddress: MetricsAddress},
		}

		ctrlManager, err := ctrl.NewManager(clusterConfig, ctrlManagerOptions)
		if err != nil {
			return fmt.Errorf("unable to create controller manager, err: %w", err)
		}

		// Add a readiness check for our xDS hooks server
		if err := ctrlManager.AddReadyzCheck("xDS hooks server", xdsHooksServer.HealthChecker()); err != nil {
			return fmt.Errorf("unable to set up xDS hooks server ready check, err: %w", err)
		}
		if err := ctrlManager.AddHealthzCheck("healthz", ctrlHealthz.Ping); err != nil {
			return fmt.Errorf("unable to set up health check, err: %w", err)
		}

		luaScriptController := luascript_controller.NewController(ctrlManager.GetClient(), irManager)
		if err := luaScriptController.SetupWithManager(ctrlManager); err != nil {
			return fmt.Errorf("unable to setup GlobalLuaScript controller, err: %w", err)
		}

		return ctrlManager.Start(gCtx)
	})

	// TODO: start any other servers or processes you need such as database
	// interactions, metrics servers, etc.
	// You can also just piggyback off of the controller runtime metrics server if you want

	// Block here until shutdown signal
	err := g.Wait()
	if err != nil {
		log.Error().Stack().Err(err).Msg("example Envoy Gateway extension service ran into an error")
	}
	log.Info().Msg("example Envoy Gateway extension service finished shutting down")

	return nil
}

// Setup logger logger config
// this example uses zerolog but there are many other good logging packages out there like zap, etc.
func setupLogger() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Set global log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Optionally, customize the time field format
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Handle errors in a way that's compatible with pkg/errors
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
}
