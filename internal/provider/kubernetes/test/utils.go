// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package test

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/utils/ptr"
)

type ObjectKindNamespacedName struct {
	Kind      string
	Namespace string
	Name      string
}

// NewEnvoyProxy returns an EnvoyProxy object with the provided ns/name.
func NewEnvoyProxy(ns, name string) *egv1a1.EnvoyProxy {
	return &egv1a1.EnvoyProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
	}
}

// GetGatewayClass returns a sample GatewayClass.
func GetGatewayClass(name string, controller gwapiv1.GatewayController) *gwapiv1.GatewayClass {
	return &gwapiv1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: gwapiv1.GatewayClassSpec{
			ControllerName: controller,
		},
	}
}

// GetGateway returns a sample Gateway with single listener.
func GetGateway(nsname types.NamespacedName, gwclass string) *gwapiv1.Gateway {
	return &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsname.Namespace,
			Name:      nsname.Name,
		},
		Spec: gwapiv1.GatewaySpec{
			GatewayClassName: gwapiv1.ObjectName(gwclass),
			Listeners: []gwapiv1.Listener{
				{
					Name:     "test",
					Port:     gwapiv1.PortNumber(int32(8080)),
					Protocol: gwapiv1.HTTPProtocolType,
				},
			},
		},
	}
}

// GetSecureGateway returns a sample Gateway with single TLS listener.
func GetSecureGateway(nsname types.NamespacedName, gwclass string, secretKindNSName ObjectKindNamespacedName) *gwapiv1.Gateway {
	secureGateway := GetGateway(nsname, gwclass)
	secureGateway.Spec.Listeners[0].TLS = &gwapiv1.GatewayTLSConfig{
		Mode: ptr.To(gwapiv1.TLSModeTerminate),
		CertificateRefs: []gwapiv1.SecretObjectReference{{
			Kind:      (*gwapiv1.Kind)(&secretKindNSName.Kind),
			Namespace: (*gwapiv1.Namespace)(&secretKindNSName.Namespace),
			Name:      gwapiv1.ObjectName(secretKindNSName.Name),
		}},
	}

	return secureGateway
}

// GetSecret returns a sample Secret object.
func GetSecret(nsname types.NamespacedName) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsname.Namespace,
			Name:      nsname.Name,
		},
	}
}

// GetHTTPRoute returns a sample HTTPRoute with a parent reference.
func GetHTTPRoute(nsName types.NamespacedName, parent string, serviceName types.NamespacedName) *gwapiv1.HTTPRoute {
	return &gwapiv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsName.Namespace,
			Name:      nsName.Name,
		},
		Spec: gwapiv1.HTTPRouteSpec{
			CommonRouteSpec: gwapiv1.CommonRouteSpec{
				ParentRefs: []gwapiv1.ParentReference{
					{Name: gwapiv1.ObjectName(parent)},
				},
			},
			Rules: []gwapiv1.HTTPRouteRule{
				{
					BackendRefs: []gwapiv1.HTTPBackendRef{
						{
							BackendRef: gwapiv1.BackendRef{
								BackendObjectReference: gwapiv1.BackendObjectReference{
									Name: gwapiv1.ObjectName(serviceName.Name),
								},
							},
						},
					},
				},
			},
		},
	}
}

// GetGRPCRoute returns a sample GRPCRoute with a parent reference.
func GetGRPCRoute(nsName types.NamespacedName, parent string, serviceName types.NamespacedName) *gwapiv1a2.GRPCRoute {
	return &gwapiv1a2.GRPCRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsName.Namespace,
			Name:      nsName.Name,
		},
		Spec: gwapiv1a2.GRPCRouteSpec{
			CommonRouteSpec: gwapiv1.CommonRouteSpec{
				ParentRefs: []gwapiv1.ParentReference{
					{Name: gwapiv1.ObjectName(parent)},
				},
			},
			Rules: []gwapiv1a2.GRPCRouteRule{
				{
					BackendRefs: []gwapiv1a2.GRPCBackendRef{
						{
							BackendRef: gwapiv1.BackendRef{
								BackendObjectReference: gwapiv1.BackendObjectReference{
									Name: gwapiv1.ObjectName(serviceName.Name),
								},
							},
						},
					},
				},
			},
		},
	}
}

