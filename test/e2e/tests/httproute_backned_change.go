// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"net/url"
	"testing"
	"time"

	"fortio.org/fortio/periodic"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteBackendChangeTest)
}

var HTTPRouteBackendChangeTest = suite.ConformanceTest{
	ShortName:   "HTTPRouteBackendChange",
	Description: "HTTPRoute with backend change",
	Manifests: []string{
		"testdata/httproute-backend.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		if !XDSNameSchemeV2() {
			t.Skip("This test is only relevant for v2 xDS scheme, skipping")
		}

		ns := ConformanceInfraNamespace
		routeNN := types.NamespacedName{Name: "backend-changed", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false,
			routeNN, types.NamespacedName{Name: "backends", Namespace: ns},
		)

		// Make sure backend is ready.
		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
			Request: http.Request{
				Path: "/v1",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Backend:   "infra-backend-v1",
			Namespace: ns,
		})

		done := make(chan struct{})
		go func() {
			for i := range 10 {
				location := ""
				if i%2 == 0 {
					location = "testdata/httproute-backend-filter.yaml"
				} else {
					location = "testdata/httproute-backend.yaml"
				}
				tlog.Logf(t, "Apply file %s", location)
				suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, location, false)
				time.Sleep(time.Second)
			}
			done <- struct{}{}
		}()

		reqURL := url.URL{Scheme: "http", Host: http.CalculateHost(t, gwAddr, "http"), Path: "/backend-changed"}

		// can be used to abort the test after deployment restart is complete or failed
		aborter := periodic.NewAborter()
		// will contain indication on success or failure of load test
		loadSuccess := make(chan bool)

		tlog.Logf(t, "Starting load generation")
		// Run load async and continue to restart deployment
		go runLoadAndWait(t, &suite.TimeoutConfig, loadSuccess, aborter, reqURL.String(), 0)
		<-done
		tlog.Logf(t, "Stopping load generation and collecting results")
		aborter.Abort(false) // abort the load either way

		// Wait for the goroutine to finish
		result := <-loadSuccess
		if !result {
			tlog.Errorf(t, "Load test failed")
		}
	},
}
