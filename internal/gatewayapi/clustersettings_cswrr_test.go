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
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestBuildLoadBalancer_ClientSideWeightedRoundRobin(t *testing.T) {
	cswrr := &egv1a1.ClientSideWeightedRoundRobin{
		EnableOOBLoadReport:                ptrBool(true),
		OOBReportingPeriod:                 ptrDuration("5s"),
		BlackoutPeriod:                     ptrDuration("10s"),
		WeightExpirationPeriod:             ptrDuration("3m"),
		WeightUpdatePeriod:                 ptrDuration("1s"),
		ErrorUtilizationPenalty:            ptrFloat32(1.5),
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
	require.Equal(t, cswrr.EnableOOBLoadReport, got.EnableOOBLoadReport)
	require.Equal(t, toMetaV1Duration(t, "5s"), got.OOBReportingPeriod)
	require.Equal(t, toMetaV1Duration(t, "10s"), got.BlackoutPeriod)
	require.Equal(t, toMetaV1Duration(t, "3m"), got.WeightExpirationPeriod)
	require.Equal(t, toMetaV1Duration(t, "1s"), got.WeightUpdatePeriod)
	require.NotNil(t, got.ErrorUtilizationPenalty)
	require.InDelta(t, 1.5, *got.ErrorUtilizationPenalty, 0.0001)
	require.Equal(t, []string{"named_metrics.foo", "cpu_utilization"}, got.MetricNamesForComputingUtilization)
}

func ptrBool(v bool) *bool                   { return &v }
func ptrFloat32(v float32) *float32          { return &v }
func ptrDuration(d string) *gwapiv1.Duration { v := gwapiv1.Duration(d); return &v }

func toMetaV1Duration(t *testing.T, d string) *metav1.Duration {
	dur, err := time.ParseDuration(d)
	require.NoError(t, err)
	return ir.MetaV1DurationPtr(dur)
}