// GetTLSRoute returns a sample TLSRoute with a parent reference.
func GetTLSRoute(nsName types.NamespacedName, parent string, serviceName types.NamespacedName) *gwapiv1a2.TLSRoute {
	return &gwapiv1a2.TLSRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsName.Namespace,
			Name:      nsName.Name,
		},
		Spec: gwapiv1a2.TLSRouteSpec{
			CommonRouteSpec: gwapiv1a2.CommonRouteSpec{
				ParentRefs: []gwapiv1a2.ParentReference{
					{Name: gwapiv1a2.ObjectName(parent)},
				},
			},
			Rules: []gwapiv1a2.TLSRouteRule{
				{
					BackendRefs: []gwapiv1a2.BackendRef{
						{
							BackendObjectReference: gwapiv1a2.BackendObjectReference{
								Name: gwapiv1a2.ObjectName(serviceName.Name),
							},
						},
					},
				},
			},
		},
	}
}

// GetTCPRoute returns a sample TCPRoute with a parent reference.
func GetTCPRoute(nsName types.NamespacedName, parent string, serviceName types.NamespacedName) *gwapiv1a2.TCPRoute {
	return &gwapiv1a2.TCPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsName.Namespace,
			Name:      nsName.Name,
		},
		Spec: gwapiv1a2.TCPRouteSpec{
			CommonRouteSpec: gwapiv1a2.CommonRouteSpec{
				ParentRefs: []gwapiv1a2.ParentReference{
					{Name: gwapiv1a2.ObjectName(parent)},
				},
			},
			Rules: []gwapiv1a2.TCPRouteRule{
				{
					BackendRefs: []gwapiv1a2.BackendRef{
						{
							BackendObjectReference: gwapiv1a2.BackendObjectReference{
								Name: gwapiv1a2.ObjectName(serviceName.Name),
							},
						},
					},
				},
			},
		},
	}
}

// GetUDPRoute returns a sample UDPRoute with a parent reference.
func GetUDPRoute(nsName types.NamespacedName, parent string, serviceName types.NamespacedName) *gwapiv1a2.UDPRoute {
	return &gwapiv1a2.UDPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsName.Namespace,
			Name:      nsName.Name,
		},
		Spec: gwapiv1a2.UDPRouteSpec{
			CommonRouteSpec: gwapiv1a2.CommonRouteSpec{
				ParentRefs: []gwapiv1a2.ParentReference{
					{Name: gwapiv1a2.ObjectName(parent)},
				},
			},
			Rules: []gwapiv1a2.UDPRouteRule{
				{
					BackendRefs: []gwapiv1a2.BackendRef{
						{
							BackendObjectReference: gwapiv1a2.BackendObjectReference{
								Name: gwapiv1a2.ObjectName(serviceName.Name),
							},
						},
					},
				},
			},
		},
	}
}

// GetGatewayDeployment returns a sample Deployment for a Gateway object.
func GetGatewayDeployment(nsname types.NamespacedName, labels map[string]string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsname.Namespace,
			Name:      nsname.Name,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "dummy",
						Image: "dummy",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
						}},
					}},
				},
			},
		},
	}
}

// GetService returns a sample Service with labels and ports.
func GetService(nsname types.NamespacedName, labels map[string]string, ports map[string]int32) *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nsname.Name,
			Namespace: nsname.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{},
		},
	}
	for name, port := range ports {
		service.Spec.Ports = append(service.Spec.Ports, corev1.ServicePort{
			Name: name,
			Port: port,
		})
	}
	return service
}

