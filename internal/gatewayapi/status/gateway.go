// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// UpdateGatewayStatusAcceptedCondition updates the status condition for the provided Gateway based on the accepted state.
func UpdateGatewayStatusAcceptedCondition(gw *gwapiv1.Gateway, accepted bool) *gwapiv1.Gateway {
	gw.Status.Conditions = MergeConditions(gw.Status.Conditions, computeGatewayAcceptedCondition(gw, accepted))
	return gw
}

// UpdateGatewayStatusProgrammedCondition updates the status addresses for the provided gateway
// based on the status IP/Hostname of svc and updates the Programmed condition based on the
// service and deployment state.
func UpdateGatewayStatusProgrammedCondition(gw *gwapiv1.Gateway, svc *corev1.Service, deployment *appsv1.Deployment, nodeAddresses ...string) {
	var addresses, hostnames []string
	// Update the status addresses field.
	if svc != nil {
		// If the addresses is explicitly set in the Gateway spec by the user, use it
		// to populate the Status
		if len(gw.Spec.Addresses) > 0 {
			// Make sure the addresses have been populated into ExternalIPs/ClusterIPs
			// and use that value
			if len(svc.Spec.ExternalIPs) > 0 {
				addresses = append(addresses, svc.Spec.ExternalIPs...)
			} else if len(svc.Spec.ClusterIPs) > 0 {
				addresses = append(addresses, svc.Spec.ClusterIPs...)
			}
		} else {
			if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
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
			}

			if svc.Spec.Type == corev1.ServiceTypeClusterIP {
				for i := range svc.Spec.ClusterIPs {
					if svc.Spec.ClusterIPs[i] != "" {
						addresses = append(addresses, svc.Spec.ClusterIPs[i])
					}
				}
			}

			if svc.Spec.Type == corev1.ServiceTypeNodePort {
				addresses = nodeAddresses
			}
		}

		var gwAddresses []gwapiv1.GatewayStatusAddress
		for i := range addresses {
			addr := gwapiv1.GatewayStatusAddress{
				Type:  ptr.To(gwapiv1.IPAddressType),
				Value: addresses[i],
			}
			gwAddresses = append(gwAddresses, addr)
		}

		for i := range hostnames {
			addr := gwapiv1.GatewayStatusAddress{
				Type:  ptr.To(gwapiv1.HostnameAddressType),
				Value: hostnames[i],
			}
			gwAddresses = append(gwAddresses, addr)
		}

		gw.Status.Addresses = gwAddresses
	} else {
		gw.Status.Addresses = nil
	}
	// Update the programmed condition.
	gw.Status.Conditions = MergeConditions(gw.Status.Conditions, computeGatewayProgrammedCondition(gw, deployment))
}

func SetGatewayListenerStatusCondition(gateway *gwapiv1.Gateway, listenerStatusIdx int,
	conditionType gwapiv1.ListenerConditionType, status metav1.ConditionStatus, reason gwapiv1.ListenerConditionReason, message string,
) {
	cond := metav1.Condition{
		Type:               string(conditionType),
		Status:             status,
		Reason:             string(reason),
		Message:            message,
		ObservedGeneration: gateway.Generation,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	gateway.Status.Listeners[listenerStatusIdx].Conditions = MergeConditions(gateway.Status.Listeners[listenerStatusIdx].Conditions, cond)
}

// computeGatewayAcceptedCondition computes the Gateway Accepted status condition.
func computeGatewayAcceptedCondition(gw *gwapiv1.Gateway, accepted bool) metav1.Condition {
	switch accepted {
	case true:
		return newCondition(string(gwapiv1.GatewayReasonAccepted), metav1.ConditionTrue,
			string(gwapiv1.GatewayReasonAccepted),
			"The Gateway has been scheduled by Envoy Gateway", time.Now(), gw.Generation)
	default:
		return newCondition(string(gwapiv1.GatewayReasonAccepted), metav1.ConditionFalse,
			string(gwapiv1.GatewayReasonAccepted),
			"The Gateway has not been scheduled by Envoy Gateway", time.Now(), gw.Generation)
	}
}

// computeGatewayProgrammedCondition computes the Gateway Programmed status condition.
// Programmed condition surfaces true when the Envoy Deployment status is ready.
func computeGatewayProgrammedCondition(gw *gwapiv1.Gateway, deployment *appsv1.Deployment) metav1.Condition {
	if len(gw.Status.Addresses) == 0 {
		return newCondition(string(gwapiv1.GatewayConditionProgrammed), metav1.ConditionFalse,
			string(gwapiv1.GatewayReasonAddressNotAssigned),
			"No addresses have been assigned to the Gateway", time.Now(), gw.Generation)
	}

	// If there are no available replicas for the Envoy Deployment, don't
	// mark the Gateway as ready yet.

	if deployment == nil || deployment.Status.AvailableReplicas == 0 {
		return newCondition(string(gwapiv1.GatewayConditionProgrammed), metav1.ConditionFalse,
			string(gwapiv1.GatewayReasonNoResources),
			"Deployment replicas unavailable", time.Now(), gw.Generation)
	}

	message := fmt.Sprintf("Address assigned to the Gateway, %d/%d envoy Deployment replicas available",
		deployment.Status.AvailableReplicas, deployment.Status.Replicas)
	return newCondition(string(gwapiv1.GatewayConditionProgrammed), metav1.ConditionTrue,
		string(gwapiv1.GatewayConditionProgrammed), message, time.Now(), gw.Generation)
}
