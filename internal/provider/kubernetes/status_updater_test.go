// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
)

// laggingCache serves reads from snapshots the test takes by hand, standing in for an informer
// cache that has not seen the latest writes yet. Objects without a snapshot are read from the
// embedded client, which plays the API server and takes the writes.
type laggingCache struct {
	client.Client
	snapshots map[string]client.Object
}

func newLaggingCache(api client.Client) *laggingCache {
	return &laggingCache{Client: api, snapshots: map[string]client.Object{}}
}

func snapshotKey(obj client.Object, key client.ObjectKey) string {
	return fmt.Sprintf("%T %s %s", obj, obj.GetObjectKind().GroupVersionKind().GroupKind(), key)
}

func (c *laggingCache) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	snapshot, ok := c.snapshots[snapshotKey(obj, key)]
	if !ok {
		return c.Client.Get(ctx, key, obj, opts...)
	}
	reflect.ValueOf(obj).Elem().Set(reflect.ValueOf(snapshot.DeepCopyObject()).Elem())
	return nil
}

// catchUp lets the cache see the current state of the object; obj is an empty object of its type.
func (c *laggingCache) catchUp(t *testing.T, obj client.Object, key client.ObjectKey) {
	t.Helper()
	snapshot := snapshotKey(obj, key)
	require.NoError(t, c.Client.Get(t.Context(), key, obj))
	c.snapshots[snapshot] = obj
}

// gatewayClassStatusWriter writes the status of one GatewayClass through an UpdateHandler whose
// cache only catches up when the test says so.
type gatewayClassStatusWriter struct {
	t     *testing.T
	key   types.NamespacedName
	api   client.Client
	cache *laggingCache
	u     *UpdateHandler
}

func newGatewayClassStatusWriter(t *testing.T, funcs *interceptor.Funcs) *gatewayClassStatusWriter {
	key := types.NamespacedName{Name: "gc"}
	api := fakeclient.NewClientBuilder().
		WithScheme(envoygateway.GetScheme()).
		WithObjects(&gwapiv1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{Name: key.Name, Generation: 2},
			Spec:       gwapiv1.GatewayClassSpec{ControllerName: egv1a1.GatewayControllerName},
		}).
		WithStatusSubresource(&gwapiv1.GatewayClass{}).
		WithInterceptorFuncs(*funcs).
		Build()
	cache := newLaggingCache(api)
	return &gatewayClassStatusWriter{t: t, key: key, api: api, cache: cache, u: NewUpdateHandler(logr.Discard(), cache, api)}
}

func (w *gatewayClassStatusWriter) write(observedGeneration int64) {
	w.u.apply(Update{
		NamespacedName: w.key,
		Resource:       new(gwapiv1.GatewayClass),
		Mutator: MutatorFunc(func(obj client.Object) client.Object {
			gc := obj.(*gwapiv1.GatewayClass).DeepCopy()
			gc.Status.Conditions = []metav1.Condition{{
				Type:               string(gwapiv1.GatewayClassConditionStatusAccepted),
				Status:             metav1.ConditionTrue,
				Reason:             string(gwapiv1.GatewayClassReasonAccepted),
				ObservedGeneration: observedGeneration,
			}}
			return gc
		}),
	})
}

func (w *gatewayClassStatusWriter) storedObservedGeneration() int64 {
	w.t.Helper()
	gc := &gwapiv1.GatewayClass{}
	require.NoError(w.t, w.api.Get(w.t.Context(), w.key, gc))
	return gc.Status.Conditions[0].ObservedGeneration
}

// writeOutOfOrder writes generation 2, lets the cache catch up, and then writes a late update
// from generation 1 that the cache does not see.
func (w *gatewayClassStatusWriter) writeOutOfOrder() {
	w.t.Helper()
	w.write(2)
	w.cache.catchUp(w.t, new(gwapiv1.GatewayClass), w.key)
	w.write(1)
	require.Equal(w.t, int64(1), w.storedObservedGeneration())
}

func TestStatusUpdateIsNotSkippedWhileCacheLagsBehindOwnWrite(t *testing.T) {
	w := newGatewayClassStatusWriter(t, &interceptor.Funcs{})
	w.writeOutOfOrder()

	// Compared with the cache, which still shows the first write, this status looks unchanged.
	w.write(2)
	require.Equal(t, int64(2), w.storedObservedGeneration())
}

func TestStatusUpdateRetriedAfterConflictIsNotSkippedWhileCacheLags(t *testing.T) {
	conflictOnce := false
	w := newGatewayClassStatusWriter(t, &interceptor.Funcs{
		SubResourceUpdate: func(ctx context.Context, c client.Client, subResource string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
			if conflictOnce {
				conflictOnce = false
				return kerrors.NewConflict(schema.GroupResource{Resource: "gatewayclasses"}, obj.GetName(), fmt.Errorf("concurrent write"))
			}
			return c.SubResource(subResource).Update(ctx, obj, opts...)
		},
	})
	w.writeOutOfOrder()

	// The write that follows the check against the API server conflicts once. The retry still
	// meets the stale cache.
	conflictOnce = true
	w.write(2)
	require.Equal(t, int64(2), w.storedObservedGeneration())
}

func TestStatusUpdateRecordsOfSameKindInDifferentGroupsAreKeptApart(t *testing.T) {
	policyA := schema.GroupVersionKind{Group: "a.example.com", Version: "v1", Kind: "Policy"}
	policyB := schema.GroupVersionKind{Group: "b.example.com", Version: "v1", Kind: "Policy"}
	key := types.NamespacedName{Namespace: "ns", Name: "policy"}
	policy := func(gvk schema.GroupVersionKind) *unstructured.Unstructured {
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(gvk)
		obj.SetNamespace(key.Namespace)
		obj.SetName(key.Name)
		return obj
	}
	mapper := meta.NewDefaultRESTMapper(nil)
	mapper.Add(policyA, meta.RESTScopeNamespace)
	mapper.Add(policyB, meta.RESTScopeNamespace)
	api := fakeclient.NewClientBuilder().
		WithRESTMapper(mapper).
		WithObjects(policy(policyA), policy(policyB)).
		WithStatusSubresource(policy(policyA), policy(policyB)).
		Build()
	cache := newLaggingCache(api)
	u := NewUpdateHandler(logr.Discard(), cache, api)
	write := func(gvk schema.GroupVersionKind, observedGeneration int64) {
		u.apply(Update{
			NamespacedName: key,
			Resource:       policy(gvk),
			Mutator: MutatorFunc(func(obj client.Object) client.Object {
				p := obj.(*unstructured.Unstructured).DeepCopy()
				p.Object["status"] = map[string]any{"observedGeneration": observedGeneration}
				return p
			}),
		})
	}
	storedObservedGeneration := func(gvk schema.GroupVersionKind) int64 {
		obj := policy(gvk)
		require.NoError(t, api.Get(t.Context(), key, obj))
		gen, _, err := unstructured.NestedInt64(obj.Object, "status", "observedGeneration")
		require.NoError(t, err)
		return gen
	}

	write(policyA, 2)
	cache.catchUp(t, policy(policyA), key)
	write(policyA, 1)
	require.Equal(t, int64(1), storedObservedGeneration(policyA))

	// A write for the other Policy, and an update that changes nothing once its cache has caught up.
	write(policyB, 7)
	cache.catchUp(t, policy(policyB), key)
	write(policyB, 7)

	write(policyA, 2)
	require.Equal(t, int64(2), storedObservedGeneration(policyA))
}
