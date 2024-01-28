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
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/get",
					Headers: map[string]string{
						"Host": "www.example.com",
					},
				},
				Response: http.Response{
					StatusCode: 200,
					Headers: map[string]string{
						"User-Agent":      "curl/7.84.0-DEV",
						"X-Forwarded-For": "https",
					},
				},
				Namespace: namespace,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTPS", "https")
			req.CertPem = []byte("-----BEGIN CERTIFICATE-----\nMIIDOzCCAiOgAwIBAgIULZP//n+WsdygxEtcJUFrsxz27wQwDQYJKoZIhvcNAQEL\nBQAwLTEVMBMGA1UECgwMZXhhbXBsZSBJbmMuMRQwEgYDVQQDDAtleGFtcGxlLmNv\nbTAeFw0yNDAxMjExNzEwMDRaFw0zNDAxMTgxNzEwMDRaMC0xFTATBgNVBAoMDGV4\nYW1wbGUgSW5jLjEUMBIGA1UEAwwLZXhhbXBsZS5jb20wggEiMA0GCSqGSIb3DQEB\nAQUAA4IBDwAwggEKAoIBAQCjtTC0OjUyYeb5N8iTWXrUYJ56aDxjmA7uHz4NK5dQ\ngGkmvTLxFbQ6mJSkwBRwtVslJfl9xR7/bJhXwcA+oha0DBOuGfeXrJzy8+ax6dAX\nwzZYdfnMzW0U2MP6mfh3TwnFTywyvCanI3dZaQ1d46chHFHcoxYAoarc0SCwj+LW\nrxvuXrvCwUKWz/UY2jEIw0WkLBZ4j7ZlxTNCQAUZMYUHZzFP0R0CxSpFCi6cRrDW\nDy7RgsLrYuAr9YHKIEXmpyTukF5/gZOUGsl8S7ndQgL4n6r8qKxRZOdIlkmi/As/\nFf2nRg5kwECxwx3HkrSJqtuLgCuCpCur88+edujU1CztAgMBAAGjUzBRMB0GA1Ud\nDgQWBBSpvaIWmBrdqoWTyAZwx58wcJgbazAfBgNVHSMEGDAWgBSpvaIWmBrdqoWT\nyAZwx58wcJgbazAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQAv\nsEbjaQiLj9KJXWh7jUB8iQZMEh4oyB1oHbsUrqf03QEMlh/cuTChn4+kmzW4K6EN\nh2mnsVMnX5ZegEFkvXcEVgAc1QFbyAB4j2L4EM4kaQ8+quKsS9N8O9qHvUmiMg3R\nVlEYwpIN/DqiHQBY22IV5KxmRZDaDqlJMOi8WjQAh5AnSo0WZC3PnDgHuMDxcOaZ\n7cw2p1/RdaFZB9Va00g5JsSl2GtaibW3oWDFJAGKH+04lz/bT3lcVv0R4ADsF3qv\noBj2wtFDiqjj5dqe6/QtJCMkNDCbbzw2buogzuPBplE0Z4WtsWaQIq4MBnfGi8AF\n+sSBiMU4B9gf4oU93U2m\n-----END CERTIFICATE-----")
			req.KeyPem = []byte("-----BEGIN PRIVATE KEY-----\nMIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQCjtTC0OjUyYeb5\nN8iTWXrUYJ56aDxjmA7uHz4NK5dQgGkmvTLxFbQ6mJSkwBRwtVslJfl9xR7/bJhX\nwcA+oha0DBOuGfeXrJzy8+ax6dAXwzZYdfnMzW0U2MP6mfh3TwnFTywyvCanI3dZ\naQ1d46chHFHcoxYAoarc0SCwj+LWrxvuXrvCwUKWz/UY2jEIw0WkLBZ4j7ZlxTNC\nQAUZMYUHZzFP0R0CxSpFCi6cRrDWDy7RgsLrYuAr9YHKIEXmpyTukF5/gZOUGsl8\nS7ndQgL4n6r8qKxRZOdIlkmi/As/Ff2nRg5kwECxwx3HkrSJqtuLgCuCpCur88+e\ndujU1CztAgMBAAECggEACrVe4MyARNIBy2e3pih/qqCW/+pNbaAEIAA+rUVenJcz\nFE6NecGB8dBKbtwaqi8OwydmAnEFwKRa8xNF2ZhS3wb9iuarcF+L6loudburOewH\neXIJe6OstjpMrYSWARXpjXTK7xb+z0bjDHVHo3kxVO7hjaAZtRkI0HDEkDrFQOYw\nzpYhQwJVCp5/ZHY4M7bzjWHySdIt/K+lyTLPSF3b6d/aNWQK1iQbQ4gPFJeW2wZP\nkj8zeSxTSxD6psahrjho/SXfGNTNq1m5jOM9fjR3BB/DPS3IBTXDl5IUOFCsFmd/\nu5CnXOX5RJJl/+gMFWEM25RNYEcFrNUpcSS67r1i+QKBgQDTlpBL9BCvTnWGl5Ru\nPL4u+AIgK4foKA1aczXAGZ5/oeCdcXpqdBvCQPBHjXZnQngbZ5tkR5fcXM376wrl\n4E+AOetbeYfJjvlRm9L6QmlFY/Nlv3McBDM4KGh6ZeiWjyR2yqgGtRO/kk3oFQ+p\npYeN+t0Fu9lO++fqUNt8yxsgtQKBgQDGEdUQfIdwAgokW/uWj5PZmOuaUkcKwcX4\nt2+4Gz/b5QE0pBCvpmrb33SriUrKBSQnzvU0r70Ig2wndY6ulfcb5sUieM/LHZiR\npxtDqVIKpQiXLHo4iycn6tY7lKqwWsCqXUxjB5JMfW1IeEzj4XexeRly4sLDapS4\nSJNe5GVWWQKBgHsY7XpC1DIpg1Z6eXBpBnxs7U+qA7edFae5v1uzi/LVSshObNni\nEwRAo4n9UxVgJmBLNqxwunkJxQz7AawbhCUljTf6zHUHKSXBck0GthgYvlJDv8Rc\n7S+O0rni8B4nyR8TaA3+6y5Y/9o15pbcJrEDcfMUBqldBN/ditRflbjBAoGAdD+D\nDWoJE3Qe/7f8sSETZWKa5LfleirARnli2Gslz6lYS8z+/hhuHx3HG+Y4PtlFnxeY\nUpPSHm0DzSTx2QWrQnTuvoypaEy2fsXU+qElxZmWsSMpmIYTNRpfIhjfFSIucc7Q\nRk7rTnlO6nmwpw5tcXvhs8vjA05Ket4doFPsJgECgYAEVrbe2182Z2bG0xacHPJw\nZ3SJAh8JEj9MG1Cj0M3COP+iCWR2JDk0F7bsDDhpEZrPtyF1Ea0i1YpFDJEAXmIw\nZaHfywv9wqkv+1/T15O9qHZ3GDkyRTqD//HgdGbVBWF3AyPzXkW5EM35A4qUVq7I\nKe1vZDEBEfK/Mq/DJ877sg==\n-----END PRIVATE KEY-----")
			cReq, cResp, err := CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}
		})
	},
}

