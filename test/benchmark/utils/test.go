// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark
// +build benchmark

package utils

type BenchmarkTest struct {
	ShortName   string
	Description string

	// Define tha benchmark targets, key is target name.
	GatewayTargets  map[string]*GatewayTarget
	HTTPRoutTargets map[string]*HTTPRouteTarget
}

type GatewayTarget struct {
	Bounds []uint16
}

type HTTPRouteTarget struct {
	Bounds        []uint16
	TargetGateway string
}
