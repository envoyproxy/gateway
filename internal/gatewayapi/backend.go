// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"net/netip"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
)

func (t *Translator) ProcessBackends(backends []*egv1a1.Backend) []*egv1a1.Backend {
	var res []*egv1a1.Backend
	for _, backend := range backends {
		backend := backend.DeepCopy()

		// Ensure Backends are enabled
		if !t.BackendEnabled {
			status.UpdateBackendStatusAcceptedCondition(backend, false,
				"The Backend was not accepted since Backend is not enabled in Envoy Gateway Config")
		} else {
			if err := validateBackend(backend); err != nil {
				status.UpdateBackendStatusAcceptedCondition(backend, false, fmt.Sprintf("The Backend was not accepted: %s", err.Error()))
			} else {
				status.UpdateBackendStatusAcceptedCondition(backend, true, "The Backend was accepted")
			}
		}

		res = append(res, backend)
	}

	return res
}

func validateBackend(backend *egv1a1.Backend) status.Error {
	backendType := egv1a1.BackendTypeEndpoints
	if backend.Spec.Type != nil {
		backendType = *backend.Spec.Type
	}

	switch backendType {
	case egv1a1.BackendTypeDynamicResolver:
		if len(backend.Spec.Endpoints) > 0 {
			return status.NewRouteStatusError(
				fmt.Errorf("DynamicResolver type cannot have endpoints specified"),
				status.RouteReasonInvalidBackendRef,
			)
		}
	case egv1a1.BackendTypeStaticResolver:
		if len(backend.Spec.Endpoints) > 0 {
			return status.NewRouteStatusError(
				fmt.Errorf("StaticResolver type cannot have endpoints specified"),
				status.RouteReasonInvalidBackendRef,
			)
		}
		if err := validateStaticResolverSettings(backend.Spec.StaticResolver); err != nil {
			return status.NewRouteStatusError(err, status.RouteReasonInvalidBackendRef)
		}
	default: // BackendTypeEndpoints
		if backend.Spec.TLS != nil {
			if backend.Spec.TLS.WellKnownCACertificates != nil {
				return status.NewRouteStatusError(
					fmt.Errorf("TLS.WellKnownCACertificates settings can only be specified for DynamicResolver backends"),
					status.RouteReasonInvalidBackendRef,
				)
			}
			if len(backend.Spec.TLS.CACertificateRefs) > 0 {
				return status.NewRouteStatusError(
					fmt.Errorf("TLS.CACertificateRefs settings can only be specified for DynamicResolver backends"),
					status.RouteReasonInvalidBackendRef,
				)
			}
		}
		if backend.Spec.StaticResolver != nil {
			return status.NewRouteStatusError(
				fmt.Errorf("StaticResolver settings can only be specified for StaticResolver backends"),
				status.RouteReasonInvalidBackendRef,
			)
		}
	}

	for _, ep := range backend.Spec.Endpoints {
		if ep.FQDN != nil {
			hostname := ep.FQDN.Hostname
			// must be a valid hostname
			if errs := validation.IsDNS1123Subdomain(hostname); errs != nil {
				return status.NewRouteStatusError(
					fmt.Errorf("hostname %s is not a valid FQDN", hostname),
					status.RouteReasonInvalidAddress,
				)
			}
			if len(strings.Split(hostname, ".")) < 2 {
				return status.NewRouteStatusError(
					fmt.Errorf("hostname %s should be a domain with at least two segments separated by dots", hostname),
					status.RouteReasonInvalidAddress,
				)
			}
			// IP addresses are not allowed so parsing the hostname as an address needs to fail
			if _, err := netip.ParseAddr(hostname); err == nil {
				return status.NewRouteStatusError(
					fmt.Errorf("hostname %s is an IP address", hostname),
					status.RouteReasonInvalidAddress,
				)
			}
		} else if ep.IP != nil {
			ip, err := netip.ParseAddr(ep.IP.Address)
			if err != nil {
				return status.NewRouteStatusError(
					fmt.Errorf("IP address %s is invalid", ep.IP.Address),
					status.RouteReasonInvalidAddress,
				)
			} else if ip.IsLoopback() {
				return status.NewRouteStatusError(
					fmt.Errorf("IP address %s in the loopback range is not supported", ep.IP.Address),
					status.RouteReasonInvalidAddress,
				)
			}
		}
	}
	return nil
}

func validateStaticResolverSettings(settings *egv1a1.StaticResolverSettings) error {
	if len(settings.OverrideHostSources) == 0 {
		return fmt.Errorf("at least one override host source must be specified")
	}

	for i, source := range settings.OverrideHostSources {
		if source.Header == nil && source.Metadata == nil {
			return fmt.Errorf("override host source %d must specify either header or metadata", i)
		}
		if source.Header != nil && source.Metadata != nil {
			return fmt.Errorf("override host source %d cannot specify both header and metadata", i)
		}
		if source.Header != nil && *source.Header == "" {
			return fmt.Errorf("override host source %d header name cannot be empty", i)
		}
		if source.Metadata != nil {
			if source.Metadata.Key == "" {
				return fmt.Errorf("override host source %d metadata key cannot be empty", i)
			}
			for j, pathSegment := range source.Metadata.Path {
				if pathSegment.Key == "" {
					return fmt.Errorf("override host source %d metadata path segment %d key cannot be empty", i, j)
				}
			}
		}
	}

	return nil
}
