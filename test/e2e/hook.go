// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package e2e

import (
	"sync"
	"testing"

	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/test/e2e/tests"
)

// dumped records the tests whose resources Hook has already collected, so that a test is dumped
// at most once when Hook runs both at the end of the test body and from the suite cleanup.
var dumped sync.Map

// Hook collects and dumps the cluster resources when the test has failed. It is safe to call more
// than once for the same test: only the first call made after the failure dumps the resources.
var Hook = func(t *testing.T, test suite.ConformanceTest, suite *suite.ConformanceTestSuite) {
	tlog.Logf(t, "Run hook for test: %s for result: %v", test.ShortName, t.Failed())
	if !t.Failed() {
		return
	}
	if _, loaded := dumped.LoadOrStore(t, struct{}{}); loaded {
		tlog.Logf(t, "Test %s failed, resources already collected and dumped", test.ShortName)
		return
	}
	tlog.Logf(t, "Test %s failed, collecting and dumping resources", test.ShortName)
	tests.CollectAndDump(t, suite.RestConfig)
}

// WrapConformanceTestsWithHook runs Hook at the end of each test body, while the resources the
// test created are still in the cluster. ConformanceOptions.Hook alone is not enough: the suite
// runs it from a t.Cleanup registered before the test applies its manifests, and cleanups run
// last-in-first-out, so it only runs once the test resources have been deleted. A deferred call
// still runs when the test fails with t.Fatal or require, and before any t.Cleanup of the test.
//
// Keep setting ConformanceOptions.Hook as well: it covers failures this wrapper cannot see, such
// as a manifest that fails to apply before the test body starts, or a failing t.Cleanup of the
// test. Hook dumps each failed test only once, so a failure is never dumped twice.
//
// TODO: remove once https://github.com/kubernetes-sigs/gateway-api/pull/5185 is released.
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
