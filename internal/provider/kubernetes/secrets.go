package kubernetes

import (
	"context"
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/envoyproxy/gateway/internal/crypto"
)

// caCertificateKey is the key name for accessing TLS CA certificate bundles
// in Kubernetes Secrets.
const caCertificateKey = "ca.crt"

func newSecret(secretType corev1.SecretType, name string, namespace string, data map[string][]byte) corev1.Secret {
	return corev1.Secret{
		Type: secretType,
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"control-plane": "envoy-gateway",
			},
		},
		Data: data,
	}
}

// CertsToSecret creates secrets in the provided namespace, in compact form, from the provided certs.
func CertsToSecret(namespace string, certs *crypto.Certificates) []corev1.Secret {
	return []corev1.Secret{
		newSecret(
			corev1.SecretTypeTLS,
			"envoy-gateway",
			namespace,
			map[string][]byte{
				caCertificateKey:        certs.CACertificate,
				corev1.TLSCertKey:       certs.EnvoyGatewayCertificate,
				corev1.TLSPrivateKeyKey: certs.EnvoyGatewayPrivateKey,
			}),
		newSecret(
			corev1.SecretTypeTLS,
			"envoy",
			namespace,
			map[string][]byte{
				caCertificateKey:        certs.CACertificate,
				corev1.TLSCertKey:       certs.EnvoyCertificate,
				corev1.TLSPrivateKeyKey: certs.EnvoyPrivateKey,
			}),
	}
}

// CreateOrUpdateSecrets creates the provided secrets if they don't exist or updates
// them if they do.
func CreateOrUpdateSecrets(ctx context.Context, client client.Client, secrets []corev1.Secret) ([]corev1.Secret, error) {
	var ret []corev1.Secret
	for i := range secrets {
		secret := secrets[i]
		current := new(corev1.Secret)
		key := types.NamespacedName{
			Namespace: secret.Namespace,
			Name:      secret.Name,
		}
		if err := client.Get(ctx, key, current); err != nil {
			// Create if not found.
			if kerrors.IsNotFound(err) {
				if err := client.Create(ctx, &secret); err != nil {
					return nil, fmt.Errorf("failed to create secret %s/%s: %w", secret.Namespace, secret.Name, err)
				}
			}
		} else {
			// Update if current value is different.
			if !reflect.DeepEqual(secret.Data, current.Data) {
				if err := client.Update(ctx, &secret); err != nil {
					return nil, fmt.Errorf("failed to update secret %s/%s: %w", secret.Namespace, secret.Name, err)
				}
			}
		}
		ret = append(ret, secret)
	}

	return ret, nil
}
