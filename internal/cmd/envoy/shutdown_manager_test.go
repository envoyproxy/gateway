// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package envoy

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// setupFakeEnvoyStats set up an HTTP server return content
func setupFakeEnvoyStats(t *testing.T, content string) *http.Server {
	// Reuse the bound listener instead of closing and re-binding the port.
	l, err := net.Listen("tcp", ":0") //nolint: gosec
	require.NoError(t, err)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte(content))
	})

	addr := l.Addr().String()
	s := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: time.Second,
	}
	t.Logf("start to listen at %s ", addr)
	go func() {
		if err := s.Serve(l); err != nil {
			fmt.Println("fail to listen: ", err)
		}
	}()

	return s
}

func TestClearShutdownReadyFile(t *testing.T) {
	cases := []struct {
		name    string
		setup   func(t *testing.T) string
		cleared bool
	}{
		{
			name: "file does not exist",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "shutdown-ready")
			},
			cleared: true,
		},
		{
			name: "stale file exists",
			setup: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "shutdown-ready")
				require.NoError(t, os.WriteFile(path, []byte("stale"), 0o600))
				return path
			},
			cleared: true,
		},
		{
			name: "removal failure is logged, not fatal",
			setup: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "shutdown-ready")
				require.NoError(t, os.Mkdir(path, 0o700))
				require.NoError(t, os.WriteFile(filepath.Join(path, "child"), []byte("x"), 0o600))
				return path
			},
			cleared: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := tc.setup(t)
			clearShutdownReadyFile(path)
			_, err := os.Stat(path)
			if tc.cleared {
				require.True(t, os.IsNotExist(err))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestShutdownReadyResponseOutlivesWriteTimeout verifies that the shutdown ready
// response is still delivered when the drain outlasts the server's write timeout.
func TestShutdownReadyResponseOutlivesWriteTimeout(t *testing.T) {
	origWriteTimeout := shutdownManagerWriteTimeout
	shutdownManagerWriteTimeout = 100 * time.Millisecond
	t.Cleanup(func() { shutdownManagerWriteTimeout = origWriteTimeout })

	cases := []struct {
		name         string
		readyTimeout time.Duration
		readyAfter   time.Duration // 0 means the ready file is never written
		expectedCode int
	}{
		{
			name:         "drain completes",
			readyTimeout: 30 * time.Second,
			readyAfter:   time.Second,
			expectedCode: http.StatusOK,
		},
		{
			name:         "ready timeout exceeded",
			readyTimeout: 500 * time.Millisecond,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			readyFile := filepath.Join(t.TempDir(), "shutdown-ready")
			ts := httptest.NewUnstartedServer(nil)
			ts.Config = newShutdownManagerServer("", tc.readyTimeout, readyFile)
			ts.Start()
			defer ts.Close()

			if tc.readyAfter > 0 {
				timer := time.AfterFunc(tc.readyAfter, func() {
					_ = os.WriteFile(readyFile, nil, 0o600)
				})
				defer timer.Stop()
			}

			resp, err := ts.Client().Get(ts.URL + ShutdownManagerReadyPath)
			require.NoError(t, err)
			defer func() {
				_ = resp.Body.Close()
			}()
			require.Equal(t, tc.expectedCode, resp.StatusCode)
		})
	}
}

func TestGetTotalConnections(t *testing.T) {
	cases := []struct {
		name  string
		input string

		expectedError error
		expectedCount *int
	}{
		{
			name: "downstream_cx_active",
			input: `{
    "stats": [
        {
            "name": "listener.0.0.0.0_8000.downstream_cx_active",
            "value": 1
        },
        {
            "name": "listener.0.0.0.0_8000.worker_0.downstream_cx_active",
            "value": 1
        },
        {
            "name": "listener.0.0.0.0_8000.worker_1.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.0.0.0.0_8000.worker_2.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.0.0.0.0_8000.worker_3.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.0.0.0.0_8000.worker_4.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.0.0.0.0_8000.worker_5.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.0.0.0.0_8000.worker_6.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.0.0.0.0_8000.worker_7.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.0.0.0.0_8000.worker_8.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.0.0.0.0_8000.worker_9.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8080.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8080.worker_0.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8080.worker_1.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8080.worker_2.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8080.worker_3.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8080.worker_4.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8080.worker_5.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8080.worker_6.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8080.worker_7.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8080.worker_8.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8080.worker_9.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8081.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8081.worker_0.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8081.worker_1.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8081.worker_2.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8081.worker_3.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8081.worker_4.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8081.worker_5.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8081.worker_6.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8081.worker_7.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8081.worker_8.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.127.0.0.1_8081.worker_9.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.admin.downstream_cx_active",
            "value": 2
        },
        {
            "name": "listener.admin.main_thread.downstream_cx_active",
            "value": 2
        }
    ]
}`,
			expectedCount: new(1),
		},
		{
			name: "udp_downstream_sess_active",
			input: `{
    "stats": [
        {"name": "listener.0.0.0.0_5300.downstream_cx_active", "value": 0},
        {"name": "listener.0.0.0.0_5300.worker_0.downstream_cx_active", "value": 0},
        {"name": "udp.service.downstream_sess_active", "value": 3}
    ]
}`,
			expectedCount: new(3),
		},
		{
			name: "tcp_and_udp",
			input: `{
    "stats": [
        {"name": "listener.0.0.0.0_8000.downstream_cx_active", "value": 1},
        {"name": "listener.0.0.0.0_8000.worker_0.downstream_cx_active", "value": 1},
        {"name": "listener.0.0.0.0_19001.downstream_cx_active", "value": 2},
        {"name": "udp.service.downstream_sess_active", "value": 2}
    ]
}`,
			expectedCount: new(3),
		},
		{
			name:          "invalid",
			input:         `{"stats":[{"name":"listener.0.0.0.0_8000.downstream_cx_active","value":1]}`,
			expectedError: errors.New("error getting active connection and UDP session stats"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := setupFakeEnvoyStats(t, tc.input)
			_, port, err := net.SplitHostPort(s.Addr)
			require.NoError(t, err)

			p, err := strconv.Atoi(port)
			require.NoError(t, err)
			defer func() {
				_ = s.Close()
			}()
			reader := strings.NewReader(tc.input)
			rc := io.NopCloser(reader)
			defer func() {
				_ = rc.Close()
			}()

			gotCount, gotError := getTotalConnections(p)
			if tc.expectedError != nil {
				require.ErrorContains(t, gotError, tc.expectedError.Error())
				return
			}
			require.NoError(t, gotError)
			require.Equal(t, tc.expectedCount, gotCount)
		})
	}
}
