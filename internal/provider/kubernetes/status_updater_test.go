// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
)

// laggingCache serves GatewayClass reads from a snapshot that the test advances by hand,
// standing in for an informer cache that has not seen the latest writes yet. Writes go
// straight to the embedded client, which plays the API server.
type laggingCache struct {
	client.Client
	snapshot *gwapiv1.GatewayClass
}

func (c *laggingCache) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if c.snapshot == nil {
		return c.Client.Get(ctx, key, obj, opts...)
	}
	c.snapshot.DeepCopyInto(obj.(*gwapiv1.GatewayClass))
	return nil
}

func (c *laggingCache) catchUp(t *testing.T, key types.NamespacedName) {
	t.Helper()
	c.snapshot = &gwapiv1.GatewayClass{}
	require.NoError(t, c.Client.Get(t.Context(), key, c.snapshot))
}

func TestStatusUpdateIsNotSkippedWhileCacheLagsBehindOwnWrite(t *testing.T) {
	key := types.NamespacedName{Name: "gc"}
	api := fakeclient.NewClientBuilder().
		WithScheme(envoygateway.GetScheme()).
		WithObjects(&gwapiv1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{Name: key.Name, Generation: 2},
			Spec:       gwapiv1.GatewayClassSpec{ControllerName: egv1a1.GatewayControllerName},
		}).
		WithStatusSubresource(&gwapiv1.GatewayClass{}).
		Build()
	cache := &laggingCache{Client: api}
	u := NewUpdateHandler(logr.Discard(), cache, api)

	write := func(observedGeneration int64) {
		u.apply(Update{
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
		})
	}
	storedObservedGeneration := func() int64 {
		gc := &gwapiv1.GatewayClass{}
		require.NoError(t, api.Get(t.Context(), key, gc))
		return gc.Status.Conditions[0].ObservedGeneration
	}

	write(2)
	cache.catchUp(t, key)

	// A late update from an older generation; the cache does not see this write.
	write(1)
	require.Equal(t, int64(1), storedObservedGeneration())

	// Compared with the cache, which still shows the first write, this status looks unchanged.
	write(2)
	require.Equal(t, int64(2), storedObservedGeneration())
}
