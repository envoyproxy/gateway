// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package conformance

import (
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

var EnvoyGatewaySuite = suite.ConformanceOptions{
	SupportedFeatures: features.AllFeatures,
	ExemptFeatures:    features.MeshCoreFeatures,
	SkipTests: []string{
		tests.GatewayStaticAddresses.ShortName,
		tests.GatewayHTTPListenerIsolation.ShortName,          // https://github.com/kubernetes-sigs/gateway-api/issues/3049
		tests.HTTPRouteBackendRequestHeaderModifier.ShortName, // https://github.com/envoyproxy/gateway/issues/3338
	},
}
