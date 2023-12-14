// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"strings"
	"time"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/tetratelabs/multierror"

	extensionTypes "github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

var (
	ErrXdsClusterExists = errors.New("xds cluster exists")
	ErrXdsSecretExists  = errors.New("xds secret exists")
)

const AuthorityHeaderKey = ":authority"

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

	// xDS translation is done in a best-effort manner, so we collect all errors
	// and return them at the end.
	//
	// Reasoning: The validation in the CRD validation and API Gateway API
	// translator should already catch most errors, there are just few rare cases
	// where xDS translation can fail, for example, failed to call an extension
	// hook or failed to patch an EnvoyPatchPolicy. In those cases, we don't want
	// to fail the entire xDS translation to panic users, but instead, we want
	// to collect all errors and reflect them in the status of the CRDs.
	var errs error
	if err := t.processHTTPListenerXdsTranslation(
		tCtx, ir.HTTP, ir.AccessLog, ir.Tracing, ir.Metrics); err != nil {
		errs = multierror.Append(errs, err)
	}

	if err := processTCPListenerXdsTranslation(tCtx, ir.TCP, ir.AccessLog); err != nil {
		errs = multierror.Append(errs, err)
	}

	if err := processUDPListenerXdsTranslation(tCtx, ir.UDP, ir.AccessLog); err != nil {
		errs = multierror.Append(errs, err)
	}

	if err := processJSONPatches(tCtx, ir.EnvoyPatchPolicies); err != nil {
		errs = multierror.Append(errs, err)
	}

	if err := processClusterForAccessLog(tCtx, ir.AccessLog); err != nil {
		errs = multierror.Append(errs, err)
	}
	if err := processClusterForTracing(tCtx, ir.Tracing); err != nil {
		errs = multierror.Append(errs, err)
	}

	// Check if an extension want to inject any clusters/secrets
	// If no extension exists (or it doesn't subscribe to this hook) then this is a quick no-op
	if err := processExtensionPostTranslationHook(tCtx, t.ExtensionManager); err != nil {
		errs = multierror.Append(errs, err)
	}

	return tCtx, errs
}

