// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Extension.go contains functions to encapsulate all of the logic in handling interacting with
// Extensions for Envoy Gateway when performing xDS translation

package translator

import (
	"errors"
	"fmt"
	"reflect"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	resourceTypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	extensionTypes "github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func processExtensionPostRouteHook(route *routev3.Route, vHost *routev3.VirtualHost, irRoute *ir.HTTPRoute, em *extensionTypes.Manager) error {
	// Do nothing unless there is an extension manager and the ir.HTTPRoute has extension filters
	if em == nil || len(irRoute.ExtensionRefs) == 0 {
		return nil
	}

	// Check if an extension want to modify the route that was just configured/created
	extManager := *em
	extRouteHookClient := extManager.GetPostXDSHookClient(v1alpha1.XDSRoute)
	if extRouteHookClient == nil {
		return nil
	}
	unstructuredResources := make([]*unstructured.Unstructured, len(irRoute.ExtensionRefs))
	for refIdx, ref := range irRoute.ExtensionRefs {
		unstructuredResources[refIdx] = ref.Object
	}
	modifiedRoute, err := extRouteHookClient.PostRouteModifyHook(
		route,
		vHost.Domains,
		unstructuredResources,
	)
	if err != nil {
		// Maybe logging the error is better here, but this only happens when an extension is in-use
		// so if modification fails then we should probably treat that as a serious problem.
		return err
	}

	// If the extension returned a modified Route, then copy its to the one that was passed in as a reference
	if modifiedRoute != nil {
		if err = deepCopyPtr(modifiedRoute, route); err != nil {
			return err
		}
	}
	return nil
}

func processExtensionPostVHostHook(vHost *routev3.VirtualHost, em *extensionTypes.Manager) error {
	// Do nothing unless there is an extension manager
	if em == nil {
		return nil
	}

	// Check if an extension want to modify the route that was just configured/created
	extManager := *em
	extVHHookClient := extManager.GetPostXDSHookClient(v1alpha1.XDSVirtualHost)
	if extVHHookClient == nil {
		return nil
	}
	modifiedVH, err := extVHHookClient.PostVirtualHostModifyHook(vHost)
	if err != nil {
		// Maybe logging the error is better here, but this only happens when an extension is in-use
		// so if modification fails then we should probably treat that as a serious problem.
		return err
	}

	// If the extension returned a modified Virtual Host, then copy its to the one that was passed in as a reference
	if modifiedVH != nil {
		if err = deepCopyPtr(modifiedVH, vHost); err != nil {
			return err
		}
	}

	return nil
}

func processExtensionPostListenerHook(tCtx *types.ResourceVersionTable, xdsListener *listenerv3.Listener, em *extensionTypes.Manager) error {
	// Do nothing unless there is an extension manager
	if em == nil {
		return nil
	}

	// Check if an extension want to modify the listener that was just configured/created
	extManager := *em
	extListenerHookClient := extManager.GetPostXDSHookClient(v1alpha1.XDSHTTPListener)
	if extListenerHookClient != nil {
		modifiedListener, err := extListenerHookClient.PostHTTPListenerModifyHook(xdsListener)
		if err != nil {
			return err
		} else if modifiedListener != nil {
			// Use the resource table to update the listener with the modified version returned by the extension
			// We're assuming that Listener names are unique.
			if err := tCtx.AddOrReplaceXdsResource(resourcev3.ListenerType, modifiedListener, func(existing resourceTypes.Resource, new resourceTypes.Resource) bool {
				oldListener := existing.(*listenerv3.Listener)
				newListener := new.(*listenerv3.Listener)
				if newListener == nil || oldListener == nil {
					return false
				}
				if oldListener.Name == newListener.Name {
					return true
				}
				return false
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func processExtensionPostTranslationHook(tCtx *types.ResourceVersionTable, em *extensionTypes.Manager) error {
	// Do nothing unless there is an extension manager
	if em == nil {
		return nil
	}
	// If there is a loaded extension that wants to inject clusters/secrets, then call it
	// while clusters can by statically added with bootstrap configuration, an extension may need to add clusters with a configuration
	// that is non-static. If a cluster definition is unlikely to change over the course of an extension's lifetime then the custom bootstrap config
	// is the preferred way of adding it.
	extManager := *em
	extensionInsertHookClient := extManager.GetPostXDSHookClient(v1alpha1.XDSTranslation)
	if extensionInsertHookClient == nil {
		return nil
	}

	clusters := tCtx.XdsResources[resourcev3.ClusterType]
	oldClusters := make([]*clusterv3.Cluster, len(clusters))
	for idx, cluster := range clusters {
		oldClusters[idx] = cluster.(*clusterv3.Cluster)
	}

	secrets := tCtx.XdsResources[resourcev3.SecretType]
	oldSecrets := make([]*tlsv3.Secret, len(secrets))
	for idx, secret := range secrets {
		oldSecrets[idx] = secret.(*tlsv3.Secret)
	}

	newClusters, newSecrets, err := extensionInsertHookClient.PostTranslateModifyHook(oldClusters, oldSecrets)
	if err != nil {
		return err
	}

	clusterResources := make([]resourceTypes.Resource, len(newClusters))
	for idx, cluster := range newClusters {
		clusterResources[idx] = cluster
	}
	tCtx.SetResources(resourcev3.ClusterType, clusterResources)

	secretResources := make([]resourceTypes.Resource, len(newSecrets))
	for idx, secret := range newSecrets {
		secretResources[idx] = secret
	}

	tCtx.SetResources(resourcev3.SecretType, secretResources)

	return nil
}

func deepCopyPtr(src interface{}, dest interface{}) error {
	if src == nil || dest == nil {
		return errors.New("cannot deep copy nil pointer")
	}
	srcVal := reflect.ValueOf(src)
	destVal := reflect.ValueOf(src)
	if srcVal.Kind() == reflect.Ptr && destVal.Kind() == reflect.Ptr {
		srcElem := srcVal.Elem()
		destVal = reflect.New(srcElem.Type())
		destElem := destVal.Elem()
		destElem.Set(srcElem)
		reflect.ValueOf(dest).Elem().Set(destVal.Elem())
	} else {
		return fmt.Errorf("cannot deep copy pointers to different kinds src %v != dest %v",
			srcVal.Kind(),
			destVal.Kind(),
		)
	}
	return nil
}
