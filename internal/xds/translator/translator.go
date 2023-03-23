// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/tetratelabs/multierror"

	resourceTypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"

	extensionTypes "github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

// Translator translates the xDS IR into xDS resources.
type Translator struct {
	// GlobalRateLimit holds the global rate limit settings
	// required during xds translation.
	GlobalRateLimit *GlobalRateLimitSettings
}

type GlobalRateLimitSettings struct {
	// ServiceURL is the URL of the global
	// rate limit service.
	ServiceURL string
}

// Translate translates the XDS IR into xDS resources
func (t *Translator) Translate(ir *ir.Xds, extManager extensionTypes.Manager) (*types.ResourceVersionTable, error) {
	if ir == nil {
		return nil, errors.New("ir is nil")
	}

	tCtx := new(types.ResourceVersionTable)

	if err := t.processHTTPListenerXdsTranslation(tCtx, ir.HTTP, extManager); err != nil {
		return nil, err
	}

	if err := processTCPListenerXdsTranslation(tCtx, ir.TCP); err != nil {
		return nil, err
	}

	if err := processUDPListenerXdsTranslation(tCtx, ir.UDP); err != nil {
		return nil, err
	}

	// If there is a loaded extension that wants to inject clusters/secrets, then call it
	// while clusters can by statically added with bootstrap configuration, an extension may need to add clusters with a configuration
	// that is non-static. If a cluster definition is unlikely to change over the course of an extension's lifetime then the custom bootstrap config
	// is the preferred way of adding it.
	extensionInsertHookClient := extManager.GetXDSHookClient(extensionTypes.PostXDSTranslation)
	if extensionInsertHookClient != nil {
		newClusters, newSecrets, err := extensionInsertHookClient.PostTranslationInsertHook()
		if err != nil {
			return nil, err
		}

		// We're assuming that Cluster names are unique.
		for _, addedCluster := range newClusters {
			tCtx.AddOrReplaceXdsResource(resourcev3.ClusterType, addedCluster, func(existing resourceTypes.Resource, new resourceTypes.Resource) bool {
				oldCluster := existing.(*clusterv3.Cluster)
				newCluster := new.(*clusterv3.Cluster)
				if newCluster == nil || oldCluster == nil {
					return false
				}
				if oldCluster.Name == newCluster.Name {
					return true
				}
				return false
			})
		}

		for _, secret := range newSecrets {
			tCtx.AddXdsResource(resourcev3.SecretType, secret)
		}
	}

	return tCtx, nil
}

