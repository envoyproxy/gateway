// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package types

import (
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type XDSHookClient interface {
	// PostRouteHook provides a way for extensions to modify a route generated by Envoy Gateway before it is finalized.
	// Doing so allows extensions to configure/modify route fields configured by Envoy Gateway and also to configure the
	// Route's TypedPerFilterConfig which may be desirable to do things such as pass settings and information to
	// ext_authz filters.
	// RouteHook also passes a list of Unstructured data for the externalRefs owned by the extension on the HTTPRoute that
	// created this xDS route
	// RouteHook will only be executed if an extension is loaded and only on Routes which were generated from an HTTPRoute
	// that uses extension resources as externalRef filters.
	PostRouteModifyHook(route *route.Route, routeHostnames []string, extensionResources []*unstructured.Unstructured) (*route.Route, error)

	// PostVirtualHostHook provides a way for extensions to modify a VirtualHost generated by Envoy Gateway before it is finalized.
	// An extension can also make use of this hook to generate and insert entirely new Routes not generated by Envoy Gateway.
	// VirtualHostHook is always executed when an extension is loaded. An extension may return an unmodified version of the VirtualHost it
	// received in order to not make any changes to it, or return nil to cause the VirtualHost to be discarded.
	PostVirtualHostModifyHook(*route.VirtualHost) (*route.VirtualHost, error)

	// PostHTTPListenerHook allows an extension to make changes to a Listener generated by Envoy Gateway before it is finalized.
	// PostHTTPListenerHook is always executed when an extension is loaded. An extension may return an unmodified version of the Listener it
	// received in order to not make any changes to it, or return nil to cause the Listener to be discarded.
	PostHTTPListenerModifyHook(*listener.Listener) (*listener.Listener, error)

	// PostTranslateModifyHook allows an extension to modify the clusters and secrets in the xDS config.
	// This allows for inserting clusters that may change along with extension specific configuration to be dynamically created rather than
	// using custom bootstrap config which would be sufficient for clusters that are static and not prone to have their configurations changed.
	// An example of how this may be used is to inject a cluster that will be used by an ext_authz http filter created by the extension.
	// The list of clusters and secrets returned by the extension are used as the final list of all clusters and secrets
	// PostTranslateModifyHook is always executed when an extension is loaded
	PostTranslateModifyHook([]*cluster.Cluster, []*tls.Secret) ([]*cluster.Cluster, []*tls.Secret, error)
}
