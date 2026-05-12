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
	ConformanceTests = append(ConformanceTests, AuthorizationGeoIPDownstreamRemoteAddressTest)
}

var AuthorizationGeoIPDownstreamRemoteAddressTest = suite.ConformanceTest{
	ShortName:   "AuthzWithGeoIPDownstreamRemoteAddress",
	Description: "Authorization with GeoIP country match using clientIPDetection.downstreamRemoteAddress",
	Manifests:   []string{"testdata/authorization-geoip-downstream-remote-address.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := ConformanceInfraNamespace
		ensureGeoIPDownstreamRemoteConfigMap(t, suite)

		routeNN := types.NamespacedName{Name: "http-authz-geoip-downstream-remote", Namespace: ns}
		gwNN := types.NamespacedName{Name: "geoip-authz-drm-gateway", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "authorization-geoip-downstream-remote", Namespace: ns}, suite.ControllerName, ancestorRef)
		ClientTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "enable-client-ip-detection-geoip-drm", Namespace: ns}, suite.ControllerName, ancestorRef)

		// In downstreamRemoteAddress mode Envoy uses the immediate TCP peer as the client IP.
		// The cluster-internal peer IP that reaches Envoy is not present in the test MMDB,
		// so the allow rule for country=US never matches and defaultAction=Deny applies.
		t.Run("xff-header-ignored-default-deny", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/geo-drm",
					Headers: map[string]string{
						// 50.114.0.1 is US in the test MMDB; with downstreamRemoteAddress
						// this header is ignored, so the request must still be denied.
						"X-Forwarded-For": "50.114.0.1",
					},
				},
				Response: http.Response{
					StatusCodes: []int{403},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("no-client-ip-default-deny", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/geo-drm",
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
func ensureGeoIPDownstreamRemoteConfigMap(t *testing.T, suite *suite.ConformanceTestSuite) {
	t.Helper()

	targetNamespace := GetGatewayResourceNamespace()
	if targetNamespace == ConformanceInfraNamespace {
		return
	}

	ctx := context.Background()
	sourceNN := types.NamespacedName{Name: "geoip-country-db-drm", Namespace: ConformanceInfraNamespace}
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
