package provider

import (
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/provider/kubernetes"
)

func Start(svr *config.Server) error {
	log := svr.Logger
	if svr.EnvoyGateway.Provider.Type == v1alpha1.ProviderTypeKubernetes {
		log.Info("Using provider", "type", v1alpha1.ProviderTypeKubernetes)
		cfg := ctrl.GetConfigOrDie()
		provider, err := kubernetes.New(cfg, svr)
		if err != nil {
			return fmt.Errorf("failed to create provider %s", v1alpha1.ProviderTypeKubernetes)
		}
		if err := provider.Start(ctrl.SetupSignalHandler()); err != nil { //lint:ignore SA4023 provider.Start currently never returns non-nil
			return fmt.Errorf("failed to start provider %s", v1alpha1.ProviderTypeKubernetes)
		}
		return nil
	}
	// Unsupported provider.
	return fmt.Errorf("unsupported provider type %v", svr.EnvoyGateway.Provider.Type)
}
