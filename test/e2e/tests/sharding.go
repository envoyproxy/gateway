// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

const (
	envShardTotal = "E2E_SHARD_TOTAL"
	envShardIndex = "E2E_SHARD_INDEX"
)

// ConformanceTestsForCurrentShard returns a deterministically ordered slice of conformance
// tests filtered for the configured shard. When sharding is disabled, the returned slice
// contains every registered test.
func ConformanceTestsForCurrentShard() ([]suite.ConformanceTest, error) {
	// Allow --run-test to bypass sharding so individual tests can still be executed directly.
	if flags.RunTest != nil && *flags.RunTest != "" {
		return ConformanceTests, nil
	}

	ordered := make([]suite.ConformanceTest, len(ConformanceTests))
	copy(ordered, ConformanceTests)

	totalStr := os.Getenv(envShardTotal)
	if totalStr == "" {
		return ordered, nil
	}

	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].ShortName < ordered[j].ShortName
	})

	total, err := strconv.Atoi(totalStr)
	if err != nil || total <= 0 {
		return nil, fmt.Errorf("invalid %s value %q: expected positive integer", envShardTotal, totalStr)
	}

	indexStr := os.Getenv(envShardIndex)
	if indexStr == "" {
		return nil, fmt.Errorf("%s must be set when %s is provided", envShardIndex, envShardTotal)
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 || index >= total {
		return nil, fmt.Errorf("invalid %s value %q: expected integer in [0,%d)", envShardIndex, indexStr, total)
	}

	var filtered []suite.ConformanceTest
	for i, test := range ordered {
		if i%total == index {
			filtered = append(filtered, test)
		}
	}

	return filtered, nil
}
