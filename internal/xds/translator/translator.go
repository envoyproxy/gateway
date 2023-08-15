// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"time"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/tetratelabs/multierror"

	extensionTypes "github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

// Translator translates the xDS IR into xDS resources.
type Translator struct {
	// GlobalRateLimit holds the global rate limit settings
	// required during xds translation.
	GlobalRateLimit *GlobalRateLimitSettings

	// ExtensionManager holds the config for interacting with extensions when generating xDS
	// resources. Only required during xds translation.
	ExtensionManager *extensionTypes.Manager
}

type GlobalRateLimitSettings struct {
	// ServiceURL is the URL of the global
	// rate limit service.
	ServiceURL string

	// Timeout specifies the timeout period for the proxy to access the ratelimit server
	// If not set, timeout is 20000000(20ms).
	Timeout time.Duration

	// FailClosed is a switch used to control the flow of traffic
	// when the response from the ratelimit server cannot be obtained.
	FailClosed bool
}

// Translate translates the XDS IR into xDS resources
func (t *Translator) Translate(ir *ir.Xds) (*types.ResourceVersionTable, error) {
	if ir == nil {
		return nil, errors.New("ir is nil")
	}

	tCtx := new(types.ResourceVersionTable)

	if err := t.processHTTPListenerXdsTranslation(tCtx, ir.HTTP, ir.AccessLog, ir.Tracing); err != nil {
		return nil, err
	}

	if err := processTCPListenerXdsTranslation(tCtx, ir.TCP, ir.AccessLog); err != nil {
		return nil, err
	}

	if err := processUDPListenerXdsTranslation(tCtx, ir.UDP, ir.AccessLog); err != nil {
		return nil, err
	}

	if err := processJSONPatches(tCtx, ir.EnvoyPatchPolicies); err != nil {
		return nil, err
	}

	if err := processClusterForAccessLog(tCtx, ir.AccessLog); err != nil {
		return nil, err
	}
	if err := processClusterForTracing(tCtx, ir.Tracing); err != nil {
		return nil, err
	}

	// Check if an extension want to inject any clusters/secrets
	// If no extension exists (or it doesn't subscribe to this hook) then this is a quick no-op
	if err := processExtensionPostTranslationHook(tCtx, t.ExtensionManager); err != nil {
		return nil, err
	}

	return tCtx, nil
}

