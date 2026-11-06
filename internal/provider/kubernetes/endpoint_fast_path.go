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
	c, err := controller.NewTyped(name, mgr, controller.TypedOptions[ir.ServiceKey]{
		// One worker: RevisionTracker.Track serializes loads anyway, so more workers
		// would only wait on its lock.
		Reconciler:         reconcile.TypedFunc[ir.ServiceKey](r.reconcileBackendEndpoints),
		SkipNameValidation: skipNameValidation(),
		NewQueue:           newWorkqueue[ir.ServiceKey],
	})
	if err != nil {
		return err
	}
	// Map handlers process both old and new objects on reassociation. Use the same
	// current resource references as the full controller to select relevant backends.
	return c.Watch(source.TypedKind(mgr.GetCache(), &discoveryv1.EndpointSlice{},
		handler.TypedEnqueueRequestsFromMapFunc(func(_ context.Context, slice *discoveryv1.EndpointSlice) []ir.ServiceKey {
			if r.namespaceLabel != nil && !r.hasMatchingNamespaceLabels(slice) {
				return nil
			}
			var keys []ir.ServiceKey
			for _, backend := range [...]struct{ kind, label string }{
				{resource.KindService, discoveryv1.LabelServiceName},
				{resource.KindServiceImport, mcsapiv1a1.LabelServiceName},
			} {
				if name := slice.Labels[backend.label]; name != "" {
					key := ir.ServiceKey{Kind: backend.kind, Namespace: slice.Namespace, Name: name}
					if r.isFastPathBackend(key, slice.Labels) {
						keys = append(keys, key)
					}
				}
			}
			return keys
		})))
}

// isFastPathBackend reports whether a backend's endpoints are used in the xDS
// output: it is referenced by a route or policy, or it is a Gateway's own Envoy
// Service, which feeds the proxy's local cluster. The labels are those of the
// Service or one of its EndpointSlices, which copy the Service's labels.
func (r *gatewayAPIReconciler) isFastPathBackend(key ir.ServiceKey, labels map[string]string) bool {
	if r.isEndpointSliceBackendReferenced(&types.NamespacedName{Namespace: key.Namespace, Name: key.Name}) {
		return true
	}
	return key.Kind == resource.KindService && r.isProxyServiceCluster(labels)
}

func (r *gatewayAPIReconciler) reconcileBackendEndpoints(ctx context.Context, key ir.ServiceKey) (reconcile.Result, error) {
	// References may have disappeared while this key was queued. The full path
	// will update any snapshot that still uses the backend.
	var labels map[string]string
	if key.Kind == resource.KindService {
		service := &corev1.Service{}
		if err := r.client.Get(ctx, types.NamespacedName{Namespace: key.Namespace, Name: key.Name}, service); err == nil {
			labels = service.Labels
		} else if !kerrors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
	}
	if !r.isFastPathBackend(key, labels) {
		return reconcile.Result{}, nil
	}
	backendEndpoints, err := r.refreshBackendEndpoints(ctx, key)
	if err != nil {
		return reconcile.Result{}, err
	}
	// Queue the update even if Track found no change. Full reconciliation may
	// have stored this revision first, and its result may not be published yet,
	// so without this the change would wait on full translation. If the full
	// result publishes first, the publisher skips this revision as already applied.
	r.resources.EndpointUpdates.Add(backendEndpoints)
	return reconcile.Result{}, nil
}

// refreshBackendEndpoints reloads one backend's EndpointSlices from Kubernetes
// through the revision tracker, which assigns a new revision only when the
// endpoints changed. Both the endpoint controller and full reconciliation call
// it, and the tracker keeps their loads of a backend in order.
func (r *gatewayAPIReconciler) refreshBackendEndpoints(ctx context.Context, key ir.ServiceKey) (*endpoint.BackendEndpoints, error) {
	return r.resources.EndpointRevisions.Track(key, func() ([]*discoveryv1.EndpointSlice, error) {
		return r.listBackendEndpointSlices(ctx, key)
	})
}

