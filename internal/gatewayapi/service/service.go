package service

import (
	"context"

	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/message"
)

type Service struct {
	config.Server
	ProviderResources *message.ProviderResources
	XdsIR             *message.XdsIR
	InfraIR           *message.InfraIR
}

func (s *Service) Name() string {
	return "gateway-api"
}

// Start starts the GatewayAPI service
func (s *Service) Start(ctx context.Context) error {
	s.Logger = s.Logger.WithValues("service", s.Name())
	// Wait until provider resources have been initialized during startup
	s.ProviderResources.Initialized.Wait()
	go s.subscribeAndTranslate(ctx)

	<-ctx.Done()
	s.Logger.Info("shutting down")
	return nil
}

func (s *Service) subscribeAndTranslate(ctx context.Context) {
	// Subscribe to resources
	gatewayClassesCh := s.ProviderResources.GatewayClasses.Subscribe(ctx)
	gatewaysCh := s.ProviderResources.Gateways.Subscribe(ctx)
	for ctx.Err() == nil {
		var r gatewayapi.Resources
		// Receive subscribed resource notifications
		select {
		case <-gatewayClassesCh:
		case <-gatewaysCh:
		}
		// Load all resources required for translation
		r.Gateways = s.ProviderResources.GetGateways()
		gatewayClasses := s.ProviderResources.GetGatewayClasses()
		// Fetch the first gateway class since there should be only 1
		// gateway class linked to this controller
		t := &gatewayapi.Translator{
			GatewayClassName: v1beta1.ObjectName(gatewayClasses[0].GetName()),
		}
		// Translate to IR
		result := t.Translate(&r)
		// Publish the IR
		// Use the service name as the key to ensure there is always
		// one element in the map
		s.XdsIR.Store(s.Name(), result.IR)
	}
}