func (t *Translator) processHTTPListenerXdsTranslation(tCtx *types.ResourceVersionTable, httpListeners []*ir.HTTPListener,
	accesslog *ir.AccessLog, tracing *ir.Tracing) error {
	for _, httpListener := range httpListeners {
		addFilterChain := true
		var xdsRouteCfg *routev3.RouteConfiguration

		// Search for an existing listener, if it does not exist, create one.
		xdsListener := findXdsListenerByHostPort(tCtx, httpListener.Address, httpListener.Port, corev3.SocketAddress_TCP)
		if xdsListener == nil {
			xdsListener = buildXdsTCPListener(httpListener.Name, httpListener.Address, httpListener.Port, accesslog)
			if err := tCtx.AddXdsResource(resourcev3.ListenerType, xdsListener); err != nil {
				return err
			}
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
			if err := t.addXdsHTTPFilterChain(xdsListener, httpListener, accesslog, tracing); err != nil {
				return err
			}
		}

		// Create a route config if we have not found one yet
		if xdsRouteCfg == nil {
			xdsRouteCfg = &routev3.RouteConfiguration{
				IgnorePortInHostMatching: true,
				Name:                     httpListener.Name,
			}

			if err := tCtx.AddXdsResource(resourcev3.RouteType, xdsRouteCfg); err != nil {
				return err
			}
		}

		// 1:1 between IR TLSListenerConfig and xDS Secret
		if httpListener.TLS != nil {
			for t := range httpListener.TLS {
				secret := buildXdsDownstreamTLSSecret(httpListener.TLS[t])
				if err := tCtx.AddXdsResource(resourcev3.SecretType, secret); err != nil {
					return err
				}
			}
		}

		protocol := DefaultProtocol
		if httpListener.IsHTTP2 {
			protocol = HTTP2
		}

		// store virtual hosts by domain
		vHosts := map[string]*routev3.VirtualHost{}
		// keep track of order by using a list as well as the map
		var vHostsList []*routev3.VirtualHost

		// Check if an extension is loaded that wants to modify xDS Routes after they have been generated
		for _, httpRoute := range httpListener.Routes {
			hostname := httpRoute.Hostname
			if hostname == "" {
				hostname = "*"
			}

			// 1:1 between IR HTTPRoute Hostname and xDS VirtualHost.
			vHost := vHosts[hostname]
			if vHost == nil {
				// Allocate virtual host for this httpRoute.
				vHost = &routev3.VirtualHost{
					Name:    fmt.Sprintf("%s-%s", httpListener.Name, hostname),
					Domains: []string{hostname},
				}
				vHosts[hostname] = vHost
				vHostsList = append(vHostsList, vHost)
			}

			// 1:1 between IR HTTPRoute and xDS config.route.v3.Route
			xdsRoute := buildXdsRoute(httpRoute, xdsListener)

			// Check if an extension want to modify the route we just generated
			// If no extension exists (or it doesn't subscribe to this hook) then this is a quick no-op.
			if err := processExtensionPostRouteHook(xdsRoute, vHost, httpRoute, t.ExtensionManager); err != nil {
				return err
			}

			vHost.Routes = append(vHost.Routes, xdsRoute)

			// Skip trying to build an IR cluster if the httpRoute only has invalid backends
			if len(httpRoute.Destinations) == 0 && httpRoute.BackendWeights.Invalid > 0 {
				continue
			}
			if err := addXdsCluster(tCtx, addXdsClusterArgs{
				name:         httpRoute.Name,
				destinations: httpRoute.Destinations,
				tSocket:      nil,
				protocol:     protocol,
				endpoint:     Static,
			}); err != nil {
				return err
			}

			// If the httpRoute has a list of mirrors create clusters for them unless they already have one
			for i, mirror := range httpRoute.Mirrors {
				mirrorClusterName := fmt.Sprintf("%s-mirror-%d", httpRoute.Name, i)
				if cluster := findXdsCluster(tCtx, mirrorClusterName); cluster == nil {
					if err := addXdsCluster(tCtx, addXdsClusterArgs{
						name:         mirrorClusterName,
						destinations: []*ir.RouteDestination{mirror},
						tSocket:      nil,
						protocol:     protocol,
						endpoint:     Static,
					}); err != nil {
						return err
					}
				}

			}
		}

		for _, vHost := range vHostsList {
			// Check if an extension want to modify the Virtual Host we just generated
			// If no extension exists (or it doesn't subscribe to this hook) then this is a quick no-op.
			if err := processExtensionPostVHostHook(vHost, t.ExtensionManager); err != nil {
				return err
			}
		}
		xdsRouteCfg.VirtualHosts = append(xdsRouteCfg.VirtualHosts, vHostsList...)

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
		// If no extension exists (or it doesn't subscribe to this hook) then this is a quick no-op
		if err := processExtensionPostListenerHook(tCtx, xdsListener, t.ExtensionManager); err != nil {
			return err
		}
	}

	return nil
}

