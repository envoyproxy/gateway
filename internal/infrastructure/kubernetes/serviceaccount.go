// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/provider/utils"
)

func expectedProxyServiceAccountName(proxyName string) string {
	svcActName := utils.GetHashedName(proxyName)
	return fmt.Sprintf("%s-%s", config.EnvoyPrefix, svcActName)
}

// expectedProxyServiceAccount returns the expected proxy serviceAccount.
func (i *Infra) expectedProxyServiceAccount(infra *ir.Infra) (*corev1.ServiceAccount, error) {
	// Set the labels based on the owning gateway name.
	labels := envoyLabels(infra.GetProxyInfra().GetProxyMetadata().Labels)
	if len(labels[gatewayapi.OwningGatewayNamespaceLabel]) == 0 || len(labels[gatewayapi.OwningGatewayNameLabel]) == 0 {
		return nil, fmt.Errorf("missing owning gateway labels")
	}

	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      expectedProxyServiceAccountName(infra.Proxy.Name),
			Labels:    labels,
		},
	}, nil
}

// createOrUpdateProxyServiceAccount creates the Envoy ServiceAccount in the kube api server,
// if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateProxyServiceAccount(ctx context.Context, infra *ir.Infra) error {
	sa, err := i.expectedProxyServiceAccount(infra)
	if err != nil {
		return err
	}

	return i.createOrUpdateServiceAccount(ctx, sa)
}

// deleteProxyServiceAccount deletes the Envoy ServiceAccount in the kube api server,
// if it exists.
func (i *Infra) deleteProxyServiceAccount(ctx context.Context, infra *ir.Infra) error {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      expectedProxyServiceAccountName(infra.Proxy.Name),
		},
	}

	return i.deleteServiceAccount(ctx, sa)
}

// expectedRateLimitServiceAccount returns the expected ratelimit serviceAccount.
func (i *Infra) expectedRateLimitServiceAccount(infra *ir.RateLimitInfra) (*corev1.ServiceAccount, error) {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      rateLimitInfraName,
		},
	}, nil
}

// createOrUpdateRateLimitServiceAccount creates the Envoy RateLimit ServiceAccount in the kube api server,
// if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateRateLimitServiceAccount(ctx context.Context, infra *ir.RateLimitInfra) error {
	sa, err := i.expectedRateLimitServiceAccount(infra)
	if err != nil {
		return err
	}

	return i.createOrUpdateServiceAccount(ctx, sa)
}

// deleteRateLimitServiceAccount deletes the Envoy RateLimit ServiceAccount in the kube api server,
// if it exists.
func (i *Infra) deleteRateLimitServiceAccount(ctx context.Context, infra *ir.RateLimitInfra) error {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      rateLimitInfraName,
		},
	}

	return i.deleteServiceAccount(ctx, sa)
}

func (i *Infra) createOrUpdateServiceAccount(ctx context.Context, sa *corev1.ServiceAccount) error {
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
		if err := i.Client.Update(ctx, sa); err != nil {
			return fmt.Errorf("failed to update serviceaccount %s/%s: %w",
				sa.Namespace, sa.Name, err)
		}
	}

	return nil
}

func (i *Infra) deleteServiceAccount(ctx context.Context, sa *corev1.ServiceAccount) error {
	if err := i.Client.Delete(ctx, sa); err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete serviceaccount %s/%s: %w", sa.Namespace, sa.Name, err)
	}

	return nil
}
