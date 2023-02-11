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

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/provider/utils"
)

// expectedProxyConfigMap returns the expected ConfigMap based on the provided infra.
func (i *Infra) expectedProxyConfigMap(infra *ir.Infra) (*corev1.ConfigMap, error) {
	// Set the labels based on the owning gateway name.
	labels := envoyLabels(infra.GetProxyInfra().GetProxyMetadata().Labels)
	if len(labels[gatewayapi.OwningGatewayNamespaceLabel]) == 0 || len(labels[gatewayapi.OwningGatewayNameLabel]) == 0 {
		return nil, fmt.Errorf("missing owning gateway labels")
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      expectedProxyConfigMapName(infra.Proxy.Name),
			Labels:    labels,
		},
		Data: map[string]string{
			sdsCAFilename:   sdsCAConfigMapData,
			sdsCertFilename: sdsCertConfigMapData,
		},
	}, nil
}

// createOrUpdateProxyConfigMap creates a ConfigMap in the Kube api server based on the provided
// infra, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateProxyConfigMap(ctx context.Context, infra *ir.Infra) error {
	cm, err := i.expectedProxyConfigMap(infra)
	if err != nil {
		return err
	}

	return i.createOrUpdateConfigMap(ctx, cm)
}

// deleteProxyConfigMap deletes the Envoy ConfigMap in the kube api server, if it exists.
func (i *Infra) deleteProxyConfigMap(ctx context.Context, infra *ir.Infra) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      expectedProxyConfigMapName(infra.Proxy.Name),
		},
	}

	return i.deleteConfigMap(ctx, cm)
}

func expectedProxyConfigMapName(proxyName string) string {
	cMapName := utils.GetHashedName(proxyName)
	return fmt.Sprintf("%s-%s", config.EnvoyPrefix, cMapName)
}