func (t *Translator) processHTTPListenerXdsTranslation(
	tCtx *types.ResourceVersionTable,
	httpListeners []*ir.HTTPListener,
	accessLog *ir.AccessLog,
	tracing *ir.Tracing,
	metrics *ir.Metrics,
) error {
	// The XDS translation is done in a best-effort manner, so we collect all
	// errors and return them at the end.
	var errs error
	for _, httpListener := range httpListeners {
		addFilterChain := true
		var xdsRouteCfg *routev3.RouteConfiguration

		// Search for an existing listener, if it does not exist, create one.
		xdsListener := findXdsListenerByHostPort(
			tCtx, httpListener.Address, httpListener.Port, corev3.SocketAddress_TCP)
		if xdsListener == nil {
			xdsListener = buildXdsTCPListener(
				httpListener.Name, httpListener.Address, httpListener.Port,
				httpListener.TCPKeepalive, accessLog)
			if err := tCtx.AddXdsResource(resourcev3.ListenerType, xdsListener); err != nil {
				// skip this listener if failed to add xds listener to the
				// resource version table. Normally, this should not happen.
				errs = multierror.Append(errs, err)
				continue
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
					// skip this listener if failed to find xds route config
					errs = multierror.Append(errs, errors.New("unable to find xds route config"))
					continue
				}
			}
		}

		if addFilterChain {
			if err := t.addXdsHTTPFilterChain(xdsListener, httpListener, accessLog, tracing); err != nil {
				errs = multierror.Append(errs, err)
			}
		}

		// Create a route config if we have not found one yet
		if xdsRouteCfg == nil {
			xdsRouteCfg = &routev3.RouteConfiguration{
				IgnorePortInHostMatching: true,
				Name:                     httpListener.Name,
			}

			if err := tCtx.AddXdsResource(resourcev3.RouteType, xdsRouteCfg); err != nil {
				errs = multierror.Append(errs, err)
			}
		}

		// 1:1 between IR TLSListenerConfig and xDS Secret
		if httpListener.TLS != nil {
			for t := range httpListener.TLS {
				secret := buildXdsDownstreamTLSSecret(httpListener.TLS[t])
				if err := tCtx.AddXdsResource(resourcev3.SecretType, secret); err != nil {
					errs = multierror.Append(errs, err)
				}
			}
		}

		// store virtual hosts by domain
		vHosts := map[string]*routev3.VirtualHost{}
		// keep track of order by using a list as well as the map
		var vHostsList []*routev3.VirtualHost

		// Check if an extension is loaded that wants to modify xDS Routes after they have been generated
		for _, httpRoute := range httpListener.Routes {
			// 1:1 between IR HTTPRoute Hostname and xDS VirtualHost.
			vHost := vHosts[httpRoute.Hostname]
			if vHost == nil {
				// Remove dots from the hostname before appending it to the virtualHost name
				// since dots are special chars used in stats tag extraction in Envoy
				underscoredHostname := strings.ReplaceAll(httpRoute.Hostname, ".", "_")
				// Allocate virtual host for this httpRoute.
				vHost = &routev3.VirtualHost{
					Name:    fmt.Sprintf("%s/%s", httpListener.Name, underscoredHostname),
					Domains: []string{httpRoute.Hostname},
				}
				if metrics != nil && metrics.EnableVirtualHostStats {
					vHost.VirtualClusters = []*routev3.VirtualCluster{
						{
							Name: underscoredHostname,
							Headers: []*routev3.HeaderMatcher{
								{
									Name: AuthorityHeaderKey,
									HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
										StringMatch: &matcherv3.StringMatcher{
											MatchPattern: &matcherv3.StringMatcher_Prefix{
												Prefix: httpRoute.Hostname,
											},
										},
									},
								},
							},
						},
					}
				}
				vHosts[httpRoute.Hostname] = vHost
				vHostsList = append(vHostsList, vHost)
			}

			// 1:1 between IR HTTPRoute and xDS config.route.v3.Route
			xdsRoute, err := buildXdsRoute(httpRoute)
			if err != nil {
				// skip this route if failed to build xds route
				errs = multierror.Append(errs, err)
				continue
			}

			// Check if an extension want to modify the route we just generated
			// If no extension exists (or it doesn't subscribe to this hook) then this is a quick no-op.
			if err = processExtensionPostRouteHook(xdsRoute, vHost, httpRoute, t.ExtensionManager); err != nil {
				if err != nil {
					errs = multierror.Append(errs, err)
				}
			}

			vHost.Routes = append(vHost.Routes, xdsRoute)

			if httpRoute.Destination != nil {
				if err = processXdsCluster(tCtx, httpRoute); err != nil {
					errs = multierror.Append(errs, err)
				}
			}

			if httpRoute.Mirrors != nil {
				for _, mirrorDest := range httpRoute.Mirrors {
					if err := addXdsCluster(tCtx, &xdsClusterArgs{
						name:         mirrorDest.Name,
						settings:     mirrorDest.Settings,
						tSocket:      nil,
						endpointType: EndpointTypeStatic,
					}); err != nil && !errors.Is(err, ErrXdsClusterExists) {
						errs = multierror.Append(errs, err)
					}
				}
			}
		}

		for _, vHost := range vHostsList {
			// Check if an extension want to modify the Virtual Host we just generated
			// If no extension exists (or it doesn't subscribe to this hook) then this is a quick no-op.
			if err := processExtensionPostVHostHook(vHost, t.ExtensionManager); err != nil {
				errs = multierror.Append(errs, err)
			}
		}
		xdsRouteCfg.VirtualHosts = append(xdsRouteCfg.VirtualHosts, vHostsList...)

		// Add per-route filter configs to the route config.
		if err := patchRouteCfgWithPerRouteConfig(xdsRouteCfg, httpListener); err != nil {
			errs = multierror.Append(errs, err)
		}

		// Add all the other needed resources referenced by this filter to the
		// resource version table.
		if err := patchResources(tCtx, httpListener.Routes); err != nil {
			return err
		}

		// RateLimit filter is handled separately because it relies on the global
		// rate limit server configuration.
		// Check if a ratelimit cluster exists, if not, add it, if it's needed.
		if err := t.createRateLimitServiceCluster(tCtx, httpListener); err != nil {
			errs = multierror.Append(errs, err)
		}

		// Check if an extension want to modify the listener that was just configured/created
		// If no extension exists (or it doesn't subscribe to this hook) then this is a quick no-op
		if err := processExtensionPostListenerHook(tCtx, xdsListener, t.ExtensionManager); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs
}

