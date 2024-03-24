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

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/types/known/anypb"
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
func perRouteFilterName(filterType, configName string) string {
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
