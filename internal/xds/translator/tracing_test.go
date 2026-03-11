// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	"github.com/stretchr/testify/require"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestBuildSampler(t *testing.T) {
	testCases := []struct {
		name          string
		sampler       *egv1a1.OTelSampler
		expected      string
		expectedError string
	}{
		{name: "nil", sampler: nil, expected: ""},
		{name: "AlwaysOn", sampler: &egv1a1.OTelSampler{Type: egv1a1.OTelSamplerTypeAlwaysOn}, expected: "envoy.tracers.opentelemetry.samplers.always_on"},
		{name: "AlwaysOff", sampler: &egv1a1.OTelSampler{Type: egv1a1.OTelSamplerTypeAlwaysOff}, expected: "envoy.tracers.opentelemetry.samplers.trace_id_ratio_based"},
		{name: "TraceIdRatioBased default arg", sampler: &egv1a1.OTelSampler{Type: egv1a1.OTelSamplerTypeTraceIDRatioBased}, expected: "envoy.tracers.opentelemetry.samplers.trace_id_ratio_based"},
		{name: "TraceIdRatioBased explicit config", sampler: &egv1a1.OTelSampler{Type: egv1a1.OTelSamplerTypeTraceIDRatioBased, SamplingPercentage: &gwapiv1.Fraction{Numerator: 50}}, expected: "envoy.tracers.opentelemetry.samplers.trace_id_ratio_based"},
		{name: "ParentBasedAlwaysOn", sampler: &egv1a1.OTelSampler{Type: egv1a1.OTelSamplerTypeParentBasedAlwaysOn}, expected: "envoy.tracers.opentelemetry.samplers.parent_based"},
		{name: "ParentBasedAlwaysOff", sampler: &egv1a1.OTelSampler{Type: egv1a1.OTelSamplerTypeParentBasedAlwaysOff}, expected: "envoy.tracers.opentelemetry.samplers.parent_based"},
		{name: "ParentBasedTraceIdRatioBased default arg", sampler: &egv1a1.OTelSampler{Type: egv1a1.OTelSamplerTypeParentBasedTraceIDRatioBased}, expected: "envoy.tracers.opentelemetry.samplers.parent_based"},
		{name: "ParentBasedTraceIdRatioBased zero", sampler: &egv1a1.OTelSampler{Type: egv1a1.OTelSamplerTypeParentBasedTraceIDRatioBased, SamplingPercentage: &gwapiv1.Fraction{Numerator: 0}}, expected: "envoy.tracers.opentelemetry.samplers.parent_based"},
		{name: "unknown type", sampler: &egv1a1.OTelSampler{Type: "Invalid"}, expectedError: "unknown sampler type: Invalid"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := buildSampler(tc.sampler)
			if tc.expectedError != "" {
				require.EqualError(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)
			if tc.expected == "" {
				require.Nil(t, actual)
				return
			}
			require.Equal(t, tc.expected, actual.Name)
		})
	}
}

func TestRandomSamplingValue(t *testing.T) {
	testCases := []struct {
		name     string
		tracing  *ir.Tracing
		expected float64
	}{
		{name: "no sampler uses SamplingRate", tracing: &ir.Tracing{SamplingRate: 50}, expected: 50},
		{name: "sampler overrides to 100", tracing: &ir.Tracing{SamplingRate: 50, Sampler: &egv1a1.OTelSampler{Type: egv1a1.OTelSamplerTypeAlwaysOn}}, expected: 100},
		{name: "AlwaysOff sampler still 100", tracing: &ir.Tracing{SamplingRate: 50, Sampler: &egv1a1.OTelSampler{Type: egv1a1.OTelSamplerTypeAlwaysOff}}, expected: 100},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, randomSamplingValue(tc.tracing))
		})
	}
}
