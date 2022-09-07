package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	podSelector := EnvoyPodSelector(infra.GetProxyInfra().Name)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      envoyServiceName,
			Labels:    podSelector.MatchLabels,
		},
		Spec: corev1.ServiceSpec{
			Type:            corev1.ServiceTypeLoadBalancer,
			Ports:           ports,
			Selector:        podSelector.MatchLabels,
			SessionAffinity: corev1.ServiceAffinityNone,
			// Preserve the client source IP and avoid a second hop for LoadBalancer.
			ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeLocal,
		},
	}

	return svc
}

// createOrUpdateService creates a Service in the kube api server based on the provided infra,
// if it doesn't exist or updates it if it does.
func (i *Infra) createOrUpdateService(ctx context.Context, infra *ir.Infra) error {
	svc := i.expectedService(infra)
	err := i.Client.Create(ctx, svc)
	if err != nil {
		if kerrors.IsAlreadyExists(err) {
			// Update service if its exists
			if err := i.Client.Update(ctx, svc); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if err := i.updateResource(svc); err != nil {
		return err
	}

	return nil
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
