// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func enqueueClass(_ context.Context, _ client.Object) []reconcile.Request {
	return []reconcile.Request{{NamespacedName: types.NamespacedName{
		Name: "controller-name",
	}}}
}

func TestSources(t *testing.T) {
	testCases := []struct {
		name              string
		ctx               context.Context
		expectedAddresses []string
		handler           handler.EventHandler
		mapFunc           handler.MapFunc
		queue             workqueue.TypedRateLimitingInterface[reconcile.Request]
		expected          bool
		obj               client.Object
	}{
		{
			name:              "Queue size should increase by one after the condition event triggered",
			expectedAddresses: []string{},
			handler:           handler.EnqueueRequestsFromMapFunc(enqueueClass),
			queue:             workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[reconcile.Request]()),
			ctx:               context.Background(),
			obj:               &gwapiv1.GatewayClass{},
			expected:          true,
		},
		{
			name:              "Confirm object is required",
			expectedAddresses: []string{},
			handler:           handler.EnqueueRequestsFromMapFunc(enqueueClass),
			queue:             workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[reconcile.Request]()),
			ctx:               context.Background(),
			obj:               nil,
			expected:          false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cond := make(chan struct{})
			store := NewWatchAndReconcileSource(cond, tc.obj, tc.handler)
			err := store.Start(tc.ctx, tc.queue)
			if !tc.expected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				close(cond)
				require.Eventually(t, func() bool {
					return tc.queue.Len() == 1
				}, time.Second*3, time.Millisecond*20)
			}
		})
	}
}
