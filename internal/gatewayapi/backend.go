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
	"k8s.io/utils/ptr"

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
	if backend.Spec.Type != nil &&
		*backend.Spec.Type == egv1a1.BackendTypeDynamicResolver {
		if len(backend.Spec.Endpoints) > 0 {
			return status.NewRouteStatusError(
				fmt.Errorf("DynamicResolver type cannot have endpoints specified"),
				status.RouteReasonInvalidBackendRef,
			)
		}

		if backend.Spec.TLS != nil &&
			!ptr.Deref(backend.Spec.TLS.InsecureSkipVerify, false) &&
			backend.Spec.TLS.WellKnownCACertificates == nil &&
			len(backend.Spec.TLS.CACertificateRefs) == 0 {
			return status.NewRouteStatusError(
				fmt.Errorf("must specify either CACertificateRefs or WellKnownCACertificates for DynamicResolver type when InsecureSkipVerify is unset or false"),
				status.RouteReasonInvalidBackendRef,
			)
		}

	} else if backend.Spec.TLS != nil {
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
