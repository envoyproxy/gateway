// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package host

import (
	"context"
	"fmt"
)

// TODO: add ratelimit support for host infra

// CreateOrUpdateRateLimitInfra creates the managed host rate limit process, if it doesn't exist.
func (i *Infra) CreateOrUpdateRateLimitInfra(ctx context.Context) error {
	return fmt.Errorf("create/update ratelimit infrastructure is not supported yet for host infrastructure")
}

// DeleteRateLimitInfra removes the managed host rate limit process, if it doesn't exist.
func (i *Infra) DeleteRateLimitInfra(context.Context) error {
	// No-op in host mode; called unconditionally due to k8s/host abstraction but not an error until CreateOrUpdateRateLimitInfra is implemented.
	return nil
}