// listBackendEndpointSlices reads a backend's EndpointSlices from the informer cache.
func (r *gatewayAPIReconciler) listBackendEndpointSlices(ctx context.Context, key ir.ServiceKey) ([]*discoveryv1.EndpointSlice, error) {
	var label, index string
	switch key.Kind {
	case resource.KindService:
		label, index = discoveryv1.LabelServiceName, serviceEndpointSliceIndex
	case resource.KindServiceImport:
		label, index = mcsapiv1a1.LabelServiceName, serviceImportEndpointSliceIndex
	default:
		return nil, fmt.Errorf("unsupported endpoint backend kind %q", key.Kind)
	}
	opts := []client.ListOption{client.InNamespace(key.Namespace)}
	if r.endpointSliceIndexEnabled() {
		opts = append(opts, client.MatchingFields{index: key.Name})
	} else {
		opts = append(opts, client.MatchingLabels{label: key.Name})
	}
	var list discoveryv1.EndpointSliceList
	if err := r.client.List(ctx, &list, opts...); err != nil {
		return nil, err
	}
	slices := make([]*discoveryv1.EndpointSlice, len(list.Items))
	for i := range list.Items {
		slices[i] = &list.Items[i]
	}
	return slices, nil
}

// populateBackendEndpoints loads the endpoints of every Service and ServiceImport
// in the resource trees and fills in each tree's BackendEndpoints and EndpointSlices.
// Trees that share a backend get the same endpoints and revision.
// It also drops stored endpoints for backends no longer referenced.
func (r *gatewayAPIReconciler) populateBackendEndpoints(ctx context.Context, resources resource.ControllerResources) error {
	var keys []ir.ServiceKey
	for _, tree := range resources {
		for _, service := range tree.Services {
			keys = append(keys, ir.ServiceKey{Kind: resource.KindService, Namespace: service.Namespace, Name: service.Name})
		}
		for _, service := range tree.ServiceImports {
			keys = append(keys, ir.ServiceKey{Kind: resource.KindServiceImport, Namespace: service.Namespace, Name: service.Name})
		}
	}

	// Remove any backends that are no longer referenced by the current resource trees from the endpoint source.
	// This ensures that the source only retains endpoints for backends that are still relevant.
	r.resources.EndpointRevisions.PruneBackends(keys)

	// Load each backend's endpoints through the revision tracker, which gives
	// them a revision, and capture the result.
	captured := make(map[ir.ServiceKey]*endpoint.BackendEndpoints, len(keys))
	for _, key := range keys {
		if _, found := captured[key]; found {
			continue
		}
		backendEndpoints, err := r.refreshBackendEndpoints(ctx, key)
		if err != nil {
			return err
		}
		captured[key] = backendEndpoints
	}

	// Rebuild each tree's EndpointSlices from the endpoints captured above, so
	// BackendEndpoints and EndpointSlices always come from the same revision.
	for _, tree := range resources {
		tree.EndpointSlices = nil
		seen := make(map[types.NamespacedName]struct{})
		add := func(key ir.ServiceKey) {
			backendEndpoints := captured[key]
			if backendEndpoints == nil {
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
			add(ir.ServiceKey{Kind: resource.KindService, Namespace: service.Namespace, Name: service.Name})
		}
		for _, service := range tree.ServiceImports {
			add(ir.ServiceKey{Kind: resource.KindServiceImport, Namespace: service.Namespace, Name: service.Name})
		}
		slices.SortFunc(tree.BackendEndpoints, func(a, b endpoint.BackendEndpoints) int {
			return strings.Compare(a.Backend.Kind+"/"+a.Backend.Namespace+"/"+a.Backend.Name,
				b.Backend.Kind+"/"+b.Backend.Namespace+"/"+b.Backend.Name)
		})
	}
	return nil
}
