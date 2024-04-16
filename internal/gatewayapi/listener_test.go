// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestProcessTracing(t *testing.T) {
	cases := []struct {
		gw         gwapiv1.Gateway
		proxy      *egcfgv1a1.EnvoyProxy
		resources  Resources
		translator Translator
		expected   *ir.Tracing
	}{
		{},
		{
			gw: gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake-gw",
					Namespace: "fake-ns",
				},
			},
			proxy: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake-ep",
					Namespace: "fake-ep-ns",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Telemetry: &egcfgv1a1.ProxyTelemetry{
						Tracing: &egcfgv1a1.ProxyTracing{},
					},
				},
			},
			expected: &ir.Tracing{
				ServiceName:  "fake-gw.fake-ns",
				SamplingRate: 100.0,
				Destination: ir.RouteDestination{
					Name:     "tracing/fake-ep-ns/fake-ep",
					Settings: []*ir.DestinationSetting{},
				},
			},
		},
		{
			gw: gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake-gw",
					Namespace: "fake-ns",
				},
			},
			proxy: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake-access-log",
					Namespace: "envoy-gateway-system",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Telemetry: &egcfgv1a1.ProxyTelemetry{
						Tracing: &egcfgv1a1.ProxyTracing{
							Provider: egcfgv1a1.TracingProvider{
								Host: ptr.To("fake-host"),
								Port: 4317,
							},
						},
					},
				},
			},
			expected: &ir.Tracing{
				ServiceName:  "fake-gw.fake-ns",
				SamplingRate: 100.0,
				Destination: ir.RouteDestination{
					Name: "tracing/envoy-gateway-system/fake-access-log",
					Settings: []*ir.DestinationSetting{
						{
							Weight:   ptr.To[uint32](1),
							Protocol: ir.GRPC,
							Endpoints: []*ir.DestinationEndpoint{
								{
									Host: "fake-host",
									Port: 4317,
								},
							},
							AddressType: ptr.To(ir.FQDN),
						},
					},
				},
			},
		},
		{
			gw: gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake-gw",
					Namespace: "fake-ns",
				},
			},
			proxy: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake-eproxy",
					Namespace: "fake-ns",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Telemetry: &egcfgv1a1.ProxyTelemetry{
						Tracing: &egcfgv1a1.ProxyTracing{
							Provider: egcfgv1a1.TracingProvider{
								BackendRefs: []egcfgv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "fake-name",
											Port: PortNumPtr(4317),
										},
									},
								},
							},
						},
					},
				},
			},
			resources: Resources{
				Services: []*corev1.Service{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "fake-name",
							Namespace: "fake-ns",
						},
						Spec: corev1.ServiceSpec{
							Ports: []corev1.ServicePort{
								{
									Name:       "otlp",
									Port:       4317,
									TargetPort: intstr.IntOrString{IntVal: 4317},
									Protocol:   corev1.ProtocolTCP,
								},
							},
						},
					},
				},
				EndpointSlices: []*discoveryv1.EndpointSlice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "fake-name-1",
							Namespace: "fake-ns",
							Labels: map[string]string{
								discoveryv1.LabelServiceName: "fake-name",
							},
						},
						AddressType: discoveryv1.AddressTypeIPv4,
						Endpoints: []discoveryv1.Endpoint{
							{
								Addresses:  []string{"1.2.3.4"},
								Conditions: discoveryv1.EndpointConditions{Ready: ptr.To(true)},
							},
							{
								Addresses:  []string{"2.3.4.5"},
								Conditions: discoveryv1.EndpointConditions{Ready: ptr.To(true)},
							},
						},
						Ports: []discoveryv1.EndpointPort{
							{
								Name:     ptr.To("otlp"),
								Port:     ptr.To[int32](4317),
								Protocol: ptr.To(corev1.ProtocolTCP),
							},
						},
					},
				},
			},
			expected: &ir.Tracing{
				ServiceName:  "fake-gw.fake-ns",
				SamplingRate: 100.0,
				Destination: ir.RouteDestination{
					Name: "tracing/fake-ns/fake-eproxy",
					Settings: []*ir.DestinationSetting{
						{
							Weight:   ptr.To[uint32](1),
							Protocol: ir.GRPC,
							Endpoints: []*ir.DestinationEndpoint{
								{
									Host: "1.2.3.4",
									Port: 4317,
								},
								{
									Host: "2.3.4.5",
									Port: 4317,
								},
							},
							AddressType: ptr.To(ir.IP),
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run("", func(t *testing.T) {
			got := c.translator.processTracing(&c.gw, c.proxy, &c.resources)
			assert.Equal(t, c.expected, got)
		})
	}
}

func TestProcessMetrics(t *testing.T) {
	cases := []struct {
		name       string
		proxy      *egcfgv1a1.EnvoyProxy
		translator Translator
		expected   *ir.Metrics
	}{
		{
			name: "nil proxy config",
		},
		{
			name: "virtual host stats enabled",
			proxy: &egcfgv1a1.EnvoyProxy{
				Spec: egcfgv1a1.EnvoyProxySpec{
					Telemetry: &egcfgv1a1.ProxyTelemetry{
						Metrics: &egcfgv1a1.ProxyMetrics{
							EnableVirtualHostStats: true,
						},
					},
				},
			},
			expected: &ir.Metrics{
				EnableVirtualHostStats: true,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.translator.processMetrics(c.proxy)
			assert.Equal(t, c.expected, got)
		})
	}
}
