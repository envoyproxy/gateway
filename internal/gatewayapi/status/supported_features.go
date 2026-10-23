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
	// Only supported in Gateway namespace mode, but the status is shared by both modes.
	// Listed by name, this package is also compiled by the test module, which uses a newer
	// Gateway API version where the feature was renamed from GatewayInfrastructurePropagation.
	"GatewayInfrastructurePropagation",
	"GatewayInfrastructure",
	// Skipped by the conformance suite.
	features.SupportGatewayHTTPSListenerDetectMisdirectedRequests,
	// No conformance tests exist for these yet.
	features.SupportHTTPRouteRetryBackendTimeout,
	features.SupportHTTPRouteRetryConnectionError,
}

// GatewaySupportedFeatures is the list of supported features sorted by name, as required by the CRD.
var GatewaySupportedFeatures = func() []gwapiv1.SupportedFeature {
	skipped := features.SetsToNamesSet(features.MeshCoreFeatures, features.MeshExtendedFeatures)
	skipped.Insert(unsupportedFeatures...)

	supported := make([]gwapiv1.SupportedFeature, 0, len(features.AllFeatures))
	for feature := range features.AllFeatures {
		if feature.Channel == features.FeatureChannelStandard && !skipped.Has(feature.Name) {
			supported = append(supported, gwapiv1.SupportedFeature{Name: gwapiv1.FeatureName(feature.Name)})
		}
	}
	slices.SortFunc(supported, func(a, b gwapiv1.SupportedFeature) int {
		return strings.Compare(string(a.Name), string(b.Name))
	})
	return supported
}()
