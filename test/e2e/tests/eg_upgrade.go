// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"net/url"
	"os"
	"testing"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

var EGUpgradeTest = suite.ConformanceTest{
	ShortName:   "EGUpgrade",
	Description: "Upgrading from the last eg version should not lead to failures",
	Manifests:   []string{"testdata/eg-upgrade.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Upgrade from an older eg release should succeed", func(t *testing.T) {
			relName := "eg"
			depNS := "envoy-gateway-system"
			lastVersionTag := os.Getenv("last_version_tag")
			if lastVersionTag == "" {
				lastVersionTag = "v1.0.0" // Default version tag if not specified
			}

			ns := "gateway-upgrade-infra"
			routeNN := types.NamespacedName{Name: "http-backend-eg-upgrade", Namespace: ns}
			gwNN := types.NamespacedName{Name: "ha-gateway", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			reqURL := url.URL{Scheme: "http", Host: http.CalculateHost(t, gwAddr, "http"), Path: "/eg-upgrade"}
			kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{depNS})
			expectOkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/eg-upgrade",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			expectOkReq := http.MakeRequest(t, &expectOkResp, gwAddr, "HTTP", "http")
			t.Log("Confirm routing works before starting to validate the eg upgrade flow")
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectOkResp)
			// fire the rest of requests
			if err := GotExactExpectedResponse(t, 5, suite.RoundTripper, expectOkReq, expectOkResp); err != nil {
				t.Errorf("failed to get expected response for the first three requests: %v", err)
			}

			t.Log("Validate route to backend is functional", reqURL.String())

			// Uninstall the current version of EG
			err := helmUninstall(relName, depNS, t)
			if err != nil {
				t.Fatalf("Failed to upgrade the release: %s", err.Error())
			}

			t.Log("Install the last version tag")
			err = helmInstall(relName, depNS, lastVersionTag, suite.TimeoutConfig.NamespacesMustBeReady, t)
			if err != nil {
				t.Fatalf("Failed to upgrade the release: %s", err.Error())
			}

			// wait for everything to startup
			kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{depNS})

			t.Log("Attempting to upgrade to current version of eg deployment")
			err = helmUpgradeChartFromPath(relName, depNS, "../../../charts/gateway-helm", suite.TimeoutConfig.NamespacesMustBeReady, t)
			if err != nil {
				t.Fatalf("Failed to upgrade the release: %s", err.Error())
			}

			kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{depNS})

			t.Log("Confirm routing works after upgrade the eg with current main version")
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectOkResp)
			// fire the rest of requests
			if err := GotExactExpectedResponse(t, 5, suite.RoundTripper, expectOkReq, expectOkResp); err != nil {
				t.Errorf("failed to get expected response for the first three requests: %v", err)
			}
		})
	},
}

func helmUpgradeChartFromPath(relName, relNamespace, chartPath string, timeout time.Duration, t *testing.T) error {
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(cli.New().RESTClientGetter(), relNamespace, "secret", t.Logf); err != nil {
		return err
	}

	// Set installation options.
	upgrade := action.NewUpgrade(actionConfig)
	upgrade.Namespace = relNamespace
	upgrade.WaitForJobs = true
	upgrade.Timeout = timeout

	// Load the chart from a local directory.
	chart, err := loader.Load(chartPath)
	if err != nil {
		return err
	}

	// Run the installation.
	values := map[string]interface{}{
		"deployment": map[string]interface{}{
			"envoyGateway": map[string]interface{}{
				"imagePullPolicy": "IfNotPresent",
			},
		},
	}
	_, err = upgrade.Run(relName, chart, values)
	if err != nil {
		return err
	}
	return nil
}

func helmInstall(relName, relNamespace string, tag string, timeout time.Duration, t *testing.T) error {
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(cli.New().RESTClientGetter(), relNamespace, "secret", t.Logf); err != nil {
		return err
	}

	// Set installation options.
	install := action.NewInstall(actionConfig)
	install.ReleaseName = relName
	install.Namespace = relNamespace
	install.CreateNamespace = true
	install.Version = tag
	install.WaitForJobs = true
	install.Timeout = timeout

	registryClient, err := registry.NewClient()
	if err != nil {
		return err
	}
	install.SetRegistryClient(registryClient)
	// todo we need to explicitly reinstall the CRDs
	chartPath, err := install.LocateChart("oci://docker.io/envoyproxy/gateway-helm", cli.New())
	if err != nil {
		return err
	}
	// Load the chart from a local directory.
	chart, err := loader.Load(chartPath)
	if err != nil {
		return err
	}
	// Run the installation.
	_, err = install.Run(chart, nil)
	if err != nil {
		return err
	}
	return nil
}

func helmUninstall(relName, relNamespace string, t *testing.T) error {
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(cli.New().RESTClientGetter(), relNamespace, "secret", t.Logf); err != nil {
		return err
	}
	uninstall := action.NewUninstall(actionConfig)
	_, err := uninstall.Run(relName) // nil can be replaced with any values you wish to override
	if err != nil {
		return err
	}
	return nil
}
