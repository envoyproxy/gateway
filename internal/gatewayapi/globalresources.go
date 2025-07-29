// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
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
	var errs error

	// Get the envoy client TLS secret. It is used for envoy to establish a TLS connection with control plane components,
	// including the rate limit server and the wasm HTTP server.
	envoyTLSSecret := resources.GetSecret(t.ControllerNamespace, envoyTLSSecretName)
	if envoyTLSSecret == nil {
		errs = errors.Join(errs, fmt.Errorf("envoy TLS secret %s/%s not found", t.ControllerNamespace, envoyTLSSecretName))
	}

	for _, xdsIR := range xdsIRs {
		switch {
		case t.MergeGateways:
			svcClusterName := string(t.GatewayClassName)
			svc := resources.GetServiceByLabels(map[string]string{OwningGatewayClassLabel: svcClusterName}, t.ControllerNamespace)
			if svc == nil {
				errs = errors.Join(errs, fmt.Errorf("envoy ServiceCluster for GatewayClass %s not found", t.GatewayClassName))
				break
			}

			bRef := gwapiv1.BackendObjectReference{
				Group:     GroupPtr(svc.GroupVersionKind().Group),
				Kind:      KindPtr(svc.Kind),
				Name:      gwapiv1.ObjectName(svc.Name),
				Namespace: NamespacePtr(svc.Namespace),
				Port:      PortNumPtr(svc.Spec.Ports[0].Port),
			}
			dst := t.processServiceDestinationSetting(svcClusterName, bRef, svc.Namespace, ir.AppProtocol(svc.Spec.Ports[0].Protocol), resources, resources.EnvoyProxyForGatewayClass)

			xdsIR.GlobalResources = &ir.GlobalResources{
				ProxyServiceClusters: []*ir.RouteDestination{{
					Name:     dst.Name,
					Settings: []*ir.DestinationSetting{dst},
				}},
			}
		default:
			for _, g := range gateways {
				svcClusterName := fmt.Sprintf("%s/%s", g.Namespace, g.Name)
				svc := resources.GetServiceByLabels(map[string]string{OwningGatewayNameLabel: g.Name, OwningGatewayNamespaceLabel: g.Namespace}, t.ControllerNamespace)
				if svc == nil {
					errs = errors.Join(errs, fmt.Errorf("envoy ServiceCluster for Gateway %s/%s not found", g.Namespace, g.Name))
					continue
				}

				bRef := gwapiv1.BackendObjectReference{
					Group:     GroupPtr(svc.GroupVersionKind().Group),
					Kind:      KindPtr(svc.Kind),
					Name:      gwapiv1.ObjectName(svc.Name),
					Namespace: NamespacePtr(svc.Namespace),
					Port:      PortNumPtr(svc.Spec.Ports[0].Port),
				}
				dst := t.processServiceDestinationSetting(svcClusterName, bRef, svc.Namespace, ir.AppProtocol(svc.Spec.Ports[0].Protocol), resources, g.envoyProxy)

				if xdsIR.GlobalResources == nil {
					xdsIR.GlobalResources = &ir.GlobalResources{}
				}

				xdsIR.GlobalResources.ProxyServiceClusters = append(
					xdsIR.GlobalResources.ProxyServiceClusters,
					&ir.RouteDestination{Name: dst.Name, Settings: []*ir.DestinationSetting{dst}},
				)

			}
		}

		if (containsGlobalRateLimit(xdsIR.HTTP) || containsWasm(xdsIR.HTTP)) && envoyTLSSecret != nil {
			if xdsIR.GlobalResources == nil {
				xdsIR.GlobalResources = &ir.GlobalResources{}
			}
			xdsIR.GlobalResources.EnvoyClientCertificate = &ir.TLSCertificate{
				Name:        irGlobalConfigName(envoyTLSSecret),
				Certificate: envoyTLSSecret.Data[corev1.TLSCertKey],
				PrivateKey:  envoyTLSSecret.Data[corev1.TLSPrivateKeyKey],
			}
		}
	}
	return errs
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
