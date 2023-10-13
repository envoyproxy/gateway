// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

	"github.com/envoyproxy/gateway/internal/ir"
)

// buildTCPSocketOptions converts listener downstream settings to xds socketOptions
func buildTCPSocketOptions(keepAlive *ir.TCPKeepalive) []*corev3.SocketOption {
	if keepAlive == nil {
		return nil
	}

	socketOptions := make([]*corev3.SocketOption, 0)
	// Enable Keep Alives
	socketOption := &corev3.SocketOption{
		Description: "socket option to enable tcp keep alive",
		Level:       0x1,                                        // syscall.SOL_SOCKET has a different value for Darwin, resulting in `go test` failing
		Name:        0x9,                                        // syscall.SO_KEEPALIVE has a different value for Darwin, resulting in `go test` failing
		Value:       &corev3.SocketOption_IntValue{IntValue: 1}, // Enable
	}

	socketOptions = append(socketOptions, socketOption)

	if keepAlive.Probes != nil {
		socketOption = &corev3.SocketOption{
			Description: "socket option for keep alive probes",
			Level:       0x6, // Darwin lacks syscall.SOL_TCP
			Name:        0x6, // Darwin lacks syscall.TCP_KEEPCNT,
			Value:       &corev3.SocketOption_IntValue{IntValue: int64(*keepAlive.Probes)},
			State:       corev3.SocketOption_STATE_PREBIND,
		}
		socketOptions = append(socketOptions, socketOption)
	}

	if keepAlive.IdleTime != nil {
		socketOption = &corev3.SocketOption{
			Description: "socket option for keep alive idle time",
			Level:       0x6, // Darwin lacks syscall.SOL_TCP
			Name:        0x4, // Darwin lacks syscall.TCP_KEEPIDLE,
			Value:       &corev3.SocketOption_IntValue{IntValue: int64(*keepAlive.IdleTime)},
			State:       corev3.SocketOption_STATE_PREBIND,
		}
		socketOptions = append(socketOptions, socketOption)
	}

	if keepAlive.Interval != nil {
		socketOption = &corev3.SocketOption{
			Description: "socket option for keep alive interval",
			Level:       0x6, // Darwin lacks syscall.SOL_TCP
			Name:        0x5, // Darwin lacks syscall.TCP_KEEPINTVL,
			Value:       &corev3.SocketOption_IntValue{IntValue: int64(*keepAlive.Interval)},
			State:       corev3.SocketOption_STATE_PREBIND,
		}
		socketOptions = append(socketOptions, socketOption)
	}

	return socketOptions

}