// GetEndpointSlice returns a sample EndpointSlice.
func GetEndpointSlice(nsName types.NamespacedName, svcName string) *discoveryv1.EndpointSlice {
	return &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nsName.Name,
			Namespace: nsName.Namespace,
			Labels:    map[string]string{discoveryv1.LabelServiceName: svcName},
		},
		Endpoints: []discoveryv1.Endpoint{
			{
				Addresses: []string{"10.0.0.1"},
				Conditions: discoveryv1.EndpointConditions{
					Ready: &[]bool{true}[0],
				},
			},
		},
		Ports: []discoveryv1.EndpointPort{
			{
				Name:     &[]string{"dummy"}[0],
				Port:     &[]int32{8080}[0],
				Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
			},
		},
	}
}

// GetAuthenticationFilter returns a pointer to an AuthenticationFilter with the
// provided ns/name. The AuthenticationFilter uses a JWT provider with dummy issuer,
// audiences, and remoteJWKS settings.
func GetAuthenticationFilter(name, ns string) *egv1a1.AuthenticationFilter {
	provider := GetAuthenticationProvider("test")
	return &egv1a1.AuthenticationFilter{
		TypeMeta: metav1.TypeMeta{
			Kind:       egv1a1.KindAuthenticationFilter,
			APIVersion: egv1a1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: egv1a1.AuthenticationFilterSpec{
			Type:         egv1a1.JwtAuthenticationFilterProviderType,
			JwtProviders: []egv1a1.JwtAuthenticationFilterProvider{provider},
		},
	}
}

// GetAuthenticationProvider returns a JwtAuthenticationFilterProvider using the provided name.
func GetAuthenticationProvider(name string) egv1a1.JwtAuthenticationFilterProvider {
	return egv1a1.JwtAuthenticationFilterProvider{
		Name:      name,
		Issuer:    "https://www.test.local",
		Audiences: []string{"test.local"},
		RemoteJWKS: egv1a1.RemoteJWKS{
			URI: "https://test.local/jwt/public-key/jwks.json",
		},
	}
}

// GetRateLimitFilter returns a pointer to an RateLimitFilter with dummy rules.
func GetRateLimitFilter(name, ns string) *egv1a1.RateLimitFilter {
	rule := GetRateLimitGlobalRule("one")
	return &egv1a1.RateLimitFilter{
		TypeMeta: metav1.TypeMeta{
			Kind:       egv1a1.KindRateLimitFilter,
			APIVersion: egv1a1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: egv1a1.RateLimitFilterSpec{
			Type: egv1a1.GlobalRateLimitType,
			Global: &egv1a1.GlobalRateLimit{
				Rules: []egv1a1.RateLimitRule{rule},
			},
		},
	}
}

// GetRateLimitGlobalRule returns a RateLimitRule using the val as the ClientSelectors
// headers value.
func GetRateLimitGlobalRule(val string) egv1a1.RateLimitRule {
	return egv1a1.RateLimitRule{
		ClientSelectors: []egv1a1.RateLimitSelectCondition{
			{
				Headers: []egv1a1.HeaderMatch{
					{
						Name:  "x-user-id",
						Value: ptr.To(val),
					},
				},
			},
		},
		Limit: egv1a1.RateLimitValue{
			Requests: 5,
			Unit:     "Second",
		},
	}
}

func ContainsAuthenFilter(hroute *gwapiv1.HTTPRoute) bool {
	if hroute == nil {
		return false
	}

	for _, rule := range hroute.Spec.Rules {
		for _, filter := range rule.Filters {
			if filter.Type == gwapiv1.HTTPRouteFilterExtensionRef &&
				filter.ExtensionRef != nil &&
				string(filter.ExtensionRef.Group) == egv1a1.GroupVersion.Group &&
				filter.ExtensionRef.Kind == egv1a1.KindAuthenticationFilter {
				return true
			}
		}
	}

	return false
}
