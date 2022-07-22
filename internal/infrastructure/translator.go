package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
	clicfg "sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes"
	"github.com/envoyproxy/gateway/internal/ir"
)

// Translate translates the provided infra ir into managed infrastructure.
func Translate(ctx context.Context, cfg *config.Server, infra *ir.Infra) (*Context, error) {
	if err := ir.ValidateInfra(infra); err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, errors.New("server config is nil")
	}

	infraCtx, err := NewContext(cfg)
	if err != nil {
		return nil, err
	}

	log := cfg.Logger

	if *infraCtx.Provider == v1alpha1.ProviderTypeKubernetes {
		log.Info("Using infra manager", "type", v1alpha1.ProviderTypeKubernetes)
		cli, err := client.New(clicfg.GetConfigOrDie(), client.Options{})
		if err != nil {
			return nil, err
		}
		kubeCtx := kubernetes.NewContext(cli, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create kube infra manager context: %v", err)
		}
		infraCtx.Kubernetes = kubeCtx

		// A nil infra proxy ir means the proxy infra should be deleted, but metadata is
		// required to know the ns/name of the resources to delete. Add support for deleting
		// the infra when https://github.com/envoyproxy/gateway/issues/173 is resolved.

		if err := infraCtx.Kubernetes.CreateIfNeeded(ctx, infra); err != nil {
			return nil, fmt.Errorf("failed to create infra: %v", err)
		}
		return infraCtx, nil
	}

	return nil, fmt.Errorf("unsupported infra manager type %v", cfg.EnvoyGateway.Provider.Type)
}
