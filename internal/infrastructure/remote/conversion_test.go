// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package remote

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

// fullyPopulatedInfra returns an ir.Infra exercising every field the
// conversion touches, including the fully-translated metadata subtree
// (annotations, labels, owner reference with nested map entries and policies),
// the structured metric sink destination (endpoints and upstream TLS), and the
// JSON-bytes-backed proxy config.
func fullyPopulatedInfra() *ir.Infra {
	return &ir.Infra{
		Proxy: &ir.ProxyInfra{
			Name:      "proxy",
			Namespace: "envoy-gateway-system",
			Metadata: &ir.InfraMetadata{
				Annotations: map[string]string{"anno": "value"},
				Labels:      map[string]string{"app": "envoy", "component": "proxy"},
				OwnerReference: &ir.ResourceMetadata{
					Kind:        "Gateway",
					Name:        "eg",
					Namespace:   "default",
					SectionName: "http",
					Annotations: []ir.MapEntry{{Key: "k", Value: "v"}},
					Policies: []*ir.PolicyMetadata{
						{Kind: "BackendTrafficPolicy", Name: "btp", Namespace: "default"},
					},
				},
			},
			Config: &egv1a1.EnvoyProxy{
				TypeMeta: metav1.TypeMeta{
					Kind:       "EnvoyProxy",
					APIVersion: "gateway.envoyproxy.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "custom-proxy",
					Namespace: "envoy-gateway-system",
				},
				Spec: egv1a1.EnvoyProxySpec{
					Logging: egv1a1.ProxyLogging{
						Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
							egv1a1.LogComponentDefault: egv1a1.LogLevelWarn,
						},
					},
				},
			},
			Listeners: []*ir.ProxyListener{
				{
					Name: "listener-1",
					Ports: []ir.ListenerPort{
						{
							Name:          "http",
							Protocol:      ir.HTTPProtocolType,
							ServicePort:   80,
							ContainerPort: 8080,
						},
						{
							Name:          "https",
							Protocol:      ir.HTTPSProtocolType,
							ServicePort:   443,
							ContainerPort: 8443,
						},
					},
				},
				{
					Name: "listener-2",
				},
			},
			Addresses: []string{"1.2.3.4", "5.6.7.8"},
			ResolvedMetricSinks: []ir.ResolvedMetricSink{
				{
					// Only the subset of ir.RouteDestination carried by the
					// contract is set here (settings -> endpoints and upstream
					// TLS); other fields are intentionally dropped by the
					// translation and would break the round-trip assertion.
					Destination: ir.RouteDestination{
						Settings: []*ir.DestinationSetting{
							{
								Endpoints: []*ir.DestinationEndpoint{
									{Host: "10.0.0.1", Port: 4317},
								},
								TLS: &ir.TLSUpstreamConfig{
									SNI:                 new("otel-collector"),
									UseSystemTrustStore: true,
									CACertificate: &ir.TLSCACertificate{
										Certificate: []byte("ca-cert-bytes"),
									},
								},
							},
						},
					},
					Authority:                "otel-collector",
					ResourceAttributes:       map[string]string{"service.name": "eg"},
					ReportCountersAsDeltas:   true,
					ReportHistogramsAsDeltas: true,
					Headers: []gwapiv1.HTTPHeader{
						{Name: "x-header", Value: "header-value"},
					},
				},
			},
		},
	}
}

func TestInfraProtoRoundTrip(t *testing.T) {
	t.Run("fully_populated", func(t *testing.T) {
		in := fullyPopulatedInfra()

		pbInfra, err := infraToProto(in)
		require.NoError(t, err)
		require.NotNil(t, pbInfra)

		out, err := protoToInfra(pbInfra)
		require.NoError(t, err)

		assert.Equal(t, in, out)
	})

	t.Run("nil_infra", func(t *testing.T) {
		pbInfra, err := infraToProto(nil)
		require.NoError(t, err)
		assert.Nil(t, pbInfra)

		out, err := protoToInfra(nil)
		require.NoError(t, err)
		assert.Nil(t, out)
	})

	t.Run("nil_proxy", func(t *testing.T) {
		in := &ir.Infra{}

		pbInfra, err := infraToProto(in)
		require.NoError(t, err)
		require.NotNil(t, pbInfra)
		assert.Nil(t, pbInfra.GetProxy())

		out, err := protoToInfra(pbInfra)
		require.NoError(t, err)
		assert.Equal(t, in, out)
	})

	t.Run("minimal_proxy", func(t *testing.T) {
		in := &ir.Infra{
			Proxy: &ir.ProxyInfra{
				Name:      "proxy",
				Namespace: "ns",
			},
		}

		pbInfra, err := infraToProto(in)
		require.NoError(t, err)

		out, err := protoToInfra(pbInfra)
		require.NoError(t, err)

		assert.Equal(t, in, out)
	})
}
