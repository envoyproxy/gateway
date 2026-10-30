// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
)

// laggingCache serves GatewayClass reads from a snapshot the test takes by hand, standing in for an
// informer cache that has not seen the latest writes yet. Until the first snapshot, reads go to the
// embedded client, which plays the API server and takes the writes.
type laggingCache struct {
	client.Client
	snapshot *gwapiv1.GatewayClass
}

func (c *laggingCache) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if gc, ok := obj.(*gwapiv1.GatewayClass); ok && c.snapshot != nil {
		c.snapshot.DeepCopyInto(gc)
		return nil
	}
	return c.Client.Get(ctx, key, obj, opts...)
}

// newGatewayClass sets a UID, which the API server would assign and the fake client does not.
func newGatewayClass(name string, generation int64) *gwapiv1.GatewayClass {
	return &gwapiv1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{Name: name, UID: types.UID(name), Generation: generation},
		Spec:       gwapiv1.GatewayClassSpec{ControllerName: egv1a1.GatewayControllerName},
	}
}

func acceptedGatewayClassUpdate(key types.NamespacedName, observedGeneration int64) Update {
	return Update{
		NamespacedName: key,
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
	}
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
		WithObjects(newGatewayClass(key.Name, 2)).
		WithStatusSubresource(&gwapiv1.GatewayClass{}).
		WithInterceptorFuncs(*funcs).
		Build()
	cache := &laggingCache{Client: api}
	return &gatewayClassStatusWriter{t: t, key: key, api: api, cache: cache, u: NewUpdateHandler(logr.Discard(), cache, api)}
}

func (w *gatewayClassStatusWriter) write(observedGeneration int64) {
	w.u.apply(acceptedGatewayClassUpdate(w.key, observedGeneration))
}

// catchUp lets the cache see the current state of the GatewayClass.
func (w *gatewayClassStatusWriter) catchUp() {
	w.t.Helper()
	gc := &gwapiv1.GatewayClass{}
	require.NoError(w.t, w.api.Get(w.t.Context(), w.key, gc))
	w.cache.snapshot = gc
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
	w.catchUp()
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

func TestStatusUpdateRecordsExpire(t *testing.T) {
	builder := fakeclient.NewClientBuilder().
		WithScheme(envoygateway.GetScheme()).
		WithStatusSubresource(&gwapiv1.GatewayClass{})
	for i := range 100 {
		builder = builder.WithObjects(newGatewayClass(fmt.Sprintf("gc-%d", i), 1))
	}
	api := builder.Build()
	u := NewUpdateHandler(logr.Discard(), api, api)
	for i := range 100 {
		u.apply(acceptedGatewayClassUpdate(types.NamespacedName{Name: fmt.Sprintf("gc-%d", i)}, 1))
	}

	u.expireLastWrites(time.Now())
	require.Len(t, u.lastWritten, 100)

	u.expireLastWrites(time.Now().Add(lastWriteTTL))
	require.Empty(t, u.lastWritten)
}