func (t *Translator) processHTTPListenerXdsTranslation(tCtx *types.ResourceVersionTable, httpListeners []*ir.HTTPListener, extManager extensionTypes.Manager) error {
	for _, httpListener := range httpListeners {
		addFilterChain := true
		var xdsRouteCfg *routev3.RouteConfiguration

		// Search for an existing listener, if it does not exist, create one.
		xdsListener := findXdsListener(tCtx, httpListener.Address, httpListener.Port, corev3.SocketAddress_TCP)
		if xdsListener == nil {
			xdsListener = buildXdsTCPListener(httpListener.Name, httpListener.Address, httpListener.Port)
			tCtx.AddXdsResource(resourcev3.ListenerType, xdsListener)
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
			if err := t.addXdsHTTPFilterChain(xdsListener, httpListener); err != nil {
				return err
			}
		}

		// Create a route config if we have not found one yet
		if xdsRouteCfg == nil {
			xdsRouteCfg = &routev3.RouteConfiguration{
				Name: httpListener.Name,
			}
			tCtx.AddXdsResource(resourcev3.RouteType, xdsRouteCfg)
		}

		// 1:1 between IR TLSListenerConfig and xDS Secret
		if httpListener.TLS != nil {
			secret := buildXdsDownstreamTLSSecret(httpListener.Name, httpListener.TLS)
			tCtx.AddXdsResource(resourcev3.SecretType, secret)
		}

		// Allocate virtual host for this httpListener.
		// 1:1 between IR HTTPListener and xDS VirtualHost
		vHost := &routev3.VirtualHost{
			Name:    httpListener.Name,
			Domains: httpListener.Hostnames,
		}
		protocol := DefaultProtocol
		if httpListener.IsHTTP2 {
			protocol = HTTP2
		}

		// Check if an extension is loaded that wants to modify xDS Routes after they have been generated
		extRouteHookClient := extManager.GetXDSHookClient(extensionTypes.PostXDSRoute)
		for _, httpRoute := range httpListener.Routes {
			// 1:1 between IR HTTPRoute and xDS config.route.v3.Route
			xdsRoute := buildXdsRoute(httpRoute, xdsListener)

			// If a loaded extension wants to modify routes and the extension's resources are used on the HTTPRoute then call the extension hook
			if extRouteHookClient != nil && len(httpRoute.ExtensionRefs) > 0 {

				modifiedRoute, err := extRouteHookClient.PostRouteModifyHook(
					xdsRoute,
					vHost.Domains,
					httpRoute.ExtensionRefs,
				)
				if err != nil {
					// Maybe logging the error is better here, but this only happens when an extension is in-use
					// so if modification fails then we should probably treat that as a serious problem.
					return err
				}

				// An extension is allowed to return a nil route to prevent it from being added
				if modifiedRoute != nil {
					vHost.Routes = append(vHost.Routes, modifiedRoute)
				}
			} else {
				vHost.Routes = append(vHost.Routes, xdsRoute)
			}

			// Skip trying to build an IR cluster if the httpRoute only has invalid backends
			if len(httpRoute.Destinations) == 0 && httpRoute.BackendWeights.Invalid > 0 {
				continue
			}
			addXdsCluster(tCtx, addXdsClusterArgs{
				name:         httpRoute.Name,
				destinations: httpRoute.Destinations,
				tSocket:      nil,
				protocol:     protocol,
				endpoint:     Static,
			})

			// If the httpRoute has a list of mirrors create clusters for them unless they already have one
			for i, mirror := range httpRoute.Mirrors {
				mirrorClusterName := fmt.Sprintf("%s-mirror-%d", httpRoute.Name, i)
				if cluster := findXdsCluster(tCtx, mirrorClusterName); cluster == nil {
					addXdsCluster(tCtx, addXdsClusterArgs{
						name:         mirrorClusterName,
						destinations: []*ir.RouteDestination{mirror},
						tSocket:      nil,
						protocol:     protocol,
						endpoint:     Static,
					})
				}

			}
		}

		// Check if an extension is loaded that wants to modify xDS VirtualHosts after they have been generated
		extVHHookClient := extManager.GetXDSHookClient(extensionTypes.PostXDSVirtualHost)
		// If no extension exists that wants to modify the VirtualHost then be done with it
		if extVHHookClient == nil {
			xdsRouteCfg.VirtualHosts = append(xdsRouteCfg.VirtualHosts, vHost)
		} else {
			modifiedVH, err := extVHHookClient.PostVirtualHostModifyHook(vHost)
			if err != nil {
				// Maybe logging the error is better here, but this only happens when an extension is in-use
				// so if modification fails then we should probably treat that as a serious problem.
				return err
			}

			if modifiedVH != nil {
				xdsRouteCfg.VirtualHosts = append(xdsRouteCfg.VirtualHosts, vHost)
			}
		}

		// TODO: Make this into a generic interface for API Gateway features.
		//       https://github.com/envoyproxy/gateway/issues/882
		// Check if a ratelimit cluster exists, if not, add it, if its needed.
		if err := t.createRateLimitServiceCluster(tCtx, httpListener); err != nil {
			return err
		}

		// Create authn jwks clusters, if needed.
		if err := createJwksClusters(tCtx, httpListener.Routes); err != nil {
			return err
		}
		// Check if an extension want to modify the listener that was just configured/created
		extListenerHookClient := extManager.GetXDSHookClient(extensionTypes.PostXDSHTTPListener)
		if extListenerHookClient != nil {
			modifiedListener, err := extListenerHookClient.PostHTTPListenerModifyHook(xdsListener)
			if err != nil {
				return err
			} else if modifiedListener != nil {
				// Use the resource table to update the listener with the modified version returned by the extension
				// We're assuming that Listener names are unique.
				tCtx.AddOrReplaceXdsResource(resourcev3.ListenerType, modifiedListener, func(existing resourceTypes.Resource, new resourceTypes.Resource) bool {
					oldListener := existing.(*listenerv3.Listener)
					newListener := new.(*listenerv3.Listener)
					if newListener == nil || oldListener == nil {
						return false
					}
					if oldListener.Name == newListener.Name {
						return true
					}
					return false
				})

			}

		}
	}

	return nil
}

