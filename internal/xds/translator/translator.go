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
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	extensionTypes "github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/protocov"
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

	// FilterOrder holds the custom order of the HTTP filters
	FilterOrder []v1alpha1.FilterPosition
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
		errs = errors.Join(errs, err)
	}

	if err := processTCPListenerXdsTranslation(tCtx, ir.TCP, ir.AccessLog, ir.Metrics); err != nil {
		errs = errors.Join(errs, err)
	}

	if err := processUDPListenerXdsTranslation(tCtx, ir.UDP, ir.AccessLog, ir.Metrics); err != nil {
		errs = errors.Join(errs, err)
	}

	if err := processJSONPatches(tCtx, ir.EnvoyPatchPolicies); err != nil {
		errs = errors.Join(errs, err)
	}

	if err := processClusterForAccessLog(tCtx, ir.AccessLog, ir.Metrics); err != nil {
		errs = errors.Join(errs, err)
	}

	if err := processClusterForTracing(tCtx, ir.Tracing, ir.Metrics); err != nil {
		errs = errors.Join(errs, err)
	}

	// Check if an extension want to inject any clusters/secrets
	// If no extension exists (or it doesn't subscribe to this hook) then this is a quick no-op
	if err := processExtensionPostTranslationHook(tCtx, t.ExtensionManager); err != nil {
		errs = errors.Join(errs, err)
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
		var (
			http3Enabled                       = httpListener.HTTP3 != nil // Whether HTTP3 is enabled
			tcpXDSListener                     *listenerv3.Listener        // TCP Listener for HTTP1/HTTP2 traffic
			quicXDSListener                    *listenerv3.Listener        // UDP(QUIC) Listener for HTTP3 traffic
			xdsListenerOnSameAddressPortExists bool                        // Whether a listener already exists on the same address + port combination
			tlsEnabled                         bool                        // Whether TLS is enabled for the listener
			xdsRouteCfg                        *routev3.RouteConfiguration // The route config is used by both the TCP and QUIC listeners
			addHCM                             bool                        // Whether to add an HCM(HTTP Connection Manager filter) to the listener's TCP filter chain
			err                                error
		)

		// Search for an existing TCP listener on the same address + port combination.
		tcpXDSListener = findXdsListenerByHostPort(tCtx, httpListener.Address, httpListener.Port, corev3.SocketAddress_TCP)
		xdsListenerOnSameAddressPortExists = tcpXDSListener != nil
		tlsEnabled = httpListener.TLS != nil

		switch {
		// If no existing listener exists, create a new one.
		case !xdsListenerOnSameAddressPortExists:
			// Create a new UDP(QUIC) listener for HTTP3 traffic if HTTP3 is enabled
			if http3Enabled {
				quicXDSListener = buildXdsQuicListener(httpListener.Name, httpListener.Address, httpListener.Port, accessLog)
				if err = tCtx.AddXdsResource(resourcev3.ListenerType, quicXDSListener); err != nil {
					errs = errors.Join(errs, err)
					continue
				}
			}

			// Create a new TCP listener for HTTP1/HTTP2 traffic.
			tcpXDSListener = buildXdsTCPListener(httpListener.Name, httpListener.Address, httpListener.Port, httpListener.TCPKeepalive, httpListener.Connection, accessLog)
			if err = tCtx.AddXdsResource(resourcev3.ListenerType, tcpXDSListener); err != nil {
				errs = errors.Join(errs, err)
				continue
			}

			// We need to add an HCM to the newly created listener.
			addHCM = true
		case xdsListenerOnSameAddressPortExists && !tlsEnabled:
			// If a xds listener exists, and Gateway HTTP Listener does not enable TLS,
			// we use the listener's default TCP filter chain because we can not
			// differentiate the HTTP traffic at the TCP filter chain level using SNI.
			//
			// A HCM(HTTP Connection Manager filter) is added to the listener's
			// default filter chain if it has not yet been added.
			//
			// The HCM is configured with a RouteConfiguration, which is used to
			// route HTTP traffic to the correct virtual host for all the domains
			// specified in the Gateway HTTP Listener's routes.
			var (
				routeName                  string
				hasHCMInDefaultFilterChain bool
			)

			// Find the route config associated with this listener that
			// maps to the default filter chain for http traffic
			// Routes for this listener will be added to this route config
			routeName = findXdsHTTPRouteConfigName(tcpXDSListener)
			hasHCMInDefaultFilterChain = routeName != ""
			addHCM = !hasHCMInDefaultFilterChain

			if routeName != "" {
				xdsRouteCfg = findXdsRouteConfig(tCtx, routeName)
				if xdsRouteCfg == nil {
					// skip this listener if failed to find xds route config
					errs = errors.Join(errs, errors.New("unable to find xds route config"))
					continue
				}
			}
		case xdsListenerOnSameAddressPortExists && tlsEnabled:
			// If an existing xds listener exists, and Gateway HTTP Listener enables
			// TLS, we need to create an HCM.
			//
			// In this case, a new filter chain is created and added to the listener,
			// and the HCM is added to the new filter chain.
			// The newly created filter chain is configured with a filter chain
			// match to match the server names(SNI) based on the listener's hostnames.
			addHCM = true
		}

		if addHCM {
			if err = t.addHCMToXDSListener(tcpXDSListener, httpListener, accessLog, tracing, false, httpListener.Connection); err != nil {
				errs = errors.Join(errs, err)
				continue
			}
			if http3Enabled {
				if err = t.addHCMToXDSListener(quicXDSListener, httpListener, accessLog, tracing, true, httpListener.Connection); err != nil {
					errs = errors.Join(errs, err)
					continue
				}
			}
		} else {
			// When the DefaultFilterChain is shared by multiple Gateway HTTP
			// Listeners, we need to add the HTTP filters associated with the
			// HTTPListener to the HCM if they have not yet been added.
			if err = t.addHTTPFiltersToHCM(tcpXDSListener.DefaultFilterChain, httpListener); err != nil {
				errs = errors.Join(errs, err)
				continue
			}
			if http3Enabled {
				if err = t.addHTTPFiltersToHCM(quicXDSListener.DefaultFilterChain, httpListener); err != nil {
					errs = errors.Join(errs, err)
					continue
				}
			}
		}

		// Add the secrets referenced by the listener's TLS configuration to the
		// resource version table.
		// 1:1 between IR TLSListenerConfig and xDS Secret
		if httpListener.TLS != nil {
			for c := range httpListener.TLS.Certificates {
				secret := buildXdsTLSCertSecret(httpListener.TLS.Certificates[c])
				if err = tCtx.AddXdsResource(resourcev3.SecretType, secret); err != nil {
					errs = errors.Join(errs, err)
				}
			}

			if httpListener.TLS.CACertificate != nil {
				caSecret := buildXdsTLSCaCertSecret(httpListener.TLS.CACertificate)
				if err = tCtx.AddXdsResource(resourcev3.SecretType, caSecret); err != nil {
					errs = errors.Join(errs, err)
				}
			}
		}

		// Create a route config if we have not found one yet
		if xdsRouteCfg == nil {
			xdsRouteCfg = &routev3.RouteConfiguration{
				IgnorePortInHostMatching: true,
				Name:                     httpListener.Name,
			}

			if err = tCtx.AddXdsResource(resourcev3.RouteType, xdsRouteCfg); err != nil {
				errs = errors.Join(errs, err)
			}
		}

		// Generate xDS virtual hosts and routes for the given HTTPListener,
		// and add them to the xDS route config.
		if err = t.addRouteToRouteConfig(tCtx, xdsRouteCfg, httpListener, metrics, http3Enabled); err != nil {
			errs = errors.Join(errs, err)
		}

		// Add all the other needed resources referenced by this filter to the
		// resource version table.
		if err = patchResources(tCtx, httpListener.Routes); err != nil {
			errs = errors.Join(errs, err)
		}

		// RateLimit filter is handled separately because it relies on the global
		// rate limit server configuration.
		// Check if a ratelimit cluster exists, if not, add it, if it's needed.
		if err = t.createRateLimitServiceCluster(tCtx, httpListener, metrics); err != nil {
			errs = errors.Join(errs, err)
		}

		// Check if an extension want to modify the listener that was just configured/created
		// If no extension exists (or it doesn't subscribe to this hook) then this is a quick no-op
		// TODO zhaohuabing should we also process the quicXDSListener?
		if err = processExtensionPostListenerHook(tCtx, tcpXDSListener, t.ExtensionManager); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

// addRouteToRouteConfig generates xDS virtual hosts and routes for the given HTTPListener,
// and adds them to the provided xDS route config.
func (t *Translator) addRouteToRouteConfig(
	tCtx *types.ResourceVersionTable,
	xdsRouteCfg *routev3.RouteConfiguration,
	httpListener *ir.HTTPListener,
	metrics *ir.Metrics,
	http3Enabled bool,
) error {
	var (
		vHosts    = map[string]*routev3.VirtualHost{} // store virtual hosts by domain
		vHostList []*routev3.VirtualHost              // keep track of order by using a list as well as the map
		errs      error                               // the accumulated errors
		err       error
	)

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
			vHostList = append(vHostList, vHost)
		}

		var xdsRoute *routev3.Route
		// 1:1 between IR HTTPRoute and xDS config.route.v3.Route
		xdsRoute, err = buildXdsRoute(httpRoute)
		if err != nil {
			// skip this route if failed to build xds route
			errs = errors.Join(errs, err)
			continue
		}

		// Check if an extension want to modify the route we just generated
		// If no extension exists (or it doesn't subscribe to this hook) then this is a quick no-op.
		if err = processExtensionPostRouteHook(xdsRoute, vHost, httpRoute, t.ExtensionManager); err != nil {
			errs = errors.Join(errs, err)
		}

		if http3Enabled {
			http3AltSvcHeader := buildHTTP3AltSvcHeader(int(httpListener.HTTP3.QUICPort))
			if xdsRoute.ResponseHeadersToAdd == nil {
				xdsRoute.ResponseHeadersToAdd = make([]*corev3.HeaderValueOption, 0)
			}
			xdsRoute.ResponseHeadersToAdd = append(xdsRoute.ResponseHeadersToAdd, http3AltSvcHeader)
		}
		vHost.Routes = append(vHost.Routes, xdsRoute)

		if httpRoute.Destination != nil {
			if err = processXdsCluster(
				tCtx,
				httpRoute,
				httpListener.HTTP1,
				metrics,
			); err != nil {
				errs = errors.Join(errs, err)
			}
		}

		if httpRoute.Mirrors != nil {
			for _, mirrorDest := range httpRoute.Mirrors {
				if err = addXdsCluster(tCtx, &xdsClusterArgs{
					name:         mirrorDest.Name,
					settings:     mirrorDest.Settings,
					tSocket:      nil,
					endpointType: EndpointTypeStatic,
					metrics:      metrics,
				}); err != nil && !errors.Is(err, ErrXdsClusterExists) {
					errs = errors.Join(errs, err)
				}
			}
		}
	}

	for _, vHost := range vHostList {
		// Check if an extension want to modify the Virtual Host we just generated
		// If no extension exists (or it doesn't subscribe to this hook) then this is a quick no-op.
		if err = processExtensionPostVHostHook(vHost, t.ExtensionManager); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	xdsRouteCfg.VirtualHosts = append(xdsRouteCfg.VirtualHosts, vHostList...)

	return errs
}

