// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"errors"

	"github.com/envoyproxy/gateway/internal/ir"
)

// CreateOrUpdateRateLimitInfra creates the managed kube rate limit infra, if it doesn't exist.
func (i *Infra) CreateOrUpdateRateLimitInfra(ctx context.Context, infra *ir.RateLimitInfra) error {
	if infra == nil {
		return errors.New("ratelimit infra ir is nil")
	}
	if err := i.createOrUpdateRateLimitServiceAccount(ctx, infra); err != nil {
		return err
	}

	if err := i.createOrUpdateRateLimitConfigMap(ctx, infra); err != nil {
		return err
	}

	if err := i.createOrUpdateRateLimitDeployment(ctx, infra); err != nil {
		return err
	}

	if err := i.createOrUpdateRateLimitService(ctx, infra); err != nil {
		return err
	}

	return nil

}

// DeleteRateLimitInfra removes the managed kube infra, if it doesn't exist.
func (i *Infra) DeleteRateLimitInfra(ctx context.Context, infra *ir.RateLimitInfra) error {
	if infra == nil {
		return errors.New("ratelimit infra ir is nil")
	}

	if err := i.deleteRateLimitService(ctx, infra); err != nil {
		return err
	}

	if err := i.deleteRateLimitDeployment(ctx, infra); err != nil {
		return err
	}

	if err := i.deleteRateLimitConfigMap(ctx, infra); err != nil {
		return err
	}

	if err := i.deleteRateLimitServiceAccount(ctx, infra); err != nil {
		return err
	}

	return nil
}
