// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package registry

import (
	"context"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/proto/extension"
)

var _ types.XDSHookClient = (*XDSHook)(nil)

type XDSHook struct {
	grpcClient extension.EnvoyGatewayExtensionClient
}

func (h *XDSHook) PostRouteModifyHook(route *route.Route, routeHostnames []string, extensionResources []*unstructured.Unstructured) (*route.Route, error) {
	// Take all of the unstructured resources for the extension and package them into bytes
	extensionResourceBytes := []*extension.ExtensionResource{}
	for _, res := range extensionResources {
		if res != nil {
			unstructuredBytes, err := res.MarshalJSON()
			// This is probably a programming error, but just return the unmodified route if so
			if err != nil {
				return route, err
			}

			extensionResourceBytes = append(extensionResourceBytes,
				&extension.ExtensionResource{
					UnstructuredBytes: unstructuredBytes,
				},
			)
		}
	}

	// Make the request to the extension server
	ctx := context.Background()
	resp, err := h.grpcClient.PostRouteModify(ctx,
		&extension.PostRouteModifyRequest{
			Route: route,
			PostRouteContext: &extension.PostRouteExtensionContext{
				Hostnames:          routeHostnames,
				ExtensionResources: extensionResourceBytes,
			},
		})

	if err != nil {
		return nil, err
	}

	return resp.Route, nil
}

func (h *XDSHook) PostVirtualHostModifyHook(vh *route.VirtualHost) (*route.VirtualHost, error) {
	// Make the request to the extension server
	ctx := context.Background()
	resp, err := h.grpcClient.PostVirtualHostModify(ctx,
		&extension.PostVirtualHostModifyRequest{
			VirtualHost:            vh,
			PostVirtualHostContext: &extension.PostVirtualHostExtensionContext{},
		})

	if err != nil {
		return nil, err
	}

	return resp.VirtualHost, nil
}

func (h *XDSHook) PostHTTPListenerModifyHook(l *listener.Listener) (*listener.Listener, error) {
	// Make the request to the extension server
	ctx := context.Background()
	resp, err := h.grpcClient.PostHTTPListenerModify(ctx,
		&extension.PostHTTPListenerModifyRequest{
			Listener:            l,
			PostListenerContext: &extension.PostHTTPListenerExtensionContext{},
		})

	if err != nil {
		return nil, err
	}

	return resp.Listener, nil
}

func (h *XDSHook) PostTranslateModifyHook(clusters []*cluster.Cluster, secrets []*tls.Secret) ([]*cluster.Cluster, []*tls.Secret, error) {
	// Make the request to the extension server
	ctx := context.Background()
	resp, err := h.grpcClient.PostTranslateModify(ctx,
		&extension.PostTranslateModifyRequest{
			PostTranslateContext: &extension.PostTranslateExtensionContext{},
			Clusters:             clusters,
			Secrets:              secrets,
		})

	if err != nil {
		return nil, nil, err
	}

	return resp.Clusters, resp.Secrets, nil
}
