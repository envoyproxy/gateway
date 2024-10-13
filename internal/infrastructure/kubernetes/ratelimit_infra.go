// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
)

// CreateOrUpdateRateLimitInfra creates the managed kube rate limit infra, if it doesn't exist.
func (i *Infra) CreateOrUpdateRateLimitInfra(ctx context.Context) error {
	if err := ratelimit.Validate(ctx, i.Client.Client, i.EnvoyGateway, i.Namespace); err != nil {
		return err
	}

	// Create ratelimit infra requires the uid of owner reference.
	ownerReferenceUID := make(map[string]types.UID)
	key := types.NamespacedName{
		Namespace: i.Namespace,
		Name:      "envoy-gateway",
	}

	serviceUID, err := i.Client.GetUID(ctx, key, &corev1.Service{})
	if err != nil {
		return err
	}
	ownerReferenceUID[ratelimit.ResourceKindService] = serviceUID

	var uid types.UID
	for _, obj := range []client.Object{&appsv1.Deployment{}, &appsv1.DaemonSet{}} {
		uid, err = i.Client.GetUID(ctx, key, obj)
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			return err
		}
		switch obj.(type) {
		case *appsv1.Deployment:
			ownerReferenceUID[ratelimit.ResourceKindDeployment] = uid
		case *appsv1.DaemonSet:
			ownerReferenceUID[ratelimit.ResourceKindDaemonset] = uid
		}
		break
	}
	if err != nil {
		return err
	}

	serviceAccountUID, err := i.Client.GetUID(ctx, key, &corev1.ServiceAccount{})
	if err != nil {
		return err
	}
	ownerReferenceUID[ratelimit.ResourceKindServiceAccount] = serviceAccountUID

	r := ratelimit.NewResourceRender(i.Namespace, i.EnvoyGateway, ownerReferenceUID)
	return i.createOrUpdate(ctx, r)
}

// DeleteRateLimitInfra removes the managed kube infra, if it doesn't exist.
func (i *Infra) DeleteRateLimitInfra(ctx context.Context) error {
	if err := ratelimit.Validate(ctx, i.Client.Client, i.EnvoyGateway, i.Namespace); err != nil {
		return err
	}

	// Delete ratelimit infra do not require the uid of owner reference.
	r := ratelimit.NewResourceRender(i.Namespace, i.EnvoyGateway, nil)
	return i.delete(ctx, r)
}
