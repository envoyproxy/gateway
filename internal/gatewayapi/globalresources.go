// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
)

const envoyTLSSecretName = "envoy"

// ProcessGlobalResources processes global resources that are not tied to a specific listener or route
func (t *Translator) ProcessGlobalResources(resources *resource.Resources, xdsIRs resource.XdsIRMap, gateways []*GatewayContext) error {
	// Add the ProxyServiceCluster information for each gateway to the IR map
	for _, gateway := range gateways {
		// Get the gateway IR key and RouteDestination representing the ProxyServiceCluster
		irKey, rDest := t.processServiceClusterForGateway(gateway, resources)

		if xdsIRs[irKey] == nil {
			continue
		}
		if xdsIRs[irKey].GlobalResources == nil {
			xdsIRs[irKey].GlobalResources = &ir.GlobalResources{}
		}
		xdsIRs[irKey].GlobalResources.ProxyServiceCluster = rDest

		// For merged gateways we only need to process once
		if t.MergeGateways {
			break
		}
	}

	// Get the envoy client TLS secret. It is used for envoy to establish a TLS connection with control plane components,
	// including the rate limit server and the wasm HTTP server.
	envoyTLSSecret := t.GetSecret(t.ControllerNamespace, envoyTLSSecretName)
	if envoyTLSSecret == nil {
		return fmt.Errorf("envoy TLS secret %s/%s not found", t.ControllerNamespace, envoyTLSSecretName)
	}

	for _, xdsIR := range xdsIRs {
		if containsGlobalRateLimit(xdsIR.HTTP) || containsWasm(xdsIR.HTTP) {
			if xdsIR.GlobalResources == nil {
				xdsIR.GlobalResources = &ir.GlobalResources{}
			}
			if containsGlobalRateLimit(xdsIR.HTTP) {
				xdsIR.GlobalResources.RateLimitServiceCluster = t.processRateLimitServiceCluster(resources)
			}
			xdsIR.GlobalResources.EnvoyClientCertificate = &ir.TLSCertificate{
				Name:        irGlobalConfigName(envoyTLSSecret),
				Certificate: envoyTLSSecret.Data[corev1.TLSCertKey],
				PrivateKey:  envoyTLSSecret.Data[corev1.TLSPrivateKeyKey],
			}
		}
	}

	return nil
}

func (t *Translator) processRateLimitServiceCluster(resources *resource.Resources) *ir.RouteDestination {
	const rateLimitServiceName = "envoy-ratelimit"

	var service *corev1.Service
	for _, candidate := range resources.Services {
		if candidate.Namespace == t.ControllerNamespace && candidate.Name == rateLimitServiceName {
			service = candidate
			break
		}
	}
	if service == nil {
		return nil
	}

	var servicePort *corev1.ServicePort
	for i := range service.Spec.Ports {
		if service.Spec.Ports[i].Name == "http" {
			servicePort = &service.Spec.Ports[i]
			break
		}
	}
	if servicePort == nil {
		return nil
	}

	endpointSlices := resources.GetEndpointSlicesForBackend(
		t.ControllerNamespace, rateLimitServiceName, resource.KindService)
	endpoints, addressType := getIREndpointsFromEndpointSlices(endpointSlices, servicePort.Name, servicePort.Protocol)
	setting := &ir.DestinationSetting{
		Name:        "ratelimit_cluster/backend/-1",
		Protocol:    ir.GRPC,
		Endpoints:   endpoints,
		AddressType: addressType,
		Metadata:    nil,
	}
	return &ir.RouteDestination{
		Name:     "ratelimit_cluster",
		Settings: []*ir.DestinationSetting{setting},
	}
}

// processServiceClusterForGateway returns the matching IR key for a gateway and builds a RouteDestination to represent the ProxyServiceCluster
func (t *Translator) processServiceClusterForGateway(gateway *GatewayContext, resources *resource.Resources) (string, *ir.RouteDestination) {
	irKey := t.getIRKey(gateway.Gateway)
	labels := OwnerLabels(gateway.Gateway, t.MergeGateways)

	svcClusterNamespace := t.ControllerNamespace
	if t.GatewayNamespaceMode {
		svcClusterNamespace = gateway.Namespace
	}
	svcCluster := resources.GetServiceByLabels(labels, svcClusterNamespace)

	// Service lookup fails on first iteration
	if svcCluster == nil {
		return "", nil
	}

	bRef := gwapiv1.BackendObjectReference{
		Group:     GroupPtr(svcCluster.GroupVersionKind().Group),
		Kind:      KindPtr(svcCluster.Kind),
		Name:      gwapiv1.ObjectName(svcCluster.Name),
		Namespace: NamespacePtr(svcCluster.Namespace),
		Port:      PortNumPtr(svcCluster.Spec.Ports[0].Port),
	}
	dst, err := t.processServiceDestinationSetting(irKey, bRef, svcCluster.Namespace, ir.AppProtocol(svcCluster.Spec.Ports[0].Protocol), resources.EnvoyProxyForGatewayClass, nil, nil)
	if err != nil {
		return "", nil
	}

	return irKey, &ir.RouteDestination{
		Name:     dst.Name,
		Settings: []*ir.DestinationSetting{dst},
		Metadata: dst.Metadata,
	}
}

func irGlobalConfigName(object metav1.Object) string {
	return fmt.Sprintf("%s/%s", object.GetNamespace(), object.GetName())
}

func containsGlobalRateLimit(httpListeners []*ir.HTTPListener) bool {
	for _, httpListener := range httpListeners {
		for _, route := range httpListener.Routes {
			if route.Traffic != nil &&
				route.Traffic.RateLimit != nil &&
				route.Traffic.RateLimit.Global != nil {
				return true
			}
		}
	}
	return false
}

func containsWasm(httpListeners []*ir.HTTPListener) bool {
	for _, httpListener := range httpListeners {
		for _, route := range httpListener.Routes {
			if route.EnvoyExtensions != nil &&
				len(route.EnvoyExtensions.Wasms) > 0 {
				return true
			}
		}
	}
	return false
}
