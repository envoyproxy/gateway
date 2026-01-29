// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wasm

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
)

func TestWasmHTTPFetchWithCACert(t *testing.T) {
	// Generate a self-signed CA certificate
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization: []string{"Envoy Gateway"},
		},
		NotBefore:             time.Now().Add(-1 * time.Minute),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	require.NoError(t, err)

	caPEM := new(bytes.Buffer)
	_ = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	// Generate a server certificate signed by the CA
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2020),
		Subject: pkix.Name{
			Organization: []string{"Envoy Gateway"},
		},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:    time.Now().Add(-1 * time.Minute),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	require.NoError(t, err)

	certPEM := new(bytes.Buffer)
	_ = pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPrivKeyPEM := new(bytes.Buffer)
	_ = pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})

	serverCert, err := tls.X509KeyPair(certPEM.Bytes(), certPrivKeyPEM.Bytes())
	require.NoError(t, err)

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("wasm binary"))
	}))
	ts.TLS = &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		MinVersion:   tls.VersionTLS12, // gosec requires TLS 1.2
	}
	ts.StartTLS()
	defer ts.Close()

	fetcher := NewHTTPFetcher(DefaultHTTPRequestTimeout, DefaultHTTPRequestMaxRetries, logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo))
	fetcher.initialBackoff = time.Microsecond

	// Fetch with CA cert should succeed
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	b, err := fetcher.Fetch(ctx, ts.URL, false, caPEM.Bytes())
	require.NoError(t, err)
	assert.Equal(t, []byte("wasm binary"), b)

	// Fetch without CA cert (and not insecure) should fail
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	_, err = fetcher.Fetch(ctx2, ts.URL, false, nil)
	require.Error(t, err)

	// Fetch with invalid CA cert should fail
	_, err = fetcher.Fetch(ctx, ts.URL, false, []byte("invalid"))
	require.Error(t, err)
}

func TestWasmHTTPFetch(t *testing.T) {
	var ts *httptest.Server

	cases := []struct {
		name           string
		handler        func(http.ResponseWriter, *http.Request, int)
		timeout        time.Duration
		wantNumRequest int
		wantErrorRegex string
	}{
		{
			name: "download ok",
			handler: func(w http.ResponseWriter, _ *http.Request, _ int) {
				fmt.Fprintln(w, "wasm")
			},
			timeout:        10 * time.Second,
			wantNumRequest: 1,
		},
		{
			name: "download retry",
			handler: func(w http.ResponseWriter, _ *http.Request, num int) {
				if num <= 2 {
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					fmt.Fprintln(w, "wasm")
				}
			},
			timeout:        10 * time.Second,
			wantNumRequest: 4,
		},
		{
			name: "download max retry",
			handler: func(w http.ResponseWriter, _ *http.Request, _ int) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			timeout:        10 * time.Second,
			wantNumRequest: 5,
			wantErrorRegex: "wasm module download failed after 5 attempts, last error: wasm module download request failed: status code 500",
		},
		{
			name: "download is never tried by immediate context timeout",
			handler: func(w http.ResponseWriter, _ *http.Request, _ int) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			timeout:        0, // Immediately timeout in the context level.
			wantNumRequest: 0, // Should not retried because it is already timed out.
			wantErrorRegex: "wasm module download failed after 1 attempts, last error: Get \"[^\"]+\": context deadline exceeded",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotNumRequest := 0
			wantWasmModule := "wasm\n"
			ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.handler(w, r, gotNumRequest)
				gotNumRequest++
			}))
			defer ts.Close()
			fetcher := NewHTTPFetcher(DefaultHTTPRequestTimeout, DefaultHTTPRequestMaxRetries, logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo))
			fetcher.initialBackoff = time.Microsecond
			ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
			defer cancel()
			b, err := fetcher.Fetch(ctx, ts.URL, false, nil)
			if c.wantNumRequest != gotNumRequest {
				t.Errorf("Wasm download request got %v, want %v", gotNumRequest, c.wantNumRequest)
			}
			if c.wantErrorRegex != "" {
				if err == nil {
					t.Errorf("Wasm download got no error, want error regex `%v`", c.wantErrorRegex)
				} else if matched, regexErr := regexp.MatchString(c.wantErrorRegex, err.Error()); regexErr != nil || !matched {
					t.Errorf("Wasm download got error `%v`, want error regex `%v`", err, c.wantErrorRegex)
				}
			} else if string(b) != wantWasmModule {
				t.Errorf("downloaded wasm module got %v, want wasm", string(b))
			}
		})
	}
}

