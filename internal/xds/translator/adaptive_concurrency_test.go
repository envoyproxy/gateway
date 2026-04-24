// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestBuildAdaptiveConcurrencyProtoDefaultsConcurrencyUpdateInterval(t *testing.T) {
	tests := []struct {
		name     string
		interval *metav1.Duration
	}{
		{
			name: "omitted",
		},
		{
			name:     "zero",
			interval: &metav1.Duration{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := buildAdaptiveConcurrencyProto(&ir.AdaptiveConcurrency{
				ConcurrencyUpdateInterval: tt.interval,
			})

			params := ac.GetGradientControllerConfig().GetConcurrencyLimitParams()
			require.NotNil(t, params)
			require.Equal(t, defaultAdaptiveConcurrencyUpdateInterval, params.GetConcurrencyUpdateInterval().AsDuration())
		})
	}
}

func TestBuildAdaptiveConcurrencyProtoUsesConfiguredConcurrencyUpdateInterval(t *testing.T) {
	ac := buildAdaptiveConcurrencyProto(&ir.AdaptiveConcurrency{
		ConcurrencyUpdateInterval: &metav1.Duration{Duration: 200 * time.Millisecond},
	})

	params := ac.GetGradientControllerConfig().GetConcurrencyLimitParams()
	require.NotNil(t, params)
	require.Equal(t, 200*time.Millisecond, params.GetConcurrencyUpdateInterval().AsDuration())
}
