// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package utils

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// RequireRandomPorts returns random available ports.
func RequireRandomPorts(t testing.TB, count int) []int {
	t.Helper()

	ports := make([]int, count)

	listeners := make([]net.Listener, 0, count)
	for i := range count {
		lc := net.ListenConfig{}
		lis, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
		require.NoError(t, err, "failed to listen on random port %d", i)
		listeners = append(listeners, lis)
		addr := lis.Addr().(*net.TCPAddr)
		ports[i] = addr.Port
	}
	for _, lis := range listeners {
		require.NoError(t, lis.Close())
	}
	return ports
}

// AwaitPortClosed waits until the port is no longer listening.
func AwaitPortClosed(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			return nil // Port closed
		}
		conn.Close()

		if time.Now().After(deadline) {
			return fmt.Errorf("port %d still listening after %v", port, timeout)
		}
		<-ticker.C
	}
}
