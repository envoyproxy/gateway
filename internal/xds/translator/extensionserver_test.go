// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"context"
	"errors"
	"fmt"
	"strings"

	clusterV3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	coreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointV3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerV3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routeV3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	tlsV3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	pb "github.com/envoyproxy/gateway/proto/extension"
)

type testingExtensionServer struct {
	pb.UnimplementedEnvoyGatewayExtensionServer
}

// PostRouteModifyHook returns a modified version of the route using context info and the passed in extensionResources
func (t *testingExtensionServer) PostRouteModify(_ context.Context, req *pb.PostRouteModifyRequest) (*pb.PostRouteModifyResponse, error) {
	// Simulate an error an extension may return
	if req.Route.Name == "extension-post-xdsroute-hook-error" {
		return &pb.PostRouteModifyResponse{
			Route: req.Route,
		}, errors.New("route hook resource error")
	}

	// Setup a new route to avoid operating directly on the passed in pointer for better test coverage that the
	// route we are returning gets used properly
	modifiedRoute := proto.Clone(req.Route).(*routeV3.Route)
	for _, extensionResourceBytes := range req.PostRouteContext.ExtensionResources {
		extensionResource := unstructured.Unstructured{}
		if err := extensionResource.UnmarshalJSON(extensionResourceBytes.UnstructuredBytes); err != nil {
			return &pb.PostRouteModifyResponse{
				Route: req.Route,
			}, err
		}
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
					Value: strings.Join(req.PostRouteContext.Hostnames, ", "),
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
	return &pb.PostRouteModifyResponse{
		Route: modifiedRoute,
	}, nil
}

// PostVirtualHostModifyHook returns a modified version of the virtualhost with a new route injected
func (t *testingExtensionServer) PostVirtualHostModify(_ context.Context, req *pb.PostVirtualHostModifyRequest) (*pb.PostVirtualHostModifyResponse, error) {
	// Only make the change when the VirtualHost's name matches the expected testdata
	// This prevents us from having to update every single testfile.out
	switch req.VirtualHost.Name {
	case "extension-post-xdsvirtualhost-hook-error/*":
		return &pb.PostVirtualHostModifyResponse{
			VirtualHost: req.VirtualHost,
		}, fmt.Errorf("extension post xds virtual host hook error")
	case "extension-listener":
		// Setup a new VirtualHost to avoid operating directly on the passed in pointer for better test coverage that the
		// VirtualHost we are returning gets used properly
		modifiedVH := proto.Clone(req.VirtualHost).(*routeV3.VirtualHost)
		modifiedVH.Routes = append(modifiedVH.Routes, &routeV3.Route{
			Name: "mock-extension-inserted-route",
			Action: &routeV3.Route_DirectResponse{
				DirectResponse: &routeV3.DirectResponseAction{
					Status: uint32(200),
				},
			},
		})
		return &pb.PostVirtualHostModifyResponse{
			VirtualHost: modifiedVH,
		}, nil
	}
	return &pb.PostVirtualHostModifyResponse{
		VirtualHost: req.VirtualHost,
	}, nil
}

// PostHTTPListenerModifyHook returns a modified version of the listener with a changed statprefix of the listener
// A more useful use-case for an extension would be looping through the FilterChains to find the
// HTTPConnectionManager(s) and inject a custom HTTPFilter, but that for testing purposes we don't need to make a complex change
func (t *testingExtensionServer) PostHTTPListenerModify(_ context.Context, req *pb.PostHTTPListenerModifyRequest) (*pb.PostHTTPListenerModifyResponse, error) {
	// Only make the change when the listener's name matches the expected testdata
	// This prevents us from having to update every single testfile.out
	switch req.Listener.Name {
	case "extension-post-xdslistener-hook-error":
		return &pb.PostHTTPListenerModifyResponse{
			Listener: req.Listener,
		}, fmt.Errorf("extension post xds listener hook error")
	case "extension-listener":
		// Setup a new Listener to avoid operating directly on the passed in pointer for better test coverage that the
		// Listener we are returning gets used properly
		modifiedListener := proto.Clone(req.Listener).(*listenerV3.Listener)
		modifiedListener.StatPrefix = "mock-extension-inserted-prefix"
		return &pb.PostHTTPListenerModifyResponse{
			Listener: modifiedListener,
		}, nil
	case "policyextension-listener":
		if len(req.PostListenerContext.ExtensionResources) == 0 {
			return nil, fmt.Errorf("expected a policy in the ext array")
		}
		extensionResource := unstructured.Unstructured{}
		if err := extensionResource.UnmarshalJSON(req.PostListenerContext.ExtensionResources[0].UnstructuredBytes); err != nil {
			return &pb.PostHTTPListenerModifyResponse{
				Listener: req.Listener,
			}, err
		}
		spec, ok := extensionResource.Object["spec"].(map[string]any)
		if !ok {
			return &pb.PostHTTPListenerModifyResponse{
				Listener: req.Listener,
			}, fmt.Errorf("can't find the spec section")
		}
		data, ok := spec["data"].(string)
		if !ok {
			return &pb.PostHTTPListenerModifyResponse{
				Listener: req.Listener,
			}, fmt.Errorf("can't find the expected information")
		}
		modifiedListener := proto.Clone(req.Listener).(*listenerV3.Listener)
		modifiedListener.StatPrefix = data
		return &pb.PostHTTPListenerModifyResponse{
			Listener: modifiedListener,
		}, nil
	case "envoy-gateway/gateway-1/http1":
		if len(req.PostListenerContext.ExtensionResources) != 1 {
			return &pb.PostHTTPListenerModifyResponse{
					Listener: req.Listener,
				}, fmt.Errorf("received %d extension policies when expecting 1: %s",
					len(req.PostListenerContext.ExtensionResources), req.Listener.Name)
		}
		modifiedListener := proto.Clone(req.Listener).(*listenerV3.Listener)
		modifiedListener.StatPrefix = req.Listener.Name
		return &pb.PostHTTPListenerModifyResponse{
			Listener: modifiedListener,
		}, nil
	case "envoy-gateway/gateway-1/tcp1":
		return &pb.PostHTTPListenerModifyResponse{
			Listener: req.Listener,
		}, fmt.Errorf("should not be called for this listener, test 'extensionpolicy-tcp-and-http' should merge tcp and http gateways to one listener")
	case "envoy-gateway/gateway-1/udp1":
		if len(req.PostListenerContext.ExtensionResources) != 1 {
			return &pb.PostHTTPListenerModifyResponse{
					Listener: req.Listener,
				}, fmt.Errorf("received %d extension policies when expecting 1: %s",
					len(req.PostListenerContext.ExtensionResources), req.Listener.Name)
		}
		modifiedListener := proto.Clone(req.Listener).(*listenerV3.Listener)
		modifiedListener.StatPrefix = req.Listener.Name
		return &pb.PostHTTPListenerModifyResponse{
			Listener: modifiedListener,
		}, nil
	case "first-listener-error":
		modifiedListener := proto.Clone(req.Listener).(*listenerV3.Listener)
		modifiedListener.StatPrefix = req.Listener.Name
		return &pb.PostHTTPListenerModifyResponse{
			Listener: modifiedListener,
		}, fmt.Errorf("simulate error when there is no default filter chain in the original resources")
	}
	return &pb.PostHTTPListenerModifyResponse{
		Listener: req.Listener,
	}, nil
}

// PostTranslateModifyHook inserts and overrides some clusters/secrets
func (t *testingExtensionServer) PostTranslateModify(_ context.Context, req *pb.PostTranslateModifyRequest) (*pb.PostTranslateModifyResponse, error) {
	for _, cluster := range req.Clusters {
		// This simulates an extension server that returns an error. It allows verifying that fail-close is working.
		if edsConfig := cluster.GetEdsClusterConfig(); edsConfig != nil {
			if strings.Contains(edsConfig.ServiceName, "fail-close-error") {
				return &pb.PostTranslateModifyResponse{
					Clusters: req.Clusters,
					Secrets:  req.Secrets,
				}, fmt.Errorf("cluster hook resource error: %s", edsConfig.ServiceName)
			}
		}
	}

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

	response := &pb.PostTranslateModifyResponse{
		Clusters: make([]*clusterV3.Cluster, len(req.Clusters)),
		Secrets:  make([]*tlsV3.Secret, len(req.Secrets)),
	}
	for idx, cluster := range req.Clusters {
		response.Clusters[idx] = proto.Clone(cluster).(*clusterV3.Cluster)
		if cluster.Name == "first-route" {
			response.Clusters[idx].ConnectTimeout = &durationpb.Duration{Seconds: 30}
		}
	}

	response.Clusters = append(response.Clusters, &clusterV3.Cluster{
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

	for idx, secret := range req.Secrets {
		response.Secrets[idx] = proto.Clone(secret).(*tlsV3.Secret)
	}
	response.Secrets = append(response.Secrets, &tlsV3.Secret{
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

	return response, nil
}
