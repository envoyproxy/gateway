// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
)

func TestDeleteRateLimitService(t *testing.T) {
	rl := &egv1a1.RateLimit{
		Backend: egv1a1.RateLimitDatabaseBackend{
			Type: egv1a1.RedisBackendType,
			Redis: &egv1a1.RateLimitRedisSettings{
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

			kube.EnvoyGateway.RateLimit = rl
			r := ratelimit.NewResourceRender(kube.Namespace, kube.EnvoyGateway, nil)
			err := kube.createOrUpdateService(context.Background(), r)
			require.NoError(t, err)

			err = kube.deleteService(context.Background(), r)
			require.NoError(t, err)
		})
	}
}
