// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark
// +build benchmark

package tests

import (
	"github.com/envoyproxy/gateway/test/benchmark/utils"
)

func init() {
	BenchmarkTests = append(BenchmarkTests, ScaleHTTPRoutes)
}

var ScaleHTTPRoutes = utils.BenchmarkTest{
	ShortName:   "ScaleHTTPRoute",
	Description: "Fixed one Gateway with different scale of HTTPRoutes",
	GatewayTargets: map[string]*utils.GatewayTarget{
		"benchmark-gateway": {
			Bounds: []uint16{1},
		},
	},
	HTTPRoutTargets: map[string]*utils.HTTPRouteTarget{
		"benchmark-route": {
			TargetGateway: "benchmark-gateway",
			Bounds:        []uint16{10, 100, 300, 500},
		},
	},
}
