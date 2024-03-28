package ratelimit

import (
	"testing"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestBuildTraceEndpoint(t *testing.T) {

	cases := []struct {
		caseName  string
		rateLimit *egv1a1.RateLimit
		expect    string
	}{
		{
			caseName: "default-endpoint",
			rateLimit: &egv1a1.RateLimit{
				Telemetry: &egv1a1.RateLimitTelemetry{
					Tracing: &egv1a1.RateLimitTracing{
						BackendRef: gwapiv1.BackendObjectReference{
							Name: "collector",
							Namespace: func() *gwapiv1.Namespace {
								var ns gwapiv1.Namespace = "observability"
								return &ns
							}(),
						},
					},
				},
			},
			expect: "collector.observability.svc.cluster.local:4318",
		},
		{
			caseName: "endpoint-with-port",
			rateLimit: &egv1a1.RateLimit{
				Telemetry: &egv1a1.RateLimitTelemetry{
					Tracing: &egv1a1.RateLimitTracing{
						BackendRef: gwapiv1.BackendObjectReference{
							Name: "collector",
							Namespace: func() *gwapiv1.Namespace {
								var ns gwapiv1.Namespace = "observability"
								return &ns
							}(),
							Port: func() *gwapiv1.PortNumber {
								var port gwapiv1.PortNumber = 4317
								return &port
							}(),
						},
					},
				},
			},
			expect: "collector.observability.svc.cluster.local:4317",
		},
		{
			caseName: "endpoint-with-domain",
			rateLimit: &egv1a1.RateLimit{
				Telemetry: &egv1a1.RateLimitTelemetry{
					Tracing: &egv1a1.RateLimitTracing{
						BackendRef: gwapiv1.BackendObjectReference{
							Name: "collector",
							Namespace: func() *gwapiv1.Namespace {
								var ns gwapiv1.Namespace = "observability"
								return &ns
							}(),
						},
						ClusterDomain: "example.com",
					},
				},
			},
			expect: "collector.observability.svc.example.com:4318",
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			actual := buildTraceEndpoint(tc.rateLimit)
			require.Equal(t, tc.expect, actual)
		})
	}

}
