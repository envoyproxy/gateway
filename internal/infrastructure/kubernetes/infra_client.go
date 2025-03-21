// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
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

func (cli *InfraClient) ServerSideApply(ctx context.Context, obj client.Object) error {
	opts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("envoy-gateway")}
	if err := cli.Client.Patch(ctx, obj, client.Apply, opts...); err != nil {
		return fmt.Errorf("failed to create/update resource with server-side apply for obj %v: %w", obj, err)
	}

	return nil
}

func (cli *InfraClient) DeleteIfExists(ctx context.Context, object client.Object) error {
	if err := cli.Delete(ctx, object); err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
	}
	return nil
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
