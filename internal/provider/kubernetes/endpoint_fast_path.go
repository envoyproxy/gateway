// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"slices"
	"strings"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	mcsapiv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/endpoint"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
)

func (r *gatewayAPIReconciler) endpointFastPathEnabled() bool {
	return r.envoyGateway != nil && r.envoyGateway.RuntimeFlags.IsEnabled(egv1a1.EndpointFastPath)
}

func (r *gatewayAPIReconciler) watchEndpointUpdates(mgr manager.Manager, name string) error {
	// TODO(robin): Add optional per-backend debouncing in a follow-up PR, with bounded
	// waiting, independent backend batches, and retries that bypass debouncing.
	c, err := controller.NewTyped(name, mgr, controller.TypedOptions[endpoint.Key]{
		// One worker: Source.Refresh serializes loads anyway, so more workers
		// would only wait on its lock.
		Reconciler:         reconcile.TypedFunc[endpoint.Key](r.reconcileBackendEndpoints),
		SkipNameValidation: skipNameValidation(),
	})
	if err != nil {
		return err
	}
	// Map handlers process both old and new objects on reassociation. Use the same
	// current resource references as the full controller to select relevant backends.
	return c.Watch(source.TypedKind(mgr.GetCache(), &discoveryv1.EndpointSlice{},
		handler.TypedEnqueueRequestsFromMapFunc(func(_ context.Context, slice *discoveryv1.EndpointSlice) []endpoint.Key {
			if r.namespaceLabel != nil && !r.hasMatchingNamespaceLabels(slice) {
				return nil
			}
			var keys []endpoint.Key
			for _, backend := range [...]struct{ kind, label string }{
				{resource.KindService, discoveryv1.LabelServiceName},
				{resource.KindServiceImport, mcsapiv1a1.LabelServiceName},
			} {
				if name := slice.Labels[backend.label]; name != "" {
					key := endpoint.Key{Kind: backend.kind, Namespace: slice.Namespace, Name: name}
					if r.isEndpointSliceBackendReferenced(&types.NamespacedName{Namespace: key.Namespace, Name: key.Name}) {
						keys = append(keys, key)
					}
				}
			}
			return keys
		})))
}

func (r *gatewayAPIReconciler) reconcileBackendEndpoints(ctx context.Context, key endpoint.Key) (reconcile.Result, error) {
	// References may have disappeared while this key was queued. The full path
	// will update any snapshot that still uses the backend.
	if !r.isEndpointSliceBackendReferenced(&types.NamespacedName{Namespace: key.Namespace, Name: key.Name}) {
		return reconcile.Result{}, nil
	}
	backendEndpoints, err := r.refreshBackendEndpoints(ctx, key)
	if err != nil {
		return reconcile.Result{}, err
	}
	// Full reconciliation may have already collected this revision without notifying.
	// Always notify here; the publisher skips revisions it has already applied.
	r.resources.EndpointUpdates.Notify(backendEndpoints)
	return reconcile.Result{}, nil
}

// refreshBackendEndpoints reloads one backend's EndpointSlices from Kubernetes and
// updates its entry in the shared endpoint source, which assigns a new revision
// only when the endpoints changed. Both the endpoint controller and full
// reconciliation call it, and the source keeps their loads of a backend in order.
func (r *gatewayAPIReconciler) refreshBackendEndpoints(ctx context.Context, key endpoint.Key) (*endpoint.BackendEndpoints, error) {
	return r.resources.EndpointSource.Refresh(key, func() (ir.EndpointBackend, []*discoveryv1.EndpointSlice, error) {
		return r.listBackendEndpointSlices(ctx, key)
	})
}

