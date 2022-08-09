package service

import (
	"context"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
)

type Service struct {
	config.Server
	InfraIR *message.InfraIR
}

func (s *Service) Name() string {
	return "infrastructure"
}

// Start starts the GatewayAPI service
func (s *Service) Start(ctx context.Context) error {
	log := s.Logger.WithValues("service", s.Name())
	go s.subscribeAndTranslate(ctx)

	<-ctx.Done()
	log.Info("shutting down")
	return nil
}

func (s *Service) subscribeAndTranslate(ctx context.Context) {
	// Subscribe to resources
	irCh := s.InfraIR.Subscribe(ctx)
	for ctx.Err() == nil {
		// Receive subscribed resource notifications
		<-irCh
		// s.InfraIR.Get()
		// TODO: Provision infra
		// infrastructure.Translate(ctx, ir)
	}
}
