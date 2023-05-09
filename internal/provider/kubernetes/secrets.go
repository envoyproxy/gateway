// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	k8smachinery "k8s.io/apimachinery/pkg/types"

	"github.com/envoyproxy/gateway/internal/crypto"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
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
	var tidySecrets []corev1.Secret
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
		tidySecrets = append(tidySecrets, secret)
	}

	return tidySecrets, nil
}

// ValidateSecretObjectReference validate secret object reference for extension tls and ratelimit tls settings.
func ValidateSecretObjectReference(ctx context.Context, client client.Client, secretObjRef gwapiv1b1.SecretObjectReference, namespace string) (*corev1.Secret, string, error) {
	if (secretObjRef.Group == nil || *secretObjRef.Group == corev1.GroupName) &&
		(secretObjRef.Kind == nil || *secretObjRef.Kind == gatewayapi.KindSecret) {
		secret := &corev1.Secret{}
		secretNamespace := namespace
		if secretObjRef.Namespace != nil && string(*secretObjRef.Namespace) != "" {
			secretNamespace = string(*secretObjRef.Namespace)
		}
		key := k8smachinery.NamespacedName{
			Namespace: secretNamespace,
			Name:      string(secretObjRef.Name),
		}
		if err := client.Get(ctx, key, secret); err != nil {
			return nil, secretNamespace, fmt.Errorf("cannot find Secret %s in namespace %s", string(secretObjRef.Name), secretNamespace)
		}

		return secret, secretNamespace, nil
	}

	return nil, "", errors.New("unsupported certificateRef group/kind")
}
