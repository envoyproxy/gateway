// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
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

	if err := processHTTPListenerXdsTranslation(tCtx, ir.HTTP); err != nil {
		return nil, err
	}

	if err := processTCPListenerXdsTranslation(tCtx, ir.TCP); err != nil {
		return nil, err
	}

	if err := processUDPListenerXdsTranslation(tCtx, ir.UDP); err != nil {
		return nil, err
	}

	return tCtx, nil
}

func processHTTPListenerXdsTranslation(tCtx *types.ResourceVersionTable, httpListeners []*ir.HTTPListener) error {
	for _, httpListener := range httpListeners {
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
					return errors.New("unable to find xds route config")
				}
			}
		}

		if addFilterChain {
			if err := addXdsHTTPFilterChain(xdsListener, httpListener); err != nil {
				return err
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

		// TODO: Make this into a generic interface for API Gateway features
		// Check if a ratelimit cluster exists, if not, add it, if its needed.
		if rlCluster := findXdsCluster(tCtx, getRateLimitServiceClusterName()); rlCluster == nil {
			rlCluster, err := buildRateLimitServiceCluster(httpListener)
			if err != nil {
				return multierror.Append(err, errors.New("error building ratelimit cluster"))
			}
			// Add cluster
			if rlCluster != nil {
				tCtx.AddXdsResource(resource.ClusterType, rlCluster)
			}
		}
	}
	return nil
}

func processTCPListenerXdsTranslation(tCtx *types.ResourceVersionTable, tcpListeners []*ir.TCPListener) error {
	for _, tcpListener := range tcpListeners {
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
			return err
		}
	}
	return nil
}

func processUDPListenerXdsTranslation(tCtx *types.ResourceVersionTable, udpListeners []*ir.UDPListener) error {
	for _, udpListener := range udpListeners {
		// 1:1 between IR UDPListener and xDS Cluster
		xdsCluster := buildXdsCluster(udpListener.Name, udpListener.Destinations, false /*isHTTP2 */)
		tCtx.AddXdsResource(resource.ClusterType, xdsCluster)

		// There won't be multiple UDP listeners on the same port since it's already been checked at the gateway api
		// translator
		xdsListener, err := buildXdsUDPListener(xdsCluster.Name, udpListener)
		if err != nil {
			return multierror.Append(err, errors.New("error building xds cluster"))
		}
		tCtx.AddXdsResource(resource.ListenerType, xdsListener)
	}
	return nil

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

// findXdsCluster finds a xds cluster with the same name, and returns nil if there is no match.
func findXdsCluster(tCtx *types.ResourceVersionTable, name string) *cluster.Cluster {
	if tCtx == nil || tCtx.XdsResources == nil || tCtx.XdsResources[resource.ClusterType] == nil {
		return nil
	}

	for _, r := range tCtx.XdsResources[resource.ClusterType] {
		cluster := r.(*cluster.Cluster)
		if cluster.Name == name {
			return cluster
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
