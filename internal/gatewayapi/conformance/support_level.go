// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package conformance

import (
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

// SupportLevel represents the level of support for a feature.
// See https://gateway-api.sigs.k8s.io/concepts/conformance/#2-support-levels.
type SupportLevel string

const (
	// Core features are portable and expected to be supported by every implementation of Gateway-API.
	Core SupportLevel = "core"

	// Extended features are those that are portable but not universally supported across implementations.
	// Those implementations that support the feature will have the same behavior and semantics.
	// It is expected that some number of roadmap features will eventually migrate into the Core.
	Extended SupportLevel = "extended"
)

// ExtendedFeatures is a list of supported Gateway-API features that are considered Extended.
var ExtendedFeatures = sets.New[features.FeatureName]()

func init() {
	featureLists := sets.New[features.Feature]().
		Insert(features.GatewayExtendedFeatures.UnsortedList()...).
		Insert(features.HTTPRouteExtendedFeatures.UnsortedList()...).
		Insert(features.MeshExtendedFeatures.UnsortedList()...)

	for _, feature := range featureLists.UnsortedList() {
		ExtendedFeatures.Insert(feature.Name)
	}
}

// GetTestSupportLevel returns the SupportLevel for a conformance test.
// The support level is determined by the highest support level of the features.
func GetTestSupportLevel(test *suite.ConformanceTest) SupportLevel {
	supportLevel := Core

	if ExtendedFeatures.HasAny(test.Features...) {
		supportLevel = Extended
	}

	return supportLevel
}

// GetFeatureSupportLevel returns the SupportLevel for a feature.
func GetFeatureSupportLevel(feature features.FeatureName) SupportLevel {
	supportLevel := Core

	if ExtendedFeatures.Has(feature) {
		supportLevel = Extended
	}

	return supportLevel
}
