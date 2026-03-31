// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ratelimit

import (
	"testing"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestUnitToSeconds(t *testing.T) {
	tests := []struct {
		name string
		unit egv1a1.RateLimitUnit
		want int64
	}{
		{
			name: "second",
			unit: egv1a1.RateLimitUnitSecond,
			want: 1,
		},
		{
			name: "minute",
			unit: egv1a1.RateLimitUnitMinute,
			want: 60,
		},
		{
			name: "hour",
			unit: egv1a1.RateLimitUnitHour,
			want: 60 * 60,
		},
		{
			name: "day",
			unit: egv1a1.RateLimitUnitDay,
			want: 60 * 60 * 24,
		},
		{
			name: "month",
			unit: egv1a1.RateLimitUnitMonth,
			want: 60 * 60 * 24 * 30,
		},
		{
			name: "year",
			unit: egv1a1.RateLimitUnitYear,
			want: 60 * 60 * 24 * 365,
		},
		{
			name: "unknown unit",
			unit: egv1a1.RateLimitUnit("Unknown"),
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UnitToSeconds(tt.unit); got != tt.want {
				t.Errorf("UnitToSeconds(%q) = %d, want %d", tt.unit, got, tt.want)
			}
		})
	}
}

func TestUnitToDuration(t *testing.T) {
	tests := []struct {
		name        string
		unit        ir.RateLimitUnit
		wantSeconds int64
	}{
		{
			name:        "second",
			unit:        ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
			wantSeconds: 1,
		},
		{
			name:        "minute",
			unit:        ir.RateLimitUnit(egv1a1.RateLimitUnitMinute),
			wantSeconds: 60,
		},
		{
			name:        "hour",
			unit:        ir.RateLimitUnit(egv1a1.RateLimitUnitHour),
			wantSeconds: 3600,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UnitToDuration(tt.unit)
			if got.Seconds != tt.wantSeconds || got.Nanos != 0 {
				t.Errorf("UnitToDuration(%q) = %v, want %ds", tt.unit, got, tt.wantSeconds)
			}
		})
	}
}
