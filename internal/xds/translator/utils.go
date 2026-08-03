// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	filterchainv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/filter_chain/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	defaultHTTPSPort                uint64 = 443
	defaultHTTPPort                 uint64 = 80
	defaultExtServiceRequestTimeout        = 10 * time.Second
)

// urlCluster is a cluster that is created from a URL.
type urlCluster struct {
	name         string
	hostname     string
	port         uint32
	endpointType EndpointType
	tls          bool
}

// url2Cluster returns a urlCluster from the provided url.
func url2Cluster(strURL string) (*urlCluster, error) {
	epType := EndpointTypeDNS

	// The URL should have already been validated in the gateway API translator.
	u, err := url.Parse(strURL)
	if err != nil {
		return nil, err
	}

	var port uint64
	if u.Scheme == "https" {
		port = defaultHTTPSPort
	} else {
		port = defaultHTTPPort
	}

	if u.Port() != "" {
		port, err = strconv.ParseUint(u.Port(), 10, 32)
		if err != nil {
			return nil, err
		}
	}

	name := clusterName(u.Hostname(), uint32(port))

	if _, err := netip.ParseAddr(u.Hostname()); err == nil {
		epType = EndpointTypeStatic
	}

	return &urlCluster{
		name:         name,
		hostname:     u.Hostname(),
		port:         uint32(port),
		endpointType: epType,
		tls:          u.Scheme == "https",
	}, nil
}

func clusterName(host string, port uint32) string {
	return fmt.Sprintf("%s_%d", strings.ReplaceAll(host, ".", "_"), port)
}

func destinationSettingName(destName string) string {
	// -1 is used here since this function is used to generate a name
	// for a backend that is defined using a scalar field that has no index.
	return fmt.Sprintf("%s/backend/-1", destName)
}

// enableFilterOnRoute enables a filterType on the provided route.
func enableFilterOnRoute(route *routev3.Route, filterName string, routeCfg proto.Message) error {
	if route == nil {
		return errors.New("xds route is nil")
	}

	filterCfg := route.GetTypedPerFilterConfig()
	if _, ok := filterCfg[filterName]; ok {
		// This should not happen since this is the only place where the filter
		// config is added in a route.
		return fmt.Errorf("route already contains filter config: %s, %+v",
			filterName, route)
	}

	// Enable the corresponding filter for this route.
	routeCfgAny, err := anypb.New(routeCfg)
	if err != nil {
		return err
	}

	if filterCfg == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}

	route.TypedPerFilterConfig[filterName] = routeCfgAny

	return nil
}

// perRouteFilterName generates a unique filter name for the provided filterType and configName.
func perRouteFilterName(filterType egv1a1.EnvoyFilter, configName string) string {
	return fmt.Sprintf("%s/%s", filterType, configName)
}

// buildHCMFilterChainFilter returns a disabled envoy.filters.http.filter_chain placeholder
// filter with the given stable name. Shared by extension types (Lua, ExtProc, Wasm,
// DynamicModule) whose EnvoyExtensionPolicy API allows an ordered list of instances per
// listener/route but whose native Envoy filter can only be overridden with a single instance
// per route. The placeholder's name never changes as instances are added/removed, so the HCM's
// filter list stays stable (no LDS update / connection drain); the actual ordered list of
// instances is delivered separately via FilterChainConfigPerRoute in TypedPerFilterConfig.
func buildHCMFilterChainFilter(filterName string) (*hcmv3.HttpFilter, error) {
	var (
		fcProto *filterchainv3.FilterChainConfig
		fcAny   *anypb.Any
		err     error
	)
	fcProto = &filterchainv3.FilterChainConfig{}

	if err = fcProto.ValidateAll(); err != nil {
		return nil, err
	}
	if fcAny, err = anypb.New(fcProto); err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name:     filterName,
		Disabled: true,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: fcAny,
		},
	}, nil
}

// filterChainFilterNamePrefixForEEP is the stable HCM-level filter name shared by every
// EnvoyExtensionPolicy extension type (Lua, ExtProc, Wasm, DynamicModule) that is delivered via
// an envoy.filters.http.filter_chain placeholder. All extension types share the same two
// placeholders (this name, and this name + ".listener") rather than each getting their own, so
// adding/removing an instance of any extension type never changes the HCM's filter list.
const filterChainFilterNamePrefixForEEP = "envoy.filters.http.filter_chain.eep"

