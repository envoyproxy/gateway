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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"

	"github.com/envoyproxy/gateway/internal/gatewayapi/conformance"
)

const (
	ReasonOlderGatewayClassExists gwapiv1.GatewayClassConditionReason = "OlderGatewayClassExists"

	MsgOlderGatewayClassExists   = "Invalid GatewayClass: another older GatewayClass with the same Spec.Controller exists"
	MsgValidGatewayClass         = "Valid GatewayClass"
	MsgGatewayClassInvalidParams = "Invalid parametersRef"
)

// SetGatewayClassAccepted inserts or updates the Accepted condition
// for the provided GatewayClass.
func SetGatewayClassAccepted(gc *gwapiv1.GatewayClass, accepted bool, reason, msg string) *gwapiv1.GatewayClass {
	gc.Status.Conditions = MergeConditions(gc.Status.Conditions, computeGatewayClassAcceptedCondition(gc, accepted, reason, msg))
	// Disable SupportedFeatures until the field moves from experimental to stable to avoid
	// status failures due to changes in the datatype. This can occur because we cannot control
	// how a CRD is installed in the cluster
	// gc.Status.SupportedFeatures = GatewaySupportedFeatures
	return gc
}

// computeGatewayClassAcceptedCondition computes the GatewayClass Accepted status condition.
func computeGatewayClassAcceptedCondition(gatewayClass *gwapiv1.GatewayClass,
	accepted bool,
	reason, msg string,
) metav1.Condition {
	switch accepted {
	case true:
		return metav1.Condition{
			Type:               string(gwapiv1.GatewayClassConditionStatusAccepted),
			Status:             metav1.ConditionTrue,
			Reason:             reason,
			Message:            msg,
			ObservedGeneration: gatewayClass.Generation,
			LastTransitionTime: metav1.NewTime(time.Now()),
		}
	default:
		return metav1.Condition{
			Type:               string(gwapiv1.GatewayClassConditionStatusAccepted),
			Status:             metav1.ConditionFalse,
			Reason:             reason,
			Message:            msg,
			ObservedGeneration: gatewayClass.Generation,
			LastTransitionTime: metav1.NewTime(time.Now()),
		}
	}
}

func getSupportedFeatures(gatewaySuite suite.ConformanceOptions, skippedTests []suite.ConformanceTest) []gwapiv1.SupportedFeature {
	supportedFeatures := gatewaySuite.SupportedFeatures.Clone()
	unsupportedFeatures := getUnsupportedFeatures(gatewaySuite, skippedTests)
	supportedFeatures.Delete(unsupportedFeatures...)

	ret := sets.New[gwapiv1.SupportedFeature]()
	for _, feature := range supportedFeatures.UnsortedList() {
		ret.Insert(gwapiv1.SupportedFeature{
			Name: gwapiv1.FeatureName(feature),
		})
	}

	var featureList []gwapiv1.SupportedFeature
	for feature := range ret {
		featureList = append(featureList, feature)
	}
	return featureList
}

func getUnsupportedFeatures(gatewaySuite suite.ConformanceOptions, skippedTests []suite.ConformanceTest) []features.FeatureName {
	unsupportedFeatures := gatewaySuite.ExemptFeatures.UnsortedList()

	for _, skippedTest := range skippedTests {
		switch conformance.GetTestSupportLevel(skippedTest) {
		case conformance.Core:
			unsupportedFeatures = append(unsupportedFeatures, skippedTest.Features...)
		case conformance.Extended:
			for _, feature := range skippedTest.Features {
				if conformance.GetFeatureSupportLevel(feature) == conformance.Extended {
					unsupportedFeatures = append(unsupportedFeatures, feature)
				}
			}
		}
	}

	return unsupportedFeatures
}
