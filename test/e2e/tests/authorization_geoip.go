// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, AuthorizationGeoIPCountryTest)
}

var AuthorizationGeoIPCountryTest = suite.ConformanceTest{
	ShortName:   "AuthzWithGeoIPCountry",
	Description: "Authorization with GeoIP country match",
	Manifests:   []string{"testdata/authorization-geoip-country.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := ConformanceInfraNamespace
		ensureGeoIPConfigMap(t, suite)

		routeNN := types.NamespacedName{Name: "http-with-authorization-geoip-country", Namespace: ns}
		gwNN := types.NamespacedName{Name: "geoip-authz-gateway", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "authorization-geoip-country", Namespace: ns}, suite.ControllerName, ancestorRef)
		ClientTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "enable-client-ip-detection-geoip", Namespace: ns}, suite.ControllerName, ancestorRef)

		t.Run("country-allowed", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/geo-country",
					Headers: map[string]string{
						"X-Forwarded-For": "50.114.0.1", // GeoIP2-Country-Test.mmdb => US
					},
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path:    "/geo-country",
						Headers: nil,
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("country-denied", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/geo-country",
					Headers: map[string]string{
						"X-Forwarded-For": "81.2.69.160", // GeoIP2-Country-Test.mmdb => GB
					},
				},
				Response: http.Response{
					StatusCodes: []int{403},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("country-missing-client-ip-default-deny", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/geo-country",
				},
				Response: http.Response{
					StatusCodes: []int{403},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}

// Ensure the GeoIP ConfigMap is present in the namespace where the Envoy fleet is deployed.
func ensureGeoIPConfigMap(t *testing.T, suite *suite.ConformanceTestSuite) {
	t.Helper()

	targetNamespace := GetGatewayResourceNamespace()
	if targetNamespace == ConformanceInfraNamespace {
		return
	}

	ctx := context.Background()
	sourceNN := types.NamespacedName{Name: "geoip-country-db", Namespace: ConformanceInfraNamespace}
	source := &corev1.ConfigMap{}
	if err := suite.Client.Get(ctx, sourceNN, source); err != nil {
		t.Fatalf("failed to get GeoIP configmap %s: %v", sourceNN, err)
	}

	targetNN := types.NamespacedName{Name: source.Name, Namespace: targetNamespace}
	target := &corev1.ConfigMap{}
	err := suite.Client.Get(ctx, targetNN, target)
	switch {
	case apierrors.IsNotFound(err):
		configMap := source.DeepCopy()
		configMap.Namespace = targetNamespace
		configMap.ResourceVersion = ""
		configMap.UID = ""
		configMap.CreationTimestamp = metav1.Time{}
		configMap.ManagedFields = nil
		if err := suite.Client.Create(ctx, configMap); err != nil {
			t.Fatalf("failed to create GeoIP configmap %s: %v", targetNN, err)
		}
	case err != nil:
		t.Fatalf("failed to get GeoIP configmap %s: %v", targetNN, err)
	default:
		sourceCopy := source.DeepCopy()
		target.BinaryData = sourceCopy.BinaryData
		target.Data = sourceCopy.Data
		if err := suite.Client.Update(ctx, target); err != nil {
			t.Fatalf("failed to update GeoIP configmap %s: %v", targetNN, err)
		}
	}
}
