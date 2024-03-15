// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import "sigs.k8s.io/gateway-api/conformance/utils/suite"

var (
	ConformanceTests []suite.ConformanceTest
	UpgradeTests     []suite.ConformanceTest
)
