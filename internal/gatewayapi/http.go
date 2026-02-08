// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"
	"time"

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
	if http2Settings == nil {
		return nil, nil
	}
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

	if http2Settings.OnInvalidMessage != nil {
		switch *http2Settings.OnInvalidMessage {
		case egv1a1.InvalidMessageActionTerminateStream:
			http2.ResetStreamOnError = ptr.To(true)
		case egv1a1.InvalidMessageActionTerminateConnection:
			http2.ResetStreamOnError = ptr.To(false)
		}
	}

	return http2, errs
}

func buildIRHTTP2ClientSettings(http2Settings *egv1a1.HTTP2ClientSettings) (*ir.HTTP2Settings, error) {
	if http2Settings == nil {
		return nil, nil
	}

	// Reuse shared builder for common fields
	http2, errs := buildIRHTTP2Settings(&http2Settings.HTTP2Settings)
	if http2 == nil {
		http2 = &ir.HTTP2Settings{}
	}

	// Handle keepalive (ClientTrafficPolicy-specific)
	if http2Settings.ConnectionKeepalive != nil {
		keepalive := &ir.HTTP2ConnectionKeepalive{}
		if http2Settings.ConnectionKeepalive.Interval != nil {
			d, err := time.ParseDuration(string(*http2Settings.ConnectionKeepalive.Interval))
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("invalid ConnectionKeepalive.Interval: %w", err))
			} else {
				keepalive.Interval = ptr.To(uint32(d.Seconds()))
			}
		}
		if http2Settings.ConnectionKeepalive.Timeout != nil {
			d, err := time.ParseDuration(string(*http2Settings.ConnectionKeepalive.Timeout))
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("invalid ConnectionKeepalive.Timeout: %w", err))
			} else {
				keepalive.Timeout = ptr.To(uint32(d.Seconds()))
			}
		}
		if http2Settings.ConnectionKeepalive.ConnectionIdleInterval != nil {
			d, err := time.ParseDuration(string(*http2Settings.ConnectionKeepalive.ConnectionIdleInterval))
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("invalid ConnectionKeepalive.ConnectionIdleInterval: %w", err))
			} else {
				keepalive.ConnectionIdleInterval = ptr.To(uint32(d.Seconds()))
			}
		}
		http2.ConnectionKeepalive = keepalive
	}

	return http2, errs
}
