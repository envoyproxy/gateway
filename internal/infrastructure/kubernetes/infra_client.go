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
				// Create if it does not exist.
				if err := cli.Client.Create(ctx, specific); err != nil {
					return errors.Wrap(err, "for Create")
				}
			}
		} else {
			// Since the client.Object does not have a specific Spec field to compare
			// just perform an update for now.
			if updateChecker() {
				specific.SetUID(current.GetUID())
				if err := cli.Client.Update(ctx, specific); err != nil {
					return errors.Wrap(err, "for Update")
				}
			}
		}

		return nil
	})
}

func (cli *InfraClient) Delete(ctx context.Context, object client.Object) error {
	if err := cli.Client.Delete(ctx, object); err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
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