func (t *Translator) addHTTPFiltersToHCM(filterChain *listenerv3.FilterChain, httpListener *ir.HTTPListener) error {
	var (
		hcm *hcmv3.HttpConnectionManager
		err error
	)

	if hcm, err = findHCMinFilterChain(filterChain); err != nil {
		return err // should not happen
	}

	// Add http filters to the HCM if they have not yet been added.
	if err = t.patchHCMWithFilters(hcm, httpListener); err != nil {
		return err
	}

	for i, filter := range filterChain.Filters {
		if filter.Name == wellknown.HTTPConnectionManager {
			var mgrAny *anypb.Any
			if mgrAny, err = protocov.ToAnyWithError(hcm); err != nil {
				return err
			}

			filterChain.Filters[i] = &listenerv3.Filter{
				Name: wellknown.HTTPConnectionManager,
				ConfigType: &listenerv3.Filter_TypedConfig{
					TypedConfig: mgrAny,
				},
			}
		}
	}
	return nil
}

func findHCMinFilterChain(filterChain *listenerv3.FilterChain) (*hcmv3.HttpConnectionManager, error) {
	for _, filter := range filterChain.Filters {
		if filter.Name == wellknown.HTTPConnectionManager {
			hcm := &hcmv3.HttpConnectionManager{}
			if err := anypb.UnmarshalTo(filter.GetTypedConfig(), hcm, proto.UnmarshalOptions{}); err != nil {
				return nil, err
			}
			return hcm, nil
		}
	}
	return nil, errors.New("http connection manager not found")
}

