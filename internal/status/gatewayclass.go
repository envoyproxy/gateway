// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// This file contains code derived from Contour,
// https://github.com/projectcontour/contour
// from the source file
// https://github.com/projectcontour/contour/blob/main/internal/status/gatewayclass.go
// and is provided here subject to the following:
// Copyright Project Contour Authors
// SPDX-License-Identifier: Apache-2.0

package status

import (
	"k8s.io/apimachinery/pkg/util/sets"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

// SetGatewayClassAccepted inserts or updates the Accepted condition
// for the provided GatewayClass.
func SetGatewayClassAccepted(gc *gwapiv1.GatewayClass, accepted bool, reason, msg string) *gwapiv1.GatewayClass {
	gc.Status.Conditions = MergeConditions(gc.Status.Conditions, computeGatewayClassAcceptedCondition(gc, accepted, reason, msg))
	return gc
}

// GetSupportedFeatures returns a list of supported Gateway-API features,
// based on the running conformance tests suite.
func GetSupportedFeatures() []gwapiv1.SupportedFeature {

	// TODO(levikobi): This must be in sync with the cSuite supported features.
	supportedFeatures := suite.AllFeatures
	supportedFeatures.Delete(suite.MeshCoreFeatures.UnsortedList()...)

	ret := sets.New[gwapiv1.SupportedFeature]()
	for _, feature := range supportedFeatures.UnsortedList() {
		ret.Insert(gwapiv1.SupportedFeature(feature))
	}
	return sets.List(ret)
}

// SetGatewayClassSupportedFeatures insert or updates the SupportedFeatures field
// for the provided GatewayClass.
func SetGatewayClassSupportedFeatures(gc *gwapiv1.GatewayClass) *gwapiv1.GatewayClass {
	gc.Status.SupportedFeatures = GetSupportedFeatures()
	return gc
}
