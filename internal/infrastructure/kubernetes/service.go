// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
)

func (i *Infra) createOrUpdateService(ctx context.Context, svc *corev1.Service) error {
	current := &corev1.Service{}
	key := types.NamespacedName{
		Namespace: svc.Namespace,
		Name:      svc.Name,
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
			svc.ResourceVersion = current.ResourceVersion
			svc.UID = current.UID
			if err := i.Client.Update(ctx, svc); err != nil {
				return fmt.Errorf("failed to update service %s/%s: %w",
					svc.Namespace, svc.Name, err)
			}
		}
	}

	return nil
}

func (i *Infra) deleteService(ctx context.Context, svc *corev1.Service) error {
	if err := i.Client.Delete(ctx, svc); err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete service %s/%s: %w", svc.Namespace, svc.Name, err)
	}

	return nil
}

func expectedServiceSpec(serviceType *egcfgv1a1.ServiceType) corev1.ServiceSpec {
	serviceSpec := corev1.ServiceSpec{}
	serviceSpec.Type = corev1.ServiceType(*serviceType)
	serviceSpec.SessionAffinity = corev1.ServiceAffinityNone
	if *serviceType == egcfgv1a1.ServiceTypeLoadBalancer {
		// Preserve the client source IP and avoid a second hop for LoadBalancer.
		serviceSpec.ExternalTrafficPolicy = corev1.ServiceExternalTrafficPolicyTypeLocal
	}
	return serviceSpec
}
