package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/envoyproxy/gateway/internal/ir"
)

const (
	envoyServiceAccountPrefix = "envoy"
)

func expectedServiceAccountName(proxyName string) string {
	return fmt.Sprintf("%s-%s", envoyServiceAccountPrefix, proxyName)
}

// expectedServiceAccount returns the expected proxy serviceAccount.
func (i *Infra) expectedServiceAccount(infra *ir.Infra) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      expectedServiceAccountName(infra.Proxy.Name),
		},
	}
}

// createOrUpdateServiceAccount creates the Envoy ServiceAccount in the kube api server,
// if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateServiceAccount(ctx context.Context, infra *ir.Infra) error {
	sa := i.expectedServiceAccount(infra)

	current := &corev1.ServiceAccount{}
	key := types.NamespacedName{
		Namespace: i.Namespace,
		Name:      expectedServiceAccountName(infra.Proxy.Name),
	}

	if err := i.Client.Get(ctx, key, current); err != nil {
		if kerrors.IsNotFound(err) {
			// Create if it does not exist.
			if err := i.Client.Create(ctx, sa); err != nil {
				return fmt.Errorf("failed to create serviceaccount %s/%s: %w",
					sa.Namespace, sa.Name, err)
			}
		}
	} else {
		// Since the ServiceAccount does not have a specific Spec field to compare
		// just perform an update for now.
		if err := i.Client.Update(ctx, sa); err != nil {
			return fmt.Errorf("failed to update serviceaccount %s/%s: %w",
				sa.Namespace, sa.Name, err)
		}
	}

	if err := i.updateResource(sa); err != nil {
		return err
	}

	return nil
}

// deleteServiceAccount deletes the Envoy ServiceAccount in the kube api server,
// if it exists.
func (i *Infra) deleteServiceAccount(ctx context.Context, infra *ir.Infra) error {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      expectedServiceAccountName(infra.Proxy.Name),
		},
	}
	if err := i.Client.Delete(ctx, sa); err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete serviceaccount %s/%s: %w", sa.Namespace, sa.Name, err)
	}

	return nil
}
