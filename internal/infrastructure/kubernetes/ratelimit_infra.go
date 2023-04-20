// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"errors"

	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	"github.com/envoyproxy/gateway/internal/ir"
)

// CreateOrUpdateRateLimitInfra creates the managed kube rate limit infra, if it doesn't exist.
func (i *Infra) CreateOrUpdateRateLimitInfra(ctx context.Context, infra *ir.RateLimitInfra) error {
	if infra == nil {
		return errors.New("ratelimit infra ir is nil")
	}

	r := ratelimit.NewResourceRender(i.Namespace, infra, i.EnvoyGateway.RateLimit, i.EnvoyGateway.GetEnvoyGatewayProvider().GetEnvoyGatewayKubeProvider().RateLimitDeployment)

	return i.createOrUpdate(ctx, r)
}

// DeleteRateLimitInfra removes the managed kube infra, if it doesn't exist.
func (i *Infra) DeleteRateLimitInfra(ctx context.Context, infra *ir.RateLimitInfra) error {
	if infra == nil {
		return errors.New("ratelimit infra ir is nil")
	}

	r := ratelimit.NewResourceRender(i.Namespace, infra, i.EnvoyGateway.RateLimit, i.EnvoyGateway.GetEnvoyGatewayProvider().GetEnvoyGatewayKubeProvider().RateLimitDeployment)
	return i.delete(ctx, r)
}