// eepFCFilterName returns the stable HCM-level filter name for the shared, per-route
// EnvoyExtensionPolicy filter_chain placeholder. Every extension type (Lua, ExtProc, Wasm,
// DynamicModule) delivers its route-scoped instances through this single placeholder rather than
// each type getting its own, so adding/removing an instance of any type never changes the HCM's
// filter list.
func eepFCFilterName() string {
	return filterChainFilterNamePrefixForEEP + ".route"
}

// eepListenerFCFilterName returns the stable HCM-level filter name for the shared, per-listener
// (per-connection) EnvoyExtensionPolicy filter_chain placeholder. See eepFCFilterName.
func eepListenerFCFilterName() string {
	return filterChainFilterNamePrefixForEEP + ".listener"
}

// eepSubFilterPriority orders sub-filters within a shared filter_chain placeholder's inner
// FilterChain across extension types: Lua runs before ExtProc, which runs before Wasm, which
// runs before DynamicModule. Instances of the same type keep the relative order they were
// appended in (stable sort), which is already the ordered EnvoyExtensionPolicy list order.
func eepSubFilterPriority(name string) int {
	switch {
	case strings.HasPrefix(name, string(egv1a1.EnvoyFilterLua)):
		return 0
	case strings.HasPrefix(name, string(egv1a1.EnvoyFilterExtProc)):
		return 1
	case strings.HasPrefix(name, string(egv1a1.EnvoyFilterWasm)):
		return 2
	case strings.HasPrefix(name, string(egv1a1.EnvoyFilterDynamicModules)):
		return 3
	default:
		return 99
	}
}

// mergeFilterChainConfigPerRoute merges newFilters into the FilterChainConfigPerRoute already
// stored at existing (nil if none yet), re-sorting the combined inner FilterChain so that
// multiple extension types sharing the same filter_chain placeholder still run in a stable,
// predictable cross-type order instead of whatever order patchRoute/patchVirtualHost happened to
// be called in for this route/virtual host.
func mergeFilterChainConfigPerRoute(existing *anypb.Any, newFilters []*corev3.TypedExtensionConfig) (*anypb.Any, error) {
	fc := &filterchainv3.FilterChainConfigPerRoute{FilterChain: &filterchainv3.FilterChain{}}
	if existing != nil {
		if err := existing.UnmarshalTo(fc); err != nil {
			return nil, err
		}
		if fc.FilterChain == nil {
			fc.FilterChain = &filterchainv3.FilterChain{}
		}
	}

	fc.FilterChain.Filters = append(fc.FilterChain.Filters, newFilters...)
	sort.SliceStable(fc.FilterChain.Filters, func(i, j int) bool {
		return eepSubFilterPriority(fc.FilterChain.Filters[i].Name) < eepSubFilterPriority(fc.FilterChain.Filters[j].Name)
	})

	return anypb.New(fc)
}

// disableFilterOnRouteOnce disables filterName on the route unless some TypedPerFilterConfig is
// already set under that name. Several extension types can all want to disable the same shared
// filter_chain placeholder for a route that fully overrides listener-scoped extensions, so unlike
// enableFilterOnRoute this must tolerate being called more than once for the same route/filter.
func disableFilterOnRouteOnce(route *routev3.Route, filterName string) error {
	if _, ok := route.GetTypedPerFilterConfig()[filterName]; ok {
		return nil
	}
	return enableFilterOnRoute(route, filterName, &routev3.FilterConfig{Disabled: true})
}

// filterChainAlreadyHasType reports whether the FilterChainConfigPerRoute stored at existing (if
// any) already contains a sub-filter of the given type. Used by patchVirtualHost to stay
// idempotent per extension type: unlike the old per-type placeholder, the shared filter_chain key
// may already carry a different type's contribution, so "the key is set" no longer implies "this
// type already delivered its filters here".
func filterChainAlreadyHasType(existing *anypb.Any, filterType egv1a1.EnvoyFilter) (bool, error) {
	if existing == nil {
		return false, nil
	}
	fc := &filterchainv3.FilterChainConfigPerRoute{}
	if err := existing.UnmarshalTo(fc); err != nil {
		return false, err
	}
	if fc.FilterChain == nil {
		return false, nil
	}
	for _, f := range fc.FilterChain.Filters {
		if strings.HasPrefix(f.Name, string(filterType)) {
			return true, nil
		}
	}
	return false, nil
}

