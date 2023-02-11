// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/envoyproxy/gateway/internal/ir"
)

// expectedRateLimitServiceAccount returns the expected ratelimit serviceAccount.
func (i *Infra) expectedRateLimitServiceAccount(_ *ir.RateLimitInfra) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      rateLimitInfraName,
		},
	}
}

// createOrUpdateRateLimitServiceAccount creates the Envoy RateLimit ServiceAccount in the kube api server,
// if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateRateLimitServiceAccount(ctx context.Context, infra *ir.RateLimitInfra) error {
	sa := i.expectedRateLimitServiceAccount(infra)
	return i.createOrUpdateServiceAccount(ctx, sa)
}

// deleteRateLimitServiceAccount deletes the Envoy RateLimit ServiceAccount in the kube api server,
// if it exists.
func (i *Infra) deleteRateLimitServiceAccount(ctx context.Context, _ *ir.RateLimitInfra) error {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      rateLimitInfraName,
		},
	}

	return i.deleteServiceAccount(ctx, sa)
}
