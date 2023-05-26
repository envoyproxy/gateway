// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"

	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
)

// CreateOrUpdateRateLimitInfra creates the managed kube rate limit infra, if it doesn't exist.
func (i *Infra) CreateOrUpdateRateLimitInfra(ctx context.Context) error {
	if err := ratelimit.Validate(ctx, i.Client.Client, i.EnvoyGateway, i.Namespace); err != nil {
		return err
	}

	r := ratelimit.NewResourceRender(i.Namespace, i.EnvoyGateway)
	return i.createOrUpdate(ctx, r)
}

// DeleteRateLimitInfra removes the managed kube infra, if it doesn't exist.
func (i *Infra) DeleteRateLimitInfra(ctx context.Context) error {
	if err := ratelimit.Validate(ctx, i.Client.Client, i.EnvoyGateway, i.Namespace); err != nil {
		return err
	}

	r := ratelimit.NewResourceRender(i.Namespace, i.EnvoyGateway)
	return i.delete(ctx, r)
}
