package infrastructure

import (
	"errors"
	"fmt"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes"
)

// Context provides the scaffolding for managing infrastructure.
type Context struct {
	Provider   *v1alpha1.ProviderType
	Kubernetes *kubernetes.Context
}

// NewContext returns a new infrastructure Context.
func NewContext(cfg *config.Server) (*Context, error) {
	if cfg == nil {
		return nil, errors.New("server config is nil")
	}

	ctx := new(Context)

	switch {
	case cfg.EnvoyGateway == nil || cfg.EnvoyGateway.Provider == nil:
		// Kube is the default provider type.
		ctx.Provider = v1alpha1.ProviderTypePtr(cfg.EnvoyGateway.Provider.Type)
	case cfg.EnvoyGateway.Provider.Type == v1alpha1.ProviderTypeKubernetes:
		ctx.Provider = v1alpha1.ProviderTypePtr(cfg.EnvoyGateway.Provider.Type)
	default:
		// Unsupported provider type.
		return nil, fmt.Errorf("unsupported provider type %v", cfg.EnvoyGateway.Provider.Type)
	}

	return ctx, nil
}
