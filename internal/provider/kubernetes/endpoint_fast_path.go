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
	c, err := controller.NewTyped(name, mgr, controller.TypedOptions[endpoint.Key]{
		Reconciler: reconcile.TypedFunc[endpoint.Key](func(ctx context.Context, key endpoint.Key) (reconcile.Result, error) {
			if !r.resources.Endpoints.Tracked(key) {
				return reconcile.Result{}, nil
			}
			_, err := r.refreshEndpointSource(ctx, key)
			return reconcile.Result{}, err
		}),
		MaxConcurrentReconciles: 4,
		SkipNameValidation:      skipNameValidation(),
	})
	if err != nil {
		return err
	}
	// Map handlers process both old and new objects on reassociation. Relevance
	// includes published consumers, even when current Routes no longer refer to them.
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
					if r.resources.Endpoints.Tracked(key) {
						keys = append(keys, key)
					}
				}
			}
			return keys
		})))
}

func (r *gatewayAPIReconciler) refreshEndpointSource(ctx context.Context, key endpoint.Key) (*endpoint.Snapshot, error) {
	r.endpointQueueMu.RLock()
	queue := r.endpointQueue
	r.endpointQueueMu.RUnlock()
	if queue == nil {
		return nil, fmt.Errorf("full reconciliation queue is not ready")
	}
	snapshot, changed, err := r.resources.Endpoints.Refresh(key, func() (ir.EndpointBackend, []*discoveryv1.EndpointSlice, error) {
		return r.loadEndpointSource(ctx, key)
	})
	if err != nil {
		return nil, err
	}
	if changed {
		queue.Add(reconcile.Request{NamespacedName: types.NamespacedName{Name: string(r.classController)}})
		r.resources.Endpoints.Notify(snapshot)
	}
	return snapshot, nil
}

func (r *gatewayAPIReconciler) loadEndpointSource(ctx context.Context, key endpoint.Key) (ir.EndpointBackend, []*discoveryv1.EndpointSlice, error) {
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

func (r *gatewayAPIReconciler) captureEndpointSnapshots(ctx context.Context, resources resource.ControllerResources) error {
	var keys []endpoint.Key
	for _, tree := range resources {
		for _, service := range tree.Services {
			keys = append(keys, endpoint.Key{Kind: resource.KindService, Namespace: service.Namespace, Name: service.Name})
		}
		for _, service := range tree.ServiceImports {
			keys = append(keys, endpoint.Key{Kind: resource.KindServiceImport, Namespace: service.Namespace, Name: service.Name})
		}
	}
	r.resources.Endpoints.SetReferences("provider", keys)
	captured := make(map[endpoint.Key]*endpoint.Snapshot, len(keys))
	for _, key := range keys {
		if _, found := captured[key]; found {
			continue
		}
		snapshot, err := r.refreshEndpointSource(ctx, key)
		if err != nil {
			if kerrors.IsNotFound(err) {
				continue
			}
			return err
		}
		captured[key] = snapshot
	}
	for _, tree := range resources {
		tree.EndpointSlices = nil
		seen := make(map[types.NamespacedName]struct{})
		add := func(key endpoint.Key, uid types.UID) {
			snapshot := captured[key]
			if snapshot == nil || snapshot.Backend.UID != string(uid) {
				return
			}
			tree.EndpointSnapshots = append(tree.EndpointSnapshots, *snapshot)
			for _, slice := range snapshot.Slices {
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
		slices.SortFunc(tree.EndpointSnapshots, func(a, b endpoint.Snapshot) int {
			return strings.Compare(a.Backend.Kind+"/"+a.Backend.Namespace+"/"+a.Backend.Name,
				b.Backend.Kind+"/"+b.Backend.Namespace+"/"+b.Backend.Name)
		})
	}
	return nil
}
