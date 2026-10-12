// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	nethttp "net/http"
	"regexp"
	"strings"
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
	ConformanceTests = append(ConformanceTests, HTTP1Settings)
}

var startLineRegex = regexp.MustCompile(`(?m)^`)

func formatDump(data []byte) string {
	data = startLineRegex.ReplaceAllLiteral(data, []byte("< "))
	return string(data)
}

// teeConn wraps a net.Conn and tees all bytes read from it into buf, so the
// raw wire response (including original header-name case) is available after
// the HTTP exchange completes.
type teeConn struct {
	net.Conn
	buf bytes.Buffer
}

func (c *teeConn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	c.buf.Write(b[:n])
	return n, err
}

// teeTransport is an http.RoundTripper that wraps an inner transport's dial
// with a teeConn so the raw response bytes are captured before net/http
// canonicalizes header names.
type teeTransport struct {
	inner *nethttp.Transport
	tc    *teeConn // set by DialContext hook; nil until first dial
}

func newTeeTransport(inner *nethttp.Transport) *teeTransport {
	if inner == nil {
		inner = &nethttp.Transport{}
	}
	t := &teeTransport{}
	// Clone so we don't mutate the caller's transport.
	cloned := inner.Clone()
	origDial := cloned.DialContext
	cloned.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		var (
			conn net.Conn
			err  error
		)
		if origDial != nil {
			conn, err = origDial(ctx, network, addr)
		} else {
			conn, err = (&net.Dialer{}).DialContext(ctx, network, addr)
		}
		if err != nil {
			return nil, err
		}
		t.tc = &teeConn{Conn: conn}
		return t.tc, nil
	}
	t.inner = cloned
	return t
}

func (t *teeTransport) RoundTrip(req *nethttp.Request) (*nethttp.Response, error) {
	return t.inner.RoundTrip(req)
}

// RawResponse returns the raw response bytes captured during the last RoundTrip.
func (t *teeTransport) RawResponse() string {
	if t.tc == nil {
		return ""
	}
	return t.tc.buf.String()
}

// customTransportRoundTrip sends an HTTP request using the provided transport
// (nil uses the default) and returns the parsed JSON response body. It is the
// standard way to send requests in tests that need a custom transport (e.g.
// custom TLS config, response-byte capture via teeTransport).
func customTransportRoundTrip(request *roundtripper.Request, transport nethttp.RoundTripper, s *suite.ConformanceTestSuite) (map[string]any, error) {
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	client := &nethttp.Client{Transport: transport}

	ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.RequestTimeout)
	defer cancel()
	req, err := nethttp.NewRequestWithContext(ctx, "GET", request.URL.String(), nil)
	if err != nil {
		return nil, err
	}
	if request.Host != "" {
		req.Host = request.Host
	}
	for k, v := range request.Headers {
		req.Header[k] = v
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result := map[string]any{}
	if err = json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unexpected response body: %w", err)
	}
	return result, nil
}

// checkHeaderCasePreserved polls until both directions show case preservation:
//   - request: "SpEcIaL: Header" is sent to Envoy and forwarded as-is to the backend,
//     which echoes it in the body as "SpEcIaL" (upstream cluster codec preserves case)
//   - response: the backend sets "ReSpOnSeHeAdEr: PrEsEnT" and Envoy forwards it
//     with case intact on the wire (verified via raw TCP bytes, before Go's HTTP
//     parser canonicalizes received headers)
func checkHeaderCasePreserved(t *testing.T, s *suite.ConformanceTestSuite, gwAddr, path string) bool {
	t.Helper()
	var preserved bool
	_ = wait.PollUntilContextTimeout(t.Context(), time.Second, s.TimeoutConfig.MaxTimeToConsistency, true, func(_ context.Context) (bool, error) {
		tt := newTeeTransport(nil)
		expectedResponse := http.ExpectedResponse{
			Request: http.Request{
				Host:    "preserve-case.example.com",
				Path:    path + "?headers=ReSpOnSeHeAdEr",
				Headers: map[string]string{"SpEcIaL": "Header"},
			},
		}
		req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
		respBody, err := customTransportRoundTrip(&req, tt, s)
		if err != nil {
			tlog.Logf(t, "request failed: %v", err)
			return false, nil
		}
		if _, found := respBody["SpEcIaL"]; !found {
			tlog.Logf(t, "SpEcIaL not echoed in body: %+v", respBody)
			return false, nil
		}
		if strings.Contains(tt.RawResponse(), "ReSpOnSeHeAdEr: PrEsEnT") {
			preserved = true
			return true, nil
		}
		tlog.Logf(t, "rawResponse=%q", tt.RawResponse())
		return false, nil
	})
	return preserved
}

