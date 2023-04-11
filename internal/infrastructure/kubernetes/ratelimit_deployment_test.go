// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestCreateOrUpdateRateLimitDeployment(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	rateLimitInfra := new(ir.RateLimitInfra)
	rl := &egcfgv1a1.RateLimit{
		Backend: egcfgv1a1.RateLimitDatabaseBackend{
			Type: egcfgv1a1.RedisBackendType,
			Redis: &egcfgv1a1.RateLimitRedisSettings{
				URL: "redis.redis.svc:6379",
			},
		},
	}

	r := ratelimit.NewResourceRender(cfg.Namespace, rateLimitInfra, rl, cfg.EnvoyGateway.GetEnvoyGatewayProvider().GetEnvoyGatewayKubeProvider().RateLimitDeployment)
	deployment, err := r.Deployment()
	require.NoError(t, err)

	testCases := []struct {
		name    string
		in      *ir.RateLimitInfra
		current *appsv1.Deployment
		want    *appsv1.Deployment
	}{
		{
			name: "create ratelimit deployment",
			in:   rateLimitInfra,
			want: deployment,
		},
		{
			name:    "ratelimit deployment exists",
			in:      rateLimitInfra,
			current: deployment,
			want:    deployment,
		},
		{
			name:    "update ratelimit deployment image",
			in:      &ir.RateLimitInfra{},
			current: deployment,
			want:    deploymentWithImage(deployment, egcfgv1a1.DefaultRateLimitImage),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var cli client.Client
			if tc.current != nil {
				cli = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.current).Build()
			} else {
				cli = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build()
			}

			kube := NewInfra(cli, cfg)
			r := ratelimit.NewResourceRender(kube.Namespace, tc.in, rl, kube.EnvoyGateway.GetEnvoyGatewayProvider().GetEnvoyGatewayKubeProvider().RateLimitDeployment)
			err := kube.createOrUpdateDeployment(context.Background(), r)
			require.NoError(t, err)

			actual := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: kube.Namespace,
					Name:      ratelimit.InfraName,
				},
			}
			require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(actual), actual))
			require.Equal(t, tc.want.Spec, actual.Spec)
		})
	}
}

func TestDeleteRateLimitDeployment(t *testing.T) {
	rl := &egcfgv1a1.RateLimit{
		Backend: egcfgv1a1.RateLimitDatabaseBackend{
			Type: egcfgv1a1.RedisBackendType,
			Redis: &egcfgv1a1.RateLimitRedisSettings{
				URL: "redis.redis.svc:6379",
			},
		},
	}

	testCases := []struct {
		name   string
		expect bool
	}{
		{
			name:   "delete ratelimit deployment",
			expect: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			kube := newTestInfra(t)
			rateLimitInfra := new(ir.RateLimitInfra)
			r := ratelimit.NewResourceRender(kube.Namespace, rateLimitInfra, rl, kube.EnvoyGateway.GetEnvoyGatewayProvider().GetEnvoyGatewayKubeProvider().RateLimitDeployment)
			err := kube.createOrUpdateDeployment(context.Background(), r)
			require.NoError(t, err)

			err = kube.deleteDeployment(context.Background(), r)
			require.NoError(t, err)
		})
	}
}
