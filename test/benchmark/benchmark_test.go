// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark
// +build benchmark

package benchmark

import (
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/envoyproxy/gateway/test/benchmark/tests"
	"github.com/envoyproxy/gateway/test/benchmark/utils"
)

func TestBenchmark(t *testing.T) {
	cfg, err := config.GetConfig()
	require.NoError(t, err)

	cli, err := client.New(cfg, client.Options{})
	require.NoError(t, err)

	// Install all the scheme for kubernetes client.
	utils.CheckInstallScheme(t, cli)

	bSuite, err := utils.NewBenchmarkTestSuite(
		cli,
		"config/gateway.yaml",
		"config/httproute.yaml",
	)
	if err != nil {
		t.Fatalf("Failed to create BenchmarkTestSuite: %v", err)
	}

	t.Logf("Running %d benchmark tests", len(tests.BenchmarkTests))
	bSuite.Run(t, tests.BenchmarkTests)
}
