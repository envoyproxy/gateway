// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

var _ AddressesTranslator = (*Translator)(nil)

type AddressesTranslator interface {
	ProcessAddresses(gateways []*GatewayContext, xdsIR resource.XdsIRMap, infraIR resource.InfraIRMap)
}

func (t *Translator) ProcessAddresses(gateways []*GatewayContext, xdsIR resource.XdsIRMap, infraIR resource.InfraIRMap) {
	for _, gateway := range gateways {
		// Infra IR already exist
		irKey := t.getIRKey(gateway.Gateway)
		gwInfraIR := infraIR[irKey]

		var ipAddr []string
		for _, addr := range gateway.Spec.Addresses {
			if *addr.Type == gwapiv1.IPAddressType {
				ipAddr = append(ipAddr, addr.Value)
			}
		}
		gwInfraIR.Proxy.Addresses = ipAddr
	}
}
