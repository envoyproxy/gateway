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
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/envoygateway"
)

func TestUpdateHandlerApplyRetries(t *testing.T) {
	gvr := schema.GroupResource{Group: gwapiv1.GroupName, Resource: "gateways"}
	key := types.NamespacedName{Namespace: "default", Name: "gw"}

	testCases := []struct {
		name string
		// getErrs and updateErrs are returned, in order, by the first calls to Get and
		// Status().Update; later calls go through to the fake client.
		getErrs        []error
		updateErrs     []error
		wantUpdates    int
		wantStatusSet  bool
		cancelCtxFirst bool
	}{
		{
			name:          "server timeout on update is retried",
			updateErrs:    []error{kerrors.NewServerTimeout(gvr, "update", 1), kerrors.NewTimeoutError("timed out", 1)},
			wantUpdates:   3,
			wantStatusSet: true,
		},
		{
			name:          "throttling and unavailable are retried",
			updateErrs:    []error{kerrors.NewTooManyRequests("slow down", 1), kerrors.NewServiceUnavailable("unavailable")},
			wantUpdates:   3,
			wantStatusSet: true,
		},
		{
			name:          "deadline exceeded on get is retried",
			getErrs:       []error{context.DeadlineExceeded},
			wantUpdates:   1,
			wantStatusSet: true,
		},
		{
			name:          "conflict is still retried",
			updateErrs:    []error{kerrors.NewConflict(gvr, "gw", nil)},
			wantUpdates:   2,
			wantStatusSet: true,
		},
		{
			name:        "forbidden is not retried",
			updateErrs:  []error{kerrors.NewForbidden(gvr, "gw", nil)},
			wantUpdates: 1,
		},
		{
			name:        "transient errors give up after the backoff runs out",
			updateErrs:  repeatErr(kerrors.NewServerTimeout(gvr, "update", 1), 100),
			wantUpdates: transientRetryBackoff.Steps,
		},
		{
			name:           "no retry once the handler is shutting down",
			updateErrs:     []error{kerrors.NewServerTimeout(gvr, "update", 1)},
			wantUpdates:    1,
			cancelCtxFirst: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gw := &gwapiv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name}}
			var gets, updates int
			cli := fakeclient.NewClientBuilder().
				WithScheme(envoygateway.GetScheme()).
				WithObjects(gw).
				WithStatusSubresource(gw).
				WithInterceptorFuncs(interceptor.Funcs{
					Get: func(ctx context.Context, c client.WithWatch, k client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						gets++
						if gets <= len(tc.getErrs) {
							return tc.getErrs[gets-1]
						}
						return c.Get(ctx, k, obj, opts...)
					},
					SubResourceUpdate: func(ctx context.Context, c client.Client, sub string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
						updates++
						if updates <= len(tc.updateErrs) {
							return tc.updateErrs[updates-1]
						}
						return c.SubResource(sub).Update(ctx, obj, opts...)
					},
				}).
				Build()

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if tc.cancelCtxFirst {
				cancel()
			}

			u := NewUpdateHandler(logr.Discard(), cli, cli)
			u.apply(ctx, Update{
				NamespacedName: key,
				Resource:       new(gwapiv1.Gateway),
				Mutator: MutatorFunc(func(obj client.Object) client.Object {
					g := obj.(*gwapiv1.Gateway).DeepCopy()
					g.Status.Conditions = []metav1.Condition{{
						Type:               string(gwapiv1.GatewayConditionAccepted),
						Status:             metav1.ConditionTrue,
						Reason:             string(gwapiv1.GatewayReasonAccepted),
						LastTransitionTime: metav1.Now(),
					}}
					return g
				}),
			})

			require.Equal(t, tc.wantUpdates, updates)

			got := &gwapiv1.Gateway{}
			require.NoError(t, cli.Get(t.Context(), key, got))
			require.Equal(t, tc.wantStatusSet, len(got.Status.Conditions) == 1)
		})
	}
}

func repeatErr(err error, n int) []error {
	errs := make([]error, n)
	for i := range errs {
		errs[i] = err
	}
	return errs
}