func hcmContainsFilter(mgr *hcmv3.HttpConnectionManager, filterName string) bool {
	for _, existingFilter := range mgr.HttpFilters {
		if existingFilter.Name == filterName {
			return true
		}
	}
	return false
}

func createExtServiceXDSCluster(rd *ir.RouteDestination, traffic *ir.TrafficFeatures, tCtx *types.ResourceVersionTable) error {
	var (
		endpointType EndpointType
		tSocket      *corev3.TransportSocket
	)

	// Make sure that there are safe defaults for the traffic
	if traffic == nil {
		traffic = &ir.TrafficFeatures{}
	}
	// Get the address type from the first setting.
	// This is safe because no mixed address types in the settings.
	addrTypeState := rd.Settings[0].AddressType
	if addrTypeState != nil && *addrTypeState == ir.FQDN {
		endpointType = EndpointTypeDNS
	} else {
		endpointType = EndpointTypeStatic
	}

	args := &xdsClusterArgs{
		name:         rd.Name,
		settings:     rd.Settings,
		tSocket:      tSocket,
		endpointType: endpointType,
		metadata:     rd.Metadata,
	}

	applyTraffic(args, traffic)

	return addXdsCluster(tCtx, args)
}

// addClusterFromURL adds a cluster to the resource version table from the provided URL.
func addClusterFromURL(url string, traffic *ir.TrafficFeatures, tCtx *types.ResourceVersionTable) error {
	var (
		uc      *urlCluster
		ds      *ir.DestinationSetting
		tSocket *corev3.TransportSocket
		err     error
	)

	if uc, err = url2Cluster(url); err != nil {
		return err
	}

	ds = &ir.DestinationSetting{
		Weight:    new(uint32(1)),
		Endpoints: []*ir.DestinationEndpoint{ir.NewDestEndpoint(nil, uc.hostname, uc.port, false, nil)},
		Name:      destinationSettingName(uc.name),
		// TODO: tracked with issue #6861
		Metadata: nil,
	}

	clusterArgs := &xdsClusterArgs{
		name:         uc.name,
		settings:     []*ir.DestinationSetting{ds},
		endpointType: uc.endpointType,
		metadata:     ds.Metadata,
	}

	if uc.tls {
		if tSocket, err = buildXdsUpstreamTLSSocket(uc.hostname); err != nil {
			return err
		}
		clusterArgs.tSocket = tSocket
	}

	applyTraffic(clusterArgs, traffic)

	return addXdsCluster(tCtx, clusterArgs)
}

func applyTraffic(args *xdsClusterArgs, traffic *ir.TrafficFeatures) {
	if traffic == nil {
		return
	}
	args.loadBalancer = traffic.LoadBalancer
	args.proxyProtocol = traffic.ProxyProtocol
	args.circuitBreaker = traffic.CircuitBreaker
	args.healthCheck = traffic.HealthCheck
	args.timeout = traffic.Timeout
	args.tcpkeepalive = traffic.TCPKeepalive
	args.backendConnection = traffic.BackendConnection
	args.dns = traffic.DNS
	args.http2Settings = traffic.HTTP2
	args.admissionControl = traffic.AdmissionControl
}

// determineIPFamily determines the IP family based on multiple destination settings
func determineIPFamily(settings []*ir.DestinationSetting) *egv1a1.IPFamily {
	// If there's only one setting, return its IPFamily directly
	if len(settings) == 1 {
		return settings[0].IPFamily
	}

	hasIPv4 := false
	hasIPv6 := false
	hasDualStack := false

	for _, setting := range settings {
		if setting.IPFamily == nil {
			continue
		}

		switch *setting.IPFamily {
		case egv1a1.IPv4:
			hasIPv4 = true
		case egv1a1.IPv6:
			hasIPv6 = true
		case egv1a1.DualStack:
			hasDualStack = true
		}
	}

	switch {
	case hasDualStack:
		return new(egv1a1.DualStack)
	case hasIPv4 && hasIPv6:
		return new(egv1a1.DualStack)
	case hasIPv4:
		return new(egv1a1.IPv4)
	case hasIPv6:
		return new(egv1a1.IPv6)
	default:
		return nil
	}
}