func processTCPListenerXdsTranslation(tCtx *types.ResourceVersionTable, tcpListeners []*ir.TCPListener, accesslog *ir.AccessLog) error {
	// The XDS translation is done in a best-effort manner, so we collect all
	// errors and return them at the end.
	var errs error
	for _, tcpListener := range tcpListeners {
		// Search for an existing listener, if it does not exist, create one.
		xdsListener := findXdsListenerByHostPort(tCtx, tcpListener.Address, tcpListener.Port, corev3.SocketAddress_TCP)
		if xdsListener == nil {
			xdsListener = buildXdsTCPListener(tcpListener.Name, tcpListener.Address, tcpListener.Port, tcpListener.TCPKeepalive, accesslog)
			if err := tCtx.AddXdsResource(resourcev3.ListenerType, xdsListener); err != nil {
				// skip this listener if failed to add xds listener to the
				errs = multierror.Append(errs, err)
				continue
			}
		}

		if err := addXdsTCPFilterChain(xdsListener, tcpListener, tcpListener.Destination.Name, accesslog); err != nil {
			errs = multierror.Append(errs, err)
		}

		// 1:1 between IR TCPListener and xDS Cluster
		if err := addXdsCluster(tCtx, &xdsClusterArgs{
			name:         tcpListener.Destination.Name,
			settings:     tcpListener.Destination.Settings,
			tSocket:      nil,
			endpointType: EndpointTypeStatic,
		}); err != nil && !errors.Is(err, ErrXdsClusterExists) {
			errs = multierror.Append(errs, err)
		}

		if tcpListener.TLS != nil && tcpListener.TLS.Terminate != nil {
			for _, s := range tcpListener.TLS.Terminate {
				secret := buildXdsDownstreamTLSSecret(s)
				if err := tCtx.AddXdsResource(resourcev3.SecretType, secret); err != nil {
					errs = multierror.Append(errs, err)
				}
			}
		}
	}
	return errs
}

func processUDPListenerXdsTranslation(tCtx *types.ResourceVersionTable, udpListeners []*ir.UDPListener, accesslog *ir.AccessLog) error {
	// The XDS translation is done in a best-effort manner, so we collect all
	// errors and return them at the end.
	var errs error

	for _, udpListener := range udpListeners {
		// There won't be multiple UDP listeners on the same port since it's already been checked at the gateway api
		// translator
		xdsListener, err := buildXdsUDPListener(udpListener.Destination.Name, udpListener, accesslog)
		if err != nil {
			// skip this listener if failed to build xds listener
			errs = multierror.Append(errs, err)
			continue
		}
		if err := tCtx.AddXdsResource(resourcev3.ListenerType, xdsListener); err != nil {
			// skip this listener if failed to add xds listener to the resource version table
			errs = multierror.Append(errs, err)
			continue
		}

		// 1:1 between IR UDPListener and xDS Cluster
		if err := addXdsCluster(tCtx, &xdsClusterArgs{
			name:         udpListener.Destination.Name,
			settings:     udpListener.Destination.Settings,
			tSocket:      nil,
			endpointType: EndpointTypeStatic,
		}); err != nil && !errors.Is(err, ErrXdsClusterExists) {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
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

// findXdsRouteConfig finds a xds route with the name and returns nil if there is no match.
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

// processXdsCluster processes a xds cluster by its endpoint address type.
func processXdsCluster(tCtx *types.ResourceVersionTable, httpRoute *ir.HTTPRoute) error {
	// Get endpoint address type for xds cluster by returning the first DestinationSetting's AddressType,
	// since there's no Mixed AddressType among all the DestinationSettings.
	addrTypeState := httpRoute.Destination.Settings[0].AddressType

	var endpointType EndpointType
	if addrTypeState != nil && *addrTypeState == ir.FQDN {
		endpointType = EndpointTypeDNS
	} else {
		endpointType = EndpointTypeStatic
	}

	if err := addXdsCluster(tCtx, &xdsClusterArgs{
		name:          httpRoute.Destination.Name,
		settings:      httpRoute.Destination.Settings,
		tSocket:       nil,
		endpointType:  endpointType,
		loadBalancer:  httpRoute.LoadBalancer,
		proxyProtocol: httpRoute.ProxyProtocol,
	}); err != nil && !errors.Is(err, ErrXdsClusterExists) {
		return err
	}

	return nil
}

// findXdsSecret finds a xds secret with the same name, and returns nil if there is no match.
func findXdsSecret(tCtx *types.ResourceVersionTable, name string) *tlsv3.Secret {
	if tCtx == nil || tCtx.XdsResources == nil || tCtx.XdsResources[resourcev3.SecretType] == nil {
		return nil
	}

	for _, r := range tCtx.XdsResources[resourcev3.SecretType] {
		secret := r.(*tlsv3.Secret)
		if secret.Name == name {
			return secret
		}
	}

	return nil
}

func addXdsSecret(tCtx *types.ResourceVersionTable, secret *tlsv3.Secret) error {
	// Return early if cluster with the same name exists
	if c := findXdsSecret(tCtx, secret.Name); c != nil {
		return ErrXdsSecretExists
	}

	if err := tCtx.AddXdsResource(resourcev3.SecretType, secret); err != nil {
		return err
	}
	return nil
}

func addXdsCluster(tCtx *types.ResourceVersionTable, args *xdsClusterArgs) error {
	// Return early if cluster with the same name exists
	if c := findXdsCluster(tCtx, args.name); c != nil {
		return ErrXdsClusterExists
	}

	xdsCluster := buildXdsCluster(args)
	xdsEndpoints := buildXdsClusterLoadAssignment(args.name, args.settings)
	// Use EDS for static endpoints
	if args.endpointType == EndpointTypeStatic {
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
