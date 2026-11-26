// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/logging"
)

func refGrant(namespace, name, fromKind, fromNamespace, toKind string, toName *string) *gwapiv1b1.ReferenceGrant {
	return &gwapiv1b1.ReferenceGrant{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: gwapiv1b1.ReferenceGrantSpec{
			From: []gwapiv1b1.ReferenceGrantFrom{{
				Group:     gwapiv1.GroupName,
				Kind:      gwapiv1.Kind(fromKind),
				Namespace: gwapiv1.Namespace(fromNamespace),
			}},
			To: []gwapiv1b1.ReferenceGrantTo{{
				Kind: gwapiv1.Kind(toKind),
				Name: (*gwapiv1.ObjectName)(toName),
			}},
		},
	}
}

func newRefGrantReconciler(t testing.TB, objs ...client.Object) *gatewayAPIReconciler {
	t.Helper()
	cli := fakeclient.NewClientBuilder().
		WithScheme(envoygateway.GetScheme()).
		WithObjects(objs...).
		WithIndex(&gwapiv1b1.ReferenceGrant{}, refGrantFromNamespaceToKindIndex, getReferenceGrantIndexerFunc).
		Build()
	return &gatewayAPIReconciler{client: cli, log: logging.DefaultLogger(nil, egv1a1.LogLevelInfo)}
}

func TestFindReferenceGrant(t *testing.T) {
	from := ObjectKindNamespacedName{kind: resource.KindHTTPRoute, namespace: "routes", name: "route"}
	to := ObjectKindNamespacedName{kind: resource.KindService, namespace: "backends", name: "svc"}

	tests := []struct {
		name   string
		grants []client.Object
		want   string // name of the expected grant, empty for none
	}{
		{
			name:   "grant in the target namespace",
			grants: []client.Object{refGrant("backends", "allow", resource.KindHTTPRoute, "routes", resource.KindService, nil)},
			want:   "allow",
		},
		{
			name:   "grant for the named target",
			grants: []client.Object{refGrant("backends", "allow-svc", resource.KindHTTPRoute, "routes", resource.KindService, new("svc"))},
			want:   "allow-svc",
		},
		{
			name:   "grant in another namespace does not apply",
			grants: []client.Object{refGrant("other", "allow", resource.KindHTTPRoute, "routes", resource.KindService, nil)},
		},
		{
			name:   "grant for another target name",
			grants: []client.Object{refGrant("backends", "allow-other", resource.KindHTTPRoute, "routes", resource.KindService, new("other"))},
		},
		{
			name:   "grant from another namespace",
			grants: []client.Object{refGrant("backends", "allow", resource.KindHTTPRoute, "elsewhere", resource.KindService, nil)},
		},
		{
			name:   "grant from another route kind",
			grants: []client.Object{refGrant("backends", "allow", resource.KindGRPCRoute, "routes", resource.KindService, nil)},
		},
		{
			name:   "grant for another kind",
			grants: []client.Object{refGrant("backends", "allow", resource.KindHTTPRoute, "routes", resource.KindSecret, nil)},
		},
		{
			name: "matching grant among grants in other namespaces",
			grants: []client.Object{
				refGrant("other", "allow-other-ns", resource.KindHTTPRoute, "routes", resource.KindService, nil),
				refGrant("backends", "allow", resource.KindHTTPRoute, "routes", resource.KindService, nil),
			},
			want: "allow",
		},
		{
			name: "matching grant among grants for other source namespaces in the target namespace",
			grants: []client.Object{
				refGrant("backends", "allow-a", resource.KindHTTPRoute, "a", resource.KindService, nil),
				refGrant("backends", "allow-b", resource.KindHTTPRoute, "b", resource.KindService, nil),
				refGrant("backends", "allow", resource.KindHTTPRoute, "routes", resource.KindService, nil),
			},
			want: "allow",
		},
		{
			name: "grant with several from and to entries",
			grants: []client.Object{&gwapiv1b1.ReferenceGrant{
				ObjectMeta: metav1.ObjectMeta{Namespace: "backends", Name: "allow-many"},
				Spec: gwapiv1b1.ReferenceGrantSpec{
					From: []gwapiv1b1.ReferenceGrantFrom{
						{Group: gwapiv1.GroupName, Kind: resource.KindGRPCRoute, Namespace: "elsewhere"},
						{Group: gwapiv1.GroupName, Kind: resource.KindHTTPRoute, Namespace: "routes"},
					},
					To: []gwapiv1b1.ReferenceGrantTo{{Kind: resource.KindSecret}, {Kind: resource.KindService}},
				},
			}},
			want: "allow-many",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newRefGrantReconciler(t, tc.grants...)
			got, err := r.findReferenceGrant(t.Context(), from, to)
			require.NoError(t, err)
			if tc.want == "" {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, "backends", got.Namespace)
			require.Equal(t, tc.want, got.Name)
		})
	}
}

// BenchmarkFindReferenceGrant looks up a grant in one namespace while many other namespaces
// also hold grants for the same kind.
func BenchmarkFindReferenceGrant(b *testing.B) {
	const namespaces = 500
	objs := make([]client.Object, 0, namespaces)
	for i := range namespaces {
		ns := fmt.Sprintf("ns-%d", i)
		objs = append(objs, refGrant(ns, "allow", resource.KindHTTPRoute, "routes", resource.KindService, nil))
	}
	r := newRefGrantReconciler(b, objs...)
	from := ObjectKindNamespacedName{kind: resource.KindHTTPRoute, namespace: "routes", name: "route"}
	to := ObjectKindNamespacedName{kind: resource.KindService, namespace: "ns-250", name: "svc"}

	b.ReportAllocs()
	for b.Loop() {
		got, err := r.findReferenceGrant(b.Context(), from, to)
		if err != nil || got == nil {
			b.Fatalf("expected a grant, got %v, %v", got, err)
		}
	}
}
