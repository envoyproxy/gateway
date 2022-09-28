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
	"github.com/envoyproxy/gateway/internal/ir"
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

// expectedConfigMap returns the expected ConfigMap based on the provided infra.
func (i *Infra) expectedConfigMap(infra *ir.Infra) *corev1.ConfigMap {
	ns := i.Namespace
	name := expectedConfigMapName(infra.Proxy.Name)

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Data: map[string]string{
			sdsCAFilename:   sdsCAConfigMapData,
			sdsCertFilename: sdsCertConfigMapData,
		},
	}
}

// createOrUpdateConfigMap creates a ConfigMap in the Kube api server based on the provided
// infra, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateConfigMap(ctx context.Context, infra *ir.Infra) (*corev1.ConfigMap, error) {
	cm := i.expectedConfigMap(infra)

	current := &corev1.ConfigMap{}
	key := types.NamespacedName{
		Namespace: i.Namespace,
		Name:      expectedConfigMapName(infra.Proxy.Name),
	}

	if err := i.Client.Get(ctx, key, current); err != nil {
		// Create if not found.
		if kerrors.IsNotFound(err) {
			if err := i.Client.Create(ctx, cm); err != nil {
				return nil, fmt.Errorf("failed to create configmap %s/%s: %w", cm.Namespace, cm.Name, err)
			}
		}
	} else {
		// Update if current value is different.
		if !reflect.DeepEqual(cm.Data, current.Data) {
			if err := i.Client.Update(ctx, cm); err != nil {
				return nil, fmt.Errorf("failed to update configmap %s/%s: %w", cm.Namespace, cm.Name, err)
			}
		}
	}

	return cm, nil
}

// deleteConfigMap deletes the Envoy ConfigMap in the kube api server, if it exists.
func (i *Infra) deleteConfigMap(ctx context.Context, infra *ir.Infra) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      expectedConfigMapName(infra.Proxy.Name),
		},
	}

	if err := i.Client.Delete(ctx, cm); err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete configmap %s/%s: %w", cm.Namespace, cm.Name, err)
	}

	return nil
}

func expectedConfigMapName(proxyName string) string {
	return fmt.Sprintf("%s-%s", config.EnvoyConfigMapPrefix, proxyName)
}
