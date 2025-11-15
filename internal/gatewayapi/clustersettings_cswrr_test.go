// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestBuildLoadBalancer_ClientSideWeightedRoundRobin(t *testing.T) {
	cswrr := &egv1a1.ClientSideWeightedRoundRobin{
		BlackoutPeriod:                     ptr.To(gwapiv1.Duration("10s")),
		WeightExpirationPeriod:             ptr.To(gwapiv1.Duration("3m")),
		WeightUpdatePeriod:                 ptr.To(gwapiv1.Duration("1s")),
		ErrorUtilizationPenalty:            ptr.To[float32](1.5),
		MetricNamesForComputingUtilization: []string{"named_metrics.foo", "cpu_utilization"},
	}

	policy := &egv1a1.ClusterSettings{
		LoadBalancer: &egv1a1.LoadBalancer{
			Type:                         egv1a1.ClientSideWeightedRoundRobinLoadBalancerType,
			ClientSideWeightedRoundRobin: cswrr,
		},
	}

	lb, err := buildLoadBalancer(policy)
	require.NoError(t, err)
	require.NotNil(t, lb)
	require.NotNil(t, lb.ClientSideWeightedRoundRobin)

	got := lb.ClientSideWeightedRoundRobin
	require.Equal(t, ptr.To(metav1.Duration{Duration: 10 * time.Second}), got.BlackoutPeriod)
	require.Equal(t, ptr.To(metav1.Duration{Duration: 3 * time.Minute}), got.WeightExpirationPeriod)
	require.Equal(t, ptr.To(metav1.Duration{Duration: 1 * time.Second}), got.WeightUpdatePeriod)
	require.NotNil(t, got.ErrorUtilizationPenalty)
	require.InDelta(t, 1.5, *got.ErrorUtilizationPenalty, 0.0001)
	require.Equal(t, []string{"named_metrics.foo", "cpu_utilization"}, got.MetricNamesForComputingUtilization)
}
