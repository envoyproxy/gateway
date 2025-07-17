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
	"strconv"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	defaultHTTPSPort                uint64 = 443
	defaultHTTPPort                 uint64 = 80
	defaultExtServiceRequestTimeout        = 10 // 10 seconds
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

	if ip, err := netip.ParseAddr(u.Hostname()); err == nil {
		if ip.Unmap().Is4() {
			epType = EndpointTypeStatic
		}
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
func enableFilterOnRoute(route *routev3.Route, filterName string) error {
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
	routeCfgAny, err := anypb.New(&routev3.FilterConfig{
		Config: &anypb.Any{},
	})
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
	return addXdsCluster(tCtx, &xdsClusterArgs{
		name:              rd.Name,
		settings:          rd.Settings,
		tSocket:           tSocket,
		loadBalancer:      traffic.LoadBalancer,
		proxyProtocol:     traffic.ProxyProtocol,
		circuitBreaker:    traffic.CircuitBreaker,
		healthCheck:       traffic.HealthCheck,
		timeout:           traffic.Timeout,
		tcpkeepalive:      traffic.TCPKeepalive,
		backendConnection: traffic.BackendConnection,
		endpointType:      endpointType,
		dns:               traffic.DNS,
		http2Settings:     traffic.HTTP2,
		metadata:          rd.Metadata,
	})
}

// addClusterFromURL adds a cluster to the resource version table from the provided URL.
func addClusterFromURL(url string, tCtx *types.ResourceVersionTable) error {
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
		Weight:    ptr.To[uint32](1),
		Endpoints: []*ir.DestinationEndpoint{ir.NewDestEndpoint(nil, uc.hostname, uc.port, false, nil)},
		Name:      destinationSettingName(uc.name),
	}

	clusterArgs := &xdsClusterArgs{
		name:         uc.name,
		settings:     []*ir.DestinationSetting{ds},
		endpointType: uc.endpointType,
	}
	if uc.tls {
		if tSocket, err = buildXdsUpstreamTLSSocket(uc.hostname); err != nil {
			return err
		}
		clusterArgs.tSocket = tSocket
	}

	return addXdsCluster(tCtx, clusterArgs)
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
		return ptr.To(egv1a1.DualStack)
	case hasIPv4 && hasIPv6:
		return ptr.To(egv1a1.DualStack)
	case hasIPv4:
		return ptr.To(egv1a1.IPv4)
	case hasIPv6:
		return ptr.To(egv1a1.IPv6)
	default:
		return nil
	}
}
