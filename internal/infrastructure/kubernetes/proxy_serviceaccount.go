// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/ir"
)

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
			Name:      expectedResourceHashedName(infra.Proxy.Name),
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
			Name:      expectedResourceHashedName(infra.Proxy.Name),
		},
	}

	return i.deleteServiceAccount(ctx, sa)
}