// listBackendEndpointSlices reads a backend's EndpointSlices from the informer cache.
func (r *gatewayAPIReconciler) listBackendEndpointSlices(ctx context.Context, key endpoint.Key) (ir.EndpointBackend, []*discoveryv1.EndpointSlice, error) {
	backend := ir.EndpointBackend{Kind: key.Kind, Namespace: key.Namespace, Name: key.Name}
	var object client.Object
	var label, index string
	switch key.Kind {
	case resource.KindService:
		object, label, index = &corev1.Service{}, discoveryv1.LabelServiceName, serviceEndpointSliceIndex
	case resource.KindServiceImport:
		object, label, index = &mcsapiv1a1.ServiceImport{}, mcsapiv1a1.LabelServiceName, serviceImportEndpointSliceIndex
	default:
		return backend, nil, fmt.Errorf("unsupported endpoint source kind %q", key.Kind)
	}
	nn := types.NamespacedName{Namespace: key.Namespace, Name: key.Name}
	if err := r.client.Get(ctx, nn, object); err != nil {
		// Backend deletion is handled by full reconciliation. Never interpret a
		// failed collection as an empty endpoint set for the old backend UID.
		return backend, nil, err
	}
	backend.UID = string(object.GetUID())
	opts := []client.ListOption{client.InNamespace(key.Namespace)}
	if r.endpointSliceIndexEnabled() {
		opts = append(opts, client.MatchingFields{index: key.Name})
	} else {
		opts = append(opts, client.MatchingLabels{label: key.Name})
	}
	var list discoveryv1.EndpointSliceList
	if err := r.client.List(ctx, &list, opts...); err != nil {
		return backend, nil, err
	}
	if err := r.client.Get(ctx, nn, object); err != nil {
		return backend, nil, err
	}
	if backend.UID != string(object.GetUID()) {
		return backend, nil, fmt.Errorf("backend %s/%s was recreated during endpoint collection", key.Namespace, key.Name)
	}
	slices := make([]*discoveryv1.EndpointSlice, len(list.Items))
	for i := range list.Items {
		slices[i] = &list.Items[i]
	}
	return backend, slices, nil
}

// populateBackendEndpoints loads the endpoints of every Service and ServiceImport
// in the resource trees and fills in each tree's BackendEndpoints and EndpointSlices.
// Trees that share a backend get the same endpoints and revision.
// It also drops stored endpoints for backends no longer referenced.
func (r *gatewayAPIReconciler) populateBackendEndpoints(ctx context.Context, resources resource.ControllerResources) error {
	var keys []endpoint.Key
	for _, tree := range resources {
		for _, service := range tree.Services {
			keys = append(keys, endpoint.Key{Kind: resource.KindService, Namespace: service.Namespace, Name: service.Name})
		}
		for _, service := range tree.ServiceImports {
			keys = append(keys, endpoint.Key{Kind: resource.KindServiceImport, Namespace: service.Namespace, Name: service.Name})
		}
	}

	// Remove any backends that are no longer referenced by the current resource trees from the endpoint source.
	// This ensures that the source only retains endpoints for backends that are still relevant.
	r.resources.EndpointSource.PruneBackends(keys)

	// Refresh each backend's endpoints in the source and capture the result.
	// The source will assign a revision to each backend's endpoints.
	captured := make(map[endpoint.Key]*endpoint.BackendEndpoints, len(keys))
	for _, key := range keys {
		if _, found := captured[key]; found {
			continue
		}
		backendEndpoints, err := r.refreshBackendEndpoints(ctx, key)
		if err != nil {
			if kerrors.IsNotFound(err) {
				continue
			}
			return err
		}
		captured[key] = backendEndpoints
	}

	// Rebuild each tree's EndpointSlices from the endpoints captured above, so
	// BackendEndpoints and EndpointSlices always come from the same revision.
	for _, tree := range resources {
		tree.EndpointSlices = nil
		seen := make(map[types.NamespacedName]struct{})
		add := func(key endpoint.Key, uid types.UID) {
			backendEndpoints := captured[key]
			if backendEndpoints == nil || backendEndpoints.Backend.UID != string(uid) {
				return
			}
			tree.BackendEndpoints = append(tree.BackendEndpoints, *backendEndpoints)
			for _, slice := range backendEndpoints.EndpointSlices {
				nn := types.NamespacedName{Namespace: slice.Namespace, Name: slice.Name}
				if _, found := seen[nn]; !found {
					seen[nn] = struct{}{}
					tree.EndpointSlices = append(tree.EndpointSlices, slice)
				}
			}
		}
		for _, service := range tree.Services {
			add(endpoint.Key{Kind: resource.KindService, Namespace: service.Namespace, Name: service.Name}, service.UID)
		}
		for _, service := range tree.ServiceImports {
			add(endpoint.Key{Kind: resource.KindServiceImport, Namespace: service.Namespace, Name: service.Name}, service.UID)
		}
		slices.SortFunc(tree.BackendEndpoints, func(a, b endpoint.BackendEndpoints) int {
			return strings.Compare(a.Backend.Kind+"/"+a.Backend.Namespace+"/"+a.Backend.Name,
				b.Backend.Kind+"/"+b.Backend.Namespace+"/"+b.Backend.Name)
		})
	}
	return nil
}
