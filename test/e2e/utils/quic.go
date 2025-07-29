// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

// copy from https://github.com/kubernetes-sigs/gateway-api/blob/0fd1805d4f97b4ee6a57339f3a411a26018ad699/conformance/utils/roundtripper/roundtripper.go
// TODO: make upstream support quic
package utils

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"regexp"

	"github.com/quic-go/quic-go/http3"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

var _ roundtripper.RoundTripper = &QuicRoundTripper{}

type QuicRoundTripper struct {
	Debug         bool
	TimeoutConfig config.TimeoutConfig
}

func (q *QuicRoundTripper) CaptureRoundTrip(request roundtripper.Request) (*roundtripper.CapturedRequest, *roundtripper.CapturedResponse, error) {
	tlsConfig := &tls.Config{
		//nolint:gosec
		InsecureSkipVerify: true,           // Skip verification of server certificate
		NextProtos:         []string{"h3"}, // Required for HTTP/3
	}

	// Create a custom HTTP/3 transport
	transport := &http3.Transport{
		TLSClientConfig: tlsConfig,
	}
	if request.Server != "" && len(request.CertPem) != 0 && len(request.KeyPem) != 0 {
		clientTLS, err := tlsClientConfig(request.Server, request.CertPem, request.KeyPem)
		if err != nil {
			return nil, nil, err
		}
		transport.TLSClientConfig = clientTLS
	}

	return q.defaultRoundTrip(request, transport)
}

func tlsClientConfig(server string, certPem, keyPem []byte) (*tls.Config, error) {
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
		NextProtos:   []string{"h3"}, // Required for HTTP/3
	}, nil
}

func (q *QuicRoundTripper) defaultRoundTrip(request roundtripper.Request, transport http.RoundTripper) (*roundtripper.CapturedRequest, *roundtripper.CapturedResponse, error) {
	client := &http.Client{}

	if request.UnfollowRedirect {
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	client.Transport = transport

	method := "GET"
	if request.Method != "" {
		method = request.Method
	}
	ctx, cancel := context.WithTimeout(context.Background(), q.TimeoutConfig.RequestTimeout)
	defer cancel()

	var reqBody io.Reader
	req, err := http.NewRequestWithContext(ctx, method, request.URL.String(), reqBody)
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

	if q.Debug {
		var dump []byte
		dump, err = httputil.DumpRequestOut(req, true)
		if err != nil {
			return nil, nil, err
		}

		tlog.Logf(request.T, "Sending Request:\n%s\n\n", formatDump(dump, "< "))
	}

	resp, err := client.Do(req)
	if err != nil {
		if q.Debug {
			var dump []byte
			if resp != nil {
				dump, err = httputil.DumpResponse(resp, true)
				if err != nil {
					return nil, nil, err
				}
				tlog.Logf(request.T, "Error sending request:\n%s\n\n", formatDump(dump, "< "))
			} else {
				tlog.Logf(request.T, "Error sending request: %v (no response)\n", err)
			}
		}
		return nil, nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if q.Debug {
		var dump []byte
		dump, err = httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, nil, err
		}

		tlog.Logf(request.T, "Received Response:\n%s\n\n", formatDump(dump, "< "))
	}

	cReq := &roundtripper.CapturedRequest{}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
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

var startLineRegex = regexp.MustCompile(`(?m)^`)

func formatDump(data []byte, prefix string) string {
	data = startLineRegex.ReplaceAllLiteral(data, []byte(prefix))
	return string(data)
}
