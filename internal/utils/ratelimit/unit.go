// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ratelimit

import (
	"time"

	"google.golang.org/protobuf/types/known/durationpb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func UnitToSeconds(unit egv1a1.RateLimitUnit) int64 {
	var seconds int64

	switch unit {
	case egv1a1.RateLimitUnitSecond:
		seconds = 1
	case egv1a1.RateLimitUnitMinute:
		seconds = 60
	case egv1a1.RateLimitUnitHour:
		seconds = 60 * 60
	case egv1a1.RateLimitUnitDay:
		seconds = 60 * 60 * 24
	case egv1a1.RateLimitUnitMonth:
		seconds = 60 * 60 * 24 * 30
	case egv1a1.RateLimitUnitYear:
		seconds = 60 * 60 * 24 * 365
	}
	return seconds
}

func UnitToDuration(unit ir.RateLimitUnit) *durationpb.Duration {
	seconds := UnitToSeconds(egv1a1.RateLimitUnit(unit))
	return durationpb.New(time.Duration(seconds) * time.Second)
}
