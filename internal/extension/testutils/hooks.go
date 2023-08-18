// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package testutils

import (
	"errors"
	"fmt"
	"strings"

	clusterV3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	coreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointV3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerV3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routeV3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	tlsV3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type XDSHookClient struct{}

// PostRouteModifyHook returns a modified version of the route using context info and the passed in extensionResources
func (c *XDSHookClient) PostRouteModifyHook(route *routeV3.Route, routeHostnames []string, extensionResources []*unstructured.Unstructured) (*routeV3.Route, error) {
	// Simulate an error an extension may return
	if route.Name == "extension-post-xdsroute-hook-error" {
		return nil, errors.New("route hook resource error")
	}

	// Setup a new route to avoid operating directly on the passed in pointer for better test coverage that the
	// route we are returning gets used properly
	modifiedRoute := proto.Clone(route).(*routeV3.Route)
	for _, extensionResource := range extensionResources {
		modifiedRoute.ResponseHeadersToAdd = append(modifiedRoute.ResponseHeadersToAdd,
			&coreV3.HeaderValueOption{
				Header: &coreV3.HeaderValue{
					Key:   "mock-extension-was-here-route-name",
					Value: modifiedRoute.Name,
				},
			},
			&coreV3.HeaderValueOption{
				Header: &coreV3.HeaderValue{
					Key:   "mock-extension-was-here-route-hostnames",
					Value: strings.Join(routeHostnames, ", "),
				},
			},
			&coreV3.HeaderValueOption{
				Header: &coreV3.HeaderValue{
					Key:   "mock-extension-was-here-extensionRef-name",
					Value: extensionResource.GetName(),
				},
			},
			&coreV3.HeaderValueOption{
				Header: &coreV3.HeaderValue{
					Key:   "mock-extension-was-here-extensionRef-namespace",
					Value: extensionResource.GetNamespace(),
				},
			},
			&coreV3.HeaderValueOption{
				Header: &coreV3.HeaderValue{
					Key:   "mock-extension-was-here-extensionRef-kind",
					Value: extensionResource.GetKind(),
				},
			},
			&coreV3.HeaderValueOption{
				Header: &coreV3.HeaderValue{
					Key:   "mock-extension-was-here-extensionRef-apiversion",
					Value: extensionResource.GetAPIVersion(),
				},
			},
		)
	}
	return modifiedRoute, nil
}

// PostVirtualHostModifyHook returns a modified version of the virtualhost with a new route injected
func (c *XDSHookClient) PostVirtualHostModifyHook(vh *routeV3.VirtualHost) (*routeV3.VirtualHost, error) {
	// Only make the change when the VirtualHost's name matches the expected testdata
	// This prevents us from having to update every single testfile.out
	if vh.Name == "extension-post-xdsvirtualhost-hook-error/*" {
		return nil, fmt.Errorf("extension post xds virtual host hook error")
	} else if vh.Name == "extension-listener" {
		// Setup a new VirtualHost to avoid operating directly on the passed in pointer for better test coverage that the
		// VirtualHost we are returning gets used properly
		modifiedVH := proto.Clone(vh).(*routeV3.VirtualHost)
		modifiedVH.Routes = append(modifiedVH.Routes, &routeV3.Route{
			Name: "mock-extension-inserted-route",
			Action: &routeV3.Route_DirectResponse{
				DirectResponse: &routeV3.DirectResponseAction{
					Status: uint32(200),
				},
			},
		})
		return modifiedVH, nil
	}
	return vh, nil
}

// PostHTTPListenerModifyHook returns a modified version of the listener with a changed statprefix of the listener
// A more useful use-case for an extension would be looping through the FilterChains to find the
// HTTPConnectionManager(s) and inject a custom HTTPFilter, but that for testing purposes we don't need to make a complex change
func (c *XDSHookClient) PostHTTPListenerModifyHook(l *listenerV3.Listener) (*listenerV3.Listener, error) {

	// Only make the change when the listener's name matches the expected testdata
	// This prevents us from having to update every single testfile.out
	if l.Name == "extension-post-xdslistener-hook-error" {
		return nil, fmt.Errorf("extension post xds listener hook error")
	} else if l.Name == "extension-listener" {
		// Setup a new Listener to avoid operating directly on the passed in pointer for better test coverage that the
		// Listener we are returning gets used properly
		modifiedListener := proto.Clone(l).(*listenerV3.Listener)
		modifiedListener.StatPrefix = "mock-extension-inserted-prefix"
		return modifiedListener, nil
	}
	return l, nil
}

// PostTranslateModifyHook inserts and overrides some clusters/secrets
func (c *XDSHookClient) PostTranslateModifyHook(clusters []*clusterV3.Cluster, secrets []*tlsV3.Secret) ([]*clusterV3.Cluster, []*tlsV3.Secret, error) {

	extensionSvcEndpoint := &endpointV3.LbEndpoint_Endpoint{
		Endpoint: &endpointV3.Endpoint{
			Address: &coreV3.Address{
				Address: &coreV3.Address_SocketAddress{
					SocketAddress: &coreV3.SocketAddress{
						Address: "exampleservice.examplenamespace.svc.cluster.local",
						PortSpecifier: &coreV3.SocketAddress_PortValue{
							PortValue: 5000,
						},
						Protocol: coreV3.SocketAddress_TCP,
					},
				},
			},
		},
	}

	for idx, cluster := range clusters {
		if cluster.Name == "first-route" {
			clusters[idx].ConnectTimeout = &durationpb.Duration{Seconds: 30}
		}
	}

	clusters = append(clusters, &clusterV3.Cluster{
		Name: "mock-extension-injected-cluster",
		LoadAssignment: &endpointV3.ClusterLoadAssignment{
			ClusterName: "mock-extension-injected-cluster",
			Endpoints: []*endpointV3.LocalityLbEndpoints{
				{
					LbEndpoints: []*endpointV3.LbEndpoint{
						{
							HostIdentifier: extensionSvcEndpoint,
						},
					},
				},
			},
		},
	})

	secrets = append(secrets, &tlsV3.Secret{
		Name: "mock-extension-injected-secret",
		Type: &tlsV3.Secret_GenericSecret{
			GenericSecret: &tlsV3.GenericSecret{
				Secret: &coreV3.DataSource{
					Specifier: &coreV3.DataSource_InlineString{
						InlineString: "super-secret-extension-secret",
					},
				},
			},
		},
	})

	return clusters, secrets, nil
}
