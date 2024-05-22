// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package wasm

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/utils/sets"
)

const (
	validWasmModule       = "valid.wasm"
	nonExistingWasmModule = "non-existing.wasm"
	resourceName          = "envoyextensionpolicy/envoy-gateway/policy-for-gateway/wasm/0"
)

func Test_httpServerWithOCIImage(t *testing.T) {
	var (
		registryURL *url.URL
		err         error
	)

	// Set up a fake registry.
	r := httptest.NewServer(registry.New())
	defer r.Close()

	if registryURL, err = url.Parse(r.URL); err != nil {
		t.Fatal(err)
	}
	if err = setupFakeRegistry(registryURL.Host); err != nil {
		t.Fatal(err)
	}

	t.Run("get wasm module from EG HTTP server", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var (
			server     *HTTPServer
			client     = newHTTPClient()
			resp       *http.Response
			servingURL string
		)

		if server, err = startLocalHTTPServer(ctx, t.TempDir()); err != nil {
			t.Fatal(err)
		}

		getOptions := GetOptions{
			ResourceName:   resourceName,
			RequestTimeout: time.Second * 10,
		}

		// Call server.Get() to initialize the local file cache.
		servingURL, err = server.Get(fmt.Sprintf("oci://%s/%s", registryURL.Host, validWasmModule), getOptions)
		require.NoError(t, err)

		// Get wasm module from the EG HTTP server.
		t.Logf("Get wasm module from the EG HTTP server: %s", servingURL)
		resp, err = client.Get(servingURL)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// Call server.Get() again to get the serving URL for the same wasm module.
		// The serving URL should be the same as the previous one.
		servingURL1, err := server.Get(fmt.Sprintf("oci://%s/%s", registryURL.Host, validWasmModule), getOptions)
		require.NoError(t, err)
		require.Equal(t, servingURL, servingURL1)

		// Get wasm module from the EG HTTP server.
		t.Logf("Get wasm module from the EG HTTP server: %s", servingURL1)
		resp, err = client.Get(servingURL1)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("get non-existing wasm module from EG HTTP server", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var server *HTTPServer

		if server, err = startLocalHTTPServer(ctx, t.TempDir()); err != nil {
			t.Fatal(err)
		}

		// Initialize the local cache.
		_, err = server.Get(fmt.Sprintf("oci://%s/%s", registryURL.Host, nonExistingWasmModule), GetOptions{
			ResourceName:   resourceName,
			RequestTimeout: time.Second * 10,
		})
		if err == nil || !strings.Contains(err.Error(), "Unknown name") {
			t.Errorf("Get() error = %v, expect error contains 'Unknown name'", err)
		}
	})
}

func Test_httpServerWithHTTP(t *testing.T) {
	var (
		fakeServerURL string
		err           error
	)

	// Set up a fake HTTP server.
	httpData := append(wasmHeader, []byte("data")...)
	r := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == fmt.Sprintf("/%s", validWasmModule) {
			w.Write(httpData)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	fakeServerURL = r.URL
	defer r.Close()

	t.Run("get wasm module from EG HTTP server", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var (
			server     *HTTPServer
			client     = newHTTPClient()
			resp       *http.Response
			servingURL string
		)

		if server, err = startLocalHTTPServer(ctx, t.TempDir()); err != nil {
			t.Fatal(err)
		}

		getOptions := GetOptions{
			ResourceName:   resourceName,
			RequestTimeout: time.Second * 10,
		}

		// Call server.Get() to initialize the local file cache.
		servingURL, err = server.Get(fmt.Sprintf("%s/%s", fakeServerURL, validWasmModule), getOptions)
		require.NoError(t, err)

		// Get wasm module from the EG HTTP server.
		t.Logf("Get wasm module from the EG HTTP server: %s", servingURL)
		resp, err = client.Get(servingURL)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// Call server.Get() again to get the serving URL for the same wasm module.
		// The serving URL should be the same as the previous one.
		servingURL1, err := server.Get(fmt.Sprintf("%s/%s", fakeServerURL, validWasmModule), getOptions)
		require.NoError(t, err)
		require.Equal(t, servingURL, servingURL1)

		// Get wasm module from the EG HTTP server.
		t.Logf("Get wasm module from the EG HTTP server: %s", servingURL1)
		resp, err = client.Get(servingURL1)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("get non-existing wasm module from EG HTTP server", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var server *HTTPServer

		if server, err = startLocalHTTPServer(ctx, t.TempDir()); err != nil {
			t.Fatal(err)
		}

		// Initialize the local cache.
		_, err = server.Get(fmt.Sprintf("%s/%s", fakeServerURL, nonExistingWasmModule), GetOptions{
			ResourceName:   resourceName,
			RequestTimeout: time.Second * 10,
		})
		if err == nil || !strings.Contains(err.Error(), "404") {
			t.Errorf("Get() error = %v, expect error contains 'Unknown name'", err)
		}
	})
}

func setupFakeRegistry(host string) error {
	var (
		l   v1.Layer
		img v1.Image
		err error
	)

	ref := fmt.Sprintf("%s/%s", host, validWasmModule)
	binary := append(wasmHeader, []byte("this is wasm plugin")...)

	// Create OCI compressed layer.
	if l, err = newMockLayer(types.OCILayer, map[string][]byte{"plugin.wasm": binary}); err != nil {
		return err
	}

	if img, err = mutate.Append(empty.Image, mutate.Addendum{Layer: l}); err != nil {
		return err
	}

	img = mutate.MediaType(img, types.OCIManifestSchema1)

	// Push image to the registry.
	if err = crane.Push(img, ref); err != nil {
		return err
	}
	return nil
}

func startLocalHTTPServer(ctx context.Context, cacheDir string) (*HTTPServer, error) {
	options := defaultOptions()
	options.InsecureRegistries = sets.New("*")
	logger := logging.DefaultLogger(v1alpha1.LogLevelInfo)
	s := NewHTTPServerWithFileCache(cacheDir, logger)
	go s.Start(ctx)

	// Wait for the server to start
	var (
		retries = 10
		err     error
	)
	for i := 0; i < retries; i++ {
		_, err = http.Get(fmt.Sprintf("http://127.0.0.1:%d", envoyGatewayHTTPServerPort))
		if err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	return s, err
}

func newHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				d := net.Dialer{}
				return d.DialContext(ctx, network, fmt.Sprintf("127.0.0.1:%d", envoyGatewayHTTPServerPort))
			},
		},
	}
}

func Test_generateRandomPath(t *testing.T) {
	t.Run("generate random path", func(t *testing.T) {
		all := make(map[string]struct{})

		for i := 0; i < 100; i++ {
			got, err := generateRandomPath()
			if err != nil {
				t.Errorf("generateRandomPath() error = %v", err)
				return
			}
			if _, ok := all[got]; ok {
				t.Errorf("generateRandomPath() = %v, want unique", got)
			}
			all[got] = struct{}{}
		}
	})
}
