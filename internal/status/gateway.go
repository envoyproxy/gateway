package status

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

// SetGatewayStatus adds or updates status for the provided Gateway.
func SetGatewayStatus(gw *gwapiv1b1.Gateway, scheduled bool, svc *corev1.Service, deployment *appsv1.Deployment) *gwapiv1b1.Gateway {
	computeGatewayStatusAddrs(gw, svc)
	gw.Status.Conditions = mergeConditions(
		gw.Status.Conditions,
		computeGatewayScheduledCondition(gw, scheduled),
		computeGatewayReadyCondition(gw, deployment),
	)
	return gw
}

// computeGatewayStatusAddrs computes status addresses for the provided gateway
// based on the status IP/Hostname of svc.
func computeGatewayStatusAddrs(gw *gwapiv1b1.Gateway, svc *corev1.Service) {
	var addrs, hostnames []string
	if svc != nil {
		for i := range svc.Status.LoadBalancer.Ingress {
			switch {
			case len(svc.Status.LoadBalancer.Ingress[i].IP) > 0:
				addrs = append(addrs, svc.Status.LoadBalancer.Ingress[i].IP)
			case len(svc.Status.LoadBalancer.Ingress[i].Hostname) > 0:
				hostnames = append(hostnames, svc.Status.LoadBalancer.Ingress[i].Hostname)
			}
		}
	}

	var gwAddrs []gwapiv1b1.GatewayAddress
	for i := range addrs {
		addr := gwapiv1b1.GatewayAddress{
			Type:  gatewayapi.GatewayAddressTypePtr(gwapiv1b1.IPAddressType),
			Value: addrs[i],
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
}
