package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/ir"
)

// Translate translates the provided infra ir into managed infrastructure.
func Translate(ctx context.Context, cfg *config.Server, infra *ir.Infra) (*Manager, error) {
	if err := ir.ValidateInfra(infra); err != nil {
		return nil, err
	}

	if cfg == nil {
		return nil, errors.New("server config is nil")
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		return nil, err
	}

	log := cfg.Logger

	if cfg.EnvoyGateway == nil ||
		cfg.EnvoyGateway.Provider == nil ||
		cfg.EnvoyGateway.Provider.Type == v1alpha1.ProviderTypeKubernetes {
		log.Info("Using infra manager", "type", v1alpha1.ProviderTypeKubernetes)

		// A nil infra proxy ir means the proxy infra should be deleted, but metadata is
		// required to know the ns/name of the resources to delete. Add support for deleting
		// the infra when https://github.com/envoyproxy/gateway/issues/173 is resolved.

		if err := mgr.Kubernetes.CreateInfra(ctx, infra); err != nil {
			return nil, fmt.Errorf("failed to create kube infra: %v", err)
		}
		return mgr, nil
	}

	return nil, fmt.Errorf("unsupported infra manager type %v", cfg.EnvoyGateway.Provider.Type)
}
