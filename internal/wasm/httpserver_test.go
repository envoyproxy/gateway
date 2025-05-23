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
	"os"
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

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
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

		if server, err = startLocalHTTPServer(
			ctx,
			t.TempDir(),
			defaultMaxFailedAttempts,
			defaultAttemptResetDelay,
			defaultAttemptsResetInterval); err != nil {
			t.Fatal(err)
		}
		defer server.close()

		// Call server.Get() to initialize the local file cache.
		servingURL, _, err = server.Get(
			fmt.Sprintf("oci://%s/%s", registryURL.Host, validWasmModule),
			GetOptions{
				ResourceName:   resourceName,
				RequestTimeout: time.Second * 1000,
			})
		require.NoError(t, err)

		// Get wasm module from the EG HTTP server.
		t.Logf("Get wasm module from the EG HTTP server: %s", servingURL)
		resp, err = client.Get(servingURL)
		_ = resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// Call server.Get() again to get the serving URL for the same wasm module.
		// The serving URL should be the same as the previous one.
		servingURL1, _, err := server.Get(
			fmt.Sprintf("oci://%s/%s", registryURL.Host, validWasmModule),
			GetOptions{
				ResourceName:   resourceName,
				RequestTimeout: time.Second * 1000,
			})
		require.NoError(t, err)
		require.Equal(t, servingURL, servingURL1)

		// Get wasm module from the EG HTTP server.
		t.Logf("Get wasm module from the EG HTTP server: %s", servingURL1)
		resp, err = client.Get(servingURL1)
		require.NoError(t, err)
		_ = resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("get non-existing wasm module from EG HTTP server", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var server *HTTPServer

		if server, err = startLocalHTTPServer(
			ctx,
			t.TempDir(),
			defaultMaxFailedAttempts,
			defaultAttemptResetDelay,
			defaultAttemptsResetInterval); err != nil {
			t.Fatal(err)
		}
		defer server.close()

		// Initialize the local cache.
		_, _, err = server.Get(fmt.Sprintf("oci://%s/%s", registryURL.Host, nonExistingWasmModule),
			GetOptions{
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
	httpData := append([]byte{}, wasmHeader...)
	httpData = append(httpData, []byte("data")...)
	r := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == fmt.Sprintf("/%s", validWasmModule) {
			_, _ = w.Write(httpData)
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

		if server, err = startLocalHTTPServer(
			ctx,
			t.TempDir(),
			defaultMaxFailedAttempts,
			defaultAttemptResetDelay,
			defaultAttemptsResetInterval); err != nil {
			t.Fatal(err)
		}
		defer server.close()

		getOptions := GetOptions{
			ResourceName:   resourceName,
			RequestTimeout: time.Second * 10,
		}

		// Call server.Get() to initialize the local file cache.
		servingURL, _, err = server.Get(fmt.Sprintf("%s/%s", fakeServerURL, validWasmModule), getOptions)
		require.NoError(t, err)

		// Get wasm module from the EG HTTP server.
		t.Logf("Get wasm module from the EG HTTP server: %s", servingURL)
		resp, err = client.Get(servingURL)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		_ = resp.Body.Close()

		// Call server.Get() again to get the serving URL for the same wasm module.
		// The serving URL should be the same as the previous one.
		servingURL1, _, err := server.Get(fmt.Sprintf("%s/%s", fakeServerURL, validWasmModule), getOptions)
		require.NoError(t, err)
		require.Equal(t, servingURL, servingURL1)

		// Get wasm module from the EG HTTP server.
		t.Logf("Get wasm module from the EG HTTP server: %s", servingURL1)
		resp, err = client.Get(servingURL1)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		_ = resp.Body.Close()
	})

	t.Run("get non-existing wasm module from EG HTTP server", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var server *HTTPServer

		if server, err = startLocalHTTPServer(
			ctx,
			t.TempDir(),
			defaultMaxFailedAttempts,
			defaultAttemptResetDelay,
			defaultAttemptsResetInterval); err != nil {
			t.Fatal(err)
		}
		defer server.close()

		// Initialize the local cache.
		_, _, err = server.Get(fmt.Sprintf("%s/%s", fakeServerURL, nonExistingWasmModule), GetOptions{
			ResourceName:   resourceName,
			RequestTimeout: time.Second * 10,
		})
		if err == nil || !strings.Contains(err.Error(), "404") {
			t.Errorf("Get() error = %v, expect error contains 'Unknown name'", err)
		}
	})
}

func Test_httpServerFailedAttempt(t *testing.T) {
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

	t.Run("failed attempts exceed the max", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var (
			server                *HTTPServer
			maxFailedAttempts     = 5
			attemptResetDelay     = time.Millisecond * 200
			attemptsResetInterval = time.Millisecond * 100
		)

		if server, err = startLocalHTTPServer(
			ctx,
			t.TempDir(),
			maxFailedAttempts,
			attemptResetDelay,
			attemptsResetInterval); err != nil {
			t.Fatal(err)
		}
		defer server.close()

		// The 6th Get() should return an error immediately because the max failed attempts is 5.
		for i := 0; i <= 6; i++ {
			_, _, err = server.Get(
				fmt.Sprintf("oci://%s/%s", registryURL.Host, nonExistingWasmModule),
				GetOptions{
					ResourceName:   resourceName,
					RequestTimeout: time.Second * 1000,
				})
		}
		require.ErrorContains(t, err, "after 5 attempts")

		// The 7th Get() should return a normal error because the failed attempts have been reset.
		err = nil
		for i := 0; i < 3; i++ {
			time.Sleep(300 * time.Millisecond)
			_, _, err = server.Get(
				fmt.Sprintf("oci://%s/%s", registryURL.Host, nonExistingWasmModule),
				GetOptions{
					ResourceName:   resourceName,
					RequestTimeout: time.Second * 1000,
				})
			if err != nil {
				break
			}
		}
		require.ErrorContains(t, err, "Unknown name")
	})
}

func setupFakeRegistry(host string) error {
	var (
		l   v1.Layer
		img v1.Image
		err error
	)

	ref := fmt.Sprintf("%s/%s", host, validWasmModule)
	binary := wasmHeader
	binary = append(binary, []byte("this is wasm plugin")...)

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

func startLocalHTTPServer(ctx context.Context, cacheDir string, maxFailedAttempts int, failedAttemptResetDelay, failedAttemptsResetInterval time.Duration) (*HTTPServer, error) {
	logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)
	s := NewHTTPServerWithFileCache(
		SeverOptions{
			Salt:                        []byte("salt"),
			MaxFailedAttempts:           maxFailedAttempts,
			FailedAttemptResetDelay:     failedAttemptResetDelay,
			FailedAttemptsResetInterval: failedAttemptsResetInterval,
		},
		CacheOptions{
			CacheDir: cacheDir,
		}, "envoy-gateway-system", logger)
	go s.Start(ctx)

	// Wait for the server to start
	var (
		retries  = 10
		response *http.Response
		err      error
	)
	for i := 0; i < retries; i++ {
		if response, err = http.Get(fmt.Sprintf("http://127.0.0.1:%d", serverPort)); err == nil {
			_ = response.Body.Close()
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
				return d.DialContext(ctx, network, fmt.Sprintf("127.0.0.1:%d", serverPort))
			},
		},
	}
}
