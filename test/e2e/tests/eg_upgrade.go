// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/kube"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
	"sigs.k8s.io/gateway-api/pkg/consts"

	"github.com/envoyproxy/gateway/internal/cmd/options"
	"github.com/envoyproxy/gateway/internal/utils/helm"
)

var EGUpgradeTest = suite.ConformanceTest{
	ShortName:   "EGUpgrade",
	Description: "Upgrading from the last eg version should not lead to failures",
	Manifests:   []string{},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Upgrade from an older eg release should succeed", func(t *testing.T) {
			chartPath := "../../../charts/gateway-helm"
			relName := "eg"
			depNS := "envoy-gateway-system"
			lastVersionTag := "v1.5.1" //  the latest prior release

			t.Logf("Upgrading from version: %s", lastVersionTag)

			// Uninstall the current version of EG
			relNamespace := "envoy-gateway-system"
			options.DefaultConfigFlags.Namespace = ptr.To(relNamespace)

			// use values file for e2e test
			valuesFile := "../../config/helm/default.yaml"
			if IsGatewayNamespaceMode() {
				valuesFile = "../../config/helm/gateway-namespace-mode.yaml"
			}

			ht := helm.NewPackageTool(valuesFile)
			if err := ht.Setup(); err != nil {
				t.Errorf("failed to setup of packageTool: %v", err)
			}

			// cleanup some resources to avoid finalizer deadlocks when deleting CRDs
			t.Log("cleanup test data")
			cleanUpResources(suite.Client, t)

			// Uninstall the version deployed for e2e test from the branch
			// Make sure to remove all CRDs and CRs, as these may be incompatible with a latestVersion
			t.Log("start uninstall envoy gateway resources")
			if err := ht.RunUninstall(&helm.PackageOptions{
				ReleaseName: relName,
				Wait:        true,
				Timeout:     suite.TimeoutConfig.NamespacesMustBeReady,
			}); err != nil {
				t.Fatalf("failed to uninstall envoy-gateway: %v", err)
			}

			err := deleteChartCRDsFromPath(depNS, chartPath, t, suite.TimeoutConfig.NamespacesMustBeReady)
			if err != nil {
				t.Fatalf("Failed to delete chart CRDs: %s", err.Error())
			}

			t.Log("success to uninstall envoy gateway resources")

			// Install latest version
			if err := ht.RunInstall(&helm.PackageOptions{
				Version:          lastVersionTag,
				ReleaseName:      relName,
				ReleaseNamespace: relNamespace,
				Wait:             true,
				Timeout:          suite.TimeoutConfig.NamespacesMustBeReady,
			}); err != nil {
				t.Fatalf("failed to  install envoy-gateway: %v", err)
			}

			// Apply base and test manifests deleted during uninstall phase; Manifests in current branch must be compatible with the latestVersion
			// Since we're applying the GWC now, we must set the controller name, as the suite cannot identify it from the GWC status
			suite.ControllerName = "gateway.envoyproxy.io/gatewayclass-controller"
			suite.Applier.ControllerName = "gateway.envoyproxy.io/gatewayclass-controller"
			suite.Applier.GatewayClass = suite.GatewayClassName
			suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, suite.BaseManifests, suite.Cleanup)

			// verify latestVersion is working
			kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{depNS})

			// let's make sure the gateway is up and running
			ns := "gateway-upgrade-infra"
			gwNN := types.NamespacedName{Name: "ha-gateway", Namespace: ns}
			_, err = kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.GatewayRef{
				NamespacedName: gwNN,
			})
			require.NoErrorf(t, err, "timed out waiting for Gateway address to be assigned")

			// Apply the test manifests
			for _, manifestLocation := range []string{"testdata/eg-upgrade.yaml"} {
				tlog.Logf(t, "Applying %s", manifestLocation)
				suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, manifestLocation, true)
			}

			// wait for everything to startup
			routeNN := types.NamespacedName{Name: "http-backend-eg-upgrade", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{depNS})
			expectOkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/eg-upgrade",
				},
				Response: http.Response{
					StatusCodes: []int{200},
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

			// Perform helm upgrade of EG
			t.Log("Attempting to upgrade to current version of eg deployment")
			// TODO: when helm tool supports upgrade-from-source action, use it
			err = upgradeChartFromPath(relName, depNS, chartPath, suite.TimeoutConfig.NamespacesMustBeReady, t)
			if err != nil {
				t.Fatalf("Failed to upgrade the release: %s", err.Error())
			}

			// wait for installation
			kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{depNS})

			t.Log("Confirm routing works after upgrading Envoy Gateway with current main version")
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectOkResp)
			// fire the rest of requests
			if err := GotExactExpectedResponse(t, 5, suite.RoundTripper, expectOkReq, expectOkResp); err != nil {
				t.Errorf("failed to get expected response for the first three requests: %v", err)
			}
		})
	},
}

func cleanUpResources(c client.Client, t *testing.T) {
	gvks := []schema.GroupVersionKind{
		{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"},
		{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "Gateway"},
		{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "GatewayClass"},
	}
	for _, gvk := range gvks {
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(gvk)

		list := &unstructured.UnstructuredList{}
		list.SetGroupVersionKind(gvk)

		if err := c.List(context.Background(), list); err != nil {
			t.Fatalf("failed fetching %s: %v", obj.GetObjectKind(), err)
		}

		for _, o := range list.Items {
			tlog.Logf(t, "deleting %s: %s/%s", o.GetObjectKind(), o.GetNamespace(), o.GetName())
			if err := c.Delete(context.Background(), &o); err != nil {
				if !kerrors.IsNotFound(err) {
					t.Fatalf("error deleting %s: %s/%s : %v", o.GetObjectKind(), o.GetNamespace(), o.GetName(), err)
				}
			}
		}

		if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
			func(ctx context.Context) (bool, error) {
				err := c.List(ctx, list)
				if err != nil {
					return false, nil
				}

				if len(list.Items) > 0 {
					tlog.Logf(t, "Waiting for deletion of %d %s", len(list.Items), gvk.String())
					return false, nil
				}

				return true, nil
			}); err != nil {
			t.Fatalf("failed to wait for %s deletion: %v", gvk.String(), err)
		}
	}
}

