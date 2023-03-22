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
			RouteContext: &extension.PostRouteExtensionContext{
				Hostnames:          routeHostnames,
				ExtensionResources: extensionResourceBytes,
			},
		})

	// If there was an error then return the original route unmodified
	if err != nil {
		return route, err
	}

	// If the returned route is nil without any error then that means the extension wants us to remove it
	// This is not an error
	if resp != nil {
		return resp.Route, nil
	}
	return nil, nil
}

func (h *XDSHook) PostVirtualHostModifyHook(vh *route.VirtualHost) (*route.VirtualHost, error) {
	// Make the request to the extension server
	ctx := context.Background()
	resp, err := h.grpcClient.PostVirtualHostModify(ctx,
		&extension.PostVirtualHostModifyRequest{
			VirtualHost:        vh,
			VirtualHostContext: &extension.PostVirtualHostExtensionContext{},
		})

	// If there was an error then return the original virtualhost unmodified
	if err != nil {
		return vh, err
	}

	// If the returned virtualhost is nil without any error then that means the extension wants us to remove it
	// This is not an error
	if resp != nil {
		return resp.VirtualHost, nil
	}
	return nil, nil
}

func (h *XDSHook) PostHTTPListenerModifyHook(l *listener.Listener) (*listener.Listener, error) {
	// Make the request to the extension server
	ctx := context.Background()
	resp, err := h.grpcClient.PostHTTPListenerModify(ctx,
		&extension.PostHTTPListenerModifyRequest{
			Listener:        l,
			ListenerContext: &extension.PostHTTPListenerExtensionContext{},
		})

	// If there was an error then return the original listener unmodified
	if err != nil {
		return l, err
	}

	// If the returned listener is nil without any error then that means the extension wants us to remove it
	// This is not an error
	if resp != nil {
		return resp.Listener, nil
	}
	return nil, nil
}

func (h *XDSHook) PostTranslationInsertHook() ([]*cluster.Cluster, []*tls.Secret, error) {
	// Make the request to the extension server
	ctx := context.Background()
	resp, err := h.grpcClient.PostTranslateInsert(ctx,
		&extension.PostTranslationInsertRequest{
			InsertContext: &extension.PostXDSInsertExtensionContext{},
		})

	if err != nil {
		return nil, nil, err
	}

	// An extension may not return anything at all to be injected. This is a no-op and not an error
	if resp != nil {
		return resp.Clusters, resp.Secrets, nil
	}
	return nil, nil, nil
}
