// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package e2e

import (
	"testing"

	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/test/e2e/tests"
)

var Hook = func(t *testing.T, test suite.ConformanceTest, suite *suite.ConformanceTestSuite) {
	tlog.Logf(t, "Run hook for test: %s for result: %v", test.ShortName, t.Failed())
	if t.Failed() {
		tlog.Logf(t, "Test %s failed, collecting and dumping resources", test.ShortName)
		tests.CollectAndDump(t, suite.RestConfig)
	}
}

// WrapConformanceTestsWithHook runs Hook at the end of each test body, while the resources the
// test created are still in the cluster. Use it instead of ConformanceOptions.Hook: the suite
// runs that from a t.Cleanup registered before the test applies its manifests, and cleanups run
// last-in-first-out, so it only runs once the test resources have been deleted. A deferred call
// still runs when the test fails with t.Fatal or require, and before any t.Cleanup of the test.
//
// TODO: remove once https://github.com/kubernetes-sigs/gateway-api/pull/5185 is released, and set
// ConformanceOptions.Hook again.
func WrapConformanceTestsWithHook(tests []suite.ConformanceTest) []suite.ConformanceTest {
	wrapped := make([]suite.ConformanceTest, 0, len(tests))

	for _, test := range tests {
		original := test.Test
		test.Test = func(t *testing.T, suite *suite.ConformanceTestSuite) {
			defer Hook(t, test, suite)
			original(t, suite)
		}
		wrapped = append(wrapped, test)
	}

	return wrapped
}
