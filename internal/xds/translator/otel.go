// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// buildGrpcInitialMetadata converts HTTP headers to gRPC initial metadata.
func buildGrpcInitialMetadata(headers []gwapiv1.HTTPHeader) []*corev3.HeaderValue {
	if len(headers) == 0 {
		return nil
	}
	result := make([]*corev3.HeaderValue, len(headers))
	for i, h := range headers {
		result[i] = &corev3.HeaderValue{
			Key:   string(h.Name),
			Value: h.Value,
		}
	}
	return result
}