func CaptureRoundTrip(request roundtripper.Request) (*roundtripper.CapturedRequest, *roundtripper.CapturedResponse, error) {
	transport, err := httpTransport(request)
	if err != nil {
		return nil, nil, err
	}
	return defaultRoundTrip(request, transport)
}

func httpTransport(request roundtripper.Request) (*http3.RoundTripper, error) {
	transport := &http3.RoundTripper{
		QuicConfig: &quic.Config{
			Tracer: qlog.DefaultTracer,
		},
	}
	if request.Server != "" && len(request.CertPem) != 0 && len(request.KeyPem) != 0 {
		tlsConfig, err := tlsClientConfig(request.Server, request.CertPem, request.KeyPem)
		if err != nil {
			return nil, err
		}
		transport.TLSClientConfig = tlsConfig
	}
	return transport, nil
}

func tlsClientConfig(server string, certPem []byte, keyPem []byte) (*tls.Config, error) {
	// Create a certificate from the provided cert and key
	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		return nil, fmt.Errorf("unexpected error creating cert: %w", err)
	}

	// Add the provided cert as a trusted CA
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(certPem) {
		return nil, fmt.Errorf("unexpected error adding trusted CA: %w", err)
	}

	if server == "" {
		return nil, fmt.Errorf("unexpected error, server name required for TLS")
	}

	// Create the tls Config for this provided host, cert, and trusted CA
	// Disable G402: TLS MinVersion too low. (gosec)
	// #nosec G402
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ServerName:   server,
		RootCAs:      certPool,
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
