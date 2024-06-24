package xds_hooks

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/exampleorg/envoygateway-extension/internal/ir"
	"github.com/rs/zerolog/log"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	egProto "github.com/envoyproxy/gateway/proto/extension"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	HooksServerAddress = "0.0.0.0:9000"
)

type HooksServer struct {
	ready     bool
	resources *ir.IR
	egProto.UnimplementedEnvoyGatewayExtensionServer
}

func NewHooksServer(resources *ir.IR) *HooksServer {
	return &HooksServer{resources: resources}
}

func (h *HooksServer) HealthChecker() healthz.Checker {
	return func(req *http.Request) error {
		if !h.ready {
			return fmt.Errorf("xds hooks server is not ready yet")
		}
		return nil
	}
}

func (h *HooksServer) Start(ctx context.Context) error {
	server := grpc.NewServer()
	listener, err := net.Listen("tcp", HooksServerAddress)
	if err != nil {
		log.Error().Stack().Err(err).Str("address", HooksServerAddress).Msg("xds hooks server setup error")
	}

	egProto.RegisterEnvoyGatewayExtensionServer(server, h)

	errChan := make(chan error)
	go func() {
		log.Info().Str("address", HooksServerAddress).Msg("starting xds hooks server")
		h.ready = true
		errChan <- server.Serve(listener)
	}()

	// Wait for grpc accept error or shutdown
	select {
	case <-ctx.Done():
		log.Info().Str("address", HooksServerAddress).Msg("xds hooks server graceful shutdown started")
		server.GracefulStop()
		log.Info().Str("address", HooksServerAddress).Msg("xds hooks server graceful shutdown without errors")
		return nil
	case err := <-errChan:
		log.Error().Stack().Err(err).Str("address", HooksServerAddress).Msg("xds hooks server encountered an error")
		return err
	}
}

// We don't actually use this hook for our purposes, but we still provide implementations to fulfill the interface requirements of the gRPC server
// If we don't subscribe to these hooks in the EnvoyGateway config, then we won't receive these requests from Envoy Gateway.
func (h *HooksServer) PostRouteModify(ctx context.Context, req *egProto.PostRouteModifyRequest) (*egProto.PostRouteModifyResponse, error) {
	log.Info().Msg("handling xDS Route modification request from Envoy Gateway")
	// We're not doing anything here so just send back the route without modifying it
	return &egProto.PostRouteModifyResponse{Route: req.Route}, nil
}

func (h *HooksServer) PostVirtualHostModify(ctx context.Context, req *egProto.PostVirtualHostModifyRequest) (*egProto.PostVirtualHostModifyResponse, error) {
	log.Info().Msg("handling xDS VirtualHost modification request from Envoy Gateway")
	// We're not doing anything here so just send back the route without modifying it
	// If you need to inject more routes, you can do so here
	return &egProto.PostVirtualHostModifyResponse{VirtualHost: req.VirtualHost}, nil
}

// This is the only hook that we need for this example service. We're going to inject an http lua filter into the Envoy config for each one of our
// GlobalLuaScript resources
func (h *HooksServer) PostHTTPListenerModify(ctx context.Context, req *egProto.PostHTTPListenerModifyRequest) (*egProto.PostHTTPListenerModifyResponse, error) {
	log.Info().Msg("handling xDS HTTP Listener modification request from Envoy Gateway")
	ret := &egProto.PostHTTPListenerModifyResponse{}

	listener := req.Listener
	if listener == nil {
		ret.Listener = req.Listener
		err := errors.New("unable to modify xDS HTTP Listener from Envoy Gateway, resource is nil")
		log.Error().Err(err).Msg("error modifying xDS HTTP Listener")
		return ret, err
	}

	// First, get the filter chains from the listener
	filterChains := listener.GetFilterChains()
	defaultFC := listener.DefaultFilterChain
	if defaultFC != nil {
		filterChains = append(filterChains, defaultFC)
	}

	// In the filter chains, we're looking for the HTTP Connection manager
	for _, fc := range filterChains {
		hcm, hcmIndex, err := findHCM(fc)
		if err != nil {
			// If this happens, there is probably something very wrong with the Envoy config
			log.Error().Err(err).Msg("unable to find HTTP connection manager in HTTP Listener, this should never happen")
			return ret, err
		}

		// First let's make sure there aren't any of our HTTP filters already in here
		existingFilters := []*hcmv3.HttpFilter{}
		for _, filter := range hcm.HttpFilters {
			if filter == nil {
				continue
			}
			// If it starts with the names we're injecting then discard it since we might need to rebuild it.
			// Rebuilding them all is probably faster than comparing them, but performance doesn't really matter too much for this demo
			if !strings.HasPrefix(filter.Name, "exampleorg.io.GlobalLuaFilter-") {
				existingFilters = append(existingFilters, filter)
			}
		}

		// Next we'll build our list of http filters from our custom resources
		httpLuaFilters, err := buildHTTPLuaFilters(h.resources)
		if err != nil {
			log.Error().Err(err).Msg("unable to create HTTP filters for xDS HTTP Listener")
			return ret, err
		}

		// We'll prepend our lua filters to the ones that were already there so they run first thing

		finalFilters := httpLuaFilters
		finalFilters = append(finalFilters, existingFilters...)

		log.Info().Msgf("filters to be set for listener: {\"number\": %d} {\"content\": %v}", len(finalFilters), finalFilters)

		hcm.HttpFilters = finalFilters

		// The Listener Filters are Protobuf "Any" and we had to unmarshall the HTTPConnectionManager
		// into a new allocated object, so we need to write that new object back to the filter chain
		anyConnectionMgr, _ := anypb.New(hcm)
		fc.Filters[hcmIndex].ConfigType = &listenerv3.Filter_TypedConfig{
			TypedConfig: anyConnectionMgr,
		}
	}

	ret.Listener = listener
	return ret, nil
}

func (h *HooksServer) PostTranslateModify(ctx context.Context, req *egProto.PostTranslateModifyRequest) (*egProto.PostTranslateModifyResponse, error) {
	log.Info().Msg("handling xDS finilization request from Envoy Gateway")
	// We're not doing anything here so just send back the clusters and secrets without modifying them
	// If you need to inject new clusters/secrets or modify existing ones you can do that here
	return &egProto.PostTranslateModifyResponse{
		Clusters: req.Clusters,
		Secrets:  req.Secrets,
	}, nil
}
