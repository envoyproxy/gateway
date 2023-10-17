// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"

	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type InfraClient struct {
	client.Client
}

func New(cli client.Client) *InfraClient {
	return &InfraClient{
		Client: cli,
	}
}

func (cli *InfraClient) CreateOrUpdate(ctx context.Context, key client.ObjectKey, current client.Object, specific client.Object, updateChecker func() bool) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		if err := cli.Client.Get(ctx, key, current); err != nil {
			if kerrors.IsNotFound(err) {
				infraManagerResourcesCreated.With(
					k8sResourceTypeLabel.Value(specific.GetObjectKind().GroupVersionKind().Kind),
					k8sResourceNameLabel.Value(key.Name),
					k8sResourceNamespaceLabel.Value(key.Namespace)).Increment()
				// Create if it does not exist.
				if err := cli.Client.Create(ctx, specific); err != nil {
					infraManagerResourcesErrors.With(
						k8sResourceTypeLabel.Value(specific.GetObjectKind().GroupVersionKind().Kind),
						operationLabel.Value("created"),
						k8sResourceNameLabel.Value(key.Name),
						k8sResourceNamespaceLabel.Value(key.Namespace)).Increment()
					return errors.Wrap(err, "for Create")
				}
			}
		} else {
			// Since the client.Object does not have a specific Spec field to compare
			// just perform an update for now.
			if updateChecker() {
				specific.SetUID(current.GetUID())
				infraManagerResourcesUpdated.With(
					k8sResourceTypeLabel.Value(specific.GetObjectKind().GroupVersionKind().Kind),
					k8sResourceNameLabel.Value(key.Name),
					k8sResourceNamespaceLabel.Value(key.Namespace)).Increment()
				if err := cli.Client.Update(ctx, specific); err != nil {
					infraManagerResourcesErrors.With(
						k8sResourceTypeLabel.Value(specific.GetObjectKind().GroupVersionKind().Kind),
						operationLabel.Value("updated"),
						k8sResourceNameLabel.Value(key.Name),
						k8sResourceNamespaceLabel.Value(key.Namespace)).Increment()
					return errors.Wrap(err, "for Update")
				}
			}
		}

		return nil
	})
}

func (cli *InfraClient) Delete(ctx context.Context, object client.Object) error {
	infraManagerResourcesDeleted.With(
		k8sResourceTypeLabel.Value(object.GetObjectKind().GroupVersionKind().Kind),
		k8sResourceNameLabel.Value(object.GetName()),
		k8sResourceNamespaceLabel.Value(object.GetNamespace())).Increment()
	if err := cli.Client.Delete(ctx, object); err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		infraManagerResourcesErrors.With(
			k8sResourceTypeLabel.Value(object.GetObjectKind().GroupVersionKind().Kind),
			operationLabel.Value("deleted"),
			k8sResourceNameLabel.Value(object.GetName()),
			k8sResourceNamespaceLabel.Value(object.GetNamespace())).Increment()
		return err
	}

	return nil
}

// GetUID retrieves the uid of one resource.
func (cli *InfraClient) GetUID(ctx context.Context, key client.ObjectKey, current client.Object) (types.UID, error) {
	if err := cli.Client.Get(ctx, key, current); err != nil {
		return "", err
	}
	return current.GetUID(), nil
}
