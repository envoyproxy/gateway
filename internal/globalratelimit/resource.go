// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package globalratelimit

import (
	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
)

// GetRateLimitURL returns the URL for the rate limit service.
func GetRateLimitURL(eg *egv1a1.EnvoyGateway, namespace, dnsDomain string) string {
	if eg != nil && eg.RateLimit != nil && eg.RateLimit.URL != nil {
		return *eg.RateLimit.URL
	}
	return ratelimit.GetServiceURL(namespace, dnsDomain)
}
