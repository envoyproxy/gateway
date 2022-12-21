// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/tetratelabs/multierror"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

// Translate translates the XDS IR into xDS resources
func Translate(ir *ir.Xds) (*types.ResourceVersionTable, error) {
	if ir == nil {
		return nil, errors.New("ir is nil")
	}

	tCtx := new(types.ResourceVersionTable)

	for _, httpListener := range ir.HTTP {
		addFilterChain := true
		var xdsRouteCfg *route.RouteConfiguration

		// Search for an existing listener, if it does not exist, create one.
		xdsListener := findXdsListener(tCtx, httpListener.Address, httpListener.Port, core.SocketAddress_TCP)
		if xdsListener == nil {
			xdsListener = buildXdsTCPListener(httpListener.Name, httpListener.Address, httpListener.Port)
			tCtx.AddXdsResource(resource.ListenerType, xdsListener)
		} else if httpListener.TLS == nil {
			// Find the route config associated with this listener that
			// maps to the default filter chain for http traffic
			routeName := findXdsHTTPRouteConfigName(xdsListener)
			if routeName != "" {
				// If an existing listener exists, dont create a new filter chain
				// for HTTP traffic, match on the Domains field within VirtualHosts
				// within the same RouteConfiguration instead
				addFilterChain = false
				xdsRouteCfg = findXdsRouteConfig(tCtx, routeName)
				if xdsRouteCfg == nil {
					return nil, errors.New("unable to find xds route config")
				}
			}
		}

		if addFilterChain {
			if err := addXdsHTTPFilterChain(xdsListener, httpListener); err != nil {
				return nil, err
			}
		}

		// Create a route config if we have not found one yet
		if xdsRouteCfg == nil {
			xdsRouteCfg = &route.RouteConfiguration{
				Name: httpListener.Name,
			}
			tCtx.AddXdsResource(resource.RouteType, xdsRouteCfg)
		}

		// 1:1 between IR TLSListenerConfig and xDS Secret
		if httpListener.TLS != nil {
			secret := buildXdsDownstreamTLSSecret(httpListener.Name, httpListener.TLS)
			tCtx.AddXdsResource(resource.SecretType, secret)
		}

		// Allocate virtual host for this httpListener.
		// 1:1 between IR HTTPListener and xDS VirtualHost
		vHost := &route.VirtualHost{
			Name:    httpListener.Name,
			Domains: httpListener.Hostnames,
		}

		for _, httpRoute := range httpListener.Routes {
			// 1:1 between IR HTTPRoute and xDS config.route.v3.Route
			xdsRoute := buildXdsRoute(httpRoute)
			vHost.Routes = append(vHost.Routes, xdsRoute)

			// Skip trying to build an IR cluster if the httpRoute only has invalid backends
			if len(httpRoute.Destinations) == 0 && httpRoute.BackendWeights.Invalid > 0 {
				continue
			}
			xdsCluster := buildXdsCluster(httpRoute.Name, httpRoute.Destinations, httpListener.IsHTTP2)
			tCtx.AddXdsResource(resource.ClusterType, xdsCluster)

		}

		xdsRouteCfg.VirtualHosts = append(xdsRouteCfg.VirtualHosts, vHost)
	}

	for _, tcpListener := range ir.TCP {
		// 1:1 between IR TCPListener and xDS Cluster
		xdsCluster := buildXdsCluster(tcpListener.Name, tcpListener.Destinations, false /*isHTTP2 */)
		tCtx.AddXdsResource(resource.ClusterType, xdsCluster)

		// Search for an existing listener, if it does not exist, create one.
		xdsListener := findXdsListener(tCtx, tcpListener.Address, tcpListener.Port, core.SocketAddress_TCP)
		if xdsListener == nil {
			xdsListener = buildXdsTCPListener(tcpListener.Name, tcpListener.Address, tcpListener.Port)
			tCtx.AddXdsResource(resource.ListenerType, xdsListener)
		}

		if err := addXdsTCPFilterChain(xdsListener, tcpListener, xdsCluster.Name); err != nil {
			return nil, err
		}
	}

	for _, udpListener := range ir.UDP {
		// 1:1 between IR UDPListener and xDS Cluster
		xdsCluster := buildXdsCluster(udpListener.Name, udpListener.Destinations, false /*isHTTP2 */)
		tCtx.AddXdsResource(resource.ClusterType, xdsCluster)

		// There won't be multiple UDP listeners on the same port since it's already been checked at the gateway api
		// translator
		xdsListener, err := buildXdsUDPListener(xdsCluster.Name, udpListener)
		if err != nil {
			return nil, multierror.Append(err, errors.New("error building xds cluster"))
		}
		tCtx.AddXdsResource(resource.ListenerType, xdsListener)
	}
	return tCtx, nil
}

// findXdsListener finds a xds listener with the same address, port and protocol, and returns nil if there is no match.
func findXdsListener(tCtx *types.ResourceVersionTable, address string, port uint32,
	protocol core.SocketAddress_Protocol) *listener.Listener {
	if tCtx == nil || tCtx.XdsResources == nil || tCtx.XdsResources[resource.ListenerType] == nil {
		return nil
	}

	for _, r := range tCtx.XdsResources[resource.ListenerType] {
		listener := r.(*listener.Listener)
		addr := listener.GetAddress()
		if addr.GetSocketAddress().GetPortValue() == port && addr.GetSocketAddress().Address == address && addr.
			GetSocketAddress().Protocol == protocol {
			return listener
		}
	}

	return nil
}

// findXdsRouteConfig finds an xds route with the name and returns nil if there is no match.
func findXdsRouteConfig(tCtx *types.ResourceVersionTable, name string) *route.RouteConfiguration {
	if tCtx == nil || tCtx.XdsResources == nil || tCtx.XdsResources[resource.RouteType] == nil {
		return nil
	}

	for _, r := range tCtx.XdsResources[resource.RouteType] {
		route := r.(*route.RouteConfiguration)
		if route.Name == name {
			return route
		}
	}

	return nil
}

// Point to xds cluster.
func makeConfigSource() *core.ConfigSource {
	source := &core.ConfigSource{}
	source.ResourceApiVersion = resource.DefaultAPIVersion
	source.ConfigSourceSpecifier = &core.ConfigSource_ApiConfigSource{
		ApiConfigSource: &core.ApiConfigSource{
			TransportApiVersion:       resource.DefaultAPIVersion,
			ApiType:                   core.ApiConfigSource_DELTA_GRPC,
			SetNodeOnFirstMessageOnly: true,
			GrpcServices: []*core.GrpcService{{
				TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: "xds_cluster"},
				},
			}},
		},
	}
	return source
}