func TestWasmHTTPInsecureServer(t *testing.T) {
	var ts *httptest.Server

	cases := []struct {
		name            string
		handler         func(http.ResponseWriter, *http.Request, int)
		insecure        bool
		wantNumRequest  int
		wantErrorSuffix string
	}{
		{
			name: "download fail",
			handler: func(w http.ResponseWriter, _ *http.Request, _ int) {
				fmt.Fprintln(w, "wasm")
			},
			insecure:        false,
			wantErrorSuffix: "x509: certificate signed by unknown authority",
			wantNumRequest:  0,
		},
		{
			name: "download ok",
			handler: func(w http.ResponseWriter, _ *http.Request, _ int) {
				fmt.Fprintln(w, "wasm")
			},
			insecure:       true,
			wantNumRequest: 1,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotNumRequest := 0
			wantWasmModule := "wasm\n"
			ts = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.handler(w, r, gotNumRequest)
				gotNumRequest++
			}))
			defer ts.Close()
			fetcher := NewHTTPFetcher(DefaultHTTPRequestTimeout, DefaultHTTPRequestMaxRetries, logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo))
			fetcher.initialBackoff = time.Microsecond
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			b, err := fetcher.Fetch(ctx, ts.URL, c.insecure, nil)
			if c.wantNumRequest != gotNumRequest {
				t.Errorf("Wasm download request got %v, want %v", gotNumRequest, c.wantNumRequest)
			}
			if c.wantErrorSuffix != "" {
				if err == nil {
					t.Errorf("Wasm download got no error, want error suffix `%v`", c.wantErrorSuffix)
				} else if !strings.HasSuffix(err.Error(), c.wantErrorSuffix) {
					t.Errorf("Wasm download got error `%v`, want error suffix `%v`", err, c.wantErrorSuffix)
				}
			} else if string(b) != wantWasmModule {
				t.Errorf("downloaded wasm module got %v, want wasm", string(b))
			}
		})
	}
}

func createTar(t *testing.T, b []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	hdr := &tar.Header{
		Name: "plugin.wasm",
		Mode: 0o600,
		Size: int64(len(b)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(b); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func createGZ(t *testing.T, b []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(b); err != nil {
		t.Fatal(err)
	}

	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

func TestWasmHTTPFetchCompressedOrTarFile(t *testing.T) {
	wasmBinary := wasmMagicNumber
	wasmBinary = append(wasmBinary, 0x00, 0x00, 0x00, 0x00)
	tarball := createTar(t, wasmBinary)
	gz := createGZ(t, wasmBinary)
	gzTarball := createGZ(t, tarball)
	cases := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request, int)
	}{
		{
			name: "plain wasm binary",
			handler: func(w http.ResponseWriter, _ *http.Request, _ int) {
				_, _ = w.Write(wasmBinary)
			},
		},
		{
			name: "tarball of wasm binary",
			handler: func(w http.ResponseWriter, _ *http.Request, _ int) {
				_, _ = w.Write(tarball)
			},
		},
		{
			name: "gzipped wasm binary",
			handler: func(w http.ResponseWriter, _ *http.Request, _ int) {
				_, _ = w.Write(gz)
			},
		},
		{
			name: "gzipped tarball of wasm binary",
			handler: func(w http.ResponseWriter, _ *http.Request, _ int) {
				_, _ = w.Write(gzTarball)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotNumRequest := 0
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.handler(w, r, gotNumRequest)
				gotNumRequest++
			}))
			defer ts.Close()
			fetcher := NewHTTPFetcher(DefaultHTTPRequestTimeout, DefaultHTTPRequestMaxRetries, logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo))
			fetcher.initialBackoff = time.Microsecond
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			b, err := fetcher.Fetch(ctx, ts.URL, false, nil)
			if err != nil {
				t.Errorf("Wasm download got an unexpected error: %v", err)
			}

			if diff := cmp.Diff(wasmBinary, b); diff != "" {
				if len(diff) > 500 {
					diff = diff[:500]
				}
				t.Errorf("unexpected binary: (-want, +got)\n%v", diff)
			}
		})
	}
}
