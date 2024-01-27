// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	nethttp "net/http"
	"net/http/httputil"
	"regexp"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, PreserveCaseTest)
}

// Copied from the conformance suite because it's needed in casePreservingRoundTrip
var startLineRegex = regexp.MustCompile(`(?m)^`)

func formatDump(data []byte, prefix string) string {
	data = startLineRegex.ReplaceAllLiteral(data, []byte(prefix))
	return string(data)
}

// Copied from the conformance suite and modified to not normalize headers before sending them
// to the remote side.
// The default HTTP client implementation in Golang also automatically normalizes received
// headers as they are parsed , so it's not possible to verify that returned headers were not normalized
func casePreservingRoundTrip(request roundtripper.Request, transport nethttp.RoundTripper, suite *suite.ConformanceTestSuite) (map[string]any, error) {
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
		for name, value := range request.Headers {
			req.Header[name] = value
		}
	}
	if suite.Debug {
		var dump []byte
		dump, err = httputil.DumpRequestOut(req, true)
		if err != nil {
			return nil, err
		}

		fmt.Printf("Sending Request:\n%s\n\n", formatDump(dump, "< "))
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

		fmt.Printf("Received Response:\n%s\n\n", formatDump(dump, "< "))
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

var PreserveCaseTest = suite.ConformanceTest{
	ShortName:   "Preserve Case",
	Description: "Preserve header cases",
	Manifests:   []string{"testdata/preserve-case.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("should preserve header cases in both directions", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "preserve-case", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			// Can't use the standard method for checking the response, since the remote side isn't the
			// conformance echo server and it returns a differently formatted response.
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/preserve?headers=ReSpOnSeHeAdEr",
					Headers: map[string]string{
						"SpEcIaL": "Header",
					},
				},
				Namespace: ns,
			}

			var rt nethttp.RoundTripper
			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			respBody, err := casePreservingRoundTrip(req, rt, suite)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if _, found := respBody["SpEcIaL"]; !found {
				t.Errorf("case was not preserved for test header: %+v", respBody)
			}
		})

	},
}
