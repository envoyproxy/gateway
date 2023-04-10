// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

// RateLimitInfra defines managed rate limit service infrastructure.
// +k8s:deepcopy-gen=true
type RateLimitInfra struct {
	// ServiceConfigs for Rate limit service configuration.
	ServiceConfigs []*RateLimitServiceConfig
}

// RateLimitServiceConfig holds the rate limit service configurations
// defined here https://github.com/envoyproxy/ratelimit#configuration-1
// +k8s:deepcopy-gen=true
type RateLimitServiceConfig struct {
	// Name of the config file.
	Name string
	// Config contents saved as a YAML string.
	Config string
}
