// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestDeleteRateLimitService(t *testing.T) {
	rl := &egcfgv1a1.RateLimit{
		Backend: egcfgv1a1.RateLimitDatabaseBackend{
			Type: egcfgv1a1.RedisBackendType,
			Redis: &egcfgv1a1.RateLimitRedisSettings{
				URL: "redis.redis.svc:6379",
			},
		},
	}

	testCases := []struct {
		name string
	}{
		{
			name: "delete ratelimit service",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			kube := newTestInfra(t)

			rateLimitInfra := new(ir.RateLimitInfra)
			r := ratelimit.NewResourceRender(kube.Namespace, rateLimitInfra, rl, kube.EnvoyGateway.GetEnvoyGatewayProvider().GetEnvoyGatewayKubeProvider().RateLimitDeployment)
			err := kube.createOrUpdateService(context.Background(), r)
			require.NoError(t, err)

			err = kube.deleteService(context.Background(), r)
			require.NoError(t, err)
		})
	}
}
