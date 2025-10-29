// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package e2e

import (
	"testing"

	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

var Hook = func(*testing.T, suite.ConformanceTest, *suite.ConformanceTestSuite) {
}
