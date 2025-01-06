// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	nethttp "net/http"
	"net/http/httputil"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, CompressionTest)
}

var CompressionTest = suite.ConformanceTest{
	ShortName:   "Compression",
	Description: "Test response compression on HTTPRoute",
	Manifests:   []string{"testdata/compression.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("HTTPRoute with compression", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "compression", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "compression", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/compression",
					Headers: map[string]string{
						"Accept-encoding": "gzip",
					},
				},
				Response: http.Response{
					StatusCode: 200,
					Headers: map[string]string{
						"content-encoding": "gzip",
					},
				},
				Namespace: ns,
			}
			roundTripper := &CompressionRoundTripper{Debug: suite.Debug, TimeoutConfig: suite.TimeoutConfig}
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, roundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("HTTPRoute without compression", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "no-compression", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "compression", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/no-compression",
					Headers: map[string]string{
						"Accept-encoding": "gzip",
					},
				},
				Response: http.Response{
					StatusCode:    200,
					AbsentHeaders: []string{"content-encoding"},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}

// CompressionRoundTripper implements roundtripper.RoundTripper and adds support for gzip encoding.
type CompressionRoundTripper struct {
	Debug             bool
	TimeoutConfig     config.TimeoutConfig
	CustomDialContext func(context.Context, string, string) (net.Conn, error)
}

func (d *CompressionRoundTripper) CaptureRoundTrip(request roundtripper.Request) (*roundtripper.CapturedRequest, *roundtripper.CapturedResponse, error) {
	return d.defaultRoundTrip(request, &nethttp.Transport{})
}

func (d *CompressionRoundTripper) defaultRoundTrip(request roundtripper.Request, transport nethttp.RoundTripper) (*roundtripper.CapturedRequest, *roundtripper.CapturedResponse, error) {
	client := &nethttp.Client{}

	client.Transport = transport

	method := "GET"
	if request.Method != "" {
		method = request.Method
	}
	ctx, cancel := context.WithTimeout(context.Background(), d.TimeoutConfig.RequestTimeout)
	defer cancel()
	req, err := nethttp.NewRequestWithContext(ctx, method, request.URL.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	if request.Host != "" {
		req.Host = request.Host
	}

	if request.Headers != nil {
		for name, value := range request.Headers {
			req.Header.Set(name, value[0])
		}
	}

	if d.Debug {
		var dump []byte
		dump, err = httputil.DumpRequestOut(req, true)
		if err != nil {
			return nil, nil, err
		}

		tlog.Logf(request.T, "Sending Request:\n%s\n\n", formatDump(dump))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if d.Debug {
		var dump []byte
		dump, err = httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, nil, err
		}

		tlog.Logf(request.T, "Received Response:\n%s\n\n", formatDump(dump))
	}

	cReq := &roundtripper.CapturedRequest{}
	var body []byte
	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, nil, err
		}
		defer reader.Close()
		if body, err = io.ReadAll(reader); err != nil {
			return nil, nil, err
		}
	} else {
		if body, err = io.ReadAll(resp.Body); err != nil {
			return nil, nil, err
		}
	}

	// we cannot assume the response is JSON
	if resp.Header.Get("Content-type") == "application/json" {
		err = json.Unmarshal(body, cReq)
		if err != nil {
			return nil, nil, fmt.Errorf("unexpected error reading response: %w", err)
		}
	} else {
		cReq.Method = method // assume it made the right request if the service being called isn't echoing
	}

	cRes := &roundtripper.CapturedResponse{
		StatusCode:    resp.StatusCode,
		ContentLength: resp.ContentLength,
		Protocol:      resp.Proto,
		Headers:       resp.Header,
	}

	if resp.TLS != nil {
		cRes.PeerCertificates = resp.TLS.PeerCertificates
	}

	return cReq, cRes, nil
}
