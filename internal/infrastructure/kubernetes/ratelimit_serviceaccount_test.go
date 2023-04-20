// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestCreateOrUpdateRateLimitServiceAccount(t *testing.T) {
	rl := &egcfgv1a1.RateLimit{
		Backend: egcfgv1a1.RateLimitDatabaseBackend{
			Type: egcfgv1a1.RedisBackendType,
			Redis: &egcfgv1a1.RateLimitRedisSettings{
				URL: "redis.redis.svc:6379",
			},
		},
	}

	testCases := []struct {
		name    string
		ns      string
		in      *ir.RateLimitInfra
		current *corev1.ServiceAccount
		want    *corev1.ServiceAccount
	}{
		{
			name: "create-ratelimit-sa",
			ns:   "envoy-gateway-system",
			in:   new(ir.RateLimitInfra),
			want: &corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "envoy-gateway-system",
					Name:      ratelimit.InfraName,
				},
			},
		},
		{
			name: "ratelimit-sa-exists",
			ns:   "envoy-gateway-system",
			in:   new(ir.RateLimitInfra),
			want: &corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "envoy-gateway-system",
					Name:      ratelimit.InfraName,
				},
			},
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

			cfg, err := config.New()
			require.NoError(t, err)
			cfg.Namespace = tc.ns

			kube := NewInfra(cli, cfg)

			r := ratelimit.NewResourceRender(kube.Namespace, tc.in, rl, kube.EnvoyGateway.GetEnvoyGatewayProvider().GetEnvoyGatewayKubeProvider().RateLimitDeployment)

			err = kube.createOrUpdateServiceAccount(context.Background(), r)
			require.NoError(t, err)

			actual := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: kube.Namespace,
					Name:      ratelimit.InfraName,
				},
			}
			require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(actual), actual))

			opts := cmpopts.IgnoreFields(metav1.ObjectMeta{}, "ResourceVersion")
			assert.Equal(t, true, cmp.Equal(tc.want, actual, opts))
		})
	}
}

func TestDeleteRateLimitServiceAccount(t *testing.T) {
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
			name: "delete ratelimit service account",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kube := newTestInfra(t)

			rateLimitInfra := new(ir.RateLimitInfra)
			r := ratelimit.NewResourceRender(kube.Namespace, rateLimitInfra, rl, kube.EnvoyGateway.GetEnvoyGatewayProvider().GetEnvoyGatewayKubeProvider().RateLimitDeployment)
			err := kube.createOrUpdateServiceAccount(context.Background(), r)
			require.NoError(t, err)

			err = kube.deleteServiceAccount(context.Background(), r)
			require.NoError(t, err)
		})
	}
}
