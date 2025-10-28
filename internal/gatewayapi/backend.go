// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"net/netip"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
)

func (t *Translator) ProcessBackends(backends []*egv1a1.Backend, backendTLSPolicies []*gwapiv1.BackendTLSPolicy) []*egv1a1.Backend {
	res := make([]*egv1a1.Backend, 0, len(backends))
	for _, backend := range backends {
		// Ensure Backends are enabled
		if !t.BackendEnabled {
			status.UpdateBackendStatusAcceptedCondition(backend, false,
				"The Backend was not accepted since Backend is not enabled in Envoy Gateway Config")
		} else {
			if err := validateBackend(backend, backendTLSPolicies); err != nil {
				status.UpdateBackendStatusAcceptedCondition(backend, false, fmt.Sprintf("The Backend was not accepted: %s", err.Error()))
			} else {
				status.UpdateBackendStatusAcceptedCondition(backend, true, "The Backend was accepted")
			}
		}

		res = append(res, backend)
	}

	return res
}

func validateBackend(backend *egv1a1.Backend, backendTLSPolicies []*gwapiv1.BackendTLSPolicy) status.Error {
	if backend.Spec.Type != nil && *backend.Spec.Type == egv1a1.BackendTypeDynamicResolver {
		if len(backend.Spec.Endpoints) > 0 {
			return status.NewRouteStatusError(
				fmt.Errorf("DynamicResolver type cannot have endpoints specified"),
				status.RouteReasonInvalidBackendRef,
			)
		}
	}

	// Validate CACert is specified if InsecureSkipVerify is false
	if err := validateBackendTLSSettings(backend, backendTLSPolicies); err != nil {
		return err
	}

	for _, ep := range backend.Spec.Endpoints {
		if ep.Hostname != nil {
			routeErr := validateHostname(*ep.Hostname, "hostname")
			if routeErr != nil {
				return routeErr
			}
		}
		if ep.FQDN != nil {
			routeErr := validateHostname(ep.FQDN.Hostname, "FQDN")
			if routeErr != nil {
				return routeErr
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

// validateBackendTLSSettings validates CACert is specified if InsecureSkipVerify is false
func validateBackendTLSSettings(backend *egv1a1.Backend, backendTLSPolicies []*gwapiv1.BackendTLSPolicy) status.Error {
	if backend.Spec.TLS == nil {
		return nil
	}

	if !ptr.Deref(backend.Spec.TLS.InsecureSkipVerify, false) {
		var (
			backendTLSHasCACerts         bool
			backendTLSPoliciesHasCACerts bool
			ports                        = make(map[string]bool, len(backend.Spec.Endpoints))
		)

		// Check if the backend has WellKnownCACertificates or CACertificateRefs
		backendTLSHasCACerts = backend.Spec.TLS.WellKnownCACertificates != nil || len(backend.Spec.TLS.CACertificateRefs) > 0

		// If the backend has no CACert, check if any associated BackendTLSPolicy has CACert
		if !backendTLSHasCACerts {
			// Collect all ports from the backend endpoints
			for _, endpoint := range backend.Spec.Endpoints {
				switch {
				case endpoint.FQDN != nil:
					ports[strconv.Itoa(int(endpoint.FQDN.Port))] = false
				case endpoint.IP != nil:
					ports[strconv.Itoa(int(endpoint.IP.Port))] = false
				}
			}

			for _, policy := range backendTLSPolicies {
				for _, target := range policy.Spec.TargetRefs {
					if string(target.Group) == egv1a1.GroupName &&
						string(target.Kind) == egv1a1.KindBackend &&
						string(target.Name) == backend.Name {
						// If a BackendTLSPolicy without a SectionName is found and it has CACertificates, then the backend is valid
						if target.SectionName == nil {
							if policy.Spec.Validation.WellKnownCACertificates != nil ||
								len(policy.Spec.Validation.CACertificateRefs) > 0 {
								backendTLSPoliciesHasCACerts = true
								break
							}
						} else {
							if _, ok := ports[string(*target.SectionName)]; ok {
								ports[string(*target.SectionName)] = true
							}
						}
					}
				}
			}
			// If any port has no BackendTLSPolicy with CACertificates, then the backend is invalid
			if !backendTLSPoliciesHasCACerts && len(ports) > 0 {
				portsWithCert := 0
				for _, hasCACert := range ports {
					if !hasCACert {
						break
					}
					portsWithCert++
				}
				// If all ports have BackendTLSPolicy with CACertificates, then the backend is valid
				if portsWithCert == len(ports) {
					backendTLSPoliciesHasCACerts = true
				}
			}

			if !backendTLSPoliciesHasCACerts {
				return status.NewRouteStatusError(
					fmt.Errorf("must specify either CACertificateRefs or WellKnownCACertificates when InsecureSkipVerify is unset or false"),
					status.RouteReasonInvalidBackendRef,
				)
			}
		}
	}

	if backend.Spec.TLS.BackendTLSConfig != nil && backend.Spec.TLS.ClientCertificateRef != nil {
		ns := string(ptr.Deref(backend.Spec.TLS.ClientCertificateRef.Namespace, "default"))
		if ns != backend.Namespace {
			return status.NewRouteStatusError(
				fmt.Errorf("clientCertificateRef Secret is not located in the same namespace as Backend. Secret namespace: %s does not match Backend namespace: %s", ns, backend.Namespace),
				status.RouteReasonInvalidBackendRef,
			)
		}
	}
	return nil
}

func validateHostname(hostname, typeName string) *status.RouteStatusError {
	// must be a valid hostname
	if errs := validation.IsDNS1123Subdomain(hostname); errs != nil {
		return status.NewRouteStatusError(
			fmt.Errorf("hostname %s is not a valid %s", hostname, typeName),
			status.RouteReasonInvalidAddress,
		)
	}
	if len(strings.Split(hostname, ".")) < 2 {
		return status.NewRouteStatusError(
			fmt.Errorf("hostname %s should be a domain with at least two segments separated by dots", hostname),
			status.RouteReasonInvalidAddress,
		)
	}
	// IP addresses are not allowed, so parsing the hostname as an address needs to fail
	if _, err := netip.ParseAddr(hostname); err == nil {
		return status.NewRouteStatusError(
			fmt.Errorf("hostname %s is an IP address", hostname),
			status.RouteReasonInvalidAddress,
		)
	}

	return nil
}