func buildHTTP3AltSvcHeader(port int) *corev3.HeaderValueOption {
	return &corev3.HeaderValueOption{
		Append: &wrapperspb.BoolValue{Value: true},
		Header: &corev3.HeaderValue{
			Key:   "alt-svc",
			Value: strings.Join([]string{fmt.Sprintf(`%s=":%d"; ma=86400`, "h3", port)}, ", "),
		},
	}
}

func processTCPListenerXdsTranslation(tCtx *types.ResourceVersionTable, tcpListeners []*ir.TCPListener, accesslog *ir.AccessLog, metrics *ir.Metrics) error {
	// The XDS translation is done in a best-effort manner, so we collect all
	// errors and return them at the end.
	var errs error
	for _, tcpListener := range tcpListeners {
		// Search for an existing listener, if it does not exist, create one.
		xdsListener := findXdsListenerByHostPort(tCtx, tcpListener.Address, tcpListener.Port, corev3.SocketAddress_TCP)
		if xdsListener == nil {
			xdsListener = buildXdsTCPListener(tcpListener.Name, tcpListener.Address, tcpListener.Port, tcpListener.TCPKeepalive, tcpListener.Connection, accesslog)
			if err := tCtx.AddXdsResource(resourcev3.ListenerType, xdsListener); err != nil {
				// skip this listener if failed to add xds listener to the
				errs = errors.Join(errs, err)
				continue
			}
		}

		for _, route := range tcpListener.Routes {
			if err := addXdsCluster(tCtx, &xdsClusterArgs{
				name:           route.Destination.Name,
				settings:       route.Destination.Settings,
				loadBalancer:   route.LoadBalancer,
				proxyProtocol:  route.ProxyProtocol,
				circuitBreaker: route.CircuitBreaker,
				tcpkeepalive:   route.TCPKeepalive,
				healthCheck:    route.HealthCheck,
				timeout:        route.Timeout,
				endpointType:   buildEndpointType(route.Destination.Settings),
				metrics:        metrics,
			}); err != nil && !errors.Is(err, ErrXdsClusterExists) {
				errs = errors.Join(errs, err)
			}
			if route.TLS != nil && route.TLS.Terminate != nil {
				for _, s := range route.TLS.Terminate.Certificates {
					secret := buildXdsTLSCertSecret(s)
					if err := tCtx.AddXdsResource(resourcev3.SecretType, secret); err != nil {
						errs = errors.Join(errs, err)
					}
				}
				if route.TLS.Terminate.CACertificate != nil {
					caSecret := buildXdsTLSCaCertSecret(route.TLS.Terminate.CACertificate)
					if err := tCtx.AddXdsResource(resourcev3.SecretType, caSecret); err != nil {
						errs = errors.Join(errs, err)
					}
				}
			}
			if err := addXdsTCPFilterChain(xdsListener, route, route.Destination.Name, accesslog, tcpListener.Connection); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}
	return errs
}

func processUDPListenerXdsTranslation(tCtx *types.ResourceVersionTable, udpListeners []*ir.UDPListener, accesslog *ir.AccessLog, metrics *ir.Metrics) error {
	// The XDS translation is done in a best-effort manner, so we collect all
	// errors and return them at the end.
	var errs error

	for _, udpListener := range udpListeners {
		// There won't be multiple UDP listeners on the same port since it's already been checked at the gateway api
		// translator
		if udpListener.Route != nil {
			route := udpListener.Route

			xdsListener, err := buildXdsUDPListener(route.Destination.Name, udpListener, accesslog)
			if err != nil {
				// skip this listener if failed to build xds listener
				errs = errors.Join(errs, err)
				continue
			}
			if err := tCtx.AddXdsResource(resourcev3.ListenerType, xdsListener); err != nil {
				// skip this listener if failed to add xds listener to the resource version table
				errs = errors.Join(errs, err)
				continue
			}

			// 1:1 between IR UDPRoute and xDS Cluster
			if err := addXdsCluster(tCtx, &xdsClusterArgs{
				name:         route.Destination.Name,
				settings:     route.Destination.Settings,
				loadBalancer: route.LoadBalancer,
				timeout:      route.Timeout,
				tSocket:      nil,
				endpointType: buildEndpointType(route.Destination.Settings),
				metrics:      metrics,
			}); err != nil && !errors.Is(err, ErrXdsClusterExists) {
				errs = errors.Join(errs, err)
			}
		}
	}
	return errs
}

// findXdsListenerByHostPort finds a xds listener with the same address, port and protocol, and returns nil if there is no match.
func findXdsListenerByHostPort(tCtx *types.ResourceVersionTable, address string, port uint32,
	protocol corev3.SocketAddress_Protocol,
) *listenerv3.Listener {
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
func processXdsCluster(tCtx *types.ResourceVersionTable, httpRoute *ir.HTTPRoute, http1Settings *ir.HTTP1Settings, metrics *ir.Metrics) error {
	clusterArgs := &xdsClusterArgs{
		name:              httpRoute.Destination.Name,
		settings:          httpRoute.Destination.Settings,
		tSocket:           nil,
		endpointType:      buildEndpointType(httpRoute.Destination.Settings),
		http1Settings:     http1Settings,
		metrics:           metrics,
		useClientProtocol: ptr.Deref(httpRoute.UseClientProtocol, false),
	}

	// Populate traffic features.
	bt := httpRoute.Traffic
	if bt != nil {
		clusterArgs.loadBalancer = bt.LoadBalancer
		clusterArgs.proxyProtocol = bt.ProxyProtocol
		clusterArgs.circuitBreaker = bt.CircuitBreaker
		clusterArgs.healthCheck = bt.HealthCheck
		clusterArgs.timeout = bt.Timeout
		clusterArgs.tcpkeepalive = bt.TCPKeepalive
	}

	if err := addXdsCluster(tCtx, clusterArgs); err != nil && !errors.Is(err, ErrXdsClusterExists) {
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
	for _, ds := range args.settings {
		if ds.TLS != nil {
			// Create a secret for the CA certificate only if it's not using the system trust store
			if !ds.TLS.UseSystemTrustStore {
				secret := buildXdsUpstreamTLSCASecret(ds.TLS)
				if err := tCtx.AddXdsResource(resourcev3.SecretType, secret); err != nil {
					return err
				}
			}
		}
	}
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

const (
	DefaultEndpointType EndpointType = iota
	Static
	EDS
)

func buildXdsUpstreamTLSCASecret(tlsConfig *ir.TLSUpstreamConfig) *tlsv3.Secret {
	// Build the tls secret
	// It's just a sanity check, we shouldn't call this function if the system trust store is used
	if tlsConfig.UseSystemTrustStore {
		return nil
	}
	return &tlsv3.Secret{
		Name: tlsConfig.CACertificate.Name,
		Type: &tlsv3.Secret_ValidationContext{
			ValidationContext: &tlsv3.CertificateValidationContext{
				TrustedCa: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineBytes{InlineBytes: tlsConfig.CACertificate.Certificate},
				},
			},
		},
	}
}

func buildXdsUpstreamTLSSocketWthCert(tlsConfig *ir.TLSUpstreamConfig) (*corev3.TransportSocket, error) {
	var tlsCtx *tlsv3.UpstreamTlsContext
	if tlsConfig.UseSystemTrustStore {
		tlsCtx = &tlsv3.UpstreamTlsContext{
			CommonTlsContext: &tlsv3.CommonTlsContext{
				TlsCertificates: nil,
				ValidationContextType: &tlsv3.CommonTlsContext_ValidationContext{
					ValidationContext: &tlsv3.CertificateValidationContext{
						TrustedCa: &corev3.DataSource{
							Specifier: &corev3.DataSource_Filename{
								// This is the default location for the system trust store
								// on Debian derivatives like the envoy-proxy image being used by the infrastructure
								// controller.
								// See https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/security/ssl
								// TODO: allow customizing this value via EnvoyGateway so that if a non-standard
								// envoy image is being used, this can be modified to match
								Filename: "/etc/ssl/certs/ca-certificates.crt",
							},
						},
					},
				},
			},
			Sni: tlsConfig.SNI,
		}
	} else {
		tlsCtx = &tlsv3.UpstreamTlsContext{
			CommonTlsContext: &tlsv3.CommonTlsContext{
				TlsCertificateSdsSecretConfigs: nil,
				ValidationContextType: &tlsv3.CommonTlsContext_ValidationContextSdsSecretConfig{
					ValidationContextSdsSecretConfig: &tlsv3.SdsSecretConfig{
						Name:      tlsConfig.CACertificate.Name,
						SdsConfig: makeConfigSource(),
					},
				},
			},
			Sni: tlsConfig.SNI,
		}
	}

	tlsParams := buildTLSParams(&tlsConfig.TLSConfig)
	if tlsParams != nil {
		tlsCtx.CommonTlsContext.TlsParams = tlsParams
	}

	if len(tlsConfig.ALPNProtocols) > 0 {
		tlsCtx.CommonTlsContext.AlpnProtocols = buildALPNProtocols(tlsConfig.ALPNProtocols)
	}
	tlsCtxAny, err := anypb.New(tlsCtx)
	if err != nil {
		return nil, err
	}

	return &corev3.TransportSocket{
		Name: wellknown.TransportSocketTLS,
		ConfigType: &corev3.TransportSocket_TypedConfig{
			TypedConfig: tlsCtxAny,
		},
	}, nil
}
