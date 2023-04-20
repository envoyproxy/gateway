// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package applier

import (
	"context"
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Instance struct {
	Client client.Client
}

func New(cli client.Client) *Instance {
	return &Instance{
		Client: cli,
	}
}

func (i *Instance) CreateOrUpdateConfigMap(ctx context.Context, cm *corev1.ConfigMap) error {
	current := &corev1.ConfigMap{}
	key := types.NamespacedName{
		Namespace: cm.Namespace,
		Name:      cm.Name,
	}

	if err := i.Client.Get(ctx, key, current); err != nil {
		// Create if not found.
		if kerrors.IsNotFound(err) {
			if err := i.Client.Create(ctx, cm); err != nil {
				return fmt.Errorf("failed to create configmap %s/%s: %w", cm.Namespace, cm.Name, err)
			}
		}
	} else {
		// Update if current value is different.
		if !reflect.DeepEqual(cm.Data, current.Data) {
			cm.ResourceVersion = current.ResourceVersion
			cm.UID = current.UID
			if err := i.Client.Update(ctx, cm); err != nil {
				return fmt.Errorf("failed to update configmap %s/%s: %w", cm.Namespace, cm.Name, err)
			}
		}
	}

	return nil
}

func (i *Instance) DeleteConfigMap(ctx context.Context, cm *corev1.ConfigMap) error {
	if err := i.Client.Delete(ctx, cm); err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete configmap %s/%s: %w", cm.Namespace, cm.Name, err)
	}

	return nil
}

func (i *Instance) CreateOrUpdateDeployment(ctx context.Context, deployment *appsv1.Deployment) error {
	current := &appsv1.Deployment{}
	key := types.NamespacedName{
		Namespace: deployment.Namespace,
		Name:      deployment.Name,
	}
	if err := i.Client.Get(ctx, key, current); err != nil {
		// Create if not found.
		if kerrors.IsNotFound(err) {
			if err := i.Client.Create(ctx, deployment); err != nil {
				return fmt.Errorf("failed to create deployment %s/%s: %w",
					deployment.Namespace, deployment.Name, err)
			}
		}
	} else {
		// Update if current value is different.
		if !reflect.DeepEqual(deployment.Spec, current.Spec) {
			deployment.ResourceVersion = current.ResourceVersion
			deployment.UID = current.UID
			if err := i.Client.Update(ctx, deployment); err != nil {
				return fmt.Errorf("failed to update deployment %s/%s: %w",
					deployment.Namespace, deployment.Name, err)
			}
		}
	}

	return nil
}

func (i *Instance) DeleteDeployment(ctx context.Context, deploy *appsv1.Deployment) error {
	if err := i.Client.Delete(ctx, deploy); err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete deployment %s/%s: %w", deploy.Namespace, deploy.Name, err)
	}
	return nil
}

func (i *Instance) CreateOrUpdateService(ctx context.Context, svc *corev1.Service) error {
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

func (i *Instance) DeleteService(ctx context.Context, svc *corev1.Service) error {
	if err := i.Client.Delete(ctx, svc); err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete service %s/%s: %w", svc.Namespace, svc.Name, err)
	}

	return nil
}

func (i *Instance) CreateOrUpdateServiceAccount(ctx context.Context, sa *corev1.ServiceAccount) error {
	current := &corev1.ServiceAccount{}
	key := types.NamespacedName{
		Namespace: sa.Namespace,
		Name:      sa.Name,
	}

	if err := i.Client.Get(ctx, key, current); err != nil {
		if kerrors.IsNotFound(err) {
			// Create if it does not exist.
			if err := i.Client.Create(ctx, sa); err != nil {
				return fmt.Errorf("failed to create serviceaccount %s/%s: %w",
					sa.Namespace, sa.Name, err)
			}
		}
	} else {
		// Since the ServiceAccount does not have a specific Spec field to compare
		// just perform an update for now.
		sa.ResourceVersion = current.ResourceVersion
		sa.UID = current.UID
		if err := i.Client.Update(ctx, sa); err != nil {
			return fmt.Errorf("failed to update serviceaccount %s/%s: %w",
				sa.Namespace, sa.Name, err)
		}
	}

	return nil
}

func (i *Instance) DeleteServiceAccount(ctx context.Context, sa *corev1.ServiceAccount) error {
	if err := i.Client.Delete(ctx, sa); err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete serviceaccount %s/%s: %w", sa.Namespace, sa.Name, err)
	}

	return nil
}
