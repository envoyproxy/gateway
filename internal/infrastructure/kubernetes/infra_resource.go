// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// createOrUpdateServiceAccount creates a ServiceAccount in the kube api server based on the
// provided ResourceRender, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateServiceAccount(ctx context.Context, r ResourceRender) error {
	sa, err := r.ServiceAccount()
	if err != nil {
		return err
	}

	current := &corev1.ServiceAccount{}
	key := types.NamespacedName{
		Namespace: sa.Namespace,
		Name:      sa.Name,
	}

	return i.Client.Create(ctx, key, current, sa)
}

// createOrUpdateConfigMap creates a ConfigMap in the Kube api server based on the provided
// ResourceRender, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateConfigMap(ctx context.Context, r ResourceRender) error {
	cm, err := r.ConfigMap()
	if err != nil {
		return err
	}

	current := &corev1.ConfigMap{}
	key := types.NamespacedName{
		Namespace: cm.Namespace,
		Name:      cm.Name,
	}

	return i.Client.Create(ctx, key, current, cm)
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

	return i.Client.Create(ctx, key, current, deployment)
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

	return i.Client.Create(ctx, key, current, svc)
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
