// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	egcrypto "github.com/envoyproxy/gateway/internal/crypto"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/logging"
)

func TestMetricServer(t *testing.T) {
	cfg := &config.Server{
		EnvoyGateway: &egv1a1.EnvoyGateway{
			EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{},
		},
		Logger: logging.NewLogger(os.Stdout, egv1a1.DefaultEnvoyGatewayLogging()),
	}

	runner := New(cfg)
	err := runner.Start(context.Background())
	require.NoError(t, err)

	// Clean up
	err = runner.Close()
	require.NoError(t, err)
}

func TestMetricServerTLS(t *testing.T) {
	cfg, err := config.New(io.Discard, io.Discard)
	require.NoError(t, err)
	cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{Type: egv1a1.ProviderTypeCustom}

	certs, err := egcrypto.GenerateCerts(cfg)
	require.NoError(t, err)

	certDir := t.TempDir()
	certFile := filepath.Join(certDir, "tls.crt")
	keyFile := filepath.Join(certDir, "tls.key")
	require.NoError(t, os.WriteFile(certFile, certs.EnvoyGatewayCertificate, 0o600))
	require.NoError(t, os.WriteFile(keyFile, certs.EnvoyGatewayPrivateKey, 0o600))

	runner := New(cfg)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("metrics"))
	})
	err = runner.start(ctx, "127.0.0.1:0", handler, metricsTLSOptions{
		enabled:  true,
		certFile: certFile,
		keyFile:  keyFile,
	})
	require.NoError(t, err)
	defer func() { require.NoError(t, runner.Close()) }()

	require.Eventually(t, func() bool {
		return runner.listener != nil
	}, time.Second, time.Millisecond)

	client := &http.Client{
		Transport: &http.Transport{
			// The test verifies TLS serving; certificate trust is covered by cert-manager integration tests.
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // test-only self-signed certificate
		},
	}
	var (
		statusCode int
		body       string
	)
	require.Eventually(t, func() bool {
		var err error
		statusCode, body, err = getMetrics(client, "https://"+runner.listener.Addr().String()+defaultEndpoint)
		if err != nil {
			return false
		}
		return statusCode == http.StatusOK
	}, time.Second, time.Millisecond)
	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, "metrics", body)
}

func getMetrics(client *http.Client, url string) (int, string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", err
	}
	return resp.StatusCode, string(body), nil
}

func TestMetricServerTLSRequiresCertificateFiles(t *testing.T) {
	cfg, err := config.New(io.Discard, io.Discard)
	require.NoError(t, err)
	cfg.MetricsTLS.Enabled = true

	runner := New(cfg)
	err = runner.Start(t.Context())
	require.EqualError(t, err, "metrics TLS requires both certificate and key files")
}
