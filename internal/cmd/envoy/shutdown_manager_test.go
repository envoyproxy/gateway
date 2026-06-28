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
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// setupFakeEnvoyStats set up an HTTP server return content
func setupFakeEnvoyStats(t *testing.T, content string) *http.Server {
	l, err := net.Listen("tcp", ":0") //nolint: gosec
	require.NoError(t, err)
	require.NoError(t, l.Close())
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
		if err := s.ListenAndServe(); err != nil {
			fmt.Println("fail to listen: ", err)
		}
	}()

	return s
}

func TestStartDrain(t *testing.T) {
	tests := []struct {
		name                  string
		readinessFailureDelay time.Duration
		expectedContains      []string
		expectedNotContains   []string
	}{
		{
			name:                  "zero readiness failure delay preserves existing healthcheck fail behavior",
			readinessFailureDelay: 0,
			expectedContains: []string{
				"healthcheck/fail",
			},
			expectedNotContains: []string{
				"drain_listeners?graceful&skip_exit",
			},
		},
		{
			name:                  "positive readiness failure delay drains then fails healthcheck",
			readinessFailureDelay: 5 * time.Millisecond,
			expectedContains: []string{
				"drain_listeners?graceful&skip_exit",
				"healthcheck/fail",
			},
		},
		{
			name:                  "readiness failure can be stopped before delay",
			readinessFailureDelay: 50 * time.Millisecond,
			expectedContains: []string{
				"drain_listeners?graceful&skip_exit",
			},
			expectedNotContains: []string{
				"healthcheck/fail",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotCh := make(chan string, len(tc.expectedContains)+len(tc.expectedNotContains)+1)
			timer := startDrain(tc.readinessFailureDelay, func(path string) error {
				gotCh <- path
				return nil
			})
			if timer != nil {
				defer timer.Stop()
			}
			if tc.readinessFailureDelay > 0 && len(tc.expectedNotContains) == 0 {
				require.Eventually(t, func() bool {
					return len(gotCh) == len(tc.expectedContains)
				}, time.Second, time.Millisecond)
			} else if timer != nil {
				timer.Stop()
			}

			var got []string
			for len(gotCh) > 0 {
				got = append(got, <-gotCh)
			}

			for _, expected := range tc.expectedContains {
				require.Contains(t, got, expected)
			}
			for _, unexpected := range tc.expectedNotContains {
				require.NotContains(t, got, unexpected)
			}
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
			name:          "invalid",
			input:         `{"stats":[{"name":"listener.0.0.0.0_8000.downstream_cx_active","value":1]}`,
			expectedError: errors.New("error getting listener downstream_cx_active stat"),
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
