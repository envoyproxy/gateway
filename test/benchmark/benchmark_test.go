// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark

package benchmark

import (
	"flag"
	"testing"

	"github.com/envoyproxy/gateway/test/benchmark/suite"
	"github.com/envoyproxy/gateway/test/benchmark/tests"
	kubetest "github.com/envoyproxy/gateway/test/utils/kubernetes"
)

func TestBenchmark(t *testing.T) {
	cli, _ := kubetest.NewClient(t)

	// Parse benchmark options.
	flag.Parse()
	options := suite.NewBenchmarkOptions(
		*suite.RPS,
		*suite.Connections,
		*suite.Duration,
		*suite.Concurrency,
	)

	bSuite, err := suite.NewBenchmarkTestSuite(
		cli,
		options,
		"config/gateway.yaml",
		"config/httproute.yaml",
		"config/nighthawk-client.yaml",
		"config/nighthawk-test-server.yaml",
		*suite.ReportSaveDir,
	)
	if err != nil {
		t.Fatalf("Failed to create BenchmarkTestSuite: %v", err)
	}

	t.Logf("Running %d benchmark tests", len(tests.BenchmarkTests))
	bSuite.Run(t, tests.BenchmarkTests)
}
