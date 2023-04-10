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
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestDesiredRateLimitService(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)
	cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
	kube := NewInfra(cli, cfg)
	rateLimitInfra := new(ir.RateLimitInfra)
	svc := kube.expectedRateLimitService(rateLimitInfra)

	// Check the service name is as expected.
	assert.Equal(t, svc.Name, rateLimitInfraName)

	checkServiceHasPort(t, svc, rateLimitInfraGRPCPort)

	// Ensure the Envoy RateLimit service has the expected labels.
	lbls := rateLimitLabels()
	checkServiceHasLabels(t, svc, lbls)

	// Make sure service type are set by default with ServiceTypeLoadBalancer
	checkServiceSpec(t, svc, expectedServiceSpec(egcfgv1a1.DefaultKubernetesServiceType()))
}

func TestDeleteRateLimitService(t *testing.T) {
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
			kube := &Infra{
				Client:    fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build(),
				Namespace: "test",
			}
			rateLimitInfra := new(ir.RateLimitInfra)

			err := kube.createOrUpdateRateLimitService(context.Background(), rateLimitInfra)
			require.NoError(t, err)

			err = kube.deleteRateLimitService(context.Background(), rateLimitInfra)
			require.NoError(t, err)
		})
	}
}
