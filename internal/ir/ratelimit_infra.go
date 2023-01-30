// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

// RateLimitInfra defines managed rate limit service infrastructure.
// +k8s:deepcopy-gen=true
type RateLimitInfra struct {
	// Rate limit service configuration
	Configs []*RateLimitServiceConfig
	// Backend holds configuration associated with the backend database.
	Backend *RateLimitDBBackend
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

// RateLimitDBBackend defines the database backend properties
// associated with the rate limit service.
// +k8s:deepcopy-gen=true
type RateLimitDBBackend struct {
	// Redis backend details.
	Redis *RateLimitRedis
}

// RateLimitRedis defines the redis database configuration.
// +k8s:deepcopy-gen=true
type RateLimitRedis struct {
	// URL of the Redis Database.
	URL string
}
