// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package e2e

import (
	"testing"

	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

var Hook = func(t *testing.T, test suite.ConformanceTest, suite *suite.ConformanceTestSuite) {
	if t.Failed() {
		tlog.Logf(t, "Test %s failed, collecting and dumping resources", test.ShortName)
	}
}
