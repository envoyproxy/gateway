// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"os"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/cmd/options"
	"github.com/envoyproxy/gateway/internal/utils/helm"
)

func init() {
	LifecycleTests = append(LifecycleTests, EGUninstallAndInstallTest, EGUpgradeTest)
}

var EGUninstallAndInstallTest = suite.ConformanceTest{
	ShortName:   "EGUninstallAndInstall",
	Description: "Uninstall and Install Envoy Gateway using Helm Package Tool",
	Manifests:   []string{"testdata/eg-lifecycle.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Uninstall Envoy Gateway should succeed", func(t *testing.T) {
			// first ensure that EG provides services normally
			ns := "gateway-lifecycle-infra"
			routeNN := types.NamespacedName{
				Name:      "http-backend-eg-lifecycle",
				Namespace: ns,
			}
			gwNN := types.NamespacedName{
				Name:      "lifecycle-gateway",
				Namespace: ns,
			}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig,
				suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/eg-lifecycle",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response before the install process started: %v", err)
			}
			if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response before the uninstall and install process started: %v", err)
			}

			// we will start the uninstallation process after ensuring that the envoy gateway normal service.
			relName := "eg"
			relNamespace := "envoy-gateway-system"
			options.DefaultConfigFlags.Namespace = ptr.To(relNamespace)

			ht := helm.NewPackageTool()
			if err = ht.Setup(); err != nil {
				t.Errorf("failed to setup of packageTool: %v", err)
			}

			t.Log("start uninstall envoy gateway resources")
			if err := ht.RunUninstall(&helm.PackageOptions{
				ReleaseName: relName,
			}); err != nil {
				t.Errorf("failed to uninstall envoy-gateway: %v", err)
			}
			t.Log("success to uninstall envoy gateway resources")
		})
		t.Run("Install Envoy Gateway should succeed", func(t *testing.T) {
			ns := "gateway-lifecycle-infra"
			routeNN := types.NamespacedName{
				Name:      "http-backend-eg-lifecycle",
				Namespace: ns,
			}
			gwNN := types.NamespacedName{
				Name:      "lifecycle-gateway",
				Namespace: ns,
			}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig,
				suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/eg-lifecycle",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")

			relName := "eg"
			relNamespace := "envoy-gateway-system"
			options.DefaultConfigFlags.Namespace = ptr.To(relNamespace)
			lastVersionTag := os.Getenv("last_version_tag")
			if lastVersionTag == "" {
				lastVersionTag = "v0.0.0-latest" // Default version tag if not specified
			}

			ht := helm.NewPackageTool()
			if err := ht.Setup(); err != nil {
				t.Errorf("failed to setup of packageTool: %v", err)
			}

			t.Log("start install envoy gateway resource")
			if err := ht.RunInstall(&helm.PackageOptions{
				SkipCRD:          true,
				ReleaseName:      relName,
				ReleaseNamespace: relNamespace,
				Version:          lastVersionTag,
				Timeout:          time.Minute * 5,
			}); err != nil {
				t.Errorf("failed to install envoy-gateway: %v", err)
			}

			// finally, ensure that the envoy-gateway is in normal service.
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response before the install process started: %v", err)
			}
			if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response before the uninstall and install process started: %v", err)
			}
			t.Log("success to install envoy gateway resources")
		})
	},
}
