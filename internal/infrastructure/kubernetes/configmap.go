// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/provider/utils"
)

const (
	sdsCAFilename   = "xds-trusted-ca.json"
	sdsCertFilename = "xds-certificate.json"
	// xdsTLSCertFilename is the fully qualified path of the file containing Envoy's
	// xDS server TLS certificate.
	xdsTLSCertFilename = "/certs/tls.crt"
	// xdsTLSKeyFilename is the fully qualified path of the file containing Envoy's
	// xDS server TLS key.
	xdsTLSKeyFilename = "/certs/tls.key"
	// xdsTLSCaFilename is the fully qualified path of the file containing Envoy's
	// trusted CA certificate.
	xdsTLSCaFilename = "/certs/ca.crt"
)

var (
	// xDS certificate rotation is supported by using SDS path-based resource files.
	sdsCAConfigMapData = fmt.Sprintf(`{"resources":[{"@type":"type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret",`+
		`"name":"xds_trusted_ca","validation_context":{"trusted_ca":{"filename":"%s"},`+
		`"match_typed_subject_alt_names":[{"san_type":"DNS","matcher":{"exact":"envoy-gateway"}}]}}]}`, xdsTLSCaFilename)
	sdsCertConfigMapData = fmt.Sprintf(`{"resources":[{"@type":"type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret",`+
		`"name":"xds_certificate","tls_certificate":{"certificate_chain":{"filename":"%s"},`+
		`"private_key":{"filename":"%s"}}}]}`, xdsTLSCertFilename, xdsTLSKeyFilename)
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

// expectedRateLimitConfigMap returns the expected ConfigMap based on the provided infra.
func (i *Infra) expectedRateLimitConfigMap(infra *ir.RateLimitInfra) *corev1.ConfigMap {
	labels := rateLimitLabels()
	data := make(map[string]string)

	for _, config := range infra.Configs {
		data[config.Name] = config.Config
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

func (i *Infra) createOrUpdateConfigMap(ctx context.Context, cm *corev1.ConfigMap) error {
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
			if err := i.Client.Update(ctx, cm); err != nil {
				return fmt.Errorf("failed to update configmap %s/%s: %w", cm.Namespace, cm.Name, err)
			}
		}
	}

	return nil
}

func (i *Infra) deleteConfigMap(ctx context.Context, cm *corev1.ConfigMap) error {
	if err := i.Client.Delete(ctx, cm); err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete configmap %s/%s: %w", cm.Namespace, cm.Name, err)
	}

	return nil
}
