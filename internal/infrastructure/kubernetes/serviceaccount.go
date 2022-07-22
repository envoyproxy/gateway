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

// createServiceAccountIfNeeded creates a serviceaccount based on the provided infra, if
// it doesn't exist in the kube api server.
func (c *Context) createServiceAccountIfNeeded(ctx context.Context, infra *ir.Infra) error {
	if infra == nil {
		return errors.New("infra ir is nil")
	}

	if infra.Proxy == nil {
		return errors.New("proxy infra ir is nil")
	}

	current, err := c.getServiceAccount(ctx, infra)
	if err != nil {
		if kerrors.IsNotFound(err) {
			sa, err := c.createServiceAccount(ctx, infra)
			if err != nil {
				return err
			}
			if err := c.addResource(KindServiceAccount, sa); err != nil {
				return err
			}
			return nil
		}
		return err
	}

	if err := c.addResource(KindServiceAccount, current); err != nil {
		return err
	}

	return nil
}

// getServiceAccount gets the ServiceAccount from the kube api for the provided infra.
func (c *Context) getServiceAccount(ctx context.Context, infra *ir.Infra) (*corev1.ServiceAccount, error) {
	ns := infra.Proxy.Namespace
	name := infra.Proxy.Name
	key := types.NamespacedName{
		Namespace: ns,
		Name:      name,
	}
	sa := new(corev1.ServiceAccount)
	if err := c.Client.Get(ctx, key, sa); err != nil {
		return nil, fmt.Errorf("failed to get serviceaccount %s/%s: %w", ns, name, err)
	}

	return sa, nil
}

// expectedServiceAccount returns the expected proxy serviceAccount based on the provided infra.
func (c *Context) expectedServiceAccount(infra *ir.Infra) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: infra.Proxy.Namespace,
			Name:      infra.Proxy.Name,
		},
	}
}

// createServiceAccount creates sa in the kube api server if it doesn't exist.
func (c *Context) createServiceAccount(ctx context.Context, infra *ir.Infra) (*corev1.ServiceAccount, error) {
	expected := c.expectedServiceAccount(infra)
	err := c.Client.Create(ctx, expected)
	if err != nil {
		if kerrors.IsAlreadyExists(err) {
			return expected, nil
		}
		return nil, fmt.Errorf("failed to create serviceaccount %s/%s: %w",
			expected.Namespace, expected.Name, err)
	}

	return expected, nil
}
