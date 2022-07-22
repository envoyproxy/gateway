package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

// Translate translates the provided infra into managed infrastructure.
func Translate(ctx context.Context, infra *ir.Infra) (*Manager, error) {
	if err := ir.ValidateInfra(infra); err != nil {
		return nil, err
	}

	if infra == nil {
		return nil, errors.New("infra is nil")
	}

	mgr, err := NewManager(infra)
	if err != nil {
		return nil, err
	}

	log, err := logr.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Kube is the only supported provider type.
	if *infra.GetProvider() == v1alpha1.ProviderTypeKubernetes {
		log.Info("Using provider", "type", v1alpha1.ProviderTypeKubernetes)

		// A nil infra proxy ir means the proxy infra should be deleted, but metadata is
		// required to know the ns/name of the resources to delete. Add support for deleting
		// the infra when https://github.com/envoyproxy/gateway/issues/173 is resolved.

		if err := mgr.Kubernetes.CreateInfra(ctx, infra); err != nil {
			return nil, fmt.Errorf("failed to create kube infra: %v", err)
		}
		return mgr, nil
	}

	return nil, fmt.Errorf("unsupported provider type %v", infra.Provider)
}
