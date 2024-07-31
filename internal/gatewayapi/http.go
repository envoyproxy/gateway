// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"

	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

const (
	MinHTTP2InitialStreamWindowSize     = 65535      // https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/protocol.proto#envoy-v3-api-field-config-core-v3-http2protocoloptions-initial-stream-window-size
	MaxHTTP2InitialStreamWindowSize     = 2147483647 // https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/protocol.proto#envoy-v3-api-field-config-core-v3-http2protocoloptions-initial-stream-window-size
	MinHTTP2InitialConnectionWindowSize = MinHTTP2InitialStreamWindowSize
	MaxHTTP2InitialConnectionWindowSize = MaxHTTP2InitialStreamWindowSize
)

func buildIRHTTP2Settings(http2Settings *egv1a1.HTTP2Settings) (*ir.HTTP2Settings, error) {
	var (
		http2 = &ir.HTTP2Settings{}
		errs  error
	)

	if http2Settings.InitialStreamWindowSize != nil {
		initialStreamWindowSize, ok := http2Settings.InitialStreamWindowSize.AsInt64()
		switch {
		case !ok:
			errs = errors.Join(errs, fmt.Errorf("invalid InitialStreamWindowSize value %s", http2Settings.InitialStreamWindowSize.String()))
		case initialStreamWindowSize < MinHTTP2InitialStreamWindowSize || initialStreamWindowSize > MaxHTTP2InitialStreamWindowSize:
			errs = errors.Join(errs, fmt.Errorf("InitialStreamWindowSize value %s is out of range, must be between %d and %d",
				http2Settings.InitialStreamWindowSize.String(),
				MinHTTP2InitialStreamWindowSize,
				MaxHTTP2InitialStreamWindowSize))
		default:
			http2.InitialStreamWindowSize = ptr.To(uint32(initialStreamWindowSize))
		}
	}

	if http2Settings.InitialConnectionWindowSize != nil {
		initialConnectionWindowSize, ok := http2Settings.InitialConnectionWindowSize.AsInt64()
		switch {
		case !ok:
			errs = errors.Join(errs, fmt.Errorf("invalid InitialConnectionWindowSize value %s", http2Settings.InitialConnectionWindowSize.String()))
		case initialConnectionWindowSize < MinHTTP2InitialConnectionWindowSize || initialConnectionWindowSize > MaxHTTP2InitialConnectionWindowSize:
			errs = errors.Join(errs, fmt.Errorf("InitialConnectionWindowSize value %s is out of range, must be between %d and %d",
				http2Settings.InitialConnectionWindowSize.String(),
				MinHTTP2InitialConnectionWindowSize,
				MaxHTTP2InitialConnectionWindowSize))
		default:
			http2.InitialConnectionWindowSize = ptr.To(uint32(initialConnectionWindowSize))
		}
	}

	http2.MaxConcurrentStreams = http2Settings.MaxConcurrentStreams

	http2.TerminateConnOnError = http2Settings.TerminateConnOnError

	return http2, errs
}
