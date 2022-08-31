package infrastructure

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
	clicfg "sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes"
)

// Manager provides the scaffolding for managing infrastructure.
type Manager struct {
	// TODO: create a common infra interface
	*kubernetes.Infra
}

// NewManager returns a new infrastructure Manager.
func NewManager(cfg *config.Server) (*Manager, error) {
	mgr := new(Manager)

	if cfg.EnvoyGateway.Provider.Type == v1alpha1.ProviderTypeKubernetes {
		cli, err := client.New(clicfg.GetConfigOrDie(), client.Options{Scheme: envoygateway.GetScheme()})
		if err != nil {
			return nil, err
		}
		mgr.Infra = kubernetes.NewInfra(cli)
	} else {
		// Kube is the only supported provider type.
		return nil, fmt.Errorf("unsupported provider type %v", cfg.EnvoyGateway.Provider.Type)
	}

	return mgr, nil
}
