// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net"
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

// sendHTTP10Request sends a raw HTTP/1.0 request to gwAddr, returning the parsed JSON
// response body. Go's net/http always uses HTTP/1.1, so we use a raw TCP connection.
// host is sent as the Host header; pass "" to omit it (tests useDefaultHost injection).
func sendHTTP10Request(gwAddr, path, host string) (map[string]string, error) {
	conn, err := net.DialTimeout("tcp", gwAddr, 5*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

	req := fmt.Sprintf("GET %s HTTP/1.0\r\n", path)
	if host != "" {
		req += fmt.Sprintf("Host: %s\r\n", host)
	}
	req += "\r\n"
	if _, err = fmt.Fprint(conn, req); err != nil {
		return nil, err
	}

	resp, err := nethttp.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != nethttp.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var headers map[string]string
	if err = json.Unmarshal(body, &headers); err != nil {
		return nil, fmt.Errorf("unexpected body: %w", err)
	}
	return headers, nil
}

// checkHTTP10Proxied polls until an HTTP/1.0 request to path succeeds (200 OK),
// proving Envoy proxies downstream HTTP/1.0 to the backend without needing
// acceptHttp10 on the upstream cluster.
func checkHTTP10Proxied(t *testing.T, s *suite.ConformanceTestSuite, gwAddr, path, host string) bool {
	t.Helper()
	var ok bool
	_ = wait.PollUntilContextTimeout(t.Context(), time.Second, s.TimeoutConfig.MaxTimeToConsistency, true, func(_ context.Context) (bool, error) {
		_, err := sendHTTP10Request(gwAddr, path, host)
		if err != nil {
			tlog.Logf(t, "HTTP/1.0 request failed: %v", err)
			return false, nil
		}
		ok = true
		return true, nil
	})
	return ok
}

// checkHTTP10DefaultHostInjected sends an HTTP/1.0 request without a Host header and
// asserts the backend received Host: wantHost — injected by the HCM default_host_for_http_10.
// This proves useDefaultHost works at the listener level and no cluster-level equivalent is needed.
func checkHTTP10DefaultHostInjected(t *testing.T, s *suite.ConformanceTestSuite, gwAddr, path, wantHost string) bool {
	t.Helper()
	var ok bool
	_ = wait.PollUntilContextTimeout(t.Context(), time.Second, s.TimeoutConfig.MaxTimeToConsistency, true, func(_ context.Context) (bool, error) {
		// send with no Host header — HCM should inject wantHost
		headers, err := sendHTTP10Request(gwAddr, path, "")
		if err != nil {
			tlog.Logf(t, "HTTP/1.0 hostless request failed: %v", err)
			return false, nil
		}
		if got := headers["Host"]; got != wantHost {
			tlog.Logf(t, "Host header: got %q, want %q", got, wantHost)
			return false, nil
		}
		ok = true
		return true, nil
	})
	return ok
}

var PreserveCase = suite.ConformanceTest{
	ShortName:   "PreserveCase",
	Description: "Verify CTP http1 preserveHeaderCase seeds the backend codec, BTP http1 overrides it, and HTTP/1.0 proxying works without cluster codec fields",
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

		// Phase 1: CTP http1 preserveHeaderCase — both listener and backend codec preserve case.
		// The backend echoes the request headers it received; the mixed-case header should appear preserved.
		t.Run("CTP http1 preserveHeaderCase seeds backend codec", func(t *testing.T) {
			ClientTrafficPolicyMustBeAccepted(t, s.Client, types.NamespacedName{Name: "preserve-case", Namespace: ns}, s.ControllerName, ancestorRef)
			require.True(t, checkHeaderCasePreserved(t, s, gwAddr, "/preserve", ns), "expected header case to be preserved with CTP http1 field")
		})

		// Phase 2: add BTP http1 preserveHeaderCase: false alongside the existing CTP.
		// BTP overrides the cluster codec — the backend now receives normalized headers,
		// proving BTP controls the backend HTTP/1 codec independently of CTP.
		t.Run("BTP http1 overrides CTP-seeded backend codec", func(t *testing.T) {
			s.Applier.MustApplyWithCleanup(t, s.Client, s.TimeoutConfig, "testdata/preserve-case-new-http1-fields.yaml", true)

			BackendTrafficPolicyMustBeAccepted(t, s.Client, types.NamespacedName{Name: "preserve-case-btp-override", Namespace: ns}, s.ControllerName, ancestorRef)

			require.True(t, checkHeaderCaseNotPreserved(t, s, gwAddr, "/preserve", ns), "expected header case to NOT be preserved at backend when BTP overrides CTP seed")
		})

		// Phase 3: HTTP/1.0 downstream request is proxied to the HTTP/1.0 backend successfully.
		// The upstream cluster has no acceptHttp10/defaultHostForHttp10 set — proving those
		// cluster fields are not needed for HTTP/1.0 proxying to work end-to-end.
		t.Run("HTTP/1.0 proxied to HTTP/1.0 backend without cluster codec fields", func(t *testing.T) {
			ClientTrafficPolicyMustBeAccepted(t, s.Client, types.NamespacedName{Name: "preserve-case", Namespace: ns}, s.ControllerName, ancestorRef)
			require.True(t, checkHTTP10Proxied(t, s, gwAddr, "/http10", "http10.example.com"), "expected HTTP/1.0 request to be proxied successfully to HTTP/1.0 backend")
		})

		// Phase 4: HTTP/1.0 request without a Host header — CTP useDefaultHost injects
		// the route hostname ("http10.example.com"). The backend echoes received headers,
		// so we can assert the injected Host arrived upstream.
		t.Run("CTP http1.http10.useDefaultHost injects Host for hostless HTTP/1.0 requests", func(t *testing.T) {
			ClientTrafficPolicyMustBeAccepted(t, s.Client, types.NamespacedName{Name: "preserve-case", Namespace: ns}, s.ControllerName, ancestorRef)
			require.True(t, checkHTTP10DefaultHostInjected(t, s, gwAddr, "/http10", "http10.example.com"), "expected HCM to inject Host: http10.example.com for hostless HTTP/1.0 request")
		})
	},
}
