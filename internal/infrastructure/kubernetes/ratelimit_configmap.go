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

// expectedRateLimitConfigMap returns the expected ConfigMap based on the provided infra.
func (i *Infra) expectedRateLimitConfigMap(infra *ir.RateLimitInfra) *corev1.ConfigMap {
	labels := rateLimitLabels()
	data := make(map[string]string)

	for _, serviceConfig := range infra.ServiceConfigs {
		data[serviceConfig.Name] = serviceConfig.Config
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      rateLimitInfraName,
			Labels:    labels,
		},
		Data: data,
	}
}

// createOrUpdateRateLimitConfigMap creates a ConfigMap in the Kube api server based on the provided
// infra, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateRateLimitConfigMap(ctx context.Context, infra *ir.RateLimitInfra) error {
	cm := i.expectedRateLimitConfigMap(infra)
	return i.createOrUpdateConfigMap(ctx, cm)
}

// deleteProxyConfigMap deletes the Envoy ConfigMap in the kube api server, if it exists.
func (i *Infra) deleteRateLimitConfigMap(ctx context.Context, _ *ir.RateLimitInfra) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      rateLimitInfraName,
		},
	}

	return i.deleteConfigMap(ctx, cm)
}
