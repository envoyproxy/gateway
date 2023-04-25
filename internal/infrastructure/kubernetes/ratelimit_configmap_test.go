// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	"github.com/envoyproxy/gateway/internal/ir"
)

var (
	rateLimitListener = "ratelimit-listener"
	rateLimitConfig   = `
domain: first-listener
descriptors:
  - key: first-route-key-rule-0-match-0
    value: first-route-value-rule-0-match-0
    rate_limit:
      requests_per_unit: 5
      unit: second
      unlimited: false
      name: ""
      replaces: []
    descriptors: []
    shadow_mode: false
`
)

func TestCreateOrUpdateRateLimitConfigMap(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	rateLimitInfra := new(ir.RateLimitInfra)
	rateLimitInfra.ServiceConfigs = append(rateLimitInfra.ServiceConfigs, &ir.RateLimitServiceConfig{
		Name:   rateLimitListener,
		Config: rateLimitConfig,
	})
	cfg.EnvoyGateway.RateLimit = &egcfgv1a1.RateLimit{
		Backend: egcfgv1a1.RateLimitDatabaseBackend{
			Type: egcfgv1a1.RedisBackendType,
			Redis: &egcfgv1a1.RateLimitRedisSettings{
				URL: "redis.redis.svc:6379",
			},
		},
	}
	r := ratelimit.NewResourceRender(cfg.Namespace, rateLimitInfra, cfg.EnvoyGateway)

	testCases := []struct {
		name    string
		current *corev1.ConfigMap
		expect  *corev1.ConfigMap
	}{
		{
			name: "create ratelimit configmap",
			expect: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: cfg.Namespace,
					Name:      ratelimit.InfraName,
					Labels: map[string]string{
						"app.gateway.envoyproxy.io/name": ratelimit.InfraName,
					},
				},
				Data: map[string]string{rateLimitListener: rateLimitConfig},
			},
		},
		{
			name: "update ratelimit configmap",
			current: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: cfg.Namespace,
					Name:      ratelimit.InfraName,
					Labels: map[string]string{
						"app.gateway.envoyproxy.io/name": ratelimit.InfraName,
					},
				},
				Data: map[string]string{"foo": "bar"},
			},
			expect: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: cfg.Namespace,
					Name:      ratelimit.InfraName,
					Labels: map[string]string{
						"app.gateway.envoyproxy.io/name": ratelimit.InfraName,
					},
				},
				Data: map[string]string{rateLimitListener: rateLimitConfig},
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

			kube := NewInfra(cli, cfg)

			err := kube.createOrUpdateConfigMap(context.Background(), r)
			require.NoError(t, err)
			actual := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tc.expect.Namespace,
					Name:      tc.expect.Name,
				},
			}
			require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(actual), actual))
			require.Equal(t, tc.expect.Data, actual.Data)
			assert.True(t, apiequality.Semantic.DeepEqual(tc.expect.Labels, actual.Labels))
		})
	}
}

func TestDeleteRateLimitConfigMap(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	cfg.EnvoyGateway.RateLimit = &egcfgv1a1.RateLimit{
		Backend: egcfgv1a1.RateLimitDatabaseBackend{
			Type: egcfgv1a1.RedisBackendType,
			Redis: &egcfgv1a1.RateLimitRedisSettings{
				URL: "redis.redis.svc:6379",
			},
		},
	}

	testCases := []struct {
		name    string
		current *corev1.ConfigMap
		expect  bool
	}{
		{
			name: "delete ratelimit configmap",
			current: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: cfg.Namespace,
					Name:      "envoy-test",
				},
			},
			expect: true,
		},
		{
			name: "ratelimit configmap not found",
			current: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: cfg.Namespace,
					Name:      "foo",
				},
			},
			expect: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.current).Build()
			kube := NewInfra(cli, cfg)

			rateLimitInfra := new(ir.RateLimitInfra)
			r := ratelimit.NewResourceRender(kube.Namespace, rateLimitInfra, cfg.EnvoyGateway)

			err := kube.createOrUpdateConfigMap(context.Background(), r)
			require.NoError(t, err)

			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: kube.Namespace,
					Name:      r.Name(),
				},
			}
			err = kube.Client.Delete(context.Background(), cm)
			require.NoError(t, err)
		})
	}
}