func processTCPListenerXdsTranslation(tCtx *types.ResourceVersionTable, tcpListeners []*ir.TCPListener, accesslog *ir.AccessLog) error {
	for _, tcpListener := range tcpListeners {
		// 1:1 between IR TCPListener and xDS Cluster
		if err := addXdsCluster(tCtx, addXdsClusterArgs{
			name:         tcpListener.Name,
			destinations: tcpListener.Destinations,
			tSocket:      nil,
			protocol:     DefaultProtocol,
			endpoint:     Static,
		}); err != nil {
			return err
		}

		if tcpListener.TLS != nil && tcpListener.TLS.Terminate != nil {
			for _, s := range tcpListener.TLS.Terminate {
				secret := buildXdsDownstreamTLSSecret(s)
				if err := tCtx.AddXdsResource(resourcev3.SecretType, secret); err != nil {
					return err
				}
			}
		}
		// Search for an existing listener, if it does not exist, create one.
		xdsListener := findXdsListenerByHostPort(tCtx, tcpListener.Address, tcpListener.Port, corev3.SocketAddress_TCP)
		if xdsListener == nil {
			xdsListener = buildXdsTCPListener(tcpListener.Name, tcpListener.Address, tcpListener.Port, accesslog)
			if err := tCtx.AddXdsResource(resourcev3.ListenerType, xdsListener); err != nil {
				return err
			}
		}

		if err := addXdsTCPFilterChain(xdsListener, tcpListener, tcpListener.Name, accesslog); err != nil {
			return err
		}
	}
	return nil
}

func processUDPListenerXdsTranslation(tCtx *types.ResourceVersionTable, udpListeners []*ir.UDPListener, accesslog *ir.AccessLog) error {
	for _, udpListener := range udpListeners {
		// 1:1 between IR UDPListener and xDS Cluster
		if err := addXdsCluster(tCtx, addXdsClusterArgs{
			name:         udpListener.Name,
			destinations: udpListener.Destinations,
			tSocket:      nil,
			protocol:     DefaultProtocol,
			endpoint:     Static,
		}); err != nil {
			return err
		}

		// There won't be multiple UDP listeners on the same port since it's already been checked at the gateway api
		// translator
		xdsListener, err := buildXdsUDPListener(udpListener.Name, udpListener, accesslog)
		if err != nil {
			return multierror.Append(err, errors.New("error building xds cluster"))
		}
		if err := tCtx.AddXdsResource(resourcev3.ListenerType, xdsListener); err != nil {
			return err
		}
	}
	return nil

}

// findXdsListenerByHostPort finds a xds listener with the same address, port and protocol, and returns nil if there is no match.
func findXdsListenerByHostPort(tCtx *types.ResourceVersionTable, address string, port uint32,
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

// findXdsListener finds a xds listener with the same name and returns nil if there is no match.
func findXdsListener(tCtx *types.ResourceVersionTable, name string) *listenerv3.Listener {
	if tCtx == nil || tCtx.XdsResources == nil || tCtx.XdsResources[resourcev3.ListenerType] == nil {
		return nil
	}

	for _, r := range tCtx.XdsResources[resourcev3.ListenerType] {
		listener := r.(*listenerv3.Listener)
		if listener.Name == name {
			return listener
		}
	}

	return nil
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

// findXdsEndpoint finds a xds endpoint with the same cluster name, and returns nil if there is no match.
func findXdsEndpoint(tCtx *types.ResourceVersionTable, name string) *endpointv3.ClusterLoadAssignment {
	if tCtx == nil || tCtx.XdsResources == nil || tCtx.XdsResources[resourcev3.EndpointType] == nil {
		return nil
	}

	for _, r := range tCtx.XdsResources[resourcev3.EndpointType] {
		endpoint := r.(*endpointv3.ClusterLoadAssignment)
		if endpoint.ClusterName == name {
			return endpoint
		}
	}

	return nil
}

func addXdsCluster(tCtx *types.ResourceVersionTable, args addXdsClusterArgs) error {
	xdsCluster := buildXdsCluster(args.name, args.tSocket, args.protocol, args.endpoint)
	xdsEndpoints := buildXdsClusterLoadAssignment(args.name, args.destinations)
	// Use EDS for static endpoints
	if args.endpoint == Static {
		if err := tCtx.AddXdsResource(resourcev3.EndpointType, xdsEndpoints); err != nil {
			return err
		}
	} else {
		xdsCluster.LoadAssignment = xdsEndpoints
	}
	if err := tCtx.AddXdsResource(resourcev3.ClusterType, xdsCluster); err != nil {
		return err
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
