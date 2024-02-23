// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"reflect"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/resource"
	"github.com/envoyproxy/gateway/internal/utils"
)

// createOrUpdateServiceAccount creates a ServiceAccount in the kube api server based on the
// provided ResourceRender, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateServiceAccount(ctx context.Context, r ResourceRender) error {
	sa, err := r.ServiceAccount()
	if err != nil {
		return err
	}

	current := &corev1.ServiceAccount{}
	key := utils.NamespacedName(sa)

	return i.Client.CreateOrUpdate(ctx, key, current, sa, func() bool {
		// the service account never changed, does not need to update
		// fixes https://github.com/envoyproxy/gateway/issues/1604
		return false
	})
}

// createOrUpdateConfigMap creates a ConfigMap in the Kube api server based on the provided
// ResourceRender, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateConfigMap(ctx context.Context, r ResourceRender) error {
	cm, err := r.ConfigMap()
	if err != nil {
		return err
	}

	if cm == nil {
		return nil
	}
	current := &corev1.ConfigMap{}
	key := types.NamespacedName{
		Namespace: cm.Namespace,
		Name:      cm.Name,
	}

	return i.Client.CreateOrUpdate(ctx, key, current, cm, func() bool {
		return !reflect.DeepEqual(cm.Data, current.Data)
	})
}

// createOrUpdateDeployment creates a Deployment in the kube api server based on the provided
// ResourceRender, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateDeployment(ctx context.Context, r ResourceRender) error {
	deployment, err := r.Deployment()
	if err != nil {
		return err
	}

	current := &appsv1.Deployment{}
	key := types.NamespacedName{
		Namespace: deployment.Namespace,
		Name:      deployment.Name,
	}

	hpa, err := r.HorizontalPodAutoscaler()
	if err != nil {
		return err
	}

	var opts cmp.Options
	if hpa != nil {
		opts = append(opts, cmpopts.IgnoreFields(appsv1.DeploymentSpec{}, "Replicas"))
	}

	return i.Client.CreateOrUpdate(ctx, key, current, deployment, func() bool {
		return !cmp.Equal(current.Spec, deployment.Spec, opts...)
	})
}

// createOrUpdateHPA creates HorizontalPodAutoscaler object in the kube api server based on
// the provided ResourceRender, if it doesn't exist and updates it if it does,
// and delete hpa if not set.
func (i *Infra) createOrUpdateHPA(ctx context.Context, r ResourceRender) error {
	hpa, err := r.HorizontalPodAutoscaler()
	if err != nil {
		return err
	}

	// when HorizontalPodAutoscaler is not set,
	// then delete the object in the kube api server if any.
	if hpa == nil {
		return i.deleteHPA(ctx, r)
	}

	current := &autoscalingv2.HorizontalPodAutoscaler{}
	key := types.NamespacedName{
		Namespace: hpa.Namespace,
		Name:      hpa.Name,
	}

	return i.Client.CreateOrUpdate(ctx, key, current, hpa, func() bool {
		return !cmp.Equal(hpa.Spec, current.Spec)
	})
}

// createOrUpdateRateLimitService creates a Service in the kube api server based on the provided ResourceRender,
// if it doesn't exist or updates it if it does.
func (i *Infra) createOrUpdateService(ctx context.Context, r ResourceRender) error {
	svc, err := r.Service()
	if err != nil {
		return err
	}

	current := &corev1.Service{}
	key := types.NamespacedName{
		Namespace: svc.Namespace,
		Name:      svc.Name,
	}

	return i.Client.CreateOrUpdate(ctx, key, current, svc, func() bool {
		return !resource.CompareSvc(svc, current)
	})
}

// deleteServiceAccount deletes the ServiceAccount in the kube api server, if it exists.
func (i *Infra) deleteServiceAccount(ctx context.Context, r ResourceRender) error {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      r.Name(),
		},
	}

	return i.Client.Delete(ctx, sa)
}

// deleteDeployment deletes the Envoy Deployment in the kube api server, if it exists.
func (i *Infra) deleteDeployment(ctx context.Context, r ResourceRender) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      r.Name(),
		},
	}

	return i.Client.Delete(ctx, deployment)
}

// deleteConfigMap deletes the ConfigMap in the kube api server, if it exists.
func (i *Infra) deleteConfigMap(ctx context.Context, r ResourceRender) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      r.Name(),
		},
	}

	return i.Client.Delete(ctx, cm)
}

// deleteService deletes the Service in the kube api server, if it exists.
func (i *Infra) deleteService(ctx context.Context, r ResourceRender) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      r.Name(),
		},
	}

	return i.Client.Delete(ctx, svc)
}

// deleteHpa deletes the Horizontal Pod Autoscaler associated to its renderer, if it exists.
func (i *Infra) deleteHPA(ctx context.Context, r ResourceRender) error {
	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      r.Name(),
		},
	}

	return i.Client.Delete(ctx, hpa)
}
