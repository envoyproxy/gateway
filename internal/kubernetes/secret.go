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

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

// ValidateSecretObjectReference validate secret object reference for extension tls and ratelimit tls settings.
func ValidateSecretObjectReference(ctx context.Context, client k8sclient.Client, secretObjRef *gwapiv1.SecretObjectReference, namespace string) (*corev1.Secret, string, error) {
	if (secretObjRef.Group == nil || *secretObjRef.Group == corev1.GroupName) &&
		(secretObjRef.Kind == nil || *secretObjRef.Kind == resource.KindSecret) {
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

// ValidateObjectReference validates an ObjectReference, ensuring it points to a Secret.
func ValidateObjectReference(ctx context.Context, client k8sclient.Client, objRef *gwapiv1.ObjectReference, namespace string) (*corev1.Secret, string, error) {
	if objRef == nil {
		return nil, "", errors.New("object reference is nil")
	}

	if objRef.Kind != resource.KindSecret {
		return nil, "", fmt.Errorf("unsupported object reference kind: %v, only 'Secret' is supported", objRef.Kind)
	}

	secretNamespace := namespace
	if objRef.Namespace != nil && string(*objRef.Namespace) != "" {
		secretNamespace = string(*objRef.Namespace)
	}

	secret := &corev1.Secret{}
	key := k8smachinery.NamespacedName{
		Namespace: secretNamespace,
		Name:      string(objRef.Name),
	}
	if err := client.Get(ctx, key, secret); err != nil {
		return nil, secretNamespace, fmt.Errorf("cannot find Secret %s in namespace %s", string(objRef.Name), secretNamespace)
	}

	return secret, secretNamespace, nil
}
