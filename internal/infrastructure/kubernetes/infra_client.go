// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	
	kerrors "k8s.io/apimachinery/pkg/api/errors"
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

func (cli *InfraClient) Create(ctx context.Context, key client.ObjectKey, current client.Object, specific client.Object, updateChecker func() bool) error {
	if err := cli.Client.Get(ctx, key, current); err != nil {
		if kerrors.IsNotFound(err) {
			// Create if it does not exist.
			if err := cli.Client.Create(ctx, specific); err != nil {
				return err
			}
		}
	} else {
		// Since the client.Object does not have a specific Spec field to compare
		// just perform an update for now.
		if updateChecker() {
			specific.SetResourceVersion(current.GetResourceVersion())
			specific.SetUID(current.GetUID())
			if err := cli.Client.Update(ctx, specific); err != nil {
				return err
			}
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
