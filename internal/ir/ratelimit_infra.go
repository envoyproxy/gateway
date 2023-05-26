// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

// RateLimitInfra defines managed rate limit service infrastructure.
// +k8s:deepcopy-gen=true
type RateLimitInfra struct {
	// ServiceNames for Rate limit service configuration name.
	ServiceNames []string
}
