// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package conformance

import (
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

var EnvoyGatewaySuite = suite.Options{
	SupportedFeatures: suite.AllFeatures,
	ExemptFeatures:    suite.MeshCoreFeatures,
	SkipTests:         []string{tests.GatewayStaticAddresses.ShortName},
}
