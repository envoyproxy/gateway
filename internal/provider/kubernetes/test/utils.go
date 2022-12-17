// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package test

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"k8s.io/apimachinery/pkg/types"
)

type ObjectKindNamespacedName struct {
	Kind      string
	Namespace string
	Name      string
}

// GetGatewayClass returns a sample GatewayClass.
func GetGatewayClass(name string, controller gwapiv1b1.GatewayController) *gwapiv1b1.GatewayClass {
	return &gwapiv1b1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: gwapiv1b1.GatewayClassSpec{
			ControllerName: controller,
		},
	}
}

// GetGateway returns a sample Gateway with single listener.
func GetGateway(nsname types.NamespacedName, gwclass string) *gwapiv1b1.Gateway {
	return &gwapiv1b1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsname.Namespace,
			Name:      nsname.Name,
		},
		Spec: gwapiv1b1.GatewaySpec{
			GatewayClassName: gwapiv1b1.ObjectName(gwclass),
			Listeners: []gwapiv1b1.Listener{
				{
					Name:     "test",
					Port:     gwapiv1b1.PortNumber(int32(8080)),
					Protocol: gwapiv1b1.HTTPProtocolType,
				},
			},
		},
	}
}

// GetSecureGateway returns a sample Gateway with single TLS listener.
func GetSecureGateway(nsname types.NamespacedName, gwclass string, secretKindNSName ObjectKindNamespacedName) *gwapiv1b1.Gateway {
	secureGateway := GetGateway(nsname, gwclass)
	secureGateway.Spec.Listeners[0].TLS = &gwapiv1b1.GatewayTLSConfig{
		Mode: gatewayapi.TLSModeTypePtr(gwapiv1b1.TLSModeTerminate),
		CertificateRefs: []gwapiv1b1.SecretObjectReference{{
			Kind:      (*gwapiv1b1.Kind)(&secretKindNSName.Kind),
			Namespace: (*gwapiv1b1.Namespace)(&secretKindNSName.Namespace),
			Name:      gwapiv1b1.ObjectName(secretKindNSName.Name),
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
func GetHTTPRoute(nsName types.NamespacedName, parent string, serviceName types.NamespacedName) *gwapiv1b1.HTTPRoute {
	return &gwapiv1b1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsName.Namespace,
			Name:      nsName.Name,
		},
		Spec: gwapiv1b1.HTTPRouteSpec{
			CommonRouteSpec: gwapiv1b1.CommonRouteSpec{
				ParentRefs: []gwapiv1b1.ParentReference{
					{Name: gwapiv1b1.ObjectName(parent)},
				},
			},
			Rules: []gwapiv1b1.HTTPRouteRule{
				{
					BackendRefs: []gwapiv1b1.HTTPBackendRef{
						{
							BackendRef: gwapiv1b1.BackendRef{
								BackendObjectReference: gwapiv1b1.BackendObjectReference{
									Name: gwapiv1b1.ObjectName(serviceName.Name),
								},
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
