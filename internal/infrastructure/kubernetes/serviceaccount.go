package kubernetes

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/envoyproxy/gateway/internal/ir"
)

const (
	envoyServiceAccountName = "envoy"
)

// createServiceAccountIfNeeded creates a serviceaccount, if it doesn't exist
// in the kube api server.
func (i *Infra) createServiceAccountIfNeeded(ctx context.Context, infra *ir.Infra) error {
	if infra == nil {
		return errors.New("infra ir is nil")
	}

	if infra.Proxy == nil {
		return errors.New("proxy infra ir is nil")
	}

	current, err := i.getServiceAccount(ctx)
	if err != nil {
		if kerrors.IsNotFound(err) {
			sa, err := i.createServiceAccount(ctx)
			if err != nil {
				return err
			}
			if err := i.addResource(sa); err != nil {
				return err
			}
			return nil
		}
		return err
	}

	if err := i.addResource(current); err != nil {
		return err
	}

	return nil
}

// getServiceAccount gets the ServiceAccount from the kube api server.
func (i *Infra) getServiceAccount(ctx context.Context) (*corev1.ServiceAccount, error) {
	key := types.NamespacedName{
		Namespace: i.Namespace,
		Name:      envoyServiceAccountName,
	}
	sa := new(corev1.ServiceAccount)
	if err := i.Client.Get(ctx, key, sa); err != nil {
		return nil, fmt.Errorf("failed to get serviceaccount %s/%s: %w",
			i.Namespace, envoyServiceAccountName, err)
	}

	return sa, nil
}

// expectedServiceAccount returns the expected proxy serviceAccount.
func (i *Infra) expectedServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      envoyServiceAccountName,
		},
	}
}

// createServiceAccount creates the Envoy ServiceAccount in the kube api server,
// if it doesn't exist.
func (i *Infra) createServiceAccount(ctx context.Context) (*corev1.ServiceAccount, error) {
	expected := i.expectedServiceAccount()
	err := i.Client.Create(ctx, expected)
	if err != nil {
		if kerrors.IsAlreadyExists(err) {
			return expected, nil
		}
		return nil, fmt.Errorf("failed to create serviceaccount %s/%s: %w",
			expected.Namespace, expected.Name, err)
	}

	return expected, nil
}

// deleteServiceAccount deletes the Envoy ServiceAccount in the kube api server,
// if it exists.
func (i *Infra) deleteServiceAccount(ctx context.Context) error {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      envoyServiceAccountName,
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
