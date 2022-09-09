package kubernetes

import (
	"context"
	"fmt"
	"reflect"

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
)

// expectedService returns the expected Service based on the provided infra.
func (i *Infra) expectedService(infra *ir.Infra) *corev1.Service {
	var ports []corev1.ServicePort
	for _, listener := range infra.Proxy.Listeners {
		for _, port := range listener.Ports {
			target := intstr.IntOrString{IntVal: port.ContainerPort}
			p := corev1.ServicePort{
				Name:       port.Name,
				Protocol:   corev1.ProtocolTCP,
				Port:       port.ServicePort,
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

	current := &corev1.Service{}
	key := types.NamespacedName{
		Namespace: i.Namespace,
		Name:      envoyServiceName,
	}

	if err := i.Client.Get(ctx, key, current); err != nil {
		// Create if not found.
		if kerrors.IsNotFound(err) {
			if err := i.Client.Create(ctx, svc); err != nil {
				return fmt.Errorf("failed to create service %s/%s: %w",
					svc.Namespace, svc.Name, err)
			}
		}
	} else {
		// Update if current value is different.
		if !reflect.DeepEqual(svc.Spec, current.Spec) {
			if err := i.Client.Update(ctx, svc); err != nil {
				return fmt.Errorf("failed to update service %s/%s: %w",
					svc.Namespace, svc.Name, err)
			}
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
