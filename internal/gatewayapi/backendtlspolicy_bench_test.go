// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/logging"
)

// backendTLSBenchResources returns a Gateway with n HTTPRoutes, each routing to its own Service
// on a named port, and n BackendTLSPolicies, each targeting one of those Services. Half the
// policies target the named port (sectionName) and half the whole Service, so both lookups run.
// envoySecrets are the base test secrets (see benchBaseSecrets).
func backendTLSBenchResources(n int, envoySecrets []*corev1.Secret) *resource.Resources {
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
		svc := fmt.Sprintf("svc-%d", i)
		res.Services = append(res.Services, &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: svc},
			Spec: corev1.ServiceSpec{
				ClusterIP: "1.1.1.1",
				Ports: []corev1.ServicePort{{
					Name:       "https",
					Port:       8443,
					TargetPort: intstr.FromInt32(8443),
					Protocol:   corev1.ProtocolTCP,
				}},
			},
		})
		res.EndpointSlices = append(res.EndpointSlices, &discoveryv1.EndpointSlice{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      svc,
				Labels:    map[string]string{discoveryv1.LabelServiceName: svc},
			},
			AddressType: discoveryv1.AddressTypeIPv4,
			Ports: []discoveryv1.EndpointPort{{
				Name:     new("https"),
				Port:     new(int32(8443)),
				Protocol: new(corev1.ProtocolTCP),
			}},
			Endpoints: []discoveryv1.Endpoint{{Addresses: []string{"7.7.7.7"}}},
		})
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
					BackendRefs: []gwapiv1.HTTPBackendRef{{
						BackendRef: gwapiv1.BackendRef{
							BackendObjectReference: gwapiv1.BackendObjectReference{
								Name: gwapiv1.ObjectName(svc),
								Port: new(gwapiv1.PortNumber(8443)),
							},
						},
					}},
				}},
			},
		})

		target := gwapiv1.LocalPolicyTargetReferenceWithSectionName{
			LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
				Kind: resource.KindService,
				Name: gwapiv1.ObjectName(svc),
			},
		}
		if i%2 == 0 {
			target.SectionName = new(gwapiv1.SectionName("https"))
		}
		res.BackendTLSPolicies = append(res.BackendTLSPolicies, &gwapiv1.BackendTLSPolicy{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: fmt.Sprintf("btlsp-%d", i)},
			Spec: gwapiv1.BackendTLSPolicySpec{
				TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{target},
				Validation: gwapiv1.BackendTLSPolicyValidation{
					Hostname:                gwapiv1.PreciseHostname(svc + ".example.com"),
					WellKnownCACertificates: new(gwapiv1.WellKnownCACertificatesSystem),
				},
			},
		})
	}
	return res
}

func BenchmarkTranslateBackendTLSPolicies(b *testing.B) {
	secrets := benchBaseSecrets(b)

	for _, n := range []int{100, 1000} {
		b.Run(fmt.Sprintf("routes-and-policies-%d", n), func(b *testing.B) {
			newTranslator := func() *Translator {
				return &Translator{
					GatewayControllerName: egv1a1.GatewayControllerName,
					GatewayClassName:      "envoy-gateway-class",
					ControllerNamespace:   "envoy-gateway-system",
					Logger:                logging.DefaultLogger(nil, egv1a1.LogLevelError),
				}
			}

			// Check once that every route picked up its BackendTLSPolicy, so the benchmark
			// measures the policy lookup rather than routes that never reach it.
			result, err := newTranslator().Translate(b.Context(), backendTLSBenchResources(n, secrets))
			if err != nil {
				b.Fatal(err)
			}
			withTLS := 0
			for _, x := range result.XdsIR {
				for _, l := range x.HTTP {
					for _, r := range l.Routes {
						if r.Destination != nil && len(r.Destination.Settings) > 0 && r.Destination.Settings[0].TLS != nil {
							withTLS++
						}
					}
				}
			}
			if withTLS != n {
				b.Fatalf("expected %d routes with backend TLS, got %d", n, withTLS)
			}

			b.ReportAllocs()
			for b.Loop() {
				b.StopTimer()
				resources := backendTLSBenchResources(n, secrets)
				t := newTranslator()
				b.StartTimer()
				if _, err := t.Translate(b.Context(), resources); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
