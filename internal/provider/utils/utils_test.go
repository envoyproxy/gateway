// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpectedContainerPortHashedName(t *testing.T) {
	tests := []struct {
		name             string
		listenerName     string
		expectedPortName string
	}{

		{
			name:             "short listener name",
			listenerName:     "http",
			expectedPortName: "http",
		},
		{
			name:             "listener name longer than 15 chars",
			listenerName:     "listener-123456789",
			expectedPortName: "listener-123456",
		},
		{
			name:             "listener name longer than 15 chars in merged gateway",
			listenerName:     "test-ns/gateway-1/listener-123456789",
			expectedPortName: fmt.Sprintf("%s-%s", "listene", HashString("test-ns/gateway-1/listener-123456789")[:8]),
		},
		{
			name:             "listener name shorter than 15 chars in merged gateway",
			listenerName:     "test-ns/gateway-1/http",
			expectedPortName: fmt.Sprintf("%s-%s", "http", HashString("test-ns/gateway-1/http")[:8]),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			portName := ExpectedContainerPortName(tc.listenerName)
			require.Equal(t, tc.expectedPortName, portName)
		})
	}
}
