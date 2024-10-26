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

func validateBackend(backend *egv1a1.Backend) error {
	for _, ep := range backend.Spec.Endpoints {
		if ep.FQDN != nil {
			hostname := ep.FQDN.Hostname
			// must be a valid hostname
			if errs := validation.IsDNS1123Subdomain(hostname); errs != nil {
				return fmt.Errorf("hostname %s is not a valid FQDN", hostname)
			}
			if len(strings.Split(hostname, ".")) < 2 {
				return fmt.Errorf("hostname %s should be a domain with at least two segments separated by dots", hostname)
			}
			// IP addresses are not allowed so parsing the hostname as an address needs to fail
			if _, err := netip.ParseAddr(hostname); err == nil {
				return fmt.Errorf("hostname %s is an IP address", hostname)
			}
		} else if ep.IP != nil {
			ip, err := netip.ParseAddr(ep.IP.Address)
			if err != nil {
				return fmt.Errorf("IP address %s is invalid", ep.IP.Address)
			} else if ip.IsLoopback() {
				return fmt.Errorf("IP address %s in the loopback range is not supported", ep.IP.Address)
			}
		}
	}
	return nil
}
