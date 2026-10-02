// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	nethttp "net/http"
	"net/http/httputil"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, PreserveCase)
}

// Copied from the conformance suite because it's needed in casePreservingRoundTrip
var startLineRegex = regexp.MustCompile(`(?m)^`)

func formatDump(data []byte) string {
	data = startLineRegex.ReplaceAllLiteral(data, []byte("< "))
	return string(data)
}

// Copied from the conformance suite and modified to not normalize headers before sending them
// to the remote side.
// The default HTTP client implementation in Golang also automatically normalizes received
// headers as they are parsed , so it's not possible to verify that returned headers were not normalized
func casePreservingRoundTrip(request *roundtripper.Request, transport nethttp.RoundTripper, suite *suite.ConformanceTestSuite) (map[string]any, error) {
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	client := &nethttp.Client{}
	client.Transport = transport

	method := "GET"
	ctx, cancel := context.WithTimeout(context.Background(), suite.TimeoutConfig.RequestTimeout)
	defer cancel()
	req, err := nethttp.NewRequestWithContext(ctx, method, request.URL.String(), nil)
	if err != nil {
		return nil, err
	}
	if request.Host != "" {
		req.Host = request.Host
	}
	if request.Headers != nil {
		maps.Copy(req.Header, request.Headers)
	}
	if suite.Debug {
		var dump []byte
		dump, err = httputil.DumpRequestOut(req, true)
		if err != nil {
			return nil, err
		}

		fmt.Printf("Sending Request:\n%s\n\n", formatDump(dump))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if suite.Debug {
		var dump []byte
		dump, err = httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, err
		}

		fmt.Printf("Received Response:\n%s\n\n", formatDump(dump))
	}

	cReq := map[string]any{}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &cReq)
	if err != nil {
		return nil, fmt.Errorf("unexpected error reading response: %w", err)
	}

	return cReq, nil
}

// checkHeaderCasePreserved polls until "SpEcIaL" appears in the echoed response body,
// returning true once observed or false if the poll times out.
func checkHeaderCasePreserved(t *testing.T, s *suite.ConformanceTestSuite, gwAddr, path, ns string) bool {
	t.Helper()
	var preserved bool
	_ = wait.PollUntilContextTimeout(t.Context(), time.Second, s.TimeoutConfig.MaxTimeToConsistency, true, func(_ context.Context) (bool, error) {
		expectedResponse := http.ExpectedResponse{
			Request: http.Request{
				Path:    path + "?headers=ReSpOnSeHeAdEr",
				Headers: map[string]string{"SpEcIaL": "Header"},
			},
			Namespace: ns,
		}
		var rt nethttp.RoundTripper
		req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
		respBody, err := casePreservingRoundTrip(&req, rt, s)
		if err != nil {
			tlog.Logf(t, "request failed: %v", err)
			return false, nil
		}
		if _, found := respBody["SpEcIaL"]; found {
			preserved = true
			return true, nil
		}
		return false, nil
	})
	return preserved
}

// checkHeaderCaseNotPreserved polls until "SpEcIaL" is absent from the echoed response body,
// returning true once confirmed or false if preservation persists until the poll times out.
func checkHeaderCaseNotPreserved(t *testing.T, s *suite.ConformanceTestSuite, gwAddr, path, ns string) bool {
	t.Helper()
	var notPreserved bool
	_ = wait.PollUntilContextTimeout(t.Context(), time.Second, s.TimeoutConfig.MaxTimeToConsistency, true, func(_ context.Context) (bool, error) {
		expectedResponse := http.ExpectedResponse{
			Request: http.Request{
				Path:    path + "?headers=ReSpOnSeHeAdEr",
				Headers: map[string]string{"SpEcIaL": "Header"},
			},
			Namespace: ns,
		}
		var rt nethttp.RoundTripper
		req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
		respBody, err := casePreservingRoundTrip(&req, rt, s)
		if err != nil {
			tlog.Logf(t, "request failed: %v", err)
			return false, nil
		}
		if _, found := respBody["SpEcIaL"]; !found {
			notPreserved = true
			return true, nil
		}
		return false, nil
	})
	return notPreserved
}

var PreserveCase = suite.ConformanceTest{
	ShortName:   "PreserveCase",
	Description: "Preserve header cases using the deprecated http1 field, then clientHttp1/backendHttp1",
	Manifests:   []string{"testdata/preserve-case.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "preserve-case", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, s.Client, s.TimeoutConfig, s.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

		WaitForPods(t, s.Client, ns, map[string]string{"app": "preserve-case"}, corev1.PodRunning, &PodReady)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(ns),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}

		// Phase 1: deprecated http1 field on CTP — case should be preserved.
		t.Run("deprecated http1 field preserves header case", func(t *testing.T) {
			ClientTrafficPolicyMustBeAccepted(t, s.Client, types.NamespacedName{Name: "preserve-case", Namespace: ns}, s.ControllerName, ancestorRef)
			require.True(t, checkHeaderCasePreserved(t, s, gwAddr, "/preserve", ns), "expected header case to be preserved with deprecated http1 field")
		})

		// Phase 2: delete the deprecated CTP — case should no longer be preserved.
		t.Run("without CTP header case is not preserved", func(t *testing.T) {
			ctpNN := types.NamespacedName{Name: "preserve-case", Namespace: ns}
			existing := &egv1a1.ClientTrafficPolicy{}
			require.NoError(t, s.Client.Get(t.Context(), ctpNN, existing))
			require.NoError(t, s.Client.Delete(t.Context(), existing))
			ClientTrafficPolicyMustNotExist(t, s.Client, ctpNN)

			require.True(t, checkHeaderCaseNotPreserved(t, s, gwAddr, "/preserve", ns), "expected header case to NOT be preserved after CTP deletion")
		})

		// Phase 3: apply clientHttp1 on CTP and backendHttp1 on BTP — case should be preserved again.
		t.Run("clientHttp1 and backendHttp1 fields preserve header case", func(t *testing.T) {
			s.Applier.MustApplyWithCleanup(t, s.Client, s.TimeoutConfig, "testdata/preserve-case-new-http1-fields.yaml", true)

			ClientTrafficPolicyMustBeAccepted(t, s.Client, types.NamespacedName{Name: "preserve-case-new", Namespace: ns}, s.ControllerName, ancestorRef)
			BackendTrafficPolicyMustBeAccepted(t, s.Client, types.NamespacedName{Name: "preserve-case-new", Namespace: ns}, s.ControllerName, ancestorRef)

			require.True(t, checkHeaderCasePreserved(t, s, gwAddr, "/preserve", ns), "expected header case to be preserved with clientHttp1 and backendHttp1 fields")
		})
	},
}
