package globalratelimit

import (
	"testing"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestGetRateLimitUrl(t *testing.T) {
	tests := []struct {
		name      string
		eg        *egv1a1.EnvoyGateway
		namespace string
		dnsDomain string
		expected  string
	}{
		{
			name: "use user supplied url",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					RateLimit: &egv1a1.RateLimit{
						URL: new("grpc://cool-rate-limiter.com:50051"),
					},
				},
			},
			namespace: "default",
			dnsDomain: "cluster.local",
			expected:  "grpc://cool-rate-limiter.com:50051",
		},
		{
			name: "no rate limit config should use cluster dns",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{},
			},
			namespace: "default",
			dnsDomain: "cluster.local",
			expected:  "grpc://envoy-ratelimit.default.svc.cluster.local:8081",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRateLimitUrl(tt.eg, tt.namespace, tt.dnsDomain)
			assert.Equal(t, tt.expected, result)
		})
	}
}