// checkHeaderCaseNotPreserved polls until both directions confirm the BTP overrides the CTP codec:
//   - request: "SpEcIaL: Header" is still forwarded preserved (shared cluster retains CTP codec),
//     so the body echo WILL contain "SpEcIaL" — this confirms the request direction is unaffected
//   - response: the backend sets "ReSpOnSeHeAdEr: PrEsEnT" but the BTP disables preserve_case
//     for this route, so Envoy lowercases it to "responseheader: PrEsEnT" on the wire
func checkHeaderCaseNotPreserved(t *testing.T, s *suite.ConformanceTestSuite, gwAddr, path string) bool {
	t.Helper()
	var notPreserved bool
	_ = wait.PollUntilContextTimeout(t.Context(), time.Second, s.TimeoutConfig.MaxTimeToConsistency, true, func(_ context.Context) (bool, error) {
		tt := newTeeTransport(nil)
		expectedResponse := http.ExpectedResponse{
			Request: http.Request{
				Host:    "preserve-case.example.com",
				Path:    path + "?headers=ReSpOnSeHeAdEr",
				Headers: map[string]string{"SpEcIaL": "Header"},
			},
		}
		req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
		respBody, err := customTransportRoundTrip(&req, tt, s)
		if err != nil {
			tlog.Logf(t, "request failed: %v", err)
			return false, nil
		}
		if _, found := respBody["SpEcIaL"]; !found {
			tlog.Logf(t, "SpEcIaL missing from body (unexpected — CTP codec should still preserve request headers): %+v", respBody)
			return false, nil
		}
		if strings.Contains(tt.RawResponse(), "responseheader: PrEsEnT") {
			notPreserved = true
			return true, nil
		}
		tlog.Logf(t, "rawResponse=%q", tt.RawResponse())
		return false, nil
	})
	return notPreserved
}

// sendHTTP10Request sends a raw HTTP/1.0 request over TCP. Go's net/http always
// upgrades to HTTP/1.1, so raw TCP is required. Pass host="" to omit the Host
// header (for testing useDefaultHost injection).
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

