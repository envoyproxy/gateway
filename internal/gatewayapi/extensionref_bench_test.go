// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/logging"
)

const (
	benchExtGroup       = "foo.example.io"
	benchExtFilterKind  = "Foo"
	benchExtBackendKind = "FooBackend"
)

func benchExtResource(kind, ns, name string) unstructured.Unstructured {
	u := unstructured.Unstructured{}
	u.SetAPIVersion(benchExtGroup + "/v1")
	u.SetKind(kind)
	u.SetNamespace(ns)
	u.SetName(name)
	return u
}

// extensionRefBenchResources returns a Gateway with n HTTPRoutes. Each route has an extensionRef
// filter and a custom backendRef, each pointing at its own extension resource, so both extension
// resource lookups run once per route against 2n resources. envoySecrets are the base test
// secrets (see benchBaseSecrets).
func extensionRefBenchResources(n int, envoySecrets []*corev1.Secret) *resource.Resources {
	const ns = "default"
	res := resource.NewResources()
	res.Secrets = append(res.Secrets, envoySecrets...)
	res.Namespaces = append(res.Namespaces, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	res.Gateways = append(res.Gateways, &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "gateway"},
		Spec: gwapiv1.GatewaySpec{
			GatewayClassName: "envoy-gateway-class",
			Listeners: []gwapiv1.Listener{{
				Name:     "http",
				Protocol: gwapiv1.HTTPProtocolType,
				Port:     80,
			}},
		},
	})

	for i := range n {
		filter := fmt.Sprintf("filter-%d", i)
		backend := fmt.Sprintf("backend-%d", i)
		res.ExtensionRefFilters = append(res.ExtensionRefFilters,
			benchExtResource(benchExtFilterKind, ns, filter),
			benchExtResource(benchExtBackendKind, ns, backend))
		res.HTTPRoutes = append(res.HTTPRoutes, &gwapiv1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: fmt.Sprintf("route-%d", i)},
			Spec: gwapiv1.HTTPRouteSpec{
				CommonRouteSpec: gwapiv1.CommonRouteSpec{
					ParentRefs: []gwapiv1.ParentReference{{Name: "gateway"}},
				},
				Rules: []gwapiv1.HTTPRouteRule{{
					Matches: []gwapiv1.HTTPRouteMatch{{
						Path: &gwapiv1.HTTPPathMatch{Value: new(fmt.Sprintf("/%d", i))},
					}},
					Filters: []gwapiv1.HTTPRouteFilter{{
						Type: gwapiv1.HTTPRouteFilterExtensionRef,
						ExtensionRef: &gwapiv1.LocalObjectReference{
							Group: benchExtGroup,
							Kind:  benchExtFilterKind,
							Name:  gwapiv1.ObjectName(filter),
						},
					}},
					BackendRefs: []gwapiv1.HTTPBackendRef{{
						BackendRef: gwapiv1.BackendRef{
							BackendObjectReference: gwapiv1.BackendObjectReference{
								Group: new(gwapiv1.Group(benchExtGroup)),
								Kind:  new(gwapiv1.Kind(benchExtBackendKind)),
								Name:  gwapiv1.ObjectName(backend),
							},
						},
					}},
				}},
			},
		})
	}
	return res
}

func BenchmarkTranslateExtensionRefs(b *testing.B) {
	secrets := benchBaseSecrets(b)
	for _, n := range []int{100, 1000} {
		b.Run(fmt.Sprintf("routes-%d", n), func(b *testing.B) {
			newTranslator := func() *Translator {
				return &Translator{
					GatewayControllerName: egv1a1.GatewayControllerName,
					GatewayClassName:      "envoy-gateway-class",
					ControllerNamespace:   "envoy-gateway-system",
					ExtensionGroupKinds: []schema.GroupKind{
						{Group: benchExtGroup, Kind: benchExtFilterKind},
						{Group: benchExtGroup, Kind: benchExtBackendKind},
					},
					Logger: logging.DefaultLogger(nil, egv1a1.LogLevelError),
				}
			}

			// Check once that every route resolved both of its extension resources, so the
			// benchmark measures the lookups rather than routes that fail before reaching them.
			result, err := newTranslator().Translate(b.Context(), extensionRefBenchResources(n, secrets))
			if err != nil {
				b.Fatal(err)
			}
			resolved := 0
			for _, x := range result.XdsIR {
				for _, l := range x.HTTP {
					for _, r := range l.Routes {
						if len(r.ExtensionRefs) == 2 {
							resolved++
						}
					}
				}
			}
			if resolved != n {
				b.Fatalf("expected %d routes with both extension resources, got %d", n, resolved)
			}

			b.ReportAllocs()
			for b.Loop() {
				b.StopTimer()
				resources := extensionRefBenchResources(n, secrets)
				t := newTranslator()
				b.StartTimer()
				if _, err := t.Translate(b.Context(), resources); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
