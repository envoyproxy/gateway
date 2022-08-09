package service

import (
	"context"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/xds/translator"
)

type Service struct {
	config.Server
	XdsIR        *message.XdsIR
	XdsResources *message.XdsResources
}

func (s *Service) Name() string {
	return "xds-translator"
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
	irCh := s.XdsIR.Subscribe(ctx)
	for ctx.Err() == nil {
		// Receive subscribed resource notifications
		<-irCh
		ir := s.XdsIR.Get()
		// Translate to xds resources
		result, err := translator.Translate(ir)
		if err != nil {
			s.Logger.Error(err, "failed to translate xds ir")
		} else {
			// Publish
			// There should always be a single element in the map
			// Use the service name as the key for now
			xdsResources := result.GetXdsResources()
			s.XdsResources.Store(s.Name(), &xdsResources)
		}
	}
}
