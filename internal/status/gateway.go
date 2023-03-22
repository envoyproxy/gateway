// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

// UpdateGatewayStatusAcceptedCondition updates the status condition for the provided Gateway based on the accepted state.
func UpdateGatewayStatusAcceptedCondition(gw *gwapiv1b1.Gateway, accepted bool) *gwapiv1b1.Gateway {
	gw.Status.Conditions = MergeConditions(gw.Status.Conditions, computeGatewayAcceptedCondition(gw, accepted))
	return gw
}

// UpdateGatewayStatusProgrammedCondition updates the status addresses for the provided gateway
// based on the status IP/Hostname of svc and updates the Programmed condition based on the
// service and deployment state.
func UpdateGatewayStatusProgrammedCondition(gw *gwapiv1b1.Gateway, svc *corev1.Service, deployment *appsv1.Deployment) {
	var addresses, hostnames []string
	// Update the status addresses field.
	if svc != nil {
		for i := range svc.Status.LoadBalancer.Ingress {
			switch {
			case len(svc.Status.LoadBalancer.Ingress[i].IP) > 0:
				addresses = append(addresses, svc.Status.LoadBalancer.Ingress[i].IP)
			case len(svc.Status.LoadBalancer.Ingress[i].Hostname) > 0:
				// Remove when the following supports the hostname address type:
				// https://github.com/kubernetes-sigs/gateway-api/blob/v0.5.0/conformance/utils/kubernetes/helpers.go#L201-L207
				if svc.Status.LoadBalancer.Ingress[i].Hostname == "localhost" {
					addresses = append(addresses, "127.0.0.1")
				}
				hostnames = append(hostnames, svc.Status.LoadBalancer.Ingress[i].Hostname)
			}
		}

		var gwAddrs []gwapiv1b1.GatewayAddress
		for i := range addresses {
			addr := gwapiv1b1.GatewayAddress{
				Type:  gatewayapi.GatewayAddressTypePtr(gwapiv1b1.IPAddressType),
				Value: addresses[i],
			}
			gwAddrs = append(gwAddrs, addr)
		}

		for i := range hostnames {
			addr := gwapiv1b1.GatewayAddress{
				Type:  gatewayapi.GatewayAddressTypePtr(gwapiv1b1.HostnameAddressType),
				Value: hostnames[i],
			}
			gwAddrs = append(gwAddrs, addr)
		}

		gw.Status.Addresses = gwAddrs
	} else {
		gw.Status.Addresses = nil
	}
	// Update the programmed condition.
	gw.Status.Conditions = MergeConditions(gw.Status.Conditions, computeGatewayProgrammedCondition(gw, deployment))
}