func processTCPListenerXdsTranslation(tCtx *types.ResourceVersionTable, tcpListeners []*ir.TCPListener) error {
	for _, tcpListener := range tcpListeners {
		// 1:1 between IR TCPListener and xDS Cluster
		addXdsCluster(tCtx, addXdsClusterArgs{
			name:         tcpListener.Name,
			destinations: tcpListener.Destinations,
			tSocket:      nil,
			protocol:     DefaultProtocol,
			endpoint:     Static,
		})

		// Search for an existing listener, if it does not exist, create one.
		xdsListener := findXdsListener(tCtx, tcpListener.Address, tcpListener.Port, corev3.SocketAddress_TCP)
		if xdsListener == nil {
			xdsListener = buildXdsTCPListener(tcpListener.Name, tcpListener.Address, tcpListener.Port)
			tCtx.AddXdsResource(resourcev3.ListenerType, xdsListener)
		}

		if err := addXdsTCPFilterChain(xdsListener, tcpListener, tcpListener.Name); err != nil {
			return err
		}
	}
	return nil
}

func processUDPListenerXdsTranslation(tCtx *types.ResourceVersionTable, udpListeners []*ir.UDPListener) error {
	for _, udpListener := range udpListeners {
		// 1:1 between IR UDPListener and xDS Cluster
		addXdsCluster(tCtx, addXdsClusterArgs{
			name:         udpListener.Name,
			destinations: udpListener.Destinations,
			tSocket:      nil,
			protocol:     DefaultProtocol,
			endpoint:     Static,
		})

		// There won't be multiple UDP listeners on the same port since it's already been checked at the gateway api
		// translator
		xdsListener, err := buildXdsUDPListener(udpListener.Name, udpListener)
		if err != nil {
			return multierror.Append(err, errors.New("error building xds cluster"))
		}
		tCtx.AddXdsResource(resourcev3.ListenerType, xdsListener)
	}
	return nil

}

// findXdsListener finds a xds listener with the same address, port and protocol, and returns nil if there is no match.
func findXdsListener(tCtx *types.ResourceVersionTable, address string, port uint32,
	protocol corev3.SocketAddress_Protocol) *listenerv3.Listener {
	if tCtx == nil || tCtx.XdsResources == nil || tCtx.XdsResources[resourcev3.ListenerType] == nil {
		return nil
	}

	for _, r := range tCtx.XdsResources[resourcev3.ListenerType] {
		listener := r.(*listenerv3.Listener)
		addr := listener.GetAddress()
		if addr.GetSocketAddress().GetPortValue() == port && addr.GetSocketAddress().Address == address && addr.
			GetSocketAddress().Protocol == protocol {
			return listener
		}
	}

	return nil
}

// findXdsCluster finds a xds cluster with the same name, and returns nil if there is no match.
func findXdsCluster(tCtx *types.ResourceVersionTable, name string) *clusterv3.Cluster {
	if tCtx == nil || tCtx.XdsResources == nil || tCtx.XdsResources[resourcev3.ClusterType] == nil {
		return nil
	}

	for _, r := range tCtx.XdsResources[resourcev3.ClusterType] {
		cluster := r.(*clusterv3.Cluster)
		if cluster.Name == name {
			return cluster
		}
	}

	return nil
}

func addXdsCluster(tCtx *types.ResourceVersionTable, args addXdsClusterArgs) {
	xdsCluster := buildXdsCluster(args.name, args.tSocket, args.protocol, args.endpoint)
	xdsEndpoints := buildXdsClusterLoadAssignment(args.name, args.destinations)
	// Use EDS for static endpoints
	if args.endpoint == Static {
		tCtx.AddXdsResource(resourcev3.EndpointType, xdsEndpoints)
	} else {
		xdsCluster.LoadAssignment = xdsEndpoints
	}
	tCtx.AddXdsResource(resourcev3.ClusterType, xdsCluster)
}

// findXdsRouteConfig finds an xds route with the name and returns nil if there is no match.
func findXdsRouteConfig(tCtx *types.ResourceVersionTable, name string) *routev3.RouteConfiguration {
	if tCtx == nil || tCtx.XdsResources == nil || tCtx.XdsResources[resourcev3.RouteType] == nil {
		return nil
	}

	for _, r := range tCtx.XdsResources[resourcev3.RouteType] {
		route := r.(*routev3.RouteConfiguration)
		if route.Name == name {
			return route
		}
	}

	return nil
}

type addXdsClusterArgs struct {
	name         string
	destinations []*ir.RouteDestination
	tSocket      *corev3.TransportSocket
	protocol     ProtocolType
	endpoint     EndpointType
}

type ProtocolType int
type EndpointType int

const (
	DefaultProtocol ProtocolType = iota
	TCP
	UDP
	HTTP
	HTTP2
)

const (
	DefaultEndpointType EndpointType = iota
	Static
	EDS
)
