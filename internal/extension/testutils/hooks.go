// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package testutils

import (
	"errors"
	"fmt"
	"strings"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"google.golang.org/protobuf/types/known/durationpb"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type XDSHookClient struct{}

// PostRouteModifyHook returns a modified version of the route using context info and the passed in extensionResources
func (c *XDSHookClient) PostRouteModifyHook(route *route.Route, routeHostnames []string, extensionResources []*unstructured.Unstructured) (*route.Route, error) {
	for _, extensionResource := range extensionResources {
		// Simulate an error an extension may return
		if route.Name == "extension-post-xdsroute-hook-error" {
			return nil, errors.New("route hook resource error")
		}
		route.ResponseHeadersToAdd = append(route.ResponseHeadersToAdd,
			&core.HeaderValueOption{
				Header: &core.HeaderValue{
					Key:   "mock-extension-was-here-route-name",
					Value: route.Name,
				},
			},
			&core.HeaderValueOption{
				Header: &core.HeaderValue{
					Key:   "mock-extension-was-here-route-hostnames",
					Value: strings.Join(routeHostnames, ", "),
				},
			},
			&core.HeaderValueOption{
				Header: &core.HeaderValue{
					Key:   "mock-extension-was-here-extensionRef-name",
					Value: extensionResource.GetName(),
				},
			},
			&core.HeaderValueOption{
				Header: &core.HeaderValue{
					Key:   "mock-extension-was-here-extensionRef-namespace",
					Value: extensionResource.GetNamespace(),
				},
			},
			&core.HeaderValueOption{
				Header: &core.HeaderValue{
					Key:   "mock-extension-was-here-extensionRef-kind",
					Value: extensionResource.GetKind(),
				},
			},
			&core.HeaderValueOption{
				Header: &core.HeaderValue{
					Key:   "mock-extension-was-here-extensionRef-apiversion",
					Value: extensionResource.GetAPIVersion(),
				},
			},
		)
	}
	return route, nil
}

// PostVirtualHostModifyHook returns a modified version of the virtualhost with a new route injected
func (c *XDSHookClient) PostVirtualHostModifyHook(vh *route.VirtualHost) (*route.VirtualHost, error) {
	// Only make the change when the VirtualHost's name matches the expected testdata
	// This prevents us from having to update every single testfile.out
	if vh.Name == "extension-post-xdsvirtualhost-hook-error" {
		return nil, fmt.Errorf("extension post xds virtual host hook error")
	} else if vh.Name == "extension-listener" {
		vh.Routes = append(vh.Routes, &route.Route{
			Name: "mock-extension-inserted-route",
			Action: &route.Route_DirectResponse{
				DirectResponse: &route.DirectResponseAction{
					Status: uint32(200),
				},
			},
		})
	}
	return vh, nil
}

// PostHTTPListenerModifyHook returns a modified version of the listener with a changed statprefix of the listener
// A more useful use-case for an extension would be looping through the FilterChains to find the
// HTTPConnectionManager(s) and inject a custom HTTPFilter, but that for testing purposes we don't need to make a complex change
func (c *XDSHookClient) PostHTTPListenerModifyHook(l *listener.Listener) (*listener.Listener, error) {

	// Only make the change when the listener's name matches the expected testdata
	// This prevents us from having to update every single testfile.out
	if l.Name == "extension-post-xdslistener-hook-error" {
		return nil, fmt.Errorf("extension post xds listener hook error")
	} else if l.Name == "extension-listener" {
		l.StatPrefix = "mock-extension-inserted-prefix"
	}
	return l, nil
}

// PostTranslationInsertHook returns a static cluster and tls secret to be inserted
func (c *XDSHookClient) PostTranslationInsertHook() ([]*cluster.Cluster, []*tls.Secret, error) {

	extensionSvcEndpoint := &endpoint.LbEndpoint_Endpoint{
		Endpoint: &endpoint.Endpoint{
			Address: &core.Address{
				Address: &core.Address_SocketAddress{
					SocketAddress: &core.SocketAddress{
						Address: "exampleservice.examplenamespace.svc.cluster.local",
						PortSpecifier: &core.SocketAddress_PortValue{
							PortValue: 5000,
						},
						Protocol: core.SocketAddress_TCP,
					},
				},
			},
		},
	}

	clusters := []*cluster.Cluster{
		{
			Name: "mock-extension-injected-cluster",
			LoadAssignment: &endpoint.ClusterLoadAssignment{
				ClusterName: "mock-extension-injected-cluster",
				Endpoints: []*endpoint.LocalityLbEndpoints{
					{
						LbEndpoints: []*endpoint.LbEndpoint{
							{
								HostIdentifier: extensionSvcEndpoint,
							},
						},
					},
				},
			},
		},
		// This is one of the default test clusters, but the connect timeout has been changed
		{
			Name: "first-route",
			CommonLbConfig: &cluster.Cluster_CommonLbConfig{
				LocalityConfigSpecifier: &cluster.Cluster_CommonLbConfig_LocalityWeightedLbConfig_{},
			},
			ConnectTimeout:  &durationpb.Duration{Seconds: 30},
			DnsLookupFamily: cluster.Cluster_V4_ONLY,
			EdsClusterConfig: &cluster.Cluster_EdsClusterConfig{
				EdsConfig: &core.ConfigSource{
					ConfigSourceSpecifier: &core.ConfigSource_Ads{},
					ResourceApiVersion:    core.ApiVersion_V3,
				},
			},
			OutlierDetection:     &cluster.OutlierDetection{},
			ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
		},
	}

	secrets := []*tls.Secret{
		{
			Name: "mock-extension-injected-secret",
			Type: &tls.Secret_GenericSecret{
				GenericSecret: &tls.GenericSecret{
					Secret: &core.DataSource{
						Specifier: &core.DataSource_InlineString{
							InlineString: "super-secret-extension-secret",
						},
					},
				},
			},
		},
	}

	return clusters, secrets, nil
}
