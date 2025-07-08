// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
)

func (t *Translator) ProcessProxyCluster(acceptedGateways []*GatewayContext, resources *resource.Resources, xdsIR resource.XdsIRMap) {
	for _, g := range acceptedGateways {
		if g == nil || g.Gateway == nil {
			continue
		}

		irKey := t.getIRKey(g.Gateway)
		/*
			if xdsIR[irKey].ProxyInfraCluster != nil {
				continue
			}
		*/
		if ptr.Deref(xdsIR[irKey].GlobalResources, ir.GlobalResources{}).ProxyInfraCluster != nil {
			continue
		}

		var svcName string
		if t.MergeGateways {
			svcName = t.expectedResourceHashedName(string(t.GatewayClassName))
		} else {
			svcName = t.expectedResourceHashedName(fmt.Sprintf("%s/%s", g.Namespace, g.Name))
		}

		svc := resources.GetService(t.ControllerNamespace, svcName)
		if svc == nil {
			return
		}

		ds := t.processEnvoyServiceDestinationSetting(svc.Name, svc, resources)
		ds.IPFamily = getServiceIPFamily(svc)

		if xdsIR[irKey].GlobalResources == nil {
			xdsIR[irKey].GlobalResources = &ir.GlobalResources{}
		}
		xdsIR[irKey].GlobalResources.ProxyInfraCluster = &ir.ProxyInfraCluster{
			Name:        svc.Name,
			Destination: ds,
		}

		if t.MergeGateways {
			return
		}
	}
}

// expectedResourceHashedName returns expected resource hashed name including up to the 48 characters of the original name.
func (t *Translator) expectedResourceHashedName(name string) string {
	hashedName := utils.GetHashedName(name, 48)
	return fmt.Sprintf("%s-%s", config.EnvoyPrefix, hashedName)
}

func (t *Translator) processEnvoyServiceDestinationSetting(
	name string,
	service *corev1.Service,
	resources *resource.Resources,
) *ir.DestinationSetting {
	var (
		endpoints []*ir.DestinationEndpoint
		addrType  *ir.DestinationAddressType
	)

	endpointSlices := resources.GetEndpointSlicesForBackend(service.Namespace, service.Name, resource.KindService)
	endpoints, addrType = getIREndpointsFromEndpointSlices(endpointSlices, service.Spec.Ports[0].Name, service.Spec.Ports[0].Protocol)

	return &ir.DestinationSetting{
		Name:        name,
		Protocol:    ir.HTTP,
		Endpoints:   endpoints,
		AddressType: addrType,
		// Use Zone Aware Lb so locality info is injected for endpoints
		PreferLocal: &ir.PreferLocalZone{MinEndpointsThreshold: ptr.To[uint64](1)},
	}
}
