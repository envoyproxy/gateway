// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package proxy

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/yaml"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	gwapiresource "github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
)

func TestServiceHTTP3Port(t *testing.T) {
	for _, tc := range []struct {
		protocol ir.ProtocolType
		http3    bool
		wantUDP  bool
	}{
		{protocol: ir.TLSProtocolType},
		{protocol: ir.TLSProtocolType, http3: true, wantUDP: true},
		{protocol: ir.HTTPSProtocolType},
		{protocol: ir.HTTPSProtocolType, http3: true, wantUDP: true},
		{protocol: ir.HTTPProtocolType, http3: true},
		{protocol: ir.TCPProtocolType, http3: true},
		{protocol: ir.UDPProtocolType, http3: true},
	} {
		t.Run(fmt.Sprintf("%s/http3=%t", tc.protocol, tc.http3), func(t *testing.T) {
			infra := newTestInfra()
			listener := &ir.ProxyListener{
				Name:  "shared",
				Ports: []ir.ListenerPort{{Name: "shared-443", Protocol: tc.protocol, ServicePort: 443, ContainerPort: 8443}},
			}
			if tc.http3 {
				listener.HTTP3 = &ir.HTTP3Settings{}
			}
			infra.Proxy.Listeners = []*ir.ProxyListener{listener}
			svc := renderHTTP3Service(t, infra)
			protocol := corev1.ProtocolTCP
			if tc.protocol == ir.UDPProtocolType {
				protocol = corev1.ProtocolUDP
			}
			want := []corev1.ServicePort{{Name: "shared-443", Protocol: protocol, Port: 443, TargetPort: intstr.FromInt32(8443)}}
			if tc.wantUDP {
				want = append(want, corev1.ServicePort{Name: "shared-443-h3", Protocol: corev1.ProtocolUDP, Port: 443, TargetPort: intstr.FromInt32(8443)})
			}
			require.Equal(t, want, svc.Spec.Ports)
		})
	}
}

func TestServiceHTTP3SharedTLSPort(t *testing.T) {
	for _, terminate := range []bool{false, true} {
		for _, tlsFirst := range []bool{true, false} {
			for _, http3 := range []bool{true, false} {
				t.Run(fmt.Sprintf("terminate=%t/TLS-first=%t/http3=%t", terminate, tlsFirst, http3), func(t *testing.T) {
					load := func(name string) *gwapiresource.Resources {
						t.Helper()
						data, err := os.ReadFile(filepath.Join("..", "..", "..", "gatewayapi", "testdata", name))
						require.NoError(t, err)
						resources := &gwapiresource.Resources{}
						require.NoError(t, yaml.UnmarshalStrict(data, resources))
						return resources
					}
					resources := load("clienttrafficpolicy-http3.in.yaml")
					resources.Secrets[0].Data["ca.crt"] = resources.Secrets[0].Data["tls.crt"]
					resources.Secrets = append(resources.Secrets, load(filepath.Join("base", "base.yaml")).Secrets...)
					resources.HTTPRoutes = nil
					resources.Namespaces = []*corev1.Namespace{
						{ObjectMeta: metav1.ObjectMeta{Name: "envoy-gateway"}},
						{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
					}
					gw := resources.Gateways[0]
					https := gw.Spec.Listeners[1]
					tlsListener := gwapiv1.Listener{
						Name: "tls-passthrough", Protocol: gwapiv1.TLSProtocolType, Port: 443,
						Hostname: new(gwapiv1.Hostname("passthrough.example.net")),
						TLS:      &gwapiv1.ListenerTLSConfig{Mode: new(gwapiv1.TLSModePassthrough)},
					}
					if terminate {
						tlsListener.Name = "tls-protected"
						tlsListener.Hostname = new(gwapiv1.Hostname("foo.example.com"))
						tlsListener.TLS = https.TLS.DeepCopy()
						https.Hostname = new(gwapiv1.Hostname("*.example.com"))
					}
					gw.Spec.Listeners = []gwapiv1.Listener{https, tlsListener}
					wantProtocol := ir.HTTPSProtocolType
					if tlsFirst {
						gw.Spec.Listeners = []gwapiv1.Listener{tlsListener, https}
						wantProtocol = ir.TLSProtocolType
					}
					policy := resources.ClientTrafficPolicies[0]
					if terminate {
						policy.Spec.TargetRef.SectionName = new(https.Name)
						validationPolicy := policy.DeepCopy()
						validationPolicy.Name = "tls-client-validation"
						validationPolicy.Spec.TargetRef.SectionName = new(tlsListener.Name)
						validationPolicy.Spec.HTTP3 = nil
						validationPolicy.Spec.TLS = &egv1a1.ClientTLSSettings{ClientValidation: &egv1a1.ClientValidationContext{
							CACertificateRefs: []gwapiv1.SecretObjectReference{{Name: "tls-secret-1"}},
						}}
						resources.ClientTrafficPolicies = append(resources.ClientTrafficPolicies, validationPolicy)
					}
					if !http3 {
						policy.Spec.HTTP3 = nil
					}
					translator := &gatewayapi.Translator{
						GatewayControllerName: egv1a1.GatewayControllerName, GatewayClassName: gatewayClassName,
						ControllerNamespace: "envoy-gateway-system", Logger: logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo),
					}
					got, err := translator.Translate(t.Context(), resources)
					require.NoError(t, err)
					require.Len(t, got.XdsIR, 1)
					for _, xds := range got.XdsIR {
						require.Len(t, xds.HTTP, 1)
						require.Equal(t, http3, xds.HTTP[0].HTTP3 != nil)
						if terminate {
							require.Len(t, xds.TCP, 1)
							require.True(t, xds.TCP[0].TLS.ClientValidationEnabled)
							require.True(t, xds.TCP[0].TLS.RequireClientCertificate)
						}
					}
					require.Len(t, got.InfraIR, 1)
					for _, infra := range got.InfraIR {
						require.Len(t, infra.Proxy.Listeners, 1)
						listener := infra.Proxy.Listeners[0]
						require.Len(t, listener.Ports, 1)
						require.Equal(t, wantProtocol, listener.Ports[0].Protocol)
						require.Equal(t, http3, listener.HTTP3 != nil)
						port := listener.Ports[0]
						want := []corev1.ServicePort{{Name: port.Name, Protocol: corev1.ProtocolTCP, Port: 443, TargetPort: intstr.FromInt32(port.ContainerPort)}}
						if http3 {
							want = append(want, corev1.ServicePort{Name: port.Name + "-h3", Protocol: corev1.ProtocolUDP, Port: 443, TargetPort: intstr.FromInt32(port.ContainerPort)})
						}
						require.Equal(t, want, renderHTTP3Service(t, infra).Spec.Ports)
					}
				})
			}
		}
	}
}

func renderHTTP3Service(t *testing.T, infra *ir.Infra) *corev1.Service {
	t.Helper()
	cfg, err := config.New(os.Stdout, os.Stderr)
	require.NoError(t, err)
	render, err := NewResourceRender(t.Context(), newFakeKubernetesInfraProvider(cfg), infra)
	require.NoError(t, err)
	svc, err := render.Service()
	require.NoError(t, err)
	return svc
}
