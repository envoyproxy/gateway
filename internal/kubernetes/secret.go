// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8smachinery "k8s.io/apimachinery/pkg/types"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

// ValidateSecretObjectReference validate secret object reference for extension tls and ratelimit tls settings.
func ValidateSecretObjectReference(ctx context.Context, client k8sclient.Client, secretObjRef *gwapiv1.SecretObjectReference, namespace string) (*corev1.Secret, string, error) {
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
