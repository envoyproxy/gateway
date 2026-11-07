// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"slices"
	"strings"

	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/pkg/features"
)

// unsupportedFeatures are the Gateway API features that are not reported in
// GatewayClass.Status.SupportedFeatures, in addition to the mesh features, which
// do not apply to a Gateway controller.
//
// Every other standard channel feature of the Gateway API module is reported, because the
// conformance suite (test/conformance) runs all of them. TestAdvertisedFeaturesAreTested
// makes sure each reported feature is run by the suite and has at least one test.
// Experimental channel features are never reported, the installed CRD channel is unknown.
var unsupportedFeatures = []features.FeatureName{
	// Skipped by the conformance suite.
	features.SupportGatewayHTTPSListenerDetectMisdirectedRequests,
	// No conformance tests exist for these yet.
	features.SupportHTTPRouteRetryBackendTimeout,
	features.SupportHTTPRouteRetryConnectionError,
}

// namespaceModeFeaturePrefix matches the infrastructure feature, which is only supported in
// Gateway namespace mode, while the status is shared by both modes. It is matched by prefix
// because the feature was renamed from GatewayInfrastructurePropagation to GatewayInfrastructure
// in Gateway API v1.6.3, and this package is compiled against both versions (the controller
// and the test module), so exactly one of the two names exists at a time.
const namespaceModeFeaturePrefix = "GatewayInfrastructure"

// GatewaySupportedFeatures is the list of supported features sorted by name, as required by the CRD.
var GatewaySupportedFeatures = func() []gwapiv1.SupportedFeature {
	skipped := features.SetsToNamesSet(features.MeshCoreFeatures, features.MeshExtendedFeatures)
	skipped.Insert(unsupportedFeatures...)

	supported := make([]gwapiv1.SupportedFeature, 0, len(features.AllFeatures))
	for feature := range features.AllFeatures {
		if feature.Channel == features.FeatureChannelStandard && !skipped.Has(feature.Name) &&
			!strings.HasPrefix(string(feature.Name), namespaceModeFeaturePrefix) {
			supported = append(supported, gwapiv1.SupportedFeature{Name: gwapiv1.FeatureName(feature.Name)})
		}
	}
	slices.SortFunc(supported, func(a, b gwapiv1.SupportedFeature) int {
		return strings.Compare(string(a.Name), string(b.Name))
	})
	return supported
}()
