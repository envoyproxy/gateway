// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"runtime"
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
	protobuf "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/apimachinery/pkg/util/sets"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	extensionTypes "github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	AuthorityHeaderKey = ":authority"
	// The dummy cluster name for TCP/UDP listeners that have no routes
	emptyClusterName = "EmptyCluster"
)

// The dummy cluster for TCP/UDP listeners that have no routes
var emptyRouteCluster = &clusterv3.Cluster{
	Name:                 emptyClusterName,
	ClusterDiscoveryType: &clusterv3.Cluster_Type{Type: clusterv3.Cluster_STATIC},
}

// Translator translates the xDS IR into xDS resources.
type Translator struct {
	// ControllerNamespace is the namespace of the Gateway API controller
	ControllerNamespace string

	// GlobalRateLimit holds the global rate limit settings
	// required during xds translation.
	GlobalRateLimit *GlobalRateLimitSettings

	// ExtensionManager holds the config for interacting with extensions when generating xDS
	// resources. Only required during xds translation.
	ExtensionManager *extensionTypes.Manager

	// FilterOrder holds the custom order of the HTTP filters
	FilterOrder []egv1a1.FilterPosition
	Logger      logging.Logger
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
func (t *Translator) Translate(xdsIR *ir.Xds) (*types.ResourceVersionTable, error) {
	if xdsIR == nil {
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

	if err := t.processHTTPReadyListenerXdsTranslation(tCtx, xdsIR.ReadyListener); err != nil {
		errs = errors.Join(errs, err)
	}

	if err := t.processHTTPListenerXdsTranslation(
		tCtx, xdsIR.HTTP, xdsIR.AccessLog, xdsIR.Tracing, xdsIR.Metrics); err != nil {
		errs = errors.Join(errs, err)
	}

	if err := t.processTCPListenerXdsTranslation(tCtx, xdsIR.TCP, xdsIR.AccessLog, xdsIR.Metrics); err != nil {
		errs = errors.Join(errs, err)
	}

	if err := processUDPListenerXdsTranslation(tCtx, xdsIR.UDP, xdsIR.AccessLog, xdsIR.Metrics); err != nil {
		errs = errors.Join(errs, err)
	}

	if err := t.notifyExtensionServerAboutListeners(tCtx, xdsIR); err != nil {
		errs = errors.Join(errs, err)
	}

	if err := processJSONPatches(tCtx, xdsIR.EnvoyPatchPolicies); err != nil {
		errs = errors.Join(errs, err)
	}

	if err := processClusterForAccessLog(tCtx, xdsIR.AccessLog, xdsIR.Metrics); err != nil {
		errs = errors.Join(errs, err)
	}

	if err := processClusterForTracing(tCtx, xdsIR.Tracing, xdsIR.Metrics); err != nil {
		errs = errors.Join(errs, err)
	}

	// Patch global resources that are shared across listeners and routes.
	// - the envoy client certificate
	// - the OIDC HMAC secret
	// - the rate limit server cluster
	if err := t.patchGlobalResources(tCtx, xdsIR); err != nil {
		errs = errors.Join(errs, err)
	}

	// Check if an extension want to inject any clusters/secrets
	// If no extension exists (or it doesn't subscribe to this hook) then this is a quick no-op
	if err := processExtensionPostTranslationHook(tCtx, t.ExtensionManager, xdsIR.ExtensionServerPolicies); err != nil {
		// If the extension server returns an error, and the extension server is not configured to fail open,
		// then propagate the error
		if !(*t.ExtensionManager).FailOpen() {
			errs = errors.Join(errs, err)
		} else {
			t.Logger.Error(err, "Extension Manager PostTranslation failure")
		}
	}

	// Validate all the xds resources in the table before returning
	// This is necessary to catch any misconfigurations that might have been missed during translation
	if err := tCtx.ValidateAll(); err != nil {
		errs = errors.Join(errs, err)
	}

	return tCtx, errs
}

func findIRListenersByXDSListener(xdsIR *ir.Xds, listener *listenerv3.Listener) []ir.Listener {
	ret := []ir.Listener{}

	addr := listener.Address.GetSocketAddress()
	if addr == nil {
		return ret
	}
	for _, l := range xdsIR.HTTP {
		if l.GetAddress() == addr.GetAddress() && l.GetPort() == addr.GetPortValue() {
			ret = append(ret, l)
		}
	}
	for _, l := range xdsIR.TCP {
		if l.GetAddress() == addr.GetAddress() && l.GetPort() == addr.GetPortValue() {
			ret = append(ret, l)
		}
	}
	for _, l := range xdsIR.UDP {
		if l.GetAddress() == addr.GetAddress() && l.GetPort() == addr.GetPortValue() {
			ret = append(ret, l)
		}
	}
	return ret
}

// notifyExtensionServerAboutListeners calls the extension server about all the translated listeners.
func (t *Translator) notifyExtensionServerAboutListeners(
	tCtx *types.ResourceVersionTable,
	xdsIR *ir.Xds,
) error {
	// Return quickly if there is no extension manager or the Listener hook is not being used.
	if t.ExtensionManager == nil {
		return nil
	}
	if postHookClient, err := (*t.ExtensionManager).GetPostXDSHookClient(egv1a1.XDSHTTPListener); postHookClient == nil && err == nil {
		return nil
	}

	var errs error
	for _, l := range tCtx.XdsResources[resourcev3.ListenerType] {
		listener := l.(*listenerv3.Listener)
		policies := []*ir.UnstructuredRef{}
		alreadyIncludedPolicies := sets.New[utils.NamespacedNameWithGroupKind]()
		for _, irListener := range findIRListenersByXDSListener(xdsIR, listener) {
			for _, pol := range irListener.GetExtensionRefs() {
				key := utils.GetNamespacedNameWithGroupKind(pol.Object)
				if !alreadyIncludedPolicies.Has(key) {
					policies = append(policies, pol)
					alreadyIncludedPolicies.Insert(key)
				}
			}
		}
		if err := processExtensionPostListenerHook(tCtx, listener, policies, t.ExtensionManager); err != nil {
			// If the extension server returns an error, and the extension server is not configured to fail open,
			// then propagate the error
			if !(*t.ExtensionManager).FailOpen() {
				errs = errors.Join(errs, err)
			} else {
				t.Logger.Error(err, "Extension Manager PostListener failure")
			}
		}
	}
	return errs
}

func (t *Translator) processHTTPReadyListenerXdsTranslation(tCtx *types.ResourceVersionTable, ready *ir.ReadyListener) error {
	// If there is no ready listener, return early.
	// TODO: update all testcases to use the new ReadyListener field
	if ready == nil {
		return nil
	}
	l, err := buildReadyListener(ready)
	if err != nil {
		return err
	}
	err = tCtx.AddXdsResource(resourcev3.ListenerType, l)
	if err != nil {
		return err
	}
	return nil
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
				if quicXDSListener, err = buildXdsQuicListener(httpListener.Name, httpListener.Address,
					httpListener.Port, httpListener.IPFamily, accessLog); err != nil {
					errs = errors.Join(errs, err)
					continue
				}

				if err = tCtx.AddXdsResource(resourcev3.ListenerType, quicXDSListener); err != nil {
					errs = errors.Join(errs, err)
					continue
				}
			}

			// Create a new TCP listener for HTTP1/HTTP2 traffic.
			if tcpXDSListener, err = buildXdsTCPListener(
				httpListener.Name, httpListener.Address, httpListener.Port, httpListener.IPFamily,
				httpListener.TCPKeepalive, httpListener.Connection, accessLog); err != nil {
				errs = errors.Join(errs, err)
				continue
			}

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

		// add http route client certs
		for _, route := range httpListener.Routes {
			if route.Destination != nil {
				for _, st := range route.Destination.Settings {
					if st.TLS != nil {
						for _, cert := range st.TLS.ClientCertificates {
							secret := buildXdsTLSCertSecret(cert)
							if err := tCtx.AddXdsResource(resourcev3.SecretType, secret); err != nil {
								errs = errors.Join(errs, err)
							}
						}
					}
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
				Name:     fmt.Sprintf("%s/%s", httpListener.Name, underscoredHostname),
				Domains:  []string{httpRoute.Hostname},
				Metadata: buildXdsMetadata(httpListener.Metadata),
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
		xdsRoute, err = buildXdsRoute(httpRoute, httpListener)
		if err != nil {
			// skip this route if failed to build xds route
			errs = errors.Join(errs, err)
			continue
		}

		// Check if an extension want to modify the route we just generated
		// If no extension exists (or it doesn't subscribe to this hook) then this is a quick no-op.
		if err = processExtensionPostRouteHook(xdsRoute, vHost, httpRoute, t.ExtensionManager); err != nil {
			// If the extension server returns an error, and the extension server is not configured to fail open,
			// then propagate the error
			if !(*t.ExtensionManager).FailOpen() {
				errs = errors.Join(errs, err)
			} else {
				t.Logger.Error(err, "Extension Manager PostRoute failure")
			}
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
			ea := &ExtraArgs{
				metrics:       metrics,
				http1Settings: httpListener.HTTP1,
				ipFamily:      determineIPFamily(httpRoute.Destination.Settings),
				statName:      httpRoute.Destination.StatName,
			}

			if httpRoute.Traffic != nil && httpRoute.Traffic.HTTP2 != nil {
				ea.http2Settings = httpRoute.Traffic.HTTP2
			}

			var err error
			// If there are no filters in the destination settings we create
			// a regular xds Cluster
			if !httpRoute.Destination.NeedsClusterPerSetting() {
				err = processXdsCluster(
					tCtx,
					httpRoute.Destination.Name,
					httpRoute.Destination.Settings,
					&HTTPRouteTranslator{httpRoute},
					ea,
					httpRoute.Destination.Metadata,
				)
				if err != nil {
					errs = errors.Join(errs, err)
				}
			} else {
				// If a filter does exist, we create a weighted cluster that's
				// attached to the route, and create a xds Cluster per setting
				for _, setting := range httpRoute.Destination.Settings {
					tSettings := []*ir.DestinationSetting{setting}
					err = processXdsCluster(
						tCtx,
						setting.Name,
						tSettings,
						&HTTPRouteTranslator{httpRoute},
						ea,
						httpRoute.Destination.Metadata)
					if err != nil {
						errs = errors.Join(errs, err)
					}
				}
			}

			if err != nil {
				errs = errors.Join(errs, err)
			}
		}

		if httpRoute.Mirrors != nil {
			for _, mrr := range httpRoute.Mirrors {
				if mrr.Destination != nil {
					if err = addXdsCluster(tCtx, &xdsClusterArgs{
						name:         mrr.Destination.Name,
						settings:     mrr.Destination.Settings,
						tSocket:      nil,
						endpointType: EndpointTypeStatic,
						metrics:      metrics,
						metadata:     mrr.Destination.Metadata,
					}); err != nil {
						errs = errors.Join(errs, err)
					}
				}
			}
		}
	}

	for _, vHost := range vHostList {
		// Check if an extension want to modify the Virtual Host we just generated
		// If no extension exists (or it doesn't subscribe to this hook) then this is a quick no-op.
		if err = processExtensionPostVHostHook(vHost, t.ExtensionManager); err != nil {
			// If the extension server returns an error, and the extension server is not configured to fail open,
			// then propagate the error
			if !(*t.ExtensionManager).FailOpen() {
				errs = errors.Join(errs, err)
			} else {
				t.Logger.Error(err, "Extension Manager PostVHost failure")
			}
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
	return replaceHCMInFilterChain(hcm, filterChain)
}

func replaceHCMInFilterChain(hcm *hcmv3.HttpConnectionManager, filterChain *listenerv3.FilterChain) error {
	var err error
	for i, filter := range filterChain.Filters {
		if filter.Name == wellknown.HTTPConnectionManager {
			var mgrAny *anypb.Any
			if mgrAny, err = proto.ToAnyWithValidation(hcm); err != nil {
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
			if err := anypb.UnmarshalTo(filter.GetTypedConfig(), hcm, protobuf.UnmarshalOptions{}); err != nil {
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

func (t *Translator) processTCPListenerXdsTranslation(
	tCtx *types.ResourceVersionTable,
	tcpListeners []*ir.TCPListener,
	accesslog *ir.AccessLog,
	metrics *ir.Metrics,
) error {
	// The XDS translation is done in a best-effort manner, so we collect all
	// errors and return them at the end.
	var errs, err error
	for _, tcpListener := range tcpListeners {
		// Search for an existing listener, if it does not exist, create one.
		xdsListener := findXdsListenerByHostPort(tCtx, tcpListener.Address, tcpListener.Port, corev3.SocketAddress_TCP)
		if xdsListener == nil {
			if xdsListener, err = buildXdsTCPListener(
				tcpListener.Name, tcpListener.Address, tcpListener.Port, tcpListener.IPFamily,
				tcpListener.TCPKeepalive, tcpListener.Connection, accesslog); err != nil {
				// skip this listener if failed to build xds listener
				errs = errors.Join(errs, err)
				continue
			}

			if err := tCtx.AddXdsResource(resourcev3.ListenerType, xdsListener); err != nil {
				// skip this listener if failed to add xds listener to the
				errs = errors.Join(errs, err)
				continue
			}
		}

		// Add the proxy protocol filter if needed
		// TODO: should make sure all listeners that will be translated into same xDS listener have
		// same EnableProxyProtocol value, otherwise listeners with EnableProxyProtocol=false will
		// never accept connection, because listeners with EnableProxyProtocol=true has configured
		// proxy protocol listener filter for xDS listener, all connection must have ProxyProtocol header.
		patchProxyProtocolFilter(xdsListener, tcpListener.EnableProxyProtocol)

		for _, route := range tcpListener.Routes {
			if err := processXdsCluster(tCtx,
				route.Destination.Name,
				route.Destination.Settings,
				&TCPRouteTranslator{route},
				&ExtraArgs{metrics: metrics},
				route.Destination.Metadata); err != nil {
				errs = errors.Join(errs, err)
			}
			if route.TLS != nil && route.TLS.Terminate != nil {
				// add tls route client certs
				for _, cert := range route.TLS.Terminate.ClientCertificates {
					secret := buildXdsTLSCertSecret(cert)
					if err := tCtx.AddXdsResource(resourcev3.SecretType, secret); err != nil {
						errs = errors.Join(errs, err)
					}
				}

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
			if err := addXdsTCPFilterChain(xdsListener, route, route.Destination.Name, accesslog, tcpListener.Timeout, tcpListener.Connection); err != nil {
				errs = errors.Join(errs, err)
			}
		}

		// If there are no routes, add a route without a destination to the listener to create a filter chain
		// This is needed because Envoy requires a filter chain to be present in the listener, otherwise it will reject the listener and report a warning
		if len(tcpListener.Routes) == 0 {
			if findXdsCluster(tCtx, emptyClusterName) == nil {
				if err := tCtx.AddXdsResource(resourcev3.ClusterType, emptyRouteCluster); err != nil {
					errs = errors.Join(errs, err)
				}
			}

			emptyRoute := &ir.TCPRoute{
				Name: emptyClusterName,
				Destination: &ir.RouteDestination{
					Name: emptyClusterName,
				},
			}
			if err := addXdsTCPFilterChain(xdsListener, emptyRoute, emptyClusterName, accesslog, tcpListener.Timeout, tcpListener.Connection); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}
	return errs
}

func processUDPListenerXdsTranslation(
	tCtx *types.ResourceVersionTable,
	udpListeners []*ir.UDPListener,
	accesslog *ir.AccessLog,
	metrics *ir.Metrics,
) error {
	// The XDS translation is done in a best-effort manner, so we collect all
	// errors and return them at the end.
	var errs error

	for _, udpListener := range udpListeners {
		// There won't be multiple UDP listeners on the same port since it's already been checked at the gateway api
		// translator
		if udpListener.Route != nil {
			// 1:1 between IR UDPRoute and xDS Cluster
			if err := processXdsCluster(tCtx,
				udpListener.Route.Destination.Name,
				udpListener.Route.Destination.Settings,
				&UDPRouteTranslator{udpListener.Route},
				&ExtraArgs{metrics: metrics},
				udpListener.Route.Destination.Metadata); err != nil {
				errs = errors.Join(errs, err)
			}
		} else {
			udpListener.Route = &ir.UDPRoute{
				Name: emptyClusterName,
				Destination: &ir.RouteDestination{
					Name: emptyClusterName,
				},
			}

			// Add empty cluster for UDP listener which have no Route, when empty cluster is not found.
			if findXdsCluster(tCtx, emptyClusterName) == nil {
				if err := tCtx.AddXdsResource(resourcev3.ClusterType, emptyRouteCluster); err != nil {
					errs = errors.Join(errs, err)
				}
			}
		}

		xdsListener, err := buildXdsUDPListener(udpListener.Route.Destination.Name, udpListener, accesslog)
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

// processXdsCluster processes xds cluster with args per route.
func processXdsCluster(tCtx *types.ResourceVersionTable,
	name string,
	settings []*ir.DestinationSetting,
	route clusterArgs,
	extras *ExtraArgs,
	metadata *ir.ResourceMetadata,
) error {
	return addXdsCluster(tCtx, route.asClusterArgs(name, settings, extras, metadata))
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

// addXdsSecret adds a xds secret with args.
// If the secret already exists, it skips adding the secret and returns nil
func addXdsSecret(tCtx *types.ResourceVersionTable, secret *tlsv3.Secret) error {
	// Return early if secret with the same name exists
	if c := findXdsSecret(tCtx, secret.Name); c != nil {
		return nil
	}

	if err := tCtx.AddXdsResource(resourcev3.SecretType, secret); err != nil {
		return err
	}
	return nil
}

// addXdsCluster adds a xds cluster with args.
// If the cluster already exists, it skips adding the cluster and returns nil.
func addXdsCluster(tCtx *types.ResourceVersionTable, args *xdsClusterArgs) error {
	// Return early if cluster with the same name exists.
	// All the current callers can all safely assume the xdsClusterArgs is the same for the clusters with the same name.
	// If this assumption changes, the callers should call findXdsCluster first to check if the cluster already exists
	// before calling addXdsCluster.
	if c := findXdsCluster(tCtx, args.name); c != nil {
		return nil
	}

	result, err := buildXdsCluster(args)
	if err != nil {
		return err
	}
	xdsCluster := result.cluster
	xdsEndpoints := buildXdsClusterLoadAssignment(args.name, args.settings)
	for _, ds := range args.settings {
		if ds.TLS != nil {
			// Create an SDS secret for the CA certificate - either with inline bytes or with a filesystem ref
			secret := buildXdsUpstreamTLSCASecret(ds.TLS)
			if err := tCtx.AddXdsResource(resourcev3.SecretType, secret); err != nil {
				return err
			}
		}
	}
	// Use EDS for static endpoints
	switch args.endpointType {
	case EndpointTypeStatic:
		if err := tCtx.AddXdsResource(resourcev3.EndpointType, xdsEndpoints); err != nil {
			return err
		}
	case EndpointTypeDNS:
		xdsCluster.LoadAssignment = xdsEndpoints
	case EndpointTypeDynamicResolver:
		// Dynamic resolver has no endpoints
		// This assignment is not necessary, but it is added for clarity.
		xdsCluster.LoadAssignment = nil
	}

	if err := tCtx.AddXdsResource(resourcev3.ClusterType, xdsCluster); err != nil {
		return err
	}

	// Add the secrets used in the cluster filters. (Currently only used for Credential Injector filter)
	for _, secret := range result.secrets {
		if err := tCtx.AddXdsResource(resourcev3.SecretType, secret); err != nil {
			return err
		}
	}
	return nil
}

const (
	DefaultEndpointType EndpointType = iota
	Static
	EDS
)

// defaultCertificateName is the default location of the system trust store, initialized at runtime once.
//
// This assumes the Envoy running in a very specific environment. For example, the default location of the system
// trust store on Debian derivatives like the envoy-proxy image being used by the infrastructure controller.
//
// TODO: this might be configurable by an env var or EnvoyGateway configuration.
var defaultCertificateName = func() string {
	switch runtime.GOOS {
	case "darwin":
		// TODO: maybe automatically get the keychain cert? That might be macOS version dependent.
		// For now, we'll just use the root cert installed by Homebrew: brew install ca-certificates.
		//
		// See:
		// * https://apple.stackexchange.com/questions/226375/where-are-the-root-cas-stored-on-os-x
		// * https://superuser.com/questions/992167/where-are-digital-certificates-physically-stored-on-a-mac-os-x-machine
		return "/opt/homebrew/etc/ca-certificates/cert.pem"
	default:
		// This is the default location for the system trust store
		// on Debian derivatives like the envoy-proxy image being used by the infrastructure
		// controller.
		// See https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/security/ssl
		return "/etc/ssl/certs/ca-certificates.crt"
	}
}()

func buildXdsUpstreamTLSCASecret(tlsConfig *ir.TLSUpstreamConfig) *tlsv3.Secret {
	// Build the tls secret
	if tlsConfig.UseSystemTrustStore {
		return &tlsv3.Secret{
			Name: tlsConfig.CACertificate.Name,
			Type: &tlsv3.Secret_ValidationContext{
				ValidationContext: &tlsv3.CertificateValidationContext{
					TrustedCa: &corev3.DataSource{
						Specifier: &corev3.DataSource_Filename{Filename: defaultCertificateName},
					},
				},
			},
		}
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

func buildValidationContext(tlsConfig *ir.TLSUpstreamConfig) *tlsv3.CommonTlsContext_CombinedValidationContext {
	validationContext := &tlsv3.CommonTlsContext_CombinedCertificateValidationContext{
		ValidationContextSdsSecretConfig: &tlsv3.SdsSecretConfig{
			Name:      tlsConfig.CACertificate.Name,
			SdsConfig: makeConfigSource(),
		},
		DefaultValidationContext: &tlsv3.CertificateValidationContext{},
	}

	if tlsConfig.SNI != nil {
		validationContext.DefaultValidationContext.MatchTypedSubjectAltNames = []*tlsv3.SubjectAltNameMatcher{
			{
				SanType: tlsv3.SubjectAltNameMatcher_DNS,
				Matcher: &matcherv3.StringMatcher{
					MatchPattern: &matcherv3.StringMatcher_Exact{
						Exact: *tlsConfig.SNI,
					},
				},
			},
		}
		for _, san := range tlsConfig.SubjectAltNames {
			var sanType tlsv3.SubjectAltNameMatcher_SanType
			var value string

			// Exactly one of san.Hostname or san.URI is guaranteed to be set
			if san.Hostname != nil {
				sanType = tlsv3.SubjectAltNameMatcher_DNS
				value = *san.Hostname
			} else if san.URI != nil {
				sanType = tlsv3.SubjectAltNameMatcher_URI
				value = *san.URI
			}
			validationContext.DefaultValidationContext.MatchTypedSubjectAltNames = append(
				validationContext.DefaultValidationContext.MatchTypedSubjectAltNames,
				&tlsv3.SubjectAltNameMatcher{
					SanType: sanType,
					Matcher: &matcherv3.StringMatcher{
						MatchPattern: &matcherv3.StringMatcher_Exact{
							Exact: value,
						},
					},
				},
			)
		}
	}

	return &tlsv3.CommonTlsContext_CombinedValidationContext{
		CombinedValidationContext: validationContext,
	}
}

func buildXdsUpstreamTLSSocketWthCert(tlsConfig *ir.TLSUpstreamConfig) (*corev3.TransportSocket, error) {
	tlsCtx := &tlsv3.UpstreamTlsContext{
		CommonTlsContext: &tlsv3.CommonTlsContext{
			TlsCertificateSdsSecretConfigs: nil,
			ValidationContextType:          nil,
		},
	}

	if !tlsConfig.InsecureSkipVerify {
		tlsCtx.CommonTlsContext.ValidationContextType = buildValidationContext(tlsConfig)
	}

	if tlsConfig.SNI != nil {
		tlsCtx.Sni = *tlsConfig.SNI
	}

	tlsParams := buildTLSParams(&tlsConfig.TLSConfig)
	if tlsParams != nil {
		tlsCtx.CommonTlsContext.TlsParams = tlsParams
	}

	if tlsConfig.ALPNProtocols != nil {
		tlsCtx.CommonTlsContext.AlpnProtocols = tlsConfig.ALPNProtocols
	}

	if len(tlsConfig.ClientCertificates) > 0 {
		for _, cert := range tlsConfig.ClientCertificates {
			tlsCtx.CommonTlsContext.TlsCertificateSdsSecretConfigs = append(
				tlsCtx.CommonTlsContext.TlsCertificateSdsSecretConfigs,
				&tlsv3.SdsSecretConfig{
					Name:      cert.Name,
					SdsConfig: makeConfigSource(),
				})
		}
	}

	tlsCtxAny, err := proto.ToAnyWithValidation(tlsCtx)
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
