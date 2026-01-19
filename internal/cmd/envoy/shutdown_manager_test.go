// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package envoy

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func TestGetTotalConnection(t *testing.T) {
	cases := []struct {
		input string

		expectedError error
		expectedCount *int
	}{
		{
			input: `{
    "stats": [
        {
            "name": "listener.0.0.0.0_8000.downstream_cx_active",
            "value": 0
        },
        {
            "name": "listener.0.0.0.0_8000.worker_0.downstream_cx_active",
            "value": 0
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
			expectedCount: ptr.To(0),
		},
	}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			reader := strings.NewReader(tc.input)
			rc := io.NopCloser(reader)
			defer func() {
				_ = rc.Close()
			}()
			gotCount, gotError := parseTotalConnection(rc)
			require.Equal(t, tc.expectedError, gotError)
			require.Equal(t, tc.expectedCount, gotCount)
		})
	}
}
