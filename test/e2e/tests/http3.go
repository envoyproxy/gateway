// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	httpv1 "net/http"
	"testing"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/qlog"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTP3Test)
}

var HTTP3Test = suite.ConformanceTest{
	ShortName:   "Http3",
	Description: "Testing http3 request",
	Manifests:   []string{"testdata/http3.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Send http3 request", func(t *testing.T) {
			namespace := "gateway-conformance-https"
			routeNN := types.NamespacedName{Name: "http3-route", Namespace: namespace}
			gwNN := types.NamespacedName{Name: "http3-gateway", Namespace: namespace}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			server := "www.example.com"
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/get",
					Headers: map[string]string{
						"User-Agent":        "quic-go HTTP/3",
						"X-Envoy-Internal":  "true",
						"X-Forwarded-Proto": "https",
					},
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: namespace,
			}

			req := http.MakeRequest(t, &expectedResponse, server+":443", "HTTPS", "https")
			req.CertPem = []byte("-----BEGIN CERTIFICATE-----\nMIIDOzCCAiOgAwIBAgIUZTLKDkhVrxfgt9megu5uiSRwRxswDQYJKoZIhvcNAQEL\nBQAwLTEVMBMGA1UECgwMZXhhbXBsZSBJbmMuMRQwEgYDVQQDDAtleGFtcGxlLmNv\nbTAeFw0yNDAzMDIxNDU3MTJaFw0yNTAzMDIxNDU3MTJaMC0xFTATBgNVBAoMDGV4\nYW1wbGUgSW5jLjEUMBIGA1UEAwwLZXhhbXBsZS5jb20wggEiMA0GCSqGSIb3DQEB\nAQUAA4IBDwAwggEKAoIBAQDJBSwf0XJkglooQbzQ6Io/M7gwmbhTjpQPX7P2/ZV6\ncdFsXBUF0X91wABtCtibv+x4if+yvPhHzBERzhjwwQitKZiewhOFoSz5ZyKT8HXd\n+Y06iRzWEnGi2i/98YiBFsG1xc5mHTLgyi6PjJDzGVdsNrL7pE8aM3R9sGrkG4PZ\n5ZWFlpbxX9eUz9dplLfLX7jKETbyKsiwcHihXAY4mdpuaUVifz35tSpnQ3P/RoNw\nNl7/gvRKqSK1a5ByYqVANV/c+O1R9MaR8kG9Cmj2PZPZm4vW/bJIs2R7z6ygGiDx\n4GyMT4MA52pk/dDkV2EtvFk5JFXoMjNPcNRGjuusRe17AgMBAAGjUzBRMB0GA1Ud\nDgQWBBQXeg9HF7TIRHUWOsF5dd0m68zbxTAfBgNVHSMEGDAWgBQXeg9HF7TIRHUW\nOsF5dd0m68zbxTAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQCS\nO7hTBUp3MjsOBGXmuuDPatz0eR05Dq4O6Etsn9KVRND0o8iULqYGPMsozkueRc0y\n/Ra1aTbNXm0n5gwNBgJeSOIOfLN0r0L5uBFSMjTPAo6qPkDzK5gXHfzpoxgDYCEy\nSuzYOIHliWoG6K2ldfMx7psMe3ZiR9SA5evOv3VKJDrrhO57niRmDhZKsADWDUFH\nyyximSiPKjFJLJ+7B4N+7TAmxDjMd2vre7qfRL/AFTc7zIkQBG+JoXbusOw1yGcm\n8IV50ZroyupxND625FgawISPYUelwS39x4jA7QNH0Tzgzp+Ao0N8Ck9H4waNcJt8\njk27Qn0mC7RsIqb981eH\n-----END CERTIFICATE-----")
			req.Server = server
			cReq, cResp, err := CaptureRoundTrip(req, gwAddr)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}
		})
	},
}

func CaptureRoundTrip(request roundtripper.Request, gwAddr string) (*roundtripper.CapturedRequest, *roundtripper.CapturedResponse, error) {
	transport, err := httpTransport(request, gwAddr)
	if err != nil {
		return nil, nil, err
	}
	return defaultRoundTrip(request, transport)
}

func httpTransport(request roundtripper.Request, gwAddr string) (*http3.RoundTripper, error) {

	transport := &http3.RoundTripper{
		QuicConfig: &quic.Config{
			Tracer: qlog.DefaultTracer,
		},
		Dial: func(ctx context.Context, addr string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlyConnection, error) {
			if request.URL.Host == "www.example.com" {
				addr = gwAddr
				return quic.DialAddrEarly(ctx, addr, tlsCfg, cfg)
			}
			return nil, nil
		},
	}
	if request.Server != "" && len(request.CertPem) != 0 {
		tlsConfig, err := tlsClientConfig(request.Server, request.CertPem)
		if err != nil {
			return nil, err
		}
		transport.TLSClientConfig = tlsConfig
	}
	return transport, nil
}

func tlsClientConfig(server string, certPem []byte) (*tls.Config, error) {
	// Add the provided cert as a trusted CA
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(certPem) {
		return nil, fmt.Errorf("unexpected error adding trusted CA")
	}

	if server == "" {
		return nil, fmt.Errorf("unexpected error, server name required for TLS")
	}

	// Create the tls Config for this provided host, cert, and trusted CA
	// Disable G402: TLS MinVersion too low. (gosec)
	// #nosec G402
	return &tls.Config{
		ServerName: server,
		RootCAs:    certPool,
	}, nil
}

func defaultRoundTrip(request roundtripper.Request, transport *http3.RoundTripper) (*roundtripper.CapturedRequest, *roundtripper.CapturedResponse, error) {
	client := &httpv1.Client{}

	if request.UnfollowRedirect {
		client.CheckRedirect = func(req *httpv1.Request, via []*httpv1.Request) error {
			return httpv1.ErrUseLastResponse
		}
	}

	client.Transport = transport

	method := "GET"
	if request.Method != "" {
		method = request.Method
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	req, err := httpv1.NewRequestWithContext(ctx, method, request.URL.String(), nil)
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

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	cReq := &roundtripper.CapturedRequest{}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	if resp.Header.Get("Content-Type") == "application/json" {
		err = json.Unmarshal(body, cReq)
		if err != nil {
			return nil, nil, fmt.Errorf("unexpected error reading response: %w", err)
		}
	} else {
		cReq.Method = method
	}

	cRes := &roundtripper.CapturedResponse{
		StatusCode:    resp.StatusCode,
		ContentLength: resp.ContentLength,
		Protocol:      resp.Proto,
		Headers:       resp.Header,
	}

	if roundtripper.IsRedirect(resp.StatusCode) {
		redirectURL, err := resp.Location()
		if err != nil {
			return nil, nil, err
		}
		cRes.RedirectRequest = &roundtripper.RedirectRequest{
			Scheme: redirectURL.Scheme,
			Host:   redirectURL.Hostname(),
			Port:   redirectURL.Port(),
			Path:   redirectURL.Path,
		}
	}

	return cReq, cRes, nil
}