func upgradeChartFromPath(relName, relNamespace, chartPath string, timeout time.Duration, t *testing.T) error {
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
	gatewayChart, err := loader.Load(chartPath)
	if err != nil {
		return err
	}

	err = migrateChartCRDs(actionConfig, gatewayChart, timeout)
	if err != nil {
		return err
	}

	err = updateChartCRDs(actionConfig, gatewayChart, timeout)
	if err != nil {
		return err
	}

	_, err = upgrade.Run(relName, gatewayChart, map[string]any{})
	if err != nil {
		return err
	}
	return nil
}

func updateChartCRDs(actionConfig *action.Configuration, gatewayChart *chart.Chart, timeout time.Duration) error {
	crds, err := extractCRDs(actionConfig, gatewayChart)
	if err != nil {
		return err
	}

	// Create all CRDs in the envoy gateway chart
	result, err := actionConfig.KubeClient.Update(crds, crds, false)
	if err != nil {
		return fmt.Errorf("failed to create CRDs: %w", err)
	}

	// We invalidate the cache and let it rebuild the cache,
	// at which point no more updated CRDs exist
	client, err := actionConfig.RESTClientGetter.ToDiscoveryClient()
	if err != nil {
		return err
	}
	client.Invalidate()

	// Wait the specified amount of time for the resource to be recognized by the cluster
	if err := actionConfig.KubeClient.Wait(result.Created, timeout); err != nil {
		return err
	}
	_, err = client.ServerGroups()
	return err
}

func migrateChartCRDs(actionConfig *action.Configuration, gatewayChart *chart.Chart, _ time.Duration) error {
	crds, err := extractCRDs(actionConfig, gatewayChart)
	if err != nil {
		return err
	}

	// https: //gateway-api.sigs.k8s.io/guides/?h=upgrade#v12-upgrade-notes
	storedVersionsMap := map[string]string{
		"referencegrants.gateway.networking.k8s.io": "v1beta1",
		"grpcroutes.gateway.networking.k8s.io":      "v1",
	}

	restCfg, err := actionConfig.RESTClientGetter.ToRESTConfig()
	if err != nil {
		return err
	}

	cli, err := client.New(restCfg, client.Options{})
	if err != nil {
		return err
	}

	for _, crd := range crds {
		storedVersion, ok := storedVersionsMap[crd.Name]
		if !ok {
			continue
		}

		newVersion, err := getGWAPIVersion(crd.Object)
		if err != nil {
			return err
		}

		if strings.HasPrefix(newVersion, "v1.2.0") {
			existingCRD := &apiextensionsv1.CustomResourceDefinition{}
			err := cli.Get(context.Background(), types.NamespacedName{Name: crd.Name}, existingCRD)
			if kerrors.IsNotFound(err) {
				continue
			}
			if err != nil {
				return fmt.Errorf("failed to get CRD: %s", err.Error())
			}

			existingCRD.Status.StoredVersions = []string{storedVersion}

			if err := cli.Status().Patch(context.Background(), existingCRD, client.MergeFrom(existingCRD)); err != nil {
				return fmt.Errorf("failed to patch CRD: %s", err.Error())
			}
		}
	}

	return nil
}

func deleteChartCRDsFromPath(relNamespace, chartPath string, t *testing.T, timeout time.Duration) error {
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(cli.New().RESTClientGetter(), relNamespace, "secret", t.Logf); err != nil {
		return err
	}

	// Load the chart from a local directory.
	gatewayChart, err := loader.Load(chartPath)
	if err != nil {
		return err
	}

	crds, err := extractCRDs(actionConfig, gatewayChart)
	if err != nil {
		return err
	}

	if _, errors := actionConfig.KubeClient.Delete(crds); len(errors) != 0 {
		return fmt.Errorf("failed to delete CRDs error: %s", util.MultipleErrors("", errors))
	}

	if kubeClient, ok := actionConfig.KubeClient.(kube.InterfaceExt); ok {
		if err := kubeClient.WaitForDelete(crds, timeout); err != nil {
			return fmt.Errorf("failed to wait for crds deletion: %s", err.Error())
		}
	}

	return nil
}

func getGWAPIVersion(object runtime.Object) (string, error) {
	accessor, err := meta.Accessor(object)
	if err != nil {
		return "", err
	}
	annotations := accessor.GetAnnotations()
	newVersion, ok := annotations[consts.BundleVersionAnnotation]
	if ok {
		return newVersion, nil
	}
	return "", fmt.Errorf("failed to determine Gateway API CRD version: %v", annotations)
}

// extractCRDs Extract the CRDs part of the chart
func extractCRDs(config *action.Configuration, ch *chart.Chart) ([]*resource.Info, error) {
	crdResInfo := make([]*resource.Info, 0, len(ch.CRDObjects()))

	for _, crd := range ch.CRDObjects() {
		resInfo, err := config.KubeClient.Build(bytes.NewBufferString(string(crd.File.Data)), false)
		if err != nil {
			return nil, err
		}
		crdResInfo = append(crdResInfo, resInfo...)
	}

	return crdResInfo, nil
}
