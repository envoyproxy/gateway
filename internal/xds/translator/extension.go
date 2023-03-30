// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Extension.go contains functions to encapsulate all of the logic in handling interacting with
// Extensions for Envoy Gateway when performing xDS translation

package translator

import (
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	resourceTypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"

	extensionTypes "github.com/envoyproxy/gateway/internal/extension/types"
)

func processExtensionPostTranslationHook(tCtx *types.ResourceVersionTable, em *extensionTypes.Manager) error {
	// Do nothing unless there is an extension manager
	if em != nil {
		// If there is a loaded extension that wants to inject clusters/secrets, then call it
		// while clusters can by statically added with bootstrap configuration, an extension may need to add clusters with a configuration
		// that is non-static. If a cluster definition is unlikely to change over the course of an extension's lifetime then the custom bootstrap config
		// is the preferred way of adding it.
		extManager := *em
		extensionInsertHookClient := extManager.GetXDSHookClient(extensionTypes.PostXDSTranslation)
		if extensionInsertHookClient != nil {
			newClusters, newSecrets, err := extensionInsertHookClient.PostTranslationInsertHook()
			if err != nil {
				return err
			}

			// We're assuming that Cluster names are unique.
			for _, addedCluster := range newClusters {
				tCtx.AddOrReplaceXdsResource(resourcev3.ClusterType, addedCluster, func(existing resourceTypes.Resource, new resourceTypes.Resource) bool {
					oldCluster := existing.(*clusterv3.Cluster)
					newCluster := new.(*clusterv3.Cluster)
					if newCluster == nil || oldCluster == nil {
						return false
					}
					if oldCluster.Name == newCluster.Name {
						return true
					}
					return false
				})
			}

			for _, secret := range newSecrets {
				tCtx.AddXdsResource(resourcev3.SecretType, secret)
			}
		}
	}
	return nil
}

func processExtensionPostRouteHook(route *routev3.Route, vHost *routev3.VirtualHost, irRoute *ir.HTTPRoute, client extensionTypes.XDSHookClient) error {
	unstructuredResources := make([]*unstructured.Unstructured, len(irRoute.ExtensionRefs))
	for refIdx, ref := range irRoute.ExtensionRefs {
		unstructuredResources[refIdx] = ref.Object
	}
	modifiedRoute, err := client.PostRouteModifyHook(
		route,
		vHost.Domains,
		unstructuredResources,
	)
	if err != nil {
		// Maybe logging the error is better here, but this only happens when an extension is in-use
		// so if modification fails then we should probably treat that as a serious problem.
		return err
	}

	// An extension is allowed to return a nil route to prevent it from being added
	if modifiedRoute != nil {
		vHost.Routes = append(vHost.Routes, modifiedRoute)
	}

	return nil
}

func processExtensionPostVHostHook(vHost *routev3.VirtualHost, routeConfig *routev3.RouteConfiguration, client extensionTypes.XDSHookClient) error {
	modifiedVH, err := client.PostVirtualHostModifyHook(vHost)
	if err != nil {
		// Maybe logging the error is better here, but this only happens when an extension is in-use
		// so if modification fails then we should probably treat that as a serious problem.
		return err
	}

	if modifiedVH != nil {
		routeConfig.VirtualHosts = append(routeConfig.VirtualHosts, vHost)
	}

	return nil
}

func processExtensionPostListenerHook(tCtx *types.ResourceVersionTable, xdsListener *listenerv3.Listener, em *extensionTypes.Manager) error {
	if em != nil {
		extManager := *em
		// Check if an extension want to modify the listener that was just configured/created
		extListenerHookClient := extManager.GetXDSHookClient(extensionTypes.PostXDSHTTPListener)
		if extListenerHookClient != nil {
			modifiedListener, err := extListenerHookClient.PostHTTPListenerModifyHook(xdsListener)
			if err != nil {
				return err
			} else if modifiedListener != nil {
				// Use the resource table to update the listener with the modified version returned by the extension
				// We're assuming that Listener names are unique.
				tCtx.AddOrReplaceXdsResource(resourcev3.ListenerType, modifiedListener, func(existing resourceTypes.Resource, new resourceTypes.Resource) bool {
					oldListener := existing.(*listenerv3.Listener)
					newListener := new.(*listenerv3.Listener)
					if newListener == nil || oldListener == nil {
						return false
					}
					if oldListener.Name == newListener.Name {
						return true
					}
					return false
				})

			}

		}
	}
	return nil
}
