package service

import (
	"context"
	"fmt"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/provider/kubernetes"
)

type Service struct {
	config.Server
	ProviderResources *message.ProviderResources
}

func (s *Service) Name() string {
	return "provider"
}

func (s *Service) Start(ctx context.Context) error {
	s.Logger = s.Logger.WithValues("service", s.Name())
	if s.EnvoyGateway.Provider.Type == v1alpha1.ProviderTypeKubernetes {
		s.Logger.Info("Using provider", "type", v1alpha1.ProviderTypeKubernetes)
		p, err := kubernetes.New(s.EnvoyGateway.Gateway.ControllerName, s.Logger, s.ProviderResources)
		if err != nil {
			return fmt.Errorf("failed to create provider %s", v1alpha1.ProviderTypeKubernetes)
		}
		if err := p.Start(ctx); err != nil { //lint:ignore SA4023 provider.Start currently never returns non-nil
			return fmt.Errorf("failed to start provider %s", v1alpha1.ProviderTypeKubernetes)
		}
		return nil
	}
	// Unsupported provider.
	return fmt.Errorf("unsupported provider type %v", s.EnvoyGateway.Provider.Type)
}
