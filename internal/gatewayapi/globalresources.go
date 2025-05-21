// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
)

const envoyTLSSecretName = "envoy"

// ProcessGlobalResources processes global resources that are not tied to a specific listener or route
func (t *Translator) ProcessGlobalResources(resources *resource.Resources, xdsIRs resource.XdsIRMap) error {
	// Get the envoy client TLS secret. It is used for envoy to establish a TLS connection with control plane components,
	// including the rate limit server and the wasm HTTP server.
	envoyTLSSecret := resources.GetSecret(t.ControllerNamespace, envoyTLSSecretName)
	if envoyTLSSecret == nil {
		return fmt.Errorf("envoy TLS secret %s/%s not found", t.ControllerNamespace, envoyTLSSecretName)
	}

	for _, xdsIR := range xdsIRs {
		// TODO zhaohuabing: this is also required by WASM
		if containsGlobalRateLimit(xdsIR.HTTP) {
			xdsIR.GlobalResources = &ir.GlobalResources{}
			xdsIR.GlobalResources.EnvoyClientCertificate = &ir.TLSCertificate{
				Name:        irGlobalConfigName(envoyTLSSecret),
				Certificate: envoyTLSSecret.Data[corev1.TLSCertKey],
				PrivateKey:  envoyTLSSecret.Data[corev1.TLSPrivateKeyKey],
			}
		}
	}
	return nil
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
