// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/resource"
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

	return i.Client.CreateOrUpdate(ctx, key, current, sa, func() bool {
		return true
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

// createOrUpdateSet creates a Deployment or DaemonSet in the kube api server based on the provided
// ResourceRender, if it doesn't exist and updates it if it does. They are mutually exclusive, so
// if a DaemonSet is provided, the corresponding Deployment is deleted, and vice versa.
func (i *Infra) createOrUpdatePodSet(ctx context.Context, r ResourceRender) error {
	dset, err := r.DaemonSet()
	if err != nil {
		return err
	}
	deployment, err := r.Deployment()
	if err != nil {
		return err
	}

	// Delete the one that is nil.
	var toDelete client.Object
	if dset == nil {
		toDelete = &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: i.Namespace,
				Name:      r.Name(),
			},
		}
	} else {
		toDelete = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: i.Namespace,
				Name:      r.Name(),
			},
		}
	}
	if err := i.Client.Delete(ctx, toDelete); err != nil {
		return err
	}

	key := types.NamespacedName{
		Namespace: i.Namespace,
		Name:      r.Name(),
	}

	// Create or update the one that is non-nil.
	if dset != nil {
		current := &appsv1.DaemonSet{}
		return i.Client.CreateOrUpdate(ctx, key, current, dset, func() bool {
			return !reflect.DeepEqual(dset.Spec, current.Spec)
		})
	} else {
		current := &appsv1.Deployment{}
		return i.Client.CreateOrUpdate(ctx, key, current, deployment, func() bool {
			return !reflect.DeepEqual(deployment.Spec, current.Spec)
		})
	}
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

// deleteDaemonSet deletes the Envoy DaemonSet in the kube api server, if it exists.
func (i *Infra) deleteDaemonSet(ctx context.Context, r ResourceRender) error {
	dset := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      r.Name(),
		},
	}

	return i.Client.Delete(ctx, dset)
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
