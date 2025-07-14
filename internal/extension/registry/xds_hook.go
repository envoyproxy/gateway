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
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/proto/extension"
)

var _ types.XDSHookClient = (*XDSHook)(nil)

type XDSHook struct {
	grpcClient extension.EnvoyGatewayExtensionClient
}

func translateUnstructuredToUnstructuredBytes(e []*unstructured.Unstructured) ([]*extension.ExtensionResource, error) {
	extensionResourceBytes := []*extension.ExtensionResource{}
	for _, res := range e {
		if res != nil {
			unstructuredBytes, err := res.MarshalJSON()
			// This is probably a programming error, but just return the unmodified route if so
			if err != nil {
				return nil, err
			}

			extensionResourceBytes = append(extensionResourceBytes,
				&extension.ExtensionResource{
					UnstructuredBytes: unstructuredBytes,
				},
			)
		}
	}
	return extensionResourceBytes, nil
}

func (h *XDSHook) PostRouteModifyHook(route *route.Route, routeHostnames []string, extensionResources []*unstructured.Unstructured) (*route.Route, error) {
	// Take all of the unstructured resources for the extension and package them into bytes
	extensionResourceBytes, err := translateUnstructuredToUnstructuredBytes(extensionResources)
	if err != nil {
		return route, err
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

func (h *XDSHook) PostClusterModifyHook(cluster *cluster.Cluster, extensionResources []*unstructured.Unstructured) (*cluster.Cluster, error) {
	// Take all of the unstructured resources for the extension and package them into bytes
	extensionResourceBytes, err := translateUnstructuredToUnstructuredBytes(extensionResources)
	if err != nil {
		return cluster, err
	}

	// Make the request to the extension server
	ctx := context.Background()
	resp, err := h.grpcClient.PostClusterModify(ctx,
		&extension.PostClusterModifyRequest{
			Cluster: cluster,
			PostClusterContext: &extension.PostClusterExtensionContext{
				BackendExtensionResources: extensionResourceBytes,
			},
		})
	if err != nil {
		return nil, err
	}

	return resp.Cluster, nil
}

func (h *XDSHook) PostVirtualHostModifyHook(vh *route.VirtualHost, extensionResources []*unstructured.Unstructured) (*route.VirtualHost, error) {
	// Take all of the unstructured resources for the extension and package them into bytes
	extensionResourceBytes, err := translateUnstructuredToUnstructuredBytes(extensionResources)
	if err != nil {
		return vh, err
	}
	// Make the request to the extension server
	ctx := context.Background()
	resp, err := h.grpcClient.PostVirtualHostModify(ctx,
		&extension.PostVirtualHostModifyRequest{
			VirtualHost: vh,
			PostVirtualHostContext: &extension.PostVirtualHostExtensionContext{
				BackendExtensionResources: extensionResourceBytes,
			},
		})
	if err != nil {
		return nil, err
	}

	return resp.VirtualHost, nil
}

func (h *XDSHook) PostHTTPListenerModifyHook(
	l *listener.Listener,
	extensionResources []*unstructured.Unstructured,
	backendExtensionResources []*unstructured.Unstructured,
) (*listener.Listener, error) {
	// Take all of the unstructured resources for the extension and package them into bytes
	extensionResourceBytes, err := translateUnstructuredToUnstructuredBytes(extensionResources)
	if err != nil {
		return l, err
	}
	// Take all of the unstructured resources for the extension and package them into bytes
	backendExtensionResourceBytes, err := translateUnstructuredToUnstructuredBytes(backendExtensionResources)
	if err != nil {
		return l, err
	}
	// Make the request to the extension server
	ctx := context.Background()
	resp, err := h.grpcClient.PostHTTPListenerModify(ctx,
		&extension.PostHTTPListenerModifyRequest{
			Listener: l,
			PostListenerContext: &extension.PostHTTPListenerExtensionContext{
				ExtensionResources:        extensionResourceBytes,
				BackendExtensionResources: backendExtensionResourceBytes,
			},
		})
	if err != nil {
		return nil, err
	}

	return resp.Listener, nil
}

func (h *XDSHook) PostTranslateModifyHook(clusters []*cluster.Cluster, secrets []*tls.Secret, extensionPolicies []*ir.UnstructuredRef) ([]*cluster.Cluster, []*tls.Secret, error) {
	// Make the request to the extension server
	// Take all of the unstructured resources for the extension and package them into bytes
	unstructuredPolicies := make([]*unstructured.Unstructured, len(extensionPolicies))
	for i, policy := range extensionPolicies {
		unstructuredPolicies[i] = policy.Object
	}
	// Convert the unstructured policies to bytes
	extensionPoliciesBytes, err := translateUnstructuredToUnstructuredBytes(unstructuredPolicies)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()
	resp, err := h.grpcClient.PostTranslateModify(ctx,
		&extension.PostTranslateModifyRequest{
			PostTranslateContext: &extension.PostTranslateExtensionContext{
				ExtensionResources: extensionPoliciesBytes,
			},
			Clusters: clusters,
			Secrets:  secrets,
		})
	if err != nil {
		return nil, nil, err
	}

	return resp.Clusters, resp.Secrets, nil
}
