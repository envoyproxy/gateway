// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (i *Infra) createOrUpdateDeployment(ctx context.Context, deployment *appsv1.Deployment) error {
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

func (i *Infra) deleteDeployment(ctx context.Context, deploy *appsv1.Deployment) error {
	if err := i.Client.Delete(ctx, deploy); err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete deployment %s/%s: %w", deploy.Namespace, deploy.Name, err)
	}
	return nil
}
