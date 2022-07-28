package infrastructure

import (
	"errors"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
	clicfg "sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes"
)

// Manager provides the scaffolding for managing infrastructure.
type Manager struct {
	Kubernetes *kubernetes.Infra
}

// NewManager returns a new infrastructure Manager.
func NewManager(cfg *config.Server) (*Manager, error) {
	if cfg == nil {
		return nil, errors.New("server config is nil")
	}

	mgr := new(Manager)

	switch {
	case cfg.EnvoyGateway == nil ||
		cfg.EnvoyGateway.Provider == nil ||
		cfg.EnvoyGateway.Provider.Type == v1alpha1.ProviderTypeKubernetes:
		cli, err := client.New(clicfg.GetConfigOrDie(), client.Options{})
		if err != nil {
			return nil, err
		}
		mgr.Kubernetes = kubernetes.NewInfra(cli, cfg)
	default:
		// Unsupported provider type.
		return nil, fmt.Errorf("unsupported provider type %v", cfg.EnvoyGateway.Provider.Type)
	}

	return mgr, nil
}