// HTTP1Settings verifies CTP http1 settings on a gateway-scoped policy:
//   - preserveHeaderCase is applied to the upstream cluster codec
//   - a BTP on a specific route overrides the CTP-seeded cluster codec
//   - HTTP/1.0 downstream requests are proxied without acceptHttp10 on the cluster
//   - useDefaultHost injects a Host for hostless HTTP/1.0 requests (dedicated gateway)
//
// Two routes share one gateway and one CTP. The "btp-override" route additionally
// has a BTP that disables preserveHeaderCase at the cluster level.
// The useDefaultHost subtest uses a dedicated gateway with a concrete hostname.
var HTTP1Settings = suite.ConformanceTest{
	ShortName:   "HTTP1Settings",
	Description: "Verify CTP http1 settings: preserveHeaderCase, enableTrailers, http10, and BTP override",
	Manifests:   []string{"testdata/http1-settings.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		routeNN := types.NamespacedName{Name: "http1-settings", Namespace: ns}
		routeBTPNN := types.NamespacedName{Name: "http1-settings-btp-override", Namespace: ns}
		routeHTTP10NN := types.NamespacedName{Name: "http1-settings-http10", Namespace: ns}

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, s.Client, s.TimeoutConfig, s.ControllerName,
			kubernetes.NewGatewayRef(gwNN), routeNN, routeBTPNN, routeHTTP10NN)

		// Dedicated gateway for useDefaultHost — needs a concrete listener hostname.
		http10GwNN := types.NamespacedName{Name: "http1-settings-http10-gtw", Namespace: ns}
		routeDefaultHostNN := types.NamespacedName{Name: "http1-settings-http10-default-host", Namespace: ns}
		http10GwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, s.Client, s.TimeoutConfig, s.ControllerName,
			kubernetes.NewGatewayRef(http10GwNN), routeDefaultHostNN)

		WaitForPods(t, s.Client, ns, map[string]string{"app": "preserve-case"}, corev1.PodRunning, &PodReady)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(ns),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		http10AncestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(ns),
			Name:      gwapiv1.ObjectName(http10GwNN.Name),
		}

		ClientTrafficPolicyMustBeAccepted(t, s.Client,
			types.NamespacedName{Name: "http1-settings", Namespace: ns},
			s.ControllerName, ancestorRef)

		ClientTrafficPolicyMustBeAccepted(t, s.Client,
			types.NamespacedName{Name: "http1-settings-http10", Namespace: ns},
			s.ControllerName, http10AncestorRef)

		BackendTrafficPolicyMustBeAccepted(t, s.Client,
			types.NamespacedName{Name: "http1-settings-btp-override", Namespace: ns},
			s.ControllerName, ancestorRef)

		t.Run("CTP preserveHeaderCase seeds the backend cluster codec", func(t *testing.T) {
			require.True(t, checkHeaderCasePreserved(t, s, gwAddr, "/preserve"),
				"expected header case to be preserved with CTP http1.preserveHeaderCase")
		})

		t.Run("BTP http1 overrides CTP-seeded backend cluster codec", func(t *testing.T) {
			require.True(t, checkHeaderCaseNotPreserved(t, s, gwAddr, "/preserve-btp-override"),
				"expected header case to NOT be preserved when BTP http1.preserveHeaderCase=false overrides CTP")
		})

		t.Run("HTTP/1.0 proxied to backend without acceptHttp10 on cluster", func(t *testing.T) {
			var ok bool
			_ = wait.PollUntilContextTimeout(t.Context(), time.Second, s.TimeoutConfig.MaxTimeToConsistency, true, func(_ context.Context) (bool, error) {
				_, err := sendHTTP10Request(gwAddr, "/http10", "http10.example.com")
				if err != nil {
					tlog.Logf(t, "HTTP/1.0 request failed: %v", err)
					return false, nil
				}
				ok = true
				return true, nil
			})
			require.True(t, ok, "expected HTTP/1.0 request to be proxied successfully")
		})

		t.Run("CTP useDefaultHost injects Host for hostless HTTP/1.0 requests", func(t *testing.T) {
			var ok bool
			_ = wait.PollUntilContextTimeout(t.Context(), time.Second, s.TimeoutConfig.MaxTimeToConsistency, true, func(_ context.Context) (bool, error) {
				headers, err := sendHTTP10Request(http10GwAddr, "/http10", "")
				if err != nil {
					tlog.Logf(t, "HTTP/1.0 hostless request failed: %v", err)
					return false, nil
				}
				if got := headers["Host"]; got != "http10.example.com" {
					tlog.Logf(t, "Host: got %q, want %q", got, "http10.example.com")
					return false, nil
				}
				ok = true
				return true, nil
			})
			require.True(t, ok, "expected HCM to inject Host: http10.example.com for hostless HTTP/1.0 request")
		})
	},
}
