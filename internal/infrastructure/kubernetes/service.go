package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/envoyproxy/gateway/internal/ir"
)

const (
	// envoyServiceName is the name of the Envoy Service resource.
	envoyServiceName = "envoy"
	// envoyServiceHTTPPort is the HTTP port number of the Envoy service.
	envoyServiceHTTPPort = 80
	// envoyServiceHTTPSPort is the HTTPS port number of the Envoy service.
	envoyServiceHTTPSPort = 443
)

// createServiceIfNeeded creates a Service based on the provided infra, if
// it doesn't exist in the kube api server.
func (i *Infra) createServiceIfNeeded(ctx context.Context, infra *ir.Infra) error {
	current, err := i.getService(ctx)
	if err != nil {
		if kerrors.IsNotFound(err) {
			svc, err := i.createService(ctx, infra)
			if err != nil {
				return err
			}
			if err := i.addResource(svc); err != nil {
				return err
			}
			return nil
		}
		return err
	}

	if err := i.addResource(current); err != nil {
		return err
	}

	return nil
}

// getService gets the Service from the kube api for the provided infra.
func (i *Infra) getService(ctx context.Context) (*corev1.Service, error) {
	key := types.NamespacedName{
		Namespace: i.Namespace,
		Name:      envoyServiceName,
	}
	svc := new(corev1.Service)
	if err := i.Client.Get(ctx, key, svc); err != nil {
		return nil, fmt.Errorf("failed to get service %s/%s: %w", i.Namespace, envoyServiceName, err)
	}

	return svc, nil
}

// expectedService returns the expected Service based on the provided infra.
func (i *Infra) expectedService(infra *ir.Infra) *corev1.Service {
	var ports []corev1.ServicePort
	for _, listener := range infra.Proxy.Listeners {
		for _, port := range listener.Ports {
			// Set the target port based on the protocol of the IR port.
			target := intstr.IntOrString{IntVal: envoyHTTPPort}
			if port.Protocol == ir.HTTPSProtocolType {
				target = intstr.IntOrString{IntVal: envoyHTTPSPort}
			}
			p := corev1.ServicePort{
				Name:       port.Name,
				Protocol:   corev1.ProtocolTCP,
				Port:       port.Port,
				TargetPort: target,
			}
			ports = append(ports, p)
		}
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{

			Namespace: i.Namespace,
			Name:      envoyServiceName,
			Labels:    envoyLabels(),
		},
		Spec: corev1.ServiceSpec{
			Type:            corev1.ServiceTypeLoadBalancer,
			Ports:           ports,
			Selector:        EnvoyPodSelector().MatchLabels,
			SessionAffinity: corev1.ServiceAffinityNone,
			// Preserve the client source IP and avoid a second hop for LoadBalancer.
			ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeLocal,
		},
	}

	return svc
}

// createService creates a Service in the kube api server based on the provided infra,
// if it doesn't exist.
func (i *Infra) createService(ctx context.Context, infra *ir.Infra) (*corev1.Service, error) {
	expected := i.expectedService(infra)
	err := i.Client.Create(ctx, expected)
	if err != nil {
		if kerrors.IsAlreadyExists(err) {
			return expected, nil
		}
		return nil, fmt.Errorf("failed to create service %s/%s: %w",
			expected.Namespace, expected.Name, err)
	}

	return expected, nil
}

// deleteService deletes the Envoy Service in the kube api server, if it exists.
func (i *Infra) deleteService(ctx context.Context) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      envoyServiceName,
		},
	}

	if err := i.Client.Delete(ctx, svc); err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete service %s/%s: %w", svc.Namespace, svc.Name, err)
	}

	return nil
}
