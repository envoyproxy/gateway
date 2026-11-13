// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Package validation is deprecated and will be removed in a future release.
// The validation logic now lives in an internal package because it is
// controller behavior, not part of the Envoy Gateway API.
package validation

import (
	"k8s.io/apimachinery/pkg/api/resource"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/validation"
)

// InvalidField is an EnvoyProxy field that failed validation.
//
// Deprecated: this package will be removed in a future release.
type InvalidField = validation.InvalidField

// ValidateEnvoyGateway validates the provided EnvoyGateway.
//
// Deprecated: this package will be removed in a future release.
func ValidateEnvoyGateway(eg *egv1a1.EnvoyGateway) error {
	return validation.ValidateEnvoyGateway(eg)
}

// ValidateRedisURL validates a rate limit Redis URL.
//
// Deprecated: this package will be removed in a future release.
func ValidateRedisURL(redisURL string) error {
	return validation.ValidateRedisURL(redisURL)
}

// ValidateEnvoyProxy validates the provided EnvoyProxy.
//
// Deprecated: this package will be removed in a future release.
func ValidateEnvoyProxy(proxy *egv1a1.EnvoyProxy) error {
	return validation.ValidateEnvoyProxy(proxy)
}

// ValidateEnvoyProxyFields returns the EnvoyProxy fields that failed validation.
//
// Deprecated: this package will be removed in a future release.
func ValidateEnvoyProxyFields(proxy *egv1a1.EnvoyProxy) ([]InvalidField, error) {
	return validation.ValidateEnvoyProxyFields(proxy)
}

// DropInvalidFields returns a copy of the EnvoyProxy with invalid fields removed.
//
// Deprecated: this package will be removed in a future release.
func DropInvalidFields(proxy *egv1a1.EnvoyProxy) *egv1a1.EnvoyProxy {
	return validation.DropInvalidFields(proxy)
}

// ValidateMinContentLength validates a compressor minimum content length.
//
// Deprecated: this package will be removed in a future release.
func ValidateMinContentLength(minContentLength *resource.Quantity) error {
	return validation.ValidateMinContentLength(minContentLength)
}

// ValidateRouteStatName validates a route stat name template.
//
// Deprecated: this package will be removed in a future release.
func ValidateRouteStatName(routeStatName string) error {
	return validation.ValidateRouteStatName(routeStatName)
}

// ValidateClusterStatName validates a cluster stat name template.
//
// Deprecated: this package will be removed in a future release.
func ValidateClusterStatName(clusterStatName string) error {
	return validation.ValidateClusterStatName(clusterStatName)
}
